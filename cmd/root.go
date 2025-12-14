package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/scaryPonens/ev-oracle/internal/db"
	"github.com/scaryPonens/ev-oracle/internal/embedding"
	"github.com/scaryPonens/ev-oracle/internal/llm"
	"github.com/scaryPonens/ev-oracle/internal/models"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ev-oracle [make] [model] [year]",
	Short: "Query EV battery specifications",
	Long: `EV Oracle is a CLI tool for retrieving electric vehicle battery specifications.
It queries a pgvector-backed knowledge base and falls back to LLM reasoning when needed.

Example:
  ev-oracle Tesla "Model 3" 2023
  ev-oracle --json Nissan Leaf 2022`,
	Args: cobra.ExactArgs(3),
	RunE: runQuery,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output result in JSON format")
}

// runQuery executes the main query logic
func runQuery(cmd *cobra.Command, args []string) error {
	fmt.Printf("Running query for %s %s %s\n", args[0], args[1], args[2])
	make := args[0]
	model := args[1]
	yearStr := args[2]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return fmt.Errorf("invalid year: %s", yearStr)
	}

	// Load configuration
	cfg, err := models.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	ctx := context.Background()

	// Initialize database client
	dbClient, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbClient.Close()

	// Try exact match first
	spec, err := dbClient.GetByMakeModelYear(ctx, make, model, year)
	if err != nil {
		return fmt.Errorf("database query error: %w", err)
	}

	// If exact match found, return it
	if spec != nil {
		return outputSpec(spec)
	}

	// Initialize embedding service
	embeddingSvc := embedding.NewWithProvider(
		embedding.ProviderType(cfg.EmbeddingProvider),
		cfg.OpenAIAPIKey,
		cfg.OllamaURL,
		cfg.OllamaModel,
	)

	// Build query text and get embedding
	queryText := embedding.BuildQueryText(make, model, year)
	embeddingVector, err := embeddingSvc.GetEmbedding(queryText)
	if err != nil {
		return fmt.Errorf("failed to get embedding: %w", err)
	}

	// Perform similarity search
	results, err := dbClient.SimilaritySearch(ctx, embeddingVector, 1)
	if err != nil {
		return fmt.Errorf("similarity search error: %w", err)
	}

	// Check if we have results with sufficient confidence
	if len(results) > 0 && results[0].Confidence >= models.ConfidenceThreshold {
		return outputSpec(&results[0])
	}

	fmt.Println("Falling back to LLM")
	// Fall back to LLM
	llmSvc := llm.NewWithProvider(
		llm.ProviderType(cfg.LLMProvider),
		cfg.AnthropicAPIKey,
		cfg.OllamaURL,
		cfg.OllamaLLMModel,
	)
	spec, err = llmSvc.QueryEVSpecs(make, model, year)
	if err != nil {
		return fmt.Errorf("LLM query error: %w", err)
	}

	return outputSpec(spec)
}

// outputSpec outputs the EV spec in the requested format
func outputSpec(spec *models.EVSpec) error {
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(spec); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	} else {
		fmt.Printf("Make:       %s\n", spec.Make)
		fmt.Printf("Model:      %s\n", spec.Model)
		fmt.Printf("Year:       %d\n", spec.Year)
		fmt.Printf("Capacity:   %.1f kWh\n", spec.Capacity)
		fmt.Printf("Power:      %.1f kW\n", spec.Power)
		fmt.Printf("Chemistry:  %s\n", spec.Chemistry)
		fmt.Printf("Confidence: %.2f\n", spec.Confidence)
		fmt.Printf("Source:     %s\n", spec.Source)
	}
	return nil
}
