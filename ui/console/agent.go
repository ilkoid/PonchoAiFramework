package console

import (
	"bufio"
	"context"
	"fmt"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// AgentConfig represents configuration for the AI agent
type AgentConfig struct {
	ModelName string            `json:"model_name"`
	System    string            `json:"system"`
	Temperature *float32        `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultAgentConfig returns default configuration for the agent
func DefaultAgentConfig() *AgentConfig {
	temperature := float32(0.7)
	return &AgentConfig{
		ModelName:   "deepseek-chat", // Default model
		System:      "You are a helpful AI assistant. Provide clear and concise answers.",
		Temperature: &temperature,
		MaxTokens:   2000,
		Stream:      true,
	}
}

// handleAgentInput processes input in chat mode with real AI model integration
func (ui *SimpleConsoleUI) handleAgentInputWithAI(ctx context.Context, line string, reader *bufio.Reader) error {
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

	// Check if we have any models registered
	models := ui.Framework.GetModelRegistry().List()
	if len(models) == 0 {
		// Fallback to echo if no model available
		fmt.Fprintf(ui.Out, "assistant: You said: %s\n", line)
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

	// Prepare request
	config := DefaultAgentConfig()

	// Use the first available model if default is not available
	if len(models) > 0 {
		config.ModelName = models[0] // Use first registered model
	}

	request := &interfaces.PonchoModelRequest{
		Model:       config.ModelName,
		Messages:    ui.buildMessagesWithSystem(config.System),
		Temperature: config.Temperature,
		MaxTokens:   &config.MaxTokens,
		Stream:      config.Stream,
		Metadata:    config.Metadata,
	}

	// Stream the response
	fmt.Fprint(ui.Out, "assistant: ")
	fullResponse := ""

	// Handle streaming response using framework
	if config.Stream {
		err := ui.Framework.GenerateStreaming(ctx, request, func(chunk *interfaces.PonchoStreamChunk) error {
			if chunk.Done {
				return nil
			}

			if chunk.Delta != nil && chunk.Delta.Content != nil {
				for _, part := range chunk.Delta.Content {
					if part.Type == interfaces.PonchoContentTypeText {
						fmt.Fprint(ui.Out, part.Text)
						fullResponse += part.Text
					}
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to generate streaming response: %w", err)
		}
		fmt.Fprintln(ui.Out) // New line after streaming
	} else {
		// Non-streaming response
		response, err := ui.Framework.Generate(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}

		if response.Message != nil && response.Message.Content != nil {
			for _, part := range response.Message.Content {
				if part.Type == interfaces.PonchoContentTypeText {
					fmt.Fprint(ui.Out, part.Text)
					fullResponse += part.Text
				}
			}
		}
		fmt.Fprintln(ui.Out) // New line after response
	}

	// Add assistant response to history
	if fullResponse != "" {
		assistantMsg := &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleAssistant,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: fullResponse,
				},
			},
		}
		ui.chatHistory = append(ui.chatHistory, assistantMsg)
	}

	return nil
}

// buildMessagesWithSystem adds system message if not already present
func (ui *SimpleConsoleUI) buildMessagesWithSystem(systemPrompt string) []*interfaces.PonchoMessage {
	messages := make([]*interfaces.PonchoMessage, 0, len(ui.chatHistory)+1)

	// Add system message if not present
	if len(ui.chatHistory) == 0 || ui.chatHistory[0].Role != interfaces.PonchoRoleSystem {
		messages = append(messages, &interfaces.PonchoMessage{
			Role: interfaces.PonchoRoleSystem,
			Content: []*interfaces.PonchoContentPart{
				{
					Type: interfaces.PonchoContentTypeText,
					Text: systemPrompt,
				},
			},
		})
	}

	// Add chat history
	messages = append(messages, ui.chatHistory...)

	return messages
}

// limitChatHistory limits the chat history to prevent context overflow
func (ui *SimpleConsoleUI) limitChatHistory(maxMessages int) {
	if len(ui.chatHistory) <= maxMessages {
		return
	}

	// Keep system message if present and trim from the start (after system)
	startIdx := 0
	if len(ui.chatHistory) > 0 && ui.chatHistory[0].Role == interfaces.PonchoRoleSystem {
		startIdx = 1
	}

	// Calculate how many to remove
	toRemove := len(ui.chatHistory) - maxMessages
	if startIdx > 0 {
		toRemove++ // Account for system message
	}

	ui.chatHistory = append(ui.chatHistory[:startIdx], ui.chatHistory[startIdx+toRemove:]...)
}