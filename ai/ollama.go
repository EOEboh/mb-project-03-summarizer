// Package ai provides a thin, reusable client for communicating with Ollama.
//
// Every project in this bootcamp imports this package.
// Two functions cover every use case:
//
//   - Chat(): non-streaming, returns the full response as a string
//   - ChatStream(): streaming, calls onChunk for each token as it arrives
//
// To swap Ollama for Groq or another OpenAI-compatible provider, only this
// file needs to change: no change to handler or main.go is neeeded.
// That's the value of keeping AI communication in its own package.
package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// DefaultModel is the model used when callers don't specify one.
	// Students with 16GB+ RAM can swap this for llama3.1:8b.
	DefaultModel = "llama3.2:3b"

	ollamaBaseURL = "http://localhost:11434"
	chatEndpoint  = ollamaBaseURL + "/api/chat"
)

// Message is a single turn in a conversation.
// Role must be one of: "system", "user" or "assistant".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// - Internal types ────────────────────────────────────────────────────────────

// chatRequest is the JSON payload we POST to Ollama
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// streamChunk is one JSON object per line that Ollama sends back when Stream:true.
// When Stream:false, Ollama sends just one JSON object with the full response,
// but we can reuse this struct for both cases.
type streamChunk struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// - Public API ────────────────────────────────────────────────────────────────

// Chat sends messages to Ollama and returns the complete response as a string.
//
// Use this for: summarizers, analyzers, document Q&A, anything where you
// want the full answer before rendering it
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

	// When Stream:false, Ollama returns a single JSON object
	var result streamChunk
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}

	return result.Message.Content, nil
}

// ChatStream sends messages to Ollama and calls onChunk for every token
// as it arrives. The stream ends when Ollama sends Done:true.
//
// Use this for: chat interfaces, writing assistants: anything that benefits
// from real-time output in the browser.
//
// onChunk receives one token at a time. Return a non-nil error from onChunk
// to abort the stream early (e.g. when the client disconnects).
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

	// Ollama sends one JSON object per line when streaming
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk streamChunk
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue // skip malformed lines
		}

		if chunk.Message.Content != "" {
			if err := onChunk(chunk.Message.Content); err != nil {
				// Caller wants to stop the stream early (e.g. client disconnected): normal, not an error
				return nil
			}
		}

		if chunk.Done {
			break
		}
	}

	return scanner.Err()
}
