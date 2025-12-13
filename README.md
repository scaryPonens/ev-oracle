# EV Oracle

A Go-powered CLI tool for retrieving electric vehicle battery specifications (capacity, power, chemistry) by make/model/year. Features intelligent similarity search using Neon PostgreSQL with pgvector and OpenAI embeddings, with Claude API fallback for low-confidence results.

## Features

- 🔍 **Vector Similarity Search**: Uses pgvector for semantic search of EV specifications
- 🤖 **OpenAI Embeddings**: Converts queries to embeddings for accurate similarity matching
- 🧠 **Claude Fallback**: Automatically falls back to Claude API when confidence < 0.8
- 📊 **Multiple Output Formats**: Supports human-readable and JSON output
- ⚡ **Fast & Efficient**: Built with Go for performance and reliability

## Architecture

The project follows a clean architecture pattern:

```
ev-oracle/
├── cmd/                    # CLI commands
│   ├── root.go            # Main query command
│   ├── init.go            # Database initialization
│   └── migrate.go         # Migration commands
├── migrations/            # Database migration files
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── internal/
│   ├── db/                # Database layer (pgx/v5, pgvector)
│   ├── embedding/         # OpenAI embeddings service
│   ├── llm/              # Claude API integration
│   └── models/           # Data models and configuration
└── main.go               # Entry point
```

## Prerequisites

- Go 1.21 or later
- PostgreSQL database with pgvector extension (Neon recommended)
- OpenAI API key
- Anthropic API key (for Claude)

## Installation

### From Source

```bash
git clone https://github.com/scaryPonens/ev-oracle.git
cd ev-oracle
go build -o ev-oracle .
```

### Using Go Install

```bash
go install github.com/scaryPonens/ev-oracle@latest
```

## Configuration

The application uses environment variables for configuration:

| Variable | Description | Required |
|----------|-------------|----------|
| `NEON_DATABASE_URL` | PostgreSQL connection string (with pgvector) | Yes |
| `EMBEDDING_PROVIDER` | Embedding provider: `openai` or `ollama` (default: `openai`) | No |
| `OPENAI_API_KEY` | OpenAI API key for embeddings (required if using OpenAI) | Conditional |
| `OLLAMA_URL` | Ollama API URL (default: `http://localhost:11434`) | No |
| `OLLAMA_MODEL` | Ollama embedding model (default: `nomic-embed-text`) | No |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude fallback | Yes |

### Example .env file

**Using OpenAI (default):**
```bash
NEON_DATABASE_URL=postgresql://user:password@host/database?sslmode=require
EMBEDDING_PROVIDER=openai
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
```

**Using Ollama:**
```bash
NEON_DATABASE_URL=postgresql://user:password@host/database?sslmode=require
EMBEDDING_PROVIDER=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=nomic-embed-text
ANTHROPIC_API_KEY=sk-ant-...
```

The application automatically loads the `.env` file if it exists. You don't need to manually export the variables.

**Note:** The `.env` file is gitignored by default to keep your secrets safe.

### Using Ollama for Embeddings

To use Ollama instead of OpenAI for embeddings:

1. **Install Ollama**: Download from [ollama.com](https://ollama.com)

2. **Pull an embedding model**:
   ```bash
   ollama pull nomic-embed-text
   ```

3. **Start Ollama** (if not running as a service):
   ```bash
   ollama serve
   ```

4. **Configure your `.env` file**:
   ```bash
   EMBEDDING_PROVIDER=ollama
   OLLAMA_URL=http://localhost:11434
   OLLAMA_MODEL=nomic-embed-text
   ```

**Important Note:** The default database schema expects 1536-dimensional vectors (OpenAI's `text-embedding-3-small`). Ollama's `nomic-embed-text` produces 768-dimensional vectors. If you want to use Ollama, you'll need to:

- Create a migration to change the embedding dimension in the database schema, OR
- Use an Ollama model that produces 1536 dimensions (if available)

To create a migration for Ollama's 768 dimensions:
```bash
# Create a new migration file
# migrations/000002_update_embedding_dimension.up.sql
ALTER TABLE ev_specs ALTER COLUMN embedding TYPE vector(768);
DROP INDEX IF EXISTS ev_specs_embedding_idx;
CREATE INDEX ev_specs_embedding_idx ON ev_specs 
 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

## Database Setup

### Initial Setup

Initialize the database schema by running:

```bash
ev-oracle init
```

This will run all pending migrations to set up the necessary tables and indexes. The pgvector extension will be automatically enabled.

For Neon databases, pgvector is typically pre-installed.

### Migrations

The project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management. Migration files are stored in the `migrations/` directory.

**Run all pending migrations:**
```bash
ev-oracle migrate up
```

**Roll back the last migration:**
```bash
ev-oracle migrate down
```

**Run a specific number of migrations:**
```bash
ev-oracle migrate --steps 2  # Run 2 migrations forward
ev-oracle migrate --steps -1 # Roll back 1 migration
```

### Creating New Migrations

To create a new migration, add files to the `migrations/` directory following the naming pattern:
- `00000N_description.up.sql` - Migration to apply
- `00000N_description.down.sql` - Migration to rollback

The migration number should be sequential and unique.

## Usage

### Basic Query

```bash
ev-oracle Tesla "Model 3" 2023
```

Output:
```
Make:       Tesla
Model:      Model 3
Year:       2023
Capacity:   75.0 kWh
Power:      283.0 kW
Chemistry:  NMC (Nickel Manganese Cobalt)
Confidence: 1.00
Source:     database
```

### JSON Output

```bash
ev-oracle --json Nissan Leaf 2022
```

Output:
```json
{
  "make": "Nissan",
  "model": "Leaf",
  "year": 2022,
  "capacity_kwh": 40.0,
  "power_kw": 110.0,
  "chemistry": "Li-ion",
  "confidence": 0.95,
  "source": "database"
}
```

### Help

```bash
ev-oracle --help
```

## How It Works

1. **Exact Match**: First tries to find an exact match in the database by make/model/year
2. **Similarity Search**: If no exact match, converts the query to an embedding and performs vector similarity search
3. **Confidence Check**: If the best match has confidence ≥ 0.8, returns it
4. **LLM Fallback**: If confidence < 0.8, queries Claude API for the information
5. **Output**: Returns the result in the requested format (text or JSON)

## Development

### Project Structure

- **cmd/root.go**: Main CLI command implementation
- **cmd/init.go**: Database initialization command
- **cmd/migrate.go**: Database migration commands
- **migrations/**: SQL migration files (up/down)
- **internal/db/**: Database operations using pgx/v5 and pgvector with migration support
- **internal/embedding/**: OpenAI embeddings integration
- **internal/llm/**: Claude API integration for fallback queries
- **internal/models/**: Data models and configuration using functional options pattern

### Building

```bash
go build -o ev-oracle .
```

### Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [pgx](https://github.com/jackc/pgx) for PostgreSQL connectivity
- Database migrations powered by [golang-migrate](https://github.com/golang-migrate/migrate)
- Powered by [OpenAI](https://openai.com) embeddings
- Falls back to [Anthropic Claude](https://www.anthropic.com) for intelligent responses
