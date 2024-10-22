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

type Field struct {
	Comment    *proto.Comment
	Index      int
	Name       string
	Type       string
	JSONName   string
	Comments   string
	Tags       string
	ImportPath string
	UsageName  string
	GoType     string // 用来映射 any 之类的
	Ptr        bool
	IsModel    bool
	Repeated   bool
	Parent     *Message
}

type Message struct {
	IsModel bool
	GoType  string
	Name    string
	RawName string // 没有后缀

	TableName  string // 模型才有
	SoftDelete string // 模型才有
	PrimaryKey string // 模型才有
	Fields     []*Field

	Relations       []*Field // 关联关系
	Template        string   // model
	Authenticatable bool     // 是否可用作登录
	ImportPath      string   // 包路径，例如：biz/models/auth
	UsageName       string   // 包名，例如：auth.UserModel
	FilePath        string   // biz/models/user.go
	Comments        []string
	Comment         *proto.Comment
}

func GenMessages(tmpl *template.Template, baseOutputDir string, messages []*Message) []string {
	var files []string
	for _, message := range messages {
		outputPath := filepath.Join(baseOutputDir, message.FilePath)

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
		err = tmpl.ExecuteTemplate(outFile, message.Template, map[string]any{
			"Imports":   DetermineMessageImports(message),
			"Model":     message,
			"Package":   filepath.Base(message.ImportPath),
			"Name":      message.Name,
			"Fields":    message.Fields,
			"Relations": message.Relations,
		})
		if err != nil {
			log.Fatal(err)
		}
		outFile.Close()

		fmt.Printf("生成模型文件：%s\n", outputPath)
		files = append(files, outputPath)
	}
	return files
}

func SDKMessages(tmpl *template.Template, baseOutputDir string, messages []*Message) []string {
	var files []string
	for _, message := range messages {
		outputPath := filepath.Join(baseOutputDir, strings.ReplaceAll(message.FilePath, ".go", ".ts"))

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
		err = tmpl.ExecuteTemplate(outFile, message.Template, map[string]any{
			"Imports":   DetermineTsMessageImports(message),
			"Model":     message,
			"Package":   filepath.Base(message.ImportPath),
			"Name":      message.Name,
			"Fields":    message.Fields,
			"Relations": message.Relations,
		})
		if err != nil {
			log.Fatal(err)
		}
		outFile.Close()

		fmt.Printf("生成模型文件：%s\n", outputPath)
		files = append(files, outputPath)
	}
	return files
}
