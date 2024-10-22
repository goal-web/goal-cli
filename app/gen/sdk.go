package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func SDK(protoFiles []string, tmplFile, outputDir string) {
	var files []string
	for _, protoFile := range protoFiles {
		// 确保 outputDir 是绝对路径
		outputDirAbs, err := filepath.Abs(outputDir)
		if err != nil {
			log.Fatal(err)
		}

		// 初始化模板，并添加函数映射
		tmpl := GetTemplate(tmplFile)

		definition := ParseProto(protoFile)
		basePackage := "@"
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("无法读取当前目录：%v", err)
		}
		// 提取数据
		data := ExtractProto(pwd, definition, basePackage, "", true)

		// 更新输出目录

		for _, messages := range data.Messages {
			files = append(files, SDKMessages(tmpl, outputDirAbs, messages)...)
		}

		// 生成服务代码
		for _, service := range data.Services {
			files = append(files, SDKServices(outputDirAbs, basePackage, tmpl, service.List)...)
		}

		files = append(files, SDKEnums(outputDirAbs, tmpl, data.Enums)...)

		fmt.Println("代码生成完成。", protoFile)
	}

}
