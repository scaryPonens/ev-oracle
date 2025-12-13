package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaryPonens/ev-oracle/internal/models"
)

// Client represents a database client
type Client struct {
	pool *pgxpool.Pool
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

	return &Client{pool: pool}, nil
}

// Close closes the database connection pool
func (c *Client) Close() {
	c.pool.Close()
}

// InitSchema initializes the database schema with pgvector extension
func (c *Client) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE TABLE IF NOT EXISTS ev_specs (
			id SERIAL PRIMARY KEY,
			make VARCHAR(100) NOT NULL,
			model VARCHAR(100) NOT NULL,
			year INTEGER NOT NULL,
			capacity_kwh FLOAT NOT NULL,
			power_kw FLOAT NOT NULL,
			chemistry VARCHAR(100) NOT NULL,
			embedding vector(1536),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(make, model, year)
		)`,
		`CREATE INDEX IF NOT EXISTS ev_specs_embedding_idx ON ev_specs 
		 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`,
	}

	for _, query := range queries {
		if _, err := c.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SimilaritySearch performs a vector similarity search
func (c *Client) SimilaritySearch(ctx context.Context, embedding []float32, limit int) ([]models.EVSpec, error) {
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

	rows, err := c.pool.Query(ctx, query, embedding, limit)
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
	query := `
		INSERT INTO ev_specs (make, model, year, capacity_kwh, power_kw, chemistry, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
		embedding,
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
