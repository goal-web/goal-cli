package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/app/gen"
	"github.com/goal-web/supports/commands"
)

func NewGen() (contracts.Command, contracts.CommandHandlerProvider) {
	return commands.Base("gen {--dir:Proto文件的路径=pro} {--out:输出的基准目录=.} {--mode:生成模式=pro} {--tmpl:模板文件路径=template.tmpl}", "通过 proto 生成代码"),
		func(application contracts.Application) contracts.CommandHandler {
			return &Proto{}
		}
}

type Proto struct {
	commands.Command
}

func (proto Proto) Handle() any {
	tmpl := proto.GetString("tmpl")
	out := proto.GetString("out")
	protoFiles, err := scanProtoFiles(proto.GetString("dir"))
	if err != nil {
		fmt.Printf("扫描目录 %s 中的 proto 文件失败: %v\n", proto.GetString("dir"), err)
		os.Exit(1)
	}

	if proto.GetString("mode") == "pro" {
		// 遍历所有找到的 proto 文件，依次调用 gen.Pro()
		for _, protoFile := range protoFiles {
			fmt.Printf("正在处理 proto 文件: %s\n", protoFile)
			gen.Pro(protoFile, tmpl, out)
		}
	} else {
		// 遍历所有找到的 proto 文件，依次调用 gen.Pro()
		gen.SDK(protoFiles, tmpl, out)
	}

	return nil
}

// scanProtoFiles 扫描指定目录，返回所有的 .proto 文件路径
func scanProtoFiles(root string) ([]string, error) {
	var protoFiles []string

	// 使用 filepath.Walk 遍历目录，查找所有 .proto 文件
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是 .proto 文件，则加入到列表中
		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return protoFiles, nil
}
