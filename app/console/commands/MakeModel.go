package commands

import (
	"fmt"
	"github.com/goal-web/console/arguments"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/app/gen"
	"github.com/goal-web/migration/migrate"
	"github.com/goal-web/supports/commands"
	"github.com/goal-web/supports/logs"
	utils2 "github.com/goal-web/supports/utils"
	"os"
	"path/filepath"
	"strings"
)

func MakeModel() (contracts.Command, contracts.CommandHandlerProvider) {
	return commands.Base("make:model {name} {path?} {table?} {m?}", "创建一个模型"),
		func(app contracts.Application) contracts.CommandHandler {
			return &makeModel{
				connection: app.Get("db").(contracts.DBConnection),
				app:        app,
			}
		}
}

type makeModel struct {
	commands.Command
	connection contracts.DBConnection
	app        contracts.Application
}

func fieldTypeToGoType(fieldType string) string {
	if strings.Contains(fieldType, "TEXT") || strings.HasPrefix(fieldType, "VARCHAR(") {
		return "string"
	}

	var SQLTypeMap = map[string]string{
		"DOUBLE":          "double", //
		"JSON":            "string", //
		"FLOAT":           "float",  //
		"TIMESTAMP":       "string", //
		"INT":             "int32",  // 32
		"BIGINT":          "int64",  // 64
		"INT UNSIGNED":    "uint32", // 32
		"BIGINT UNSIGNED": "uint64", // 64
		"BOOLEAN":         "bool",   //
		"VARCHAR(255)":    "string", //
		"BLOB":            "bytes",  //
	}
	return SQLTypeMap[fieldType]
}

func (cmd makeModel) Handle() any {
	name := cmd.GetString("name")
	table := cmd.StringOptional("table", gen.ConvertCamelToSnake(name))
	path := cmd.StringOptional("path", "pro")
	pkg := filepath.Base(path)
	m := cmd.GetBool("m")
	path = fmt.Sprintf("%s/%s.proto", path, name)

	if utils2.ExistsPath(path) {
		logs.Default().WithFields(contracts.Fields{
			"path":  path,
			"pkg":   pkg,
			"table": table,
			"name":  name,
			"m":     m,
		}).Error("model file is already exists.")
		return nil
	}

	var existsColumns = make([]migrate.ColumnInfo, 0)
	exception := cmd.connection.Select(&existsColumns, fmt.Sprintf("describe %s", table))
	if exception != nil {
		logs.Default().WithError(exception).Error("table is not exists.")
	}

	var columns []string
	for i, column := range existsColumns {
		columns = append(columns, fmt.Sprintf("    %s %s = %d;", fieldTypeToGoType(strings.ToUpper(column.Type)), column.Field, i+1))
	}

	if len(columns) == 0 {
		columns = append(columns, `
		//@pk
		uint64 id = 1;

		string created_at = 100;
		string updated_at = 101;`)
	}

	err := os.WriteFile(path, []byte(fmt.Sprintf(`syntax = "proto3";

option go_package = ".";

//@timestamps
//@table: %s
message %sModel {
%s
}
`, table, name, strings.Join(columns, "\n"))), os.ModePerm)

	if err != nil {
		panic(err)
	}

	// create migration files
	if m && len(existsColumns) == 0 {
		cmd.app.Get("console").(contracts.Console).Call("make:migration", arguments.NewArguments([]string{
			fmt.Sprintf("create_%s", table),
		}, contracts.Fields{}))
	}

	logs.Default().WithFields(contracts.Fields{
		"path":  path,
		"pkg":   pkg,
		"table": table,
		"name":  name,
		"m":     m,
	}).Info(name)

	return nil
}
