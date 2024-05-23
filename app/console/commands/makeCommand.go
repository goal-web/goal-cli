package commands

import (
	"fmt"
	"github.com/goal-web/contracts"
	"github.com/goal-web/goal-cli/utils"
	"github.com/goal-web/supports/commands"
	"github.com/goal-web/supports/logs"
	"os"
	"path/filepath"
	"strings"
)

func MakeCommand(app contracts.Application) contracts.Command {
	return &makeCommand{
		Command: commands.Base("make:command {name} {path?}", "make a command"),
	}
}

type makeCommand struct {
	commands.Command
}

func (hello makeCommand) Handle() any {
	path := hello.StringOptional("path", "app/console/commands")
	name := hello.GetString("name")

	pkg := filepath.Base(path)
	className := strings.ToLower(string(name[0])) + name[1:]

	err := os.WriteFile(fmt.Sprintf("%s/%s.go", path, name), []byte(fmt.Sprintf(`package %s

import (
	"github.com/goal-web/contracts"
	"github.com/goal-web/supports/commands"
)

func %s(app contracts.Application) contracts.Command {
	return &%s{
		Command: commands.Base("%s", "description"),
	}
}

type %s struct {
	commands.Command
}

func (cmd %s) Handle() any {
	return nil
}
`, pkg, name, className, utils.CamelToColonHyphen(name), className, className)), os.ModePerm)

	if err != nil {
		panic(err)
	}

	logs.Default().WithFields(contracts.Fields{
		"pkg":  pkg,
		"path": path,
	}).Info("路径是" + path)

	return nil
}
