package commands

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/upgrade"
	"github.com/goal-web/supports/commands"
	"github.com/goal-web/supports/logs"
)

func NewUpgrade() (contracts.Command, contracts.CommandHandlerProvider) {
	return commands.Base("upgrade {-v:目标版本，如 v0.4} {-mod:go.mod路径=./go.mod}", "通过 proto 生成代码"),
		func(application contracts.Application) contracts.CommandHandler {
			return &Upgrade{}
		}
}

type Upgrade struct {
	commands.Command
}

func (cmd Upgrade) Handle() any {
	err := upgrade.UpdateGoalWebDependencies(cmd.GetString("-mod"), cmd.GetString("-v"))
	if err != nil {
		logs.Default().WithError(err).Error("升级失败")
	}
	return nil
}
