package providers

import (
	"github.com/goal-web/contracts"
)

type Console struct {
	Commands []contracts.CommandProvider
}

func NewConsoleService(commands []contracts.CommandProvider) contracts.ServiceProvider {
	return Console{
		Commands: commands,
	}
}

func (c Console) Register(application contracts.Application) {
	application.Call(func(console contracts.Console) {
		for _, provider := range c.Commands {
			console.RegisterCommand(provider)
		}
	})
}

func (c Console) Start() error {
	return nil
}

func (c Console) Stop() {
}
