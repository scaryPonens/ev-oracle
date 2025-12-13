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

// Service handles text-to-vector embedding operations
type Service struct {
	apiKey string
	client *http.Client
}

// New creates a new embedding service
func New(apiKey string) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{},
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

// GetEmbedding converts text to a vector embedding using OpenAI
func (s *Service) GetEmbedding(text string) ([]float32, error) {
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	
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

// BuildQueryText creates a search query text from make, model, and year
func BuildQueryText(make, model string, year int) string {
	return fmt.Sprintf("%s %s %d battery specifications", make, model, year)
}
