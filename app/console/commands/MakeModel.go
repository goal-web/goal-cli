package commands

import (
	"fmt"
	"github.com/goal-web/console/arguments"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/utils"
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
	switch fieldType {
	case "int", "bigint":
		return "int64"
	case "float", "double":
		return "float64"
	case "varchar", "char", "text", "json":
		return "string"
	case "binary":
		return "[]byte"
	case "date", "datetime", "timestamp":
		return "time.Time"
	case "boolean":
		return "bool"
	default:
		if strings.HasPrefix(fieldType, "varchar") ||
			strings.HasPrefix(fieldType, "nvarchar") ||
			strings.HasPrefix(fieldType, "text") {
			return "string"
		}
		return "any" // Fallback type
	}
}

func (cmd makeModel) Handle() any {
	name := cmd.GetString("name")
	table := cmd.StringOptional("table", utils.ModelNameToTable(name))
	path := cmd.StringOptional("path", "app/models")
	pkg := filepath.Base(path)
	m := cmd.GetBool("m")
	path = fmt.Sprintf("%s/%s.go", path, name)

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
	_ = cmd.connection.Select(&existsColumns, fmt.Sprintf("describe %s", table))

	var columns []string
	var primaryColumn migrate.ColumnInfo
	for _, column := range existsColumns {
		columns = append(columns,
			fmt.Sprintf("\n\t%s     %s    `json:\"%s\"`", utils.ConvertToCamelCase(column.Field), fieldTypeToGoType(column.Type), column.Field))
		if column.Key == "PRI" {
			primaryColumn = column
		}
	}

	err := os.WriteFile(path, []byte(fmt.Sprintf("package models"+
		"\n"+
		"\nimport ("+
		"\n\t\"github.com/goal-web/database/table\""+
		"\n\t\"github.com/goal-web/supports/class\""+
		"\n\t\"time\""+
		"\n)"+
		"\n"+
		"\nvar ("+
		fmt.Sprintf("\n\t%sClass = class.Make[%s]()", name, name)+
		"\n)"+
		"\n"+
		fmt.Sprintf("\nfunc %s() *table.Table[%s] {", utils.ToPlural(name), name)+
		fmt.Sprintf("\n\treturn table.Class(%sClass, \"%s\").SetPrimaryKey(\"%s\")", name, table, primaryColumn.Field)+
		"\n}"+
		"\n"+
		fmt.Sprintf("\ntype %s struct {", name)+
		fmt.Sprintf("\n\ttable.Model[%s] `json:\"-\"`", name)+
		"\n"+
		strings.Join(columns, "")+
		"\n}"+
		"\n")), os.ModePerm)

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
