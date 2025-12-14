package console

import (
	"context"
	"io"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// ConsoleUI defines the interface for console-based user interface
type ConsoleUI interface {
	Run(ctx context.Context) error
}

// Command represents a parsed command with its name and arguments
type Command struct {
	Name string
	Args []string
}

// SimpleConsoleUI implements ConsoleUI with a basic REPL interface
type SimpleConsoleUI struct {
	In         io.Reader
	Out        io.Writer
	Framework  interfaces.PonchoFramework
	ArticleFlow interface{} // Will be *articleflow.ArticleFlow once imported
	Logger     interfaces.Logger

	// Internal state
	chatHistory      []*interfaces.PonchoMessage
	lastFlowState    interface{} // Will hold the last ArticleFlowState
	lastFlowEvent    *FlowEvent
	isInChatMode     bool
}

// NewSimpleConsoleUI creates a new instance of SimpleConsoleUI
func NewSimpleConsoleUI(
	in io.Reader,
	out io.Writer,
	framework interfaces.PonchoFramework,
	articleFlow interface{},
	logger interfaces.Logger,
) *SimpleConsoleUI {
	return &SimpleConsoleUI{
		In:           in,
		Out:          out,
		Framework:    framework,
		ArticleFlow:  articleFlow,
		Logger:       logger,
		chatHistory:  make([]*interfaces.PonchoMessage, 0),
	}
}