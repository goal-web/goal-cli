package commands

import (
	"database/sql"
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/utils"
	"github.com/goal-web/supports/commands"
	"github.com/goal-web/supports/logs"
	"os"
	"path/filepath"
	"strings"
)

func MakeModel(app contracts.Application) contracts.Command {
	return &makeModel{
		Command:    commands.Base("make:model {name} {path?} {table?} {m?}", "创建一个模型"),
		connection: app.Get("db").(contracts.DBConnection),
	}
}

type makeModel struct {
	commands.Command
	connection contracts.DBConnection
}

// ColumnInfo 结构体用于存储表的列信息
type ColumnInfo struct {
	Field   string         `db:"Field"`
	Type    string         `db:"Type"`
	Null    string         `db:"Null"`
	Key     string         `db:"Key"`
	Default sql.NullString `db:"Default"`
	Extra   string         `db:"Extra"`
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
	m := cmd.BoolOptional("m", false)

	var dest = make([]ColumnInfo, 0)
	exception := cmd.connection.Select(&dest, fmt.Sprintf("describe %s", table))
	if exception != nil {
		panic(exception)
	}

	var columns []string
	var primaryColumn ColumnInfo
	for _, column := range dest {
		columns = append(columns,
			fmt.Sprintf("\n\t%s     %s    `json:\"%s\"`", utils.ConvertToCamelCase(column.Field), fieldTypeToGoType(column.Type), column.Field))
		if column.Key == "PRI" {
			primaryColumn = column
		}
	}

	err := os.WriteFile(fmt.Sprintf("%s/%s.go", path, name), []byte(fmt.Sprintf("package models"+
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

	logs.Default().WithFields(contracts.Fields{
		"path":  path,
		"pkg":   pkg,
		"table": table,
		"name":  name,
		"m":     m,
	}).Info(name)

	return nil
}
