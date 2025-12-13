package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/scaryPonens/ev-oracle/internal/models"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	claudeModel     = "claude-3-5-sonnet-20241022"
)

// Service handles LLM operations for fallback queries
type Service struct {
	apiKey string
	client *http.Client
}

// New creates a new LLM service
func New(apiKey string) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{},
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

// QueryEVSpecs queries Claude API for EV battery specifications
func (s *Service) QueryEVSpecs(make, model string, year int) (*models.EVSpec, error) {
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
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
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

// parseEVSpecs parses the Claude response text into an EVSpec
func parseEVSpecs(text, make, model string, year int) (*models.EVSpec, error) {
	spec := &models.EVSpec{
		Make:       make,
		Model:      model,
		Year:       year,
		Confidence: 0.5, // LLM responses get lower confidence
		Source:     "llm",
	}

	// Extract capacity
	capacityRe := regexp.MustCompile(`(?i)capacity:\s*([0-9.]+)\s*kWh`)
	if matches := capacityRe.FindStringSubmatch(text); len(matches) > 1 {
		if capacity, err := strconv.ParseFloat(matches[1], 64); err == nil {
			spec.Capacity = capacity
		}
	}

	// Extract power
	powerRe := regexp.MustCompile(`(?i)power:\s*([0-9.]+)\s*kW`)
	if matches := powerRe.FindStringSubmatch(text); len(matches) > 1 {
		if power, err := strconv.ParseFloat(matches[1], 64); err == nil {
			spec.Power = power
		}
	}

	// Extract chemistry
	chemistryRe := regexp.MustCompile(`(?i)chemistry:\s*([^\n]+)`)
	if matches := chemistryRe.FindStringSubmatch(text); len(matches) > 1 {
		spec.Chemistry = matches[1]
	}

	// Validate that we got at least some data
	if spec.Capacity == 0 && spec.Power == 0 && spec.Chemistry == "" {
		return nil, fmt.Errorf("failed to extract any specifications from response")
	}

	return spec, nil
}
