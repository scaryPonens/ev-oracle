# ⚡ EV Oracle

> Ask about any EV battery, get instant answers.

A blazing-fast CLI tool that retrieves electric vehicle battery specifications using vector similarity search with intelligent LLM fallback.

## 🎯 Features

- **Vector-First Search**: Query a pgvector knowledge base for sub-second lookups
- **Smart Fallback**: Claude AI fills gaps when the database doesn't have an answer
- **Rich Battery Data**: Capacity (kWh), power nameplate (kW), chemistry, and more
- **Multiple Output Formats**: Human-readable text, JSON, or YAML
- **Confidence Scoring**: Transparent similarity scores show answer reliability

## 🚀 Quick Start
```bash
# Install
go install github.com/scaryPonens/ev-oracle@latest

# Query a vehicle
ev-oracle tesla "model 3" 2023

# Output:
# Tesla Model 3 (2023)
# Battery Capacity: 75 kWh
# Power Nameplate: 208 kW
# Chemistry: NMC (Nickel Manganese Cobalt)
# Source: vector_db (confidence: 0.94)
```

## 📦 Installation

### From Source
```bash
git clone https://github.com/scaryPonens/ev-oracle.git
cd ev-oracle
go build -o ev-oracle
```

### Prerequisites

- Go 1.21+
- Neon PostgreSQL database with pgvector extension
- OpenAI API key (for embeddings)
- Anthropic API key (for LLM fallback)

## ⚙️ Configuration

Set environment variables:
```bash
export NEON_DATABASE_URL="postgresql://user:pass@host/dbname"
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
```

Or create `.env` file:
```env
NEON_DATABASE_URL=postgresql://user:pass@host/dbname
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
CONFIDENCE_THRESHOLD=0.8
```

## 🎮 Usage

### Basic Query
```bash
ev-oracle <make> <model> <year>
```

### With Flags
```bash
# JSON output
ev-oracle nissan leaf 2022 --json

# Custom confidence threshold
ev-oracle ford "f-150 lightning" 2023 --threshold 0.85

# Verbose mode (show embedding/query details)
ev-oracle rivian r1t 2024 --verbose
```

## 🏗️ How It Works
```
User Input → Generate Embedding → Vector Search (Neon/pgvector)
                                          ↓
                                   Confidence > 0.8?
                                    ↙           ↘
                              Return Result    Claude API
                                                    ↓
                                              Parse Response
                                                    ↓
                                              Cache to Vector DB
```

1. **Embedding Generation**: Input text converted to 1536-dim vector via OpenAI
2. **Similarity Search**: Cosine similarity query against pgvector index
3. **Confidence Check**: Results above threshold returned immediately
4. **LLM Fallback**: Claude analyzes the query and returns structured battery data
5. **Caching**: LLM responses stored in vector DB for future queries

## 🗄️ Database Schema
```sql
CREATE TABLE ev_specs (
    id SERIAL PRIMARY KEY,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER NOT NULL,
    battery_capacity_kwh DECIMAL,
    power_nameplate_kw DECIMAL,
    chemistry TEXT,
    metadata JSONB,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX ON ev_specs USING hnsw (embedding vector_cosine_ops);
```

## 🧪 Example Output

### Text Format (Default)
```
Chevrolet Bolt EV (2022)
━━━━━━━━━━━━━━━━━━━━━━━━━━
Battery Capacity:     65 kWh
Power Nameplate:      150 kW
Chemistry:            NMC
Range (EPA):          259 miles
Source:               vector_db
Confidence:           0.92
```

### JSON Format
```json
{
  "make": "Chevrolet",
  "model": "Bolt EV",
  "year": 2022,
  "battery": {
    "capacity_kwh": 65,
    "power_nameplate_kw": 150,
    "chemistry": "NMC",
    "range_miles": 259
  },
  "source": "vector_db",
  "confidence": 0.92
}
```

## 🛠️ Development

### Project Structure
```
ev-oracle/
├── cmd/
│   └── root.go              # Cobra CLI setup
├── internal/
│   ├── db/
│   │   └── postgres.go      # Neon connection & queries
│   ├── embedding/
│   │   └── openai.go        # Embedding generation
│   ├── llm/
│   │   └── claude.go        # LLM fallback logic
│   └── models/
│       └── vehicle.go       # Data structures
├── go.mod
├── go.sum
└── README.md
```

### Run Tests
```bash
go test ./...
```

### Build
```bash
go build -o ev-oracle
```

## 🌟 Roadmap

- [ ] Hybrid search (vector + keyword filters)
- [ ] Batch lookup from CSV
- [ ] Web API mode
- [ ] Support for plug-in hybrid vehicles
- [ ] Historical battery degradation data
- [ ] Export to multiple formats (CSV, XML)

## 🤝 Contributing

Contributions welcome! Please open an issue or PR.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details

## 🔗 Related Projects

- [bidirectional.energy](https://bidirectional.energy) - Vehicle-to-grid platform
- [pgvector](https://github.com/pgvector/pgvector) - Vector similarity search for Postgres
- [Neon](https://neon.tech) - Serverless Postgres

---

Built with ⚡ by [@scaryPonens](https://github.com/scaryPonens)
