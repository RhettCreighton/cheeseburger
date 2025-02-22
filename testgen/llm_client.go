// File: cheeseburger/testgen/llm_client.go
package testgen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// ChatMessage represents a single message in the chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequestBody represents the request payload for the chat completions API.
type ChatRequestBody struct {
	Model               string            `json:"model"`
	Messages            []ChatMessage     `json:"messages"`
	MaxCompletionTokens int               `json:"max_completion_tokens"`
	ResponseFormat      map[string]string `json:"response_format"`
	ReasoningEffort     string            `json:"reasoning_effort"`
}

// ChatChoice represents one of the returned completions.
type ChatChoice struct {
	Message ChatMessage `json:"message"`
}

// ChatResponseBody represents the structure of the API response.
type ChatResponseBody struct {
	Choices []ChatChoice `json:"choices"`
}

// CallLLM sends the prompt to the OpenAI chat completions API and returns the generated test code.
func CallLLM(prompt string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY not set")
	}

	// Get model from environment variable MODEL or default.
	model := os.Getenv("MODEL")
	if model == "" {
		model = "o3-mini-2025-01-31"
	}

	// Adapt the reasoning effort based on the model name.
	var reasoningEffort string
	if strings.HasPrefix(model, "o3-mini") {
		switch model {
		case "o3-mini-high":
			reasoningEffort = "high"
		case "o3-mini-medium":
			reasoningEffort = "medium"
		default:
			reasoningEffort = "low"
		}
		// Use the default chat model version.
		model = "o3-mini-2025-01-31"
	} else {
		reasoningEffort = "high"
	}

	// Build the chat conversation:
	// First message: a developer message stating that you are a helpful assistant.
	// Second message: a user message containing the prompt.
	messages := []ChatMessage{
		{
			Role:    "developer",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	reqBody := ChatRequestBody{
		Model:               model,
		Messages:            messages,
		MaxCompletionTokens: 10000, // generous output length
		ResponseFormat:      map[string]string{"type": "text"},
		ReasoningEffort:     reasoningEffort,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response from OpenAI API: %d; response: %s", resp.StatusCode, string(bodyBytes))
	}

	var responseBody ChatResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	if len(responseBody.Choices) == 0 {
		return "", errors.New("no completions returned")
	}

	// Return the content string.
	if responseBody.Choices[0].Message.Content == "" {
		return "", errors.New("no content in response message")
	}
	return responseBody.Choices[0].Message.Content, nil
}
