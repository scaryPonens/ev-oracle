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
│   └── root.go            # Main query command
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
| `OPENAI_API_KEY` | OpenAI API key for embeddings | Yes |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude fallback | Yes |

### Example .env file

```bash
NEON_DATABASE_URL=postgresql://user:password@host/database?sslmode=require
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
```

Load environment variables:

```bash
export $(cat .env | xargs)
```

## Database Setup

The application will automatically create the necessary schema on first use. However, you need to ensure the pgvector extension is available:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

For Neon databases, pgvector is typically pre-installed.

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
- **internal/db/**: Database operations using pgx/v5 and pgvector
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
- Powered by [OpenAI](https://openai.com) embeddings
- Falls back to [Anthropic Claude](https://www.anthropic.com) for intelligent responses
