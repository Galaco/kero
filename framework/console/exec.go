package console

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// FileSystemProvider defines the interface for file access
// This allows for dependency injection and testing
type FileSystemProvider interface {
	GetFile(string) (io.Reader, error)
}

var fileSystemProvider FileSystemProvider

// SetFileSystemProvider sets the filesystem provider for config file loading
func SetFileSystemProvider(provider FileSystemProvider) {
	fileSystemProvider = provider
}

// ExecFile executes a config file from the cfg/ directory
func ExecFile(filename string) error {
	if fileSystemProvider == nil {
		return fmt.Errorf("filesystem not initialized")
	}

	// Security: prevent path traversal attacks
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename: path separators not allowed")
	}

	// Add .cfg extension if not present
	if !strings.HasSuffix(filename, ".cfg") {
		filename = filename + ".cfg"
	}

	// Try to load from cfg/ directory
	path := "cfg/" + filename
	reader, err := fileSystemProvider.GetFile(path)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", filename, err)
	}

	return execReader(reader, filename)
}

// execReader executes commands from an io.Reader
func execReader(reader io.Reader, source string) error {
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	executedCount := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Parse the line into individual commands
		commands := parseConfigLine(line)

		// Execute each command
		for _, cmd := range commands {
			if err := ExecuteCommand(cmd); err != nil {
				PrintString(LevelWarning, fmt.Sprintf("%s:%d: error executing '%s': %v", source, lineNum, cmd, err))
				// Continue executing other commands even if one fails
			} else {
				executedCount++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading %s: %w", source, err)
	}

	PrintString(LevelInfo, fmt.Sprintf("Executed %d commands from %s", executedCount, source))
	return nil
}

// parseConfigLine parses a single line from a config file
// Returns a slice of commands to execute
func parseConfigLine(line string) []string {
	// Step 1: Strip comments (// to end of line)
	if idx := strings.Index(line, "//"); idx != -1 {
		line = line[:idx]
	}

	// Step 2: Trim whitespace
	line = strings.TrimSpace(line)

	// Step 3: Skip empty lines
	if line == "" {
		return nil
	}

	// Step 4: Split by semicolon for multiple commands
	parts := strings.Split(line, ";")

	// Step 5: Trim and filter each command
	commands := make([]string, 0, len(parts))
	for _, cmd := range parts {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

func init() {
	// Register the exec command
	AddCommand("exec", "Execute a config file from cfg/ directory", "exec <filename>", func(options string) error {
		if options == "" {
			PrintString(LevelWarning, "Usage: exec <filename>")
			return nil
		}

		// Remove any extra whitespace and quotes
		filename := strings.TrimSpace(options)
		filename = strings.Trim(filename, "\"'")

		if err := ExecFile(filename); err != nil {
			PrintString(LevelError, fmt.Sprintf("Failed to exec %s: %v", filename, err))
			return err
		}

		return nil
	})
}
