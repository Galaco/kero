package console

import (
	"fmt"
	"sort"
	"strconv"
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

// ExecuteCommand parses a command string and executes the assigned callback if found.
// If not found as a command, it checks if it's a convar and attempts to set its value.
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
	commandName := parts[0]

	// Case 1: Single word input (no arguments)
	if len(parts) < 2 {
		// Check if it's a command first
		if cmd, ok := commandListSingleton.commands[input]; ok {
			PrintString(LevelInfo, fmt.Sprintf("> %s", input))
			return cmd.callback("")
		}

		// Not a command, check if it's a convar (print current value)
		if convar := GetConvar(input); convar != nil {
			PrintString(LevelInfo, fmt.Sprintf("%s: %v", input, convar.Value))
			return nil
		}

		// Not a command or convar, silently ignore
		return nil
	}

	// Case 2: Input with arguments
	args := parts[1]

	// Check if it's a command first
	if cmd, ok := commandListSingleton.commands[commandName]; ok {
		PrintString(LevelInfo, fmt.Sprintf("> %s", input))
		return cmd.callback(args)
	}

	// Not a command, check if it's a convar and try to set it
	if convar := GetConvar(commandName); convar != nil {
		return setConvarFromString(commandName, args)
	}

	// Not a command or convar, silently ignore
	return nil
}

// setConvarFromString attempts to parse and set a convar value from a string
func setConvarFromString(key, value string) error {
	convar := GetConvar(key)
	if convar == nil {
		return nil
	}

	// Try to parse as boolean
	if value == "true" {
		SetConvarBoolean(key, true)
		PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
		return nil
	}
	if value == "false" {
		SetConvarBoolean(key, false)
		PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
		return nil
	}

	// Try to parse as integer
	if i, err := strconv.Atoi(value); err == nil {
		// Special case: if the convar is a boolean, treat integers as bool (0=false, 1=true)
		if convar.Type == ConvarTypeBool {
			SetConvarBooleanInt(key, i)
			PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
			return nil
		}
		SetConvarInt(key, i)
		PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
		return nil
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(value, 32); err == nil {
		SetConvarFloat(key, float32(f))
		PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
		return nil
	}

	// Default to string
	SetConvarString(key, value)
	PrintString(LevelInfo, fmt.Sprintf("> %s %s", key, value))
	return nil
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
