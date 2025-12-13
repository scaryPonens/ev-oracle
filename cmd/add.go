package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/scaryPonens/ev-oracle/internal/db"
	"github.com/scaryPonens/ev-oracle/internal/embedding"
	"github.com/scaryPonens/ev-oracle/internal/models"
	"github.com/spf13/cobra"
)

var (
	capacity  float64
	power     float64
	chemistry string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [make] [model] [year]",
	Short: "Add an EV specification to the database",
	Long: `Add an electric vehicle specification to the database with embedding.
This command is useful for populating the database with known EV specs.

Example:
  ev-oracle add Tesla "Model 3" 2023 --capacity 75.0 --power 283.0 --chemistry "NMC"`,
	Args: cobra.ExactArgs(3),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().Float64Var(&capacity, "capacity", 0, "Battery capacity in kWh (required)")
	addCmd.Flags().Float64Var(&power, "power", 0, "Power output in kW (required)")
	addCmd.Flags().StringVar(&chemistry, "chemistry", "", "Battery chemistry type (required)")
	addCmd.MarkFlagRequired("capacity")
	addCmd.MarkFlagRequired("power")
	addCmd.MarkFlagRequired("chemistry")
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Initialize embedding service
	embeddingSvc := embedding.New(cfg.OpenAIAPIKey)

	// Create the EV spec
	spec := &models.EVSpec{
		Make:      make,
		Model:     model,
		Year:      year,
		Capacity:  capacity,
		Power:     power,
		Chemistry: chemistry,
	}

	// Generate embedding
	queryText := embedding.BuildQueryText(make, model, year)
	embeddingVector, err := embeddingSvc.GetEmbedding(queryText)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Insert into database
	if err := dbClient.InsertEVSpec(ctx, spec, embeddingVector); err != nil {
		return fmt.Errorf("failed to insert spec: %w", err)
	}

	fmt.Printf("Successfully added %d %s %s to the database!\n", year, make, model)
	fmt.Printf("  Capacity: %.1f kWh\n", capacity)
	fmt.Printf("  Power: %.1f kW\n", power)
	fmt.Printf("  Chemistry: %s\n", chemistry)

	return nil
}
