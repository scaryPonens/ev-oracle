package cmd

import (
	"context"
	"fmt"

	"github.com/scaryPonens/ev-oracle/internal/db"
	"github.com/scaryPonens/ev-oracle/internal/models"
	"github.com/spf13/cobra"
)

var (
	migrateSteps int
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate [direction]",
	Short: "Run database migrations",
	Long: `Run database migrations to update the database schema.

Direction can be:
  up   - Run all pending migrations (default)
  down - Roll back the last migration

Alternatively, use the --steps flag to run a specific number of migrations:
  --steps N  - Run N migrations forward (positive number)
  --steps -N - Roll back N migrations (negative number)

Examples:
  ev-oracle migrate up
  ev-oracle migrate down
  ev-oracle migrate --steps 2
  ev-oracle migrate --steps -1`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().IntVar(&migrateSteps, "steps", 0, "Number of migration steps to run (positive for up, negative for down)")
}

func runMigrate(cmd *cobra.Command, args []string) error {
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

	// If steps flag is set, use it (takes precedence)
	if migrateSteps != 0 {
		if err := dbClient.MigrateSteps(ctx, migrateSteps); err != nil {
			return fmt.Errorf("failed to run migration steps: %w", err)
		}
		fmt.Printf("Successfully ran %d migration step(s)\n", migrateSteps)
		return nil
	}

	// Otherwise use direction argument
	direction := "up"
	if len(args) > 0 {
		direction = args[0]
	}

	switch direction {
	case "up":
		if err := dbClient.MigrateUp(ctx); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		fmt.Println("Migrations applied successfully!")
	case "down":
		if err := dbClient.MigrateDown(ctx); err != nil {
			return fmt.Errorf("failed to rollback migration: %w", err)
		}
		fmt.Println("Migration rolled back successfully!")
	default:
		return fmt.Errorf("invalid direction: %s. Use 'up' or 'down'", direction)
	}

	return nil
}

