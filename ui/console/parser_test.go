package console

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Command
	}{
		{
			name:  "empty command",
			input: "",
			expected: Command{
				Name: "",
				Args: []string{},
			},
		},
		{
			name:  "whitespace only",
			input: "   \t\n  ",
			expected: Command{
				Name: "",
				Args: []string{},
			},
		},
		{
			name:  "single word command",
			input: "help",
			expected: Command{
				Name: "help",
				Args: []string{},
			},
		},
		{
			name:  "command with args",
			input: "article 12345",
			expected: Command{
				Name: "article",
				Args: []string{"12345"},
			},
		},
		{
			name:  "command with multiple args",
			input: "article 12345 extra",
			expected: Command{
				Name: "article",
				Args: []string{"12345", "extra"},
			},
		},
		{
			name:  "uppercase command",
			input: "HELP",
			expected: Command{
				Name: "help",
				Args: []string{},
			},
		},
		{
			name:  "mixed case command",
			input: "Agent",
			expected: Command{
				Name: "agent",
				Args: []string{},
			},
		},
		{
			name:  "command with extra whitespace",
			input: "  article   12345  ",
			expected: Command{
				Name: "article",
				Args: []string{"12345"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommand(tt.input)
			if result.Name != tt.expected.Name {
				t.Errorf("ParseCommand() name = %v, want %v", result.Name, tt.expected.Name)
			}
			if len(result.Args) != len(tt.expected.Args) {
				t.Errorf("ParseCommand() args length = %v, want %v", len(result.Args), len(tt.expected.Args))
				return
			}
			for i, arg := range result.Args {
				if arg != tt.expected.Args[i] {
					t.Errorf("ParseCommand() arg[%d] = %v, want %v", i, arg, tt.expected.Args[i])
				}
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		wantErr bool
	}{
		{
			name:    "empty command",
			cmd:     Command{},
			wantErr: false,
		},
		{
			name:    "valid help command",
			cmd:     Command{Name: "help"},
			wantErr: false,
		},
		{
			name:    "help with args",
			cmd:     Command{Name: "help", Args: []string{"extra"}},
			wantErr: true,
		},
		{
			name:    "valid agent command",
			cmd:     Command{Name: "agent"},
			wantErr: false,
		},
		{
			name:    "article without args",
			cmd:     Command{Name: "article"},
			wantErr: true,
		},
		{
			name:    "article with one arg",
			cmd:     Command{Name: "article", Args: []string{"12345"}},
			wantErr: false,
		},
		{
			name:    "article with multiple args",
			cmd:     Command{Name: "article", Args: []string{"12345", "extra"}},
			wantErr: true,
		},
		{
			name:    "unknown command",
			cmd:     Command{Name: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}