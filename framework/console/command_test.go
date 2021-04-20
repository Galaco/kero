package console

import (
	"log"
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

func TestBuiltinCommands(t *testing.T) {
	sut := make([]string, 0)
	AddOutputPipe(func(f LogLevel, a interface{}) {
		sut = append(sut, a.(string))
		log.Println(sut)
	})

	err := ExecuteCommand("listcommands")
	if err != nil {
		t.Error(err)
	}

	if len(sut) < 3 {
		t.Error("unexpected number of lines printed by listcommands")
		return
	}

	if sut[0] != "> listcommands" || sut[1] != "  describe: Explains a specific command" || sut[2] != "  foo: bar" {
		t.Error("unexpected output from listcommands")
	}
}
