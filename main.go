package main

import (
	"github.com/goal-web/application"
	"github.com/goal-web/cache"
	"github.com/goal-web/config"
	"github.com/goal-web/console"
	"github.com/goal-web/console/inputs"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database"
	"github.com/goal-web/email"
	"github.com/goal-web/encryption"
	"github.com/goal-web/events"
	"github.com/goal-web/filesystem"
	config2 "github.com/goal-web/goal-cli/config"
	"github.com/goal-web/goal-cli/providers"
	"github.com/goal-web/hashing"
	"github.com/goal-web/migration"
	"github.com/goal-web/redis"
	"github.com/goal-web/serialization"
	"github.com/goal-web/supports/exceptions"
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
		events.NewService(),
		filesystem.NewService(),
		serialization.NewService(),
		redis.NewService(),
		cache.NewService(),
		database.NewService(false), // 不需要立刻链接数据，只有当用到了 migration 的时候才会连数据库
		email.NewService(),
		console.NewService(),
		migration.NewService(),
		providers.NewApp(),
	)

	app.Call(func(console3 contracts.Console) {
		console3.Run(inputs.NewOSArgsInput())
	})
}
