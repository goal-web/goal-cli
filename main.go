package main

import (
	"github.com/goal-web/application"
	"github.com/goal-web/cache"
	"github.com/goal-web/config"
	"github.com/goal-web/console/inputs"
	"github.com/goal-web/console/scheduling"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database"
	"github.com/goal-web/email"
	"github.com/goal-web/encryption"
	"github.com/goal-web/filesystem"
	"github.com/goal-web/goal-cli/app/console"
	config2 "github.com/goal-web/goal-cli/config"
	"github.com/goal-web/hashing"
	"github.com/goal-web/migration"
	"github.com/goal-web/queue"
	"github.com/goal-web/redis"
	"github.com/goal-web/serialization"
	"github.com/goal-web/supports/exceptions"
	"github.com/golang-module/carbon/v2"
)

func main() {
	env := config.NewToml(config.File("env.toml"))
	app := application.Singleton(env.GetBool("app.debug"))

	app.Singleton("exceptions.handler", func() contracts.ExceptionHandler {
		return exceptions.DefaultExceptionHandler{}
	})

	// 设置异常处理器

	app.RegisterServices(
		config.NewService(env, config2.GetConfigProviders()),
		hashing.NewService(),
		encryption.NewService(),
		filesystem.NewService(),
		serialization.NewService(),
		redis.NewService(),
		cache.NewService(),
		scheduling.NewService(),
		database.NewService(),
		queue.NewService(true),
		email.NewService(),
		console.NewService(),
		migration.NewService(),
	)

	app.Call(func(config contracts.Config, console3 contracts.Console) {
		appConfig := config.Get("app").(application.Config)
		carbon.SetLocale(appConfig.Locale)
		carbon.SetTimezone(appConfig.Timezone)

		console3.Run(inputs.NewOSArgsInput())
	})
}
