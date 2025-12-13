package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openaiEmbeddingURL = "https://api.openai.com/v1/embeddings"
	// embeddingModel is the OpenAI model used for generating embeddings
	// This model produces 1536-dimensional vectors
	embeddingModel = "text-embedding-3-small"
)

// ProviderType represents the embedding provider
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderOllama ProviderType = "ollama"
)

// Service handles text-to-vector embedding operations
type Service struct {
	provider    ProviderType
	openAIKey   string
	ollamaURL   string
	ollamaModel string
	client      *http.Client
}

// New creates a new embedding service with OpenAI
func New(apiKey string) *Service {
	return &Service{
		provider:  ProviderOpenAI,
		openAIKey: apiKey,
		client:    &http.Client{},
	}
}

// NewWithProvider creates a new embedding service with the specified provider
func NewWithProvider(provider ProviderType, openAIKey, ollamaURL, ollamaModel string) *Service {
	return &Service{
		provider:    provider,
		openAIKey:   openAIKey,
		ollamaURL:   ollamaURL,
		ollamaModel: ollamaModel,
		client:      &http.Client{},
	}
}

// openAIEmbeddingRequest represents the request to OpenAI's embedding API
type openAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// openAIEmbeddingResponse represents the response from OpenAI's embedding API
type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// GetEmbedding converts text to a vector embedding
func (s *Service) GetEmbedding(text string) ([]float32, error) {
	switch s.provider {
	case ProviderOllama:
		return s.getOllamaEmbedding(text)
	case ProviderOpenAI:
		fallthrough
	default:
		return s.getOpenAIEmbedding(text)
	}
}

// getOpenAIEmbedding converts text to a vector embedding using OpenAI
func (s *Service) getOpenAIEmbedding(text string) ([]float32, error) {
	reqBody := openAIEmbeddingRequest{
		Input: text,
		Model: embeddingModel,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openaiEmbeddingURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.openAIKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embeddingResp openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// ollamaEmbeddingRequest represents the request to Ollama's embedding API
type ollamaEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// ollamaEmbeddingResponse represents the response from Ollama's embedding API
type ollamaEmbeddingResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

// getOllamaEmbedding converts text to a vector embedding using Ollama
func (s *Service) getOllamaEmbedding(text string) ([]float32, error) {
	reqBody := ollamaEmbeddingRequest{
		Model: s.ollamaModel,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embed", s.ollamaURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embeddingResp ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddingResp.Embeddings) == 0 || len(embeddingResp.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// Convert []float64 to []float32
	// Ollama returns embeddings as an array of arrays, we take the first one
	embedding := make([]float32, len(embeddingResp.Embeddings[0]))
	for i, v := range embeddingResp.Embeddings[0] {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// BuildQueryText creates a search query text from make, model, and year
func BuildQueryText(make, model string, year int) string {
	return fmt.Sprintf("%s %s %d battery specifications", make, model, year)
}
