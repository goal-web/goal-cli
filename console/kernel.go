package console

import (
	"github.com/goal-web/config"
	"github.com/goal-web/console"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/console/commands"
)

func NewService() contracts.ServiceProvider {
	return console.NewService(NewKernel)
}

func NewKernel(app contracts.Application) contracts.Console {
	return &Kernel{console.NewKernel(app, []contracts.CommandProvider{
		commands.NewHello,
		config.EncryptionCommand,
	}), app}
}

type Kernel struct {
	*console.Kernel
	app contracts.Application
}

func (kernel *Kernel) Schedule(schedule contracts.Schedule) {
}
