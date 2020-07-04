package console

import (
	"fmt"
	"sort"
	"strings"
)

type CommandCallback func(options string) error

type Command struct {
	description string
	usage       string
	callback    CommandCallback
}

type commandList struct {
	commands map[string]Command
}

var commandListSingleton commandList

func AddCommand(key, description, usage string, callback CommandCallback) {
	commandListSingleton.commands[key] = Command{
		description: description,
		usage:       usage,
		callback:    callback,
	}
}

func ExecuteCommand(input string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			PrintString(LevelError, e.(error).Error())
			err = e.(error)
		}
	}()

	if input == "" {
		return nil
	}

	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		if _, ok := commandListSingleton.commands[input]; !ok {
			return nil
		}
		PrintString(LevelInfo, fmt.Sprintf("> %s", input))
		return commandListSingleton.commands[input].callback("")
	}

	if _, ok := commandListSingleton.commands[parts[0]]; !ok {
		return nil
	}

	PrintString(LevelInfo, fmt.Sprintf("> %s", input))

	return commandListSingleton.commands[parts[0]].callback(parts[1])
}

func init() {
	commandListSingleton.commands = map[string]Command{}

	AddCommand("listcommands", "Displays a list of all available commands", "", func(options string) error {
		keys := make([]string, 0, len(commandListSingleton.commands))
		for k := range commandListSingleton.commands {
			keys = append(keys, k)
		}

		sort.Sort(sort.StringSlice(keys))
		for _, k := range keys {
			PrintString(LevelInfo, fmt.Sprintf("  %s: %s", k, commandListSingleton.commands[k].description))
		}

		return nil
	})

	AddCommand("describe", "Explains a specific command", "describe <command>", func(options string) error {
		if options == "" {
			return nil
		}

		if k, ok := commandListSingleton.commands[options]; ok {
			PrintString(LevelInfo, fmt.Sprintf("  %s.\n  Usage: %s", k.description, k.usage))
		} else {
			PrintString(LevelWarning, fmt.Sprintf("%s is not a recognized command", options))
		}

		return nil
	})
}
