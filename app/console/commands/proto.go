package commands

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/gen"
	"github.com/goal-web/supports/commands"
)

func NewGen() (contracts.Command, contracts.CommandHandlerProvider) {
	return commands.Base("gen {proto:Proto 文件的路径} {--out:输出的基准目录=.} {--tmpl:模板文件路径=template.tmpl}", "通过 proto 生成代码"),
		func(application contracts.Application) contracts.CommandHandler {
			return &Proto{}
		}
}

type Proto struct {
	commands.Command
}

func (proto Proto) Handle() any {
	gen.Pro(proto.GetString("proto"), proto.GetString("tmpl"), proto.GetString("out"))
	return nil
}
