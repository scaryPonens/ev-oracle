package db

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // PostgreSQL driver for golang-migrate
	"github.com/scaryPonens/ev-oracle/internal/models"
)

// Client represents a database client
type Client struct {
	pool        *pgxpool.Pool
	databaseURL string
}

// New creates a new database client
func New(ctx context.Context, databaseURL string) (*Client, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		pool:        pool,
		databaseURL: databaseURL,
	}, nil
}

// Close closes the database connection pool
func (c *Client) Close() {
	c.pool.Close()
}

// InitSchema initializes the database schema by running all pending migrations
// This is a convenience method that calls MigrateUp
func (c *Client) InitSchema(ctx context.Context) error {
	return c.MigrateUp(ctx)
}

// MigrateUp runs all pending migrations
func (c *Client) MigrateUp(ctx context.Context) error {
	m, err := c.getMigrateInstance()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func (c *Client) MigrateDown(ctx context.Context) error {
	m, err := c.getMigrateInstance()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// MigrateSteps runs n migrations (positive for up, negative for down)
func (c *Client) MigrateSteps(ctx context.Context, n int) error {
	m, err := c.getMigrateInstance()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}

	return nil
}

// getMigrateInstance creates a migrate instance for the database
func (c *Client) getMigrateInstance() (*migrate.Migrate, error) {
	// Get migrations directory path (relative to project root)
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations path: %w", err)
	}

	// Use the database URL directly
	// golang-migrate accepts both postgres:// and postgresql:// formats
	dbURL := c.databaseURL

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

// SimilaritySearch performs a vector similarity search
func (c *Client) SimilaritySearch(ctx context.Context, embedding []float32, limit int) ([]models.EVSpec, error) {
	// Format embedding as a string in pgvector format: [1.0,2.0,3.0]
	embeddingStrs := make([]string, len(embedding))
	for i, v := range embedding {
		embeddingStrs[i] = fmt.Sprintf("%g", v)
	}
	embeddingStr := "[" + strings.Join(embeddingStrs, ",") + "]"

	query := `
		SELECT 
			make, 
			model, 
			year, 
			capacity_kwh, 
			power_kw, 
			chemistry,
			1 - (embedding <=> $1::vector) as confidence
		FROM ev_specs
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`

	rows, err := c.pool.Query(ctx, query, embeddingStr, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var specs []models.EVSpec
	for rows.Next() {
		var spec models.EVSpec
		err := rows.Scan(
			&spec.Make,
			&spec.Model,
			&spec.Year,
			&spec.Capacity,
			&spec.Power,
			&spec.Chemistry,
			&spec.Confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		spec.Source = "database"
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return specs, nil
}

// InsertEVSpec inserts a new EV specification with its embedding
func (c *Client) InsertEVSpec(ctx context.Context, spec *models.EVSpec, embedding []float32) error {
	// Format embedding as a string in pgvector format: [1.0,2.0,3.0]
	embeddingStrs := make([]string, len(embedding))
	for i, v := range embedding {
		embeddingStrs[i] = fmt.Sprintf("%g", v)
	}
	embeddingStr := "[" + strings.Join(embeddingStrs, ",") + "]"

	query := `
		INSERT INTO ev_specs (make, model, year, capacity_kwh, power_kw, chemistry, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7::vector)
		ON CONFLICT (make, model, year) 
		DO UPDATE SET 
			capacity_kwh = EXCLUDED.capacity_kwh,
			power_kw = EXCLUDED.power_kw,
			chemistry = EXCLUDED.chemistry,
			embedding = EXCLUDED.embedding
	`

	_, err := c.pool.Exec(ctx, query,
		spec.Make,
		spec.Model,
		spec.Year,
		spec.Capacity,
		spec.Power,
		spec.Chemistry,
		embeddingStr,
	)
	if err != nil {
		return fmt.Errorf("failed to insert spec: %w", err)
	}

	return nil
}

// GetByMakeModelYear retrieves an EV spec by exact make, model, and year
func (c *Client) GetByMakeModelYear(ctx context.Context, make, model string, year int) (*models.EVSpec, error) {
	query := `
		SELECT make, model, year, capacity_kwh, power_kw, chemistry
		FROM ev_specs
		WHERE LOWER(make) = LOWER($1) AND LOWER(model) = LOWER($2) AND year = $3
	`

	var spec models.EVSpec
	err := c.pool.QueryRow(ctx, query, make, model, year).Scan(
		&spec.Make,
		&spec.Model,
		&spec.Year,
		&spec.Capacity,
		&spec.Power,
		&spec.Chemistry,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query spec: %w", err)
	}

	spec.Confidence = 1.0
	spec.Source = "database"

	return &spec, nil
}
