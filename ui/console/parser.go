package console

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ParseCommand parses a line of text into a Command struct
func ParseCommand(line string) Command {
	// Trim whitespace and skip empty lines
	line = strings.TrimSpace(line)
	if line == "" {
		return Command{}
	}

	// Split by whitespace to get command name and arguments
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Command{}
	}

	return Command{
		Name: strings.ToLower(parts[0]),
		Args: parts[1:],
	}
}

// ReadLine reads a single line from the provided reader
func ReadLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// ValidateCommand checks if a command is valid
func ValidateCommand(cmd Command) error {
	switch cmd.Name {
	case "agent", "status", "help", "quit", "exit":
		// These commands take no arguments
		if len(cmd.Args) > 0 {
			return fmt.Errorf("command '%s' takes no arguments", cmd.Name)
		}
	case "article":
		// Article command requires exactly one argument
		if len(cmd.Args) == 0 {
			return fmt.Errorf("command 'article' requires an article ID")
		}
		if len(cmd.Args) > 1 {
			return fmt.Errorf("command 'article' takes only one argument (article ID)")
		}
	case "":
		// Empty command is valid (just skip)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return nil
}