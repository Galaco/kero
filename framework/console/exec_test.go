package console

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// TestParseConfigLine tests the config line parser
func TestParseConfigLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "map de_dust2",
			expected: []string{"map de_dust2"},
		},
		{
			name:     "command with comment",
			input:    "sv_cheats 1 // enable cheats",
			expected: []string{"sv_cheats 1"},
		},
		{
			name:     "comment only",
			input:    "// this is a comment",
			expected: nil,
		},
		{
			name:     "empty line",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "multiple commands with semicolon",
			input:    "map de_dust2; sv_cheats 1; god",
			expected: []string{"map de_dust2", "sv_cheats 1", "god"},
		},
		{
			name:     "semicolon with comment",
			input:    "map de_dust2; sv_cheats 1 // comment here",
			expected: []string{"map de_dust2", "sv_cheats 1"},
		},
		{
			name:     "trailing semicolon",
			input:    "map de_dust2;",
			expected: []string{"map de_dust2"},
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  map de_dust2  ",
			expected: []string{"map de_dust2"},
		},
		{
			name:     "semicolon with spaces",
			input:    "cmd1 ; cmd2 ; cmd3",
			expected: []string{"cmd1", "cmd2", "cmd3"},
		},
		{
			name:     "command with quotes",
			input:    `echo "hello world"`,
			expected: []string{`echo "hello world"`},
		},
		{
			name:     "convar assignment",
			input:    "r_drawdisplacements 1",
			expected: []string{"r_drawdisplacements 1"},
		},
		{
			name:     "complex multiline scenario",
			input:    "cl_showfps 1; mat_picmip 0 // graphics settings",
			expected: []string{"cl_showfps 1", "mat_picmip 0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConfigLine(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseConfigLine(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// mockFileSystem is a mock implementation of FileSystemProvider for testing
type mockFileSystem struct {
	files map[string]string
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]string),
	}
}

func (m *mockFileSystem) addFile(path, content string) {
	m.files[path] = content
}

func (m *mockFileSystem) GetFile(path string) (io.Reader, error) {
	content, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return strings.NewReader(content), nil
}

// TestExecFile_Security tests path traversal prevention
func TestExecFile_Security(t *testing.T) {
	mockFS := newMockFileSystem()
	SetFileSystemProvider(mockFS)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid filename",
			filename: "autoexec.cfg",
			wantErr:  false,
		},
		{
			name:     "valid filename without extension",
			filename: "autoexec",
			wantErr:  false,
		},
		{
			name:     "path traversal with ..",
			filename: "../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "path with forward slash",
			filename: "subdir/config.cfg",
			wantErr:  true,
		},
		{
			name:     "path with backslash",
			filename: "subdir\\config.cfg",
			wantErr:  true,
		},
		{
			name:     "hidden file attempt",
			filename: "../.hidden",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add a valid file to mock filesystem
			mockFS.addFile("cfg/"+tt.filename, "// test file")
			if !strings.HasSuffix(tt.filename, ".cfg") {
				mockFS.addFile("cfg/"+tt.filename+".cfg", "// test file")
			}

			err := ExecFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecFile(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}

// TestExecReader tests the config file execution
func TestExecReader(t *testing.T) {
	// Store original command list to restore after test
	originalCommands := commandListSingleton.commands
	defer func() {
		commandListSingleton.commands = originalCommands
	}()

	// Create a fresh command list for testing
	commandListSingleton.commands = make(map[string]command)

	// Create a test convar
	AddConvarInt("test_convar", "Test convar", 0)

	// Track executed commands
	var executedCommands []string
	AddCommand("testcmd", "Test command", "", func(options string) error {
		executedCommands = append(executedCommands, "testcmd "+options)
		return nil
	})

	configContent := `// This is a test config file
testcmd arg1
test_convar 42
// Comment line
testcmd arg2; testcmd arg3

testcmd arg4 // inline comment
`

	reader := strings.NewReader(configContent)
	err := execReader(reader, "test.cfg")
	if err != nil {
		t.Fatalf("execReader failed: %v", err)
	}

	expectedCommands := []string{
		"testcmd arg1",
		"testcmd arg2",
		"testcmd arg3",
		"testcmd arg4",
	}

	if !reflect.DeepEqual(executedCommands, expectedCommands) {
		t.Errorf("executed commands = %v, want %v", executedCommands, expectedCommands)
	}

	// Verify convar was set
	if GetConvarInt("test_convar") != 42 {
		t.Errorf("test_convar = %d, want 42", GetConvarInt("test_convar"))
	}
}

// TestExecReader_ErrorHandling tests that errors in individual commands don't stop execution
func TestExecReader_ErrorHandling(t *testing.T) {
	// Store original command list to restore after test
	originalCommands := commandListSingleton.commands
	defer func() {
		commandListSingleton.commands = originalCommands
	}()

	// Create a fresh command list for testing
	commandListSingleton.commands = make(map[string]command)

	var executedCommands []string

	// Command that always succeeds
	AddCommand("goodcmd", "Good command", "", func(options string) error {
		executedCommands = append(executedCommands, "goodcmd")
		return nil
	})

	// Command that always fails
	AddCommand("badcmd", "Bad command", "", func(options string) error {
		return fmt.Errorf("intentional error")
	})

	configContent := `goodcmd
badcmd
goodcmd
`

	reader := strings.NewReader(configContent)
	err := execReader(reader, "test.cfg")
	if err != nil {
		t.Fatalf("execReader failed: %v", err)
	}

	// Verify both goodcmd calls executed despite badcmd error
	expectedCommands := []string{"goodcmd", "goodcmd"}
	if !reflect.DeepEqual(executedCommands, expectedCommands) {
		t.Errorf("executed commands = %v, want %v", executedCommands, expectedCommands)
	}
}

// TestExecFile_FileNotFound tests missing file handling
func TestExecFile_FileNotFound(t *testing.T) {
	mockFS := newMockFileSystem()
	SetFileSystemProvider(mockFS)

	err := ExecFile("nonexistent.cfg")
	if err == nil {
		t.Error("ExecFile should return error for nonexistent file")
	}
}

// TestExecFile_NoFileSystem tests behavior when filesystem not initialized
func TestExecFile_NoFileSystem(t *testing.T) {
	// Save current provider
	originalProvider := fileSystemProvider
	defer func() {
		fileSystemProvider = originalProvider
	}()

	// Set to nil to simulate uninitialized state
	SetFileSystemProvider(nil)

	err := ExecFile("test.cfg")
	if err == nil {
		t.Error("ExecFile should return error when filesystem not initialized")
	}
	if !strings.Contains(err.Error(), "filesystem not initialized") {
		t.Errorf("Expected 'filesystem not initialized' error, got: %v", err)
	}
}

// TestExecFile_AutoAddExtension tests automatic .cfg extension addition
func TestExecFile_AutoAddExtension(t *testing.T) {
	mockFS := newMockFileSystem()
	SetFileSystemProvider(mockFS)

	testContent := "// test file"
	mockFS.addFile("cfg/myconfig.cfg", testContent)

	// Test with and without extension
	tests := []string{"myconfig", "myconfig.cfg"}
	for _, filename := range tests {
		err := ExecFile(filename)
		if err != nil {
			t.Errorf("ExecFile(%q) failed: %v", filename, err)
		}
	}
}

// Benchmark for parseConfigLine
func BenchmarkParseConfigLine(b *testing.B) {
	line := "map de_dust2; sv_cheats 1; god // enable everything"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseConfigLine(line)
	}
}

// Benchmark for execReader
func BenchmarkExecReader(b *testing.B) {
	// Setup
	originalCommands := commandListSingleton.commands
	defer func() {
		commandListSingleton.commands = originalCommands
	}()
	commandListSingleton.commands = make(map[string]command)
	AddConvarInt("test_var", "Test", 0)

	configContent := `// Config file
test_var 1
test_var 2
test_var 3
test_var 4
test_var 5
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader([]byte(configContent))
		execReader(reader, "bench.cfg")
	}
}
