package console

import (
	"github.com/goal-web/config"
	"github.com/goal-web/console"
	"github.com/goal-web/contracts"
	commands2 "github.com/goal-web/goal-cli/app/console/commands"
)

func NewService() contracts.ServiceProvider {
	return console.NewService(NewKernel)
}

func NewKernel(app contracts.Application) contracts.Console {
	return &Kernel{console.NewKernel(app, []contracts.CommandProvider{
		commands2.NewHello,
		config.EncryptionCommand,
		commands2.MakeCommand,
		commands2.MakeModel,
	}), app}
}

type Kernel struct {
	*console.Kernel
	app contracts.Application
}

func (kernel *Kernel) Schedule(schedule contracts.Schedule) {
}
