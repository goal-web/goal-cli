package providers

import (
	"github.com/goal-web/application"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/app/console"
)

type appServiceProvider struct {
	serviceProviders []contracts.ServiceProvider
}

func NewApp() contracts.ServiceProvider {
	return &appServiceProvider{
		serviceProviders: []contracts.ServiceProvider{
			NewConsoleService(
				console.Commands,
			),
		},
	}
}

func (app appServiceProvider) Register(instance contracts.Application) {
	instance.RegisterServices(app.serviceProviders...)

	instance.Call(func(config contracts.Config) {
		appConfig := config.Get("app").(application.Config)
		instance.Instance("app.env", appConfig.Env)
	})
}

func (app appServiceProvider) Start() error {
	return nil
}

func (app appServiceProvider) Stop() {
}
