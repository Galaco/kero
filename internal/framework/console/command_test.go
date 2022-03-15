package console

import (
	"testing"
)

func TestAddCommand(t *testing.T) {
	AddCommand("foo", "bar", "baz", func(options string) error {
		return nil
	})

	if _, ok := commandListSingleton.commands["foo"]; !ok {
		t.Error("could not find added command")
	}
}

func TestBuiltinCommands(t *testing.T) {
	sut := make([]string, 0)
	ClearOutputPipes()
	AddOutputPipe(func(f LogLevel, a interface{}) {
		sut = append(sut, a.(string))
	})

	err := ExecuteCommand("listcommands")
	if err != nil {
		t.Error(err)
	}

	if len(sut) < 3 {
		t.Error("unexpected number of lines printed by listcommands")
		return
	}

	if sut[0] != "> listcommands" || sut[1] != "  describe: Explains a specific command" || sut[3] != "  listcommands: Displays a list of all available commands" {
		t.Error("unexpected output from listcommands")
	}
}

func TestExecuteCommand(t *testing.T) {
	sut := false
	AddCommand("foo", "bar", "baz", func(options string) error {
		sut = true
		return nil
	})

	if _, ok := commandListSingleton.commands["foo"]; !ok {
		t.Error("could not find added command")
		return
	}

	err := ExecuteCommand("foo")
	if err != nil {
		t.Error(err)
	}

	if sut != true {
		t.Error("executed command failed to run")
	}

	sut = false

	err = ExecuteCommand("foo an_arg")
	if err != nil {
		t.Error(err)
	}

	if sut != true {
		t.Error("executed command with parameter failed to run")
	}
}
