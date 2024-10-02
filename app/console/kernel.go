package console

import (
	"github.com/goal-web/config"
	"github.com/goal-web/contracts"
	commands2 "github.com/goal-web/goal-cli/app/console/commands"
)

var Commands = []contracts.CommandProvider{
	commands2.NewHello,
	commands2.NewGen,
	config.EncryptionCommand,
	commands2.MakeCommand,
	commands2.MakeModel,
}
