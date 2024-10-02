package gen

import (
	"fmt"
	"github.com/emicklei/proto"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Enum struct {
	Package  string
	Name     string
	Comments []string
	Values   []*EnumValue
	FilePath string
}

type EnumValue struct {
	Name     string
	Value    int
	Message  string
	Comments []string
}

func ExtractEnums(p *proto.Proto, basePackage, dir string) []*Enum {
	var list []*Enum
	for _, e := range p.Elements {
		if enum, ok := e.(*proto.Enum); ok {
			enumPath := strings.Join(trim("enums", dir, replaceSuffix(enum.Name, "Enum")+"_gen.go"), "/")
			enumInstance := Enum{
				Name:     enum.Name,
				FilePath: enumPath,
			}

			if pkgs := trim("enums", dir); len(pkgs) == 1 {
				enumInstance.Package = "enums"
			} else {
				enumInstance.Package = pkgs[len(pkgs)-1]
			}

			if enum.Comment != nil {
				enumInstance.Comments = enum.Comment.Lines
			}

			for _, item := range enum.Elements {
				if v, ok := item.(*proto.EnumField); ok {
					value := &EnumValue{
						Name:    v.Name,
						Value:   v.Integer,
						Message: getComment(v.Comment, "@msg", v.Name),
					}
					if v.Comment != nil {
						value.Comments = v.Comment.Lines
					}
					enumInstance.Values = append(enumInstance.Values, value)
				}
			}
			list = append(list, &enumInstance)
		}
	}

	return list
}

func GenEnums(baseOutputDir string, tmpl *template.Template, enums []*Enum) []string {
	var files []string
	for _, enum := range enums {
		outputPath := filepath.Join(baseOutputDir, enum.FilePath)

		// 创建目录
		err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// 创建输出文件
		outFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}

		// 执行模板，传入 moduleName 和 outputPackageName
		err = tmpl.ExecuteTemplate(outFile, "enum", map[string]any{
			"Package": enum.Package,
			"Name":    enum.Name,
			"Values":  enum.Values,
		})
		if err != nil {
			log.Fatal(err)
		}
		outFile.Close()
		files = append(files, outputPath)
		fmt.Printf("生成枚举文件：%s\n", outputPath)
	}
	return files
}
