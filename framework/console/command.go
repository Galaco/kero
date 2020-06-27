package console

import "strings"

type Command func(options string) error

type commandList struct {
	commands map[string]Command
}

var commandListSingleton commandList


func AddCommand(key string, callback Command) {
	commandListSingleton.commands[key] = callback
}

func ExecuteCommand(input string) error {
	if input == "" {
		return nil
	}

	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		if _,ok := commandListSingleton.commands[input]; !ok {
			return nil
		}
		return commandListSingleton.commands[input]("")
	}

	if _,ok := commandListSingleton.commands[parts[0]]; !ok {
		return nil
	}

	return commandListSingleton.commands[parts[0]](parts[1])
}

func init() {
	commandListSingleton.commands = map[string]Command{}
}