package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/scaryPonens/ev-oracle/internal/models"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	claudeModel     = "claude-3-5-sonnet-20241022"
)

// ProviderType represents the LLM provider
type ProviderType string

const (
	ProviderClaude ProviderType = "claude"
	ProviderOllama ProviderType = "ollama"
)

// Compile regular expressions once at package initialization
var (
	capacityRe  = regexp.MustCompile(`(?i)capacity:\s*([0-9.]+)\s*kWh`)
	powerRe     = regexp.MustCompile(`(?i)power:\s*([0-9.]+)\s*kW`)
	chemistryRe = regexp.MustCompile(`(?i)chemistry:\s*([^\n]+)`)
)

// Service handles LLM operations for fallback queries
type Service struct {
	provider     ProviderType
	anthropicKey string
	ollamaURL    string
	ollamaModel  string
	client       *http.Client
}

// New creates a new LLM service with Claude (legacy)
func New(apiKey string) *Service {
	return &Service{
		provider:     ProviderClaude,
		anthropicKey: apiKey,
		client:       &http.Client{},
	}
}

// NewWithProvider creates a new LLM service with the specified provider
func NewWithProvider(provider ProviderType, anthropicKey, ollamaURL, ollamaModel string) *Service {
	return &Service{
		provider:     provider,
		anthropicKey: anthropicKey,
		ollamaURL:    ollamaURL,
		ollamaModel:  ollamaModel,
		client:       &http.Client{},
	}
}

// claudeRequest represents the request to Claude API
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

// claudeMessage represents a message in the Claude API request
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse represents the response from Claude API
type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// QueryEVSpecs queries the LLM API for EV battery specifications
func (s *Service) QueryEVSpecs(make, model string, year int) (*models.EVSpec, error) {
	switch s.provider {
	case ProviderOllama:
		return s.queryOllama(make, model, year)
	case ProviderClaude:
		fallthrough
	default:
		return s.queryClaude(make, model, year)
	}
}

// queryClaude queries Claude API for EV battery specifications
func (s *Service) queryClaude(make, model string, year int) (*models.EVSpec, error) {
	prompt := fmt.Sprintf(`Please provide the battery specifications for the %d %s %s electric vehicle.

Return ONLY the following information in this exact format:
Capacity: [number] kWh
Power: [number] kW
Chemistry: [chemistry type]

If you don't have exact information, provide your best estimate based on similar models and clearly indicate it's an estimate.`, year, make, model)

	reqBody := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 1024,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Parse the response text
	spec, err := parseEVSpecs(claudeResp.Content[0].Text, make, model, year)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return spec, nil
}

// ollamaRequest represents the request to Ollama API
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaResponse represents the response from Ollama API
type ollamaResponse struct {
	Response string `json:"response"`
}

// queryOllama queries Ollama API for EV battery specifications
func (s *Service) queryOllama(make, model string, year int) (*models.EVSpec, error) {
	fmt.Println("Querying Ollama for", year, make, model)
	prompt := fmt.Sprintf(`Please provide the DC fast charging capabilities of the %d %s %s. 
Where "Power" is the peak rate at which the vehicle can DC fast charge.  

Return ONLY the following information in this exact format: 
Capacity: [number] 
kWh Power: [number] kW 
Chemistry: [chemistry type]

If you don't have exact information, provide your best estimate based on similar models.`, year, make, model)

	reqBody := ollamaRequest{
		Model:  s.ollamaModel,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", s.ollamaURL)
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
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Response == "" {
		return nil, fmt.Errorf("no response from ollama")
	}
	fmt.Println("Ollama response:", ollamaResp.Response)
	// Parse the response text
	spec, err := parseEVSpecs(ollamaResp.Response, make, model, year)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return spec, nil
}

// parseEVSpecs parses the Claude response text into an EVSpec
func parseEVSpecs(text, make, model string, year int) (*models.EVSpec, error) {
	spec := &models.EVSpec{
		Make:       make,
		Model:      model,
		Year:       year,
		Confidence: models.LLMConfidenceScore,
		Source:     "llm",
	}

	// Extract capacity using pre-compiled regex
	if matches := capacityRe.FindStringSubmatch(text); len(matches) > 1 {
		if capacity, err := strconv.ParseFloat(matches[1], 64); err == nil {
			spec.Capacity = capacity
		}
	}

	// Extract power using pre-compiled regex
	if matches := powerRe.FindStringSubmatch(text); len(matches) > 1 {
		if power, err := strconv.ParseFloat(matches[1], 64); err == nil {
			spec.Power = power
		}
	}

	// Extract chemistry using pre-compiled regex and trim whitespace
	if matches := chemistryRe.FindStringSubmatch(text); len(matches) > 1 {
		spec.Chemistry = strings.TrimSpace(matches[1])
	}

	// Validate that we got at least some data
	if spec.Capacity == 0 && spec.Power == 0 && spec.Chemistry == "" {
		return nil, fmt.Errorf("failed to extract any specifications from response")
	}

	return spec, nil
}
