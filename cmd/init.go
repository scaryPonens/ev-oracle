package cmd

import (
	"context"
	"fmt"

	"github.com/scaryPonens/ev-oracle/internal/db"
	"github.com/scaryPonens/ev-oracle/internal/models"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database schema",
	Long: `Initialize the database schema by creating the necessary tables and indexes.
This command should be run once before using the query command.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
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

	// Initialize schema
	if err := dbClient.InitSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	fmt.Println("Database schema initialized successfully!")
	fmt.Println("You can now use 'ev-oracle' to query EV specifications.")
	
	return nil
}
