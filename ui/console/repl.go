package console

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Run starts the console UI REPL loop
func (ui *SimpleConsoleUI) Run(ctx context.Context) error {
	reader := bufio.NewReader(ui.In)

	// Print welcome message
	ui.printWelcome()

	// Main REPL loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Print prompt
			if ui.isInChatMode {
				fmt.Fprint(ui.Out, "agent> ")
			} else {
				fmt.Fprint(ui.Out, "poncho> ")
			}

			// Read command
			line, err := ReadLine(reader)
			if err != nil {
				if err == io.EOF {
					fmt.Fprintln(ui.Out, "\nGoodbye!")
					return nil
				}
				return fmt.Errorf("failed to read input: %w", err)
			}

			// Handle chat mode specially
			if ui.isInChatMode {
				if err := ui.handleAgentInputWithAI(ctx, line, reader); err != nil {
					fmt.Fprintf(ui.Out, "Error: %v\n", err)
					// Fall back to simple echo mode on error
					if err := ui.handleChatInput(ctx, line, reader); err != nil {
						fmt.Fprintf(ui.Out, "Error in fallback mode: %v\n", err)
					}
				}
				continue
			}

			// Parse command
			cmd := ParseCommand(line)
			if err := ValidateCommand(cmd); err != nil {
				fmt.Fprintf(ui.Out, "Error: %v\n", err)
				continue
			}

			// Skip empty commands
			if cmd.Name == "" {
				continue
			}

			// Execute command
			if err := ui.executeCommand(ctx, cmd); err != nil {
				fmt.Fprintf(ui.Out, "Error: %v\n", err)
			}
		}
	}
}

// printWelcome displays the welcome message and help
func (ui *SimpleConsoleUI) printWelcome() {
	fmt.Fprintln(ui.Out, "Welcome to Poncho Framework Console UI")
	fmt.Fprintln(ui.Out, "=====================================")
	fmt.Fprintln(ui.Out, "")
	ui.printHelp()
}

// printHelp displays available commands
func (ui *SimpleConsoleUI) printHelp() {
	fmt.Fprintln(ui.Out, "Available commands:")
	fmt.Fprintln(ui.Out, "  agent          - Enter chat mode with AI agent")
	fmt.Fprintln(ui.Out, "  article <ID>   - Run article processing flow for given ID")
	fmt.Fprintln(ui.Out, "  status         - Show status of last flow execution")
	fmt.Fprintln(ui.Out, "  help           - Show this help message")
	fmt.Fprintln(ui.Out, "  quit/exit      - Exit the console")
	fmt.Fprintln(ui.Out, "")
}

// executeCommand executes the given command
func (ui *SimpleConsoleUI) executeCommand(ctx context.Context, cmd Command) error {
	switch cmd.Name {
	case "help":
		ui.printHelp()
		return nil

	case "quit", "exit":
		fmt.Fprintln(ui.Out, "Goodbye!")
		return fmt.Errorf("exit requested")

	case "agent":
		return ui.enterChatMode(ctx)

	case "article":
		return ui.runArticleFlow(ctx, cmd.Args[0])

	case "status":
		return ui.showStatus()

	default:
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
}

// handleChatInput processes input in chat mode
func (ui *SimpleConsoleUI) handleChatInput(ctx context.Context, line string, reader *bufio.Reader) error {
	// Check for exit command
	if line == "exit" {
		ui.isInChatMode = false
		fmt.Fprintln(ui.Out, "Exiting chat mode")
		ui.chatHistory = make([]*interfaces.PonchoMessage, 0) // Clear history
		return nil
	}

	// Skip empty lines
	if line == "" {
		return nil
	}

	// Add user message to history
	userMsg := &interfaces.PonchoMessage{
		Role: interfaces.PonchoRoleUser,
		Content: []*interfaces.PonchoContentPart{
			{
				Type: interfaces.PonchoContentTypeText,
				Text: line,
			},
		},
	}
	ui.chatHistory = append(ui.chatHistory, userMsg)

	// TODO: Implement actual LLM call with streaming
	// For now, just echo back
	fmt.Fprintf(ui.Out, "assistant: You said: %s\n", line)

	// Add assistant response to history (placeholder)
	assistantMsg := &interfaces.PonchoMessage{
		Role: interfaces.PonchoRoleAssistant,
		Content: []*interfaces.PonchoContentPart{
			{
				Type: interfaces.PonchoContentTypeText,
				Text: "You said: " + line,
			},
		},
	}
	ui.chatHistory = append(ui.chatHistory, assistantMsg)

	return nil
}

// enterChatMode switches to chat mode
func (ui *SimpleConsoleUI) enterChatMode(ctx context.Context) error {
	ui.isInChatMode = true
	fmt.Fprintln(ui.Out, "Entering agent chat mode. Type 'exit' to return to main menu.")
	ui.chatHistory = make([]*interfaces.PonchoMessage, 0) // Clear history when entering chat mode
	return nil
}

// runArticleFlow executes the article processing flow
func (ui *SimpleConsoleUI) runArticleFlow(ctx context.Context, articleID string) error {
	fmt.Fprintf(ui.Out, "Starting article flow for %s...\n", articleID)

	// TODO: Implement actual ArticleFlow execution
	// For now, just simulate
	event := FlowEvent{
		Time:   time.Now(),
		Step:   "s3_load",
		Status: "started",
		Detail: fmt.Sprintf("loading article %s", articleID),
	}
	ui.OnEvent(event)

	event.Time = time.Now()
	event.Status = "completed"
	event.Detail = fmt.Sprintf("found 4 images for article %s", articleID)
	ui.OnEvent(event)

	fmt.Fprintf(ui.Out, "Article flow completed for %s\n", articleID)
	return nil
}

// showStatus displays the status of the last flow execution
func (ui *SimpleConsoleUI) showStatus() error {
	if ui.lastFlowEvent == nil {
		fmt.Fprintln(ui.Out, "No flow has been executed yet.")
		return nil
	}

	fmt.Fprintln(ui.Out, "Last Flow Status:")
	fmt.Fprintf(ui.Out, "  Step: %s\n", ui.lastFlowEvent.Step)
	fmt.Fprintf(ui.Out, "  Status: %s\n", ui.lastFlowEvent.Status)
	if ui.lastFlowEvent.Detail != "" {
		fmt.Fprintf(ui.Out, "  Detail: %s\n", ui.lastFlowEvent.Detail)
	}
	fmt.Fprintf(ui.Out, "  Time: %s\n", ui.lastFlowEvent.Time.Format("2006-01-02 15:04:05"))

	return nil
}

// OnEvent implements the FlowObserver interface
func (ui *SimpleConsoleUI) OnEvent(event FlowEvent) {
	// Store the event
	ui.lastFlowEvent = &event

	// Print the event
	fmt.Fprintf(ui.Out, "[%s] %-16s %-10s %s\n",
		event.Time.Format("15:04:05"),
		event.Step,
		event.Status,
		event.Detail)
}