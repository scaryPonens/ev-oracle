# EV Oracle Usage Examples

This file contains example commands and use cases for the EV Oracle CLI.

## Setup

1. Set up your environment variables:
```bash
export NEON_DATABASE_URL="postgresql://user:password@host/database?sslmode=require"
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
```

2. Initialize the database schema:
```bash
ev-oracle init
```

## Adding EV Specifications

Add some known EV specifications to the database:

```bash
# Tesla Model 3 (2023)
ev-oracle add Tesla "Model 3" 2023 --capacity 75.0 --power 283.0 --chemistry "NMC (Nickel Manganese Cobalt)"

# Nissan Leaf (2022)
ev-oracle add Nissan Leaf 2022 --capacity 40.0 --power 110.0 --chemistry "Li-ion"

# Chevrolet Bolt EV (2023)
ev-oracle add Chevrolet "Bolt EV" 2023 --capacity 65.0 --power 150.0 --chemistry "NMC"

# Ford Mustang Mach-E (2023)
ev-oracle add Ford "Mustang Mach-E" 2023 --capacity 91.0 --power 258.0 --chemistry "NMC"

# Volkswagen ID.4 (2023)
ev-oracle add Volkswagen "ID.4" 2023 --capacity 82.0 --power 150.0 --chemistry "NMC"
```

## Querying EV Specifications

### Exact Match Query (Plain Text Output)
```bash
ev-oracle Tesla "Model 3" 2023
```

Expected output:
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

### Exact Match Query (JSON Output)
```bash
ev-oracle --json Nissan Leaf 2022
```

Expected output:
```json
{
  "make": "Nissan",
  "model": "Leaf",
  "year": 2022,
  "capacity_kwh": 40.0,
  "power_kw": 110.0,
  "chemistry": "Li-ion",
  "confidence": 1.0,
  "source": "database"
}
```

### Similarity Search
Query for a similar model that might not be an exact match:

```bash
ev-oracle Tesla "Model Y" 2023
```

This will use vector similarity search to find the closest match in the database.

### LLM Fallback
Query for a vehicle not in the database (confidence < 0.8):

```bash
ev-oracle Rivian R1T 2023
```

This will fall back to Claude API to generate the specifications.

## Piping JSON Output

You can pipe the JSON output to other tools:

```bash
# Pretty print with jq
ev-oracle --json Tesla "Model 3" 2023 | jq '.'

# Extract specific fields
ev-oracle --json Tesla "Model 3" 2023 | jq '.capacity_kwh'

# Filter by confidence
ev-oracle --json Tesla "Model 3" 2023 | jq 'select(.confidence > 0.9)'
```

## Integration with Scripts

```bash
#!/bin/bash
# Example script to query multiple EVs

MAKES=("Tesla" "Nissan" "Chevrolet")
MODELS=("Model 3" "Leaf" "Bolt EV")
YEARS=(2023 2022 2023)

for i in "${!MAKES[@]}"; do
    echo "Querying ${MAKES[$i]} ${MODELS[$i]} ${YEARS[$i]}..."
    ev-oracle --json "${MAKES[$i]}" "${MODELS[$i]}" "${YEARS[$i]}"
    echo ""
done
```

## Troubleshooting

### Database Connection Issues
If you see database connection errors:
```bash
# Test your database connection string
psql "$NEON_DATABASE_URL" -c "SELECT 1"

# Ensure pgvector extension is available
psql "$NEON_DATABASE_URL" -c "CREATE EXTENSION IF NOT EXISTS vector"
```

### API Key Issues
Verify your API keys are set correctly:
```bash
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY
```

### Re-initialize Database
If you need to reset the database schema:
```bash
# Drop the table (WARNING: This deletes all data)
psql "$NEON_DATABASE_URL" -c "DROP TABLE IF EXISTS ev_specs"

# Re-initialize
ev-oracle init
```
