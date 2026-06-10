// Package ai provides a thin, reusable client for communicating with Ollama.
// Identical across all bootcamp projects — only this file changes when
// swapping AI providers.
package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// DefaultModel is the general-purpose model used across all projects.
	DefaultModel = "llama3.2:3b"

	ollamaBaseURL = "http://localhost:11434"
	chatEndpoint  = ollamaBaseURL + "/api/chat"
)

// Message is a single turn in a conversation.
// Role must be one of: "system", "user", or "assistant".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type streamChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// Chat sends messages to Ollama and returns the complete response as a string.
// Used in Project 02 so far, any task where complete output is better
// than incremental output — for example, when we need to parse the full response as JSON, or when
// the response is short enough that streaming isn't worth the complexity
func Chat(model string, messages []Message) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return "", fmt.Errorf("ai: marshal request: %w", err)
	}

	resp, err := http.Post(chatEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ai: ollama unreachable — is `ollama serve` running? %w", err)
	}
	defer resp.Body.Close()

	var result streamChunk
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}

	return result.Message.Content, nil
}

// ChatStream sends messages to Ollama and calls onChunk for every token.
// Used in Projects 01 so far, tasks where partial output has immediate value.
func ChatStream(model string, messages []Message, onChunk func(string) error) error {
	body, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("ai: marshal request: %w", err)
	}

	resp, err := http.Post(chatEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ai: ollama unreachable — is `ollama serve` running? %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk streamChunk
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue
		}
		if chunk.Message.Content != "" {
			if err := onChunk(chunk.Message.Content); err != nil {
				return nil
			}
		}
		if chunk.Done {
			break
		}
	}

	return scanner.Err()
}
