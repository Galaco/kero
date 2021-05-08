package console

import (
	"fmt"
	"sort"
	"strings"
)

// Provides a mechanism for creating and calling console commands.
// Does this need to be a in /framework? Maybe not, but its pretty integral to every aspect of the engine

// CommandCallback defines what a valid ConVar function should look like
type CommandCallback func(options string) error

// command
type command struct {
	description string
	usage       string
	callback    CommandCallback
}

// commandList
type commandList struct {
	commands map[string]command
}

// Singleton for command storage. There's no real reason to ever want more than 1 instance; and this
// needs to be easily accessible to a lot of higher level code.
var commandListSingleton commandList

// AddCommand registers a new ConVar that can be executed.
func AddCommand(key, description, usage string, callback CommandCallback) {
	commandListSingleton.commands[key] = command{
		description: description,
		usage:       usage,
		callback:    callback,
	}
}

// GetCommandList collates all available commands
func GetCommandList(prefix string) []string {
	commands := make([]string, 0, len(commandListSingleton.commands))

	for key := range commandListSingleton.commands {
		if len(prefix) == 0 {
			commands = append(commands, key)
			continue
		}

		if strings.HasPrefix(key, prefix) {
			commands = append(commands, key)
		}
	}

	return commands
}

// ExecuteCommand parses a command string and executes the assigned callback if found
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
	commandListSingleton.commands = map[string]command{}

	// Register helper commands
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
