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

	// Reset sut to clear any buffered entries from previous tests
	sut = make([]string, 0)

	err := ExecuteCommand("listcommands")
	if err != nil {
		t.Error(err)
	}

	// Commands are sorted alphabetically, so we should have:
	// 0: "> listcommands"
	// 1: "  describe: Explains a specific command"
	// 2: "  exec: Execute a config file from cfg/ directory"
	// 3: "  foo: bar"
	// 4: "  listcommands: Displays a list of all available commands"
	if len(sut) < 5 {
		t.Errorf("unexpected number of lines printed by listcommands: got %d, want at least 5", len(sut))
		return
	}

	if sut[0] != "> listcommands" {
		t.Errorf("sut[0] = %q, want \"> listcommands\"", sut[0])
	}
	if sut[1] != "  describe: Explains a specific command" {
		t.Errorf("sut[1] = %q, want \"  describe: Explains a specific command\"", sut[1])
	}
	if sut[2] != "  exec: Execute a config file from cfg/ directory" {
		t.Errorf("sut[2] = %q, want \"  exec: Execute a config file from cfg/ directory\"", sut[2])
	}
	if sut[3] != "  foo: bar" {
		t.Errorf("sut[3] = %q, want \"  foo: bar\"", sut[3])
	}
}

func TestExecuteCommandWithBooleanConvarAndInt(t *testing.T) {
	// Setup: Create a boolean convar
	AddConvarBool("test_bool_cmd", "Test boolean convar for command execution", false)

	// Test: Execute command "test_bool_cmd 1" (should set to true)
	err := ExecuteCommand("test_bool_cmd 1")
	if err != nil {
		t.Error(err)
	}
	if !GetConvarBoolean("test_bool_cmd") {
		t.Error("Expected true after executing 'test_bool_cmd 1'")
	}

	// Test: Execute command "test_bool_cmd 0" (should set to false)
	err = ExecuteCommand("test_bool_cmd 0")
	if err != nil {
		t.Error(err)
	}
	if GetConvarBoolean("test_bool_cmd") {
		t.Error("Expected false after executing 'test_bool_cmd 0'")
	}

	// Test: Execute command "test_bool_cmd true" (should still work)
	err = ExecuteCommand("test_bool_cmd true")
	if err != nil {
		t.Error(err)
	}
	if !GetConvarBoolean("test_bool_cmd") {
		t.Error("Expected true after executing 'test_bool_cmd true'")
	}

	// Test: Execute command "test_bool_cmd 2" (should be ignored, remain true)
	err = ExecuteCommand("test_bool_cmd 2")
	if err != nil {
		t.Error(err)
	}
	if !GetConvarBoolean("test_bool_cmd") {
		t.Error("Expected true (unchanged) after executing 'test_bool_cmd 2'")
	}
}
