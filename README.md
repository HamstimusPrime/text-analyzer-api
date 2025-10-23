# Text Analyzer API

A REST API service built with Go that analyzes and stores text strings with comprehensive analytical properties. The API automatically calculates various text metrics and provides powerful filtering capabilities.

## Features

### Text Analysis

- **Palindrome Detection**: Automatically detects if a string is a palindrome
- **Character Frequency**: Tracks the count of each unique character in the text
- **Word Count**: Counts the number of words in the text
- **Length Calculation**: Measures character length of the text
- **Hash Generation**: Creates SHA256 hash for each text entry
- **Timestamp Tracking**: Records creation time for all entries

### API Capabilities

- **CRUD Operations**: Create, read, update, and delete text entries
- **Advanced Filtering**: Filter texts by multiple criteria including:
  - Palindrome status (`is_palindrome`)
  - Minimum/Maximum length (`min_length`, `max_length`)
  - Word count (`word_count`)
  - Character presence (`contains_character`)
- **Natural Language Queries**: Query texts using natural language descriptions
- **Unique String Management**: Prevents duplicate entries using SHA256 hashing

## Tech Stack

- **Language**: Go 1.24.3
- **Database**: PostgreSQL
- **Query Builder**: SQLC for type-safe SQL queries
- **Migration**: Goose (SQL migrations)
- **Dependencies**:
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/joho/godotenv` - Environment variable management
  - `github.com/google/uuid` - UUID generation

## API Endpoints

### Create Text

```http
POST /strings
Content-Type: application/json

{
  "value": "racecar"
}
```

### Get Single Text

```http
GET /strings/{string_value}
```

### Get Filtered Texts

```http
GET /strings?is_palindrome=true&min_length=5&max_length=20
```

### Natural Language Query

```http
GET /strings/filter-by-natural-language?query=palindromes with more than 5 characters
```

### Delete Text

```http
DELETE /strings/{string_value}
```

## Database Schema

### Texts Table

```sql
CREATE TABLE texts(
    id UUID PRIMARY KEY,
    value TEXT NOT NULL,
    length INT NOT NULL,
    is_palindrome BOOLEAN NOT NULL,
    word_count INT NOT NULL,
    sha256_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Character Count Table

```sql
CREATE TABLE character_count(
    id UUID PRIMARY KEY,
    string_id UUID NOT NULL,
    character TEXT NOT NULL,
    unique_char_count INTEGER NOT NULL,
    FOREIGN KEY(string_id) REFERENCES texts(id) ON DELETE CASCADE,
    UNIQUE(string_id, character)
);
```

## Setup and Installation

### Prerequisites

- Go 1.24.3 or higher
- PostgreSQL database
- Environment variables configured

### Environment Variables

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost/text_analyzer_db?sslmode=disable
PORT=8080
```

### Installation Steps

1. **Clone the repository**

   ```bash
   git clone https://github.com/HamstimusPrime/text-analyzer-api.git
   cd text-analyzer-api
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up the database**

   ```bash
   # Run migrations (assuming goose is installed)
   goose -dir sql/schema postgres $DB_URL up
   ```

4. **Build the application**

   ```bash
   go build .
   ```

5. **Run the server**
   ```bash
   ./text-analyzer-api
   ```

The server will start on the port specified in your `.env` file (default: 8080).

## Usage Examples

### Analyzing a Palindrome

```bash
curl -X POST http://localhost:8080/strings \
  -H "Content-Type: application/json" \
  -d '{"value": "racecar"}'
```

### Filtering Palindromes

```bash
curl "http://localhost:8080/strings?is_palindrome=true"
```

### Finding Texts with Specific Character Count

```bash
curl "http://localhost:8080/strings?min_length=10&contains_character=a"
```

### Natural Language Query

```bash
curl "http://localhost:8080/strings/filter-by-natural-language?query=find palindromes longer than 5 characters"
```

## Response Format

### Single Text Response

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "value": "racecar",
  "properties": {
    "length": 7,
    "is_palindrome": "true",
    "word_count": "1",
    "sha256_hash": "abc123...",
    "character_frequency_map": {
      "r": 2,
      "a": 2,
      "c": 2,
      "e": 1
    }
  },
  "created_at": "2025-10-23T10:30:00Z"
}
```

## Project Structure

```
.
├── main.go                 # Application entry point and server setup
├── handlers.go            # HTTP request handlers
├── models.go              # Data structures and types
├── utils.go               # Utility functions (palindrome check, hashing, etc.)
├── go.mod                 # Go module dependencies
├── sqlc.yaml             # SQLC configuration
├── internal/
│   └── database/         # Generated database queries and models
├── sql/
│   ├── queries/          # SQL query definitions
│   │   └── texts.sql
│   └── schema/           # Database migration files
│       ├── 001_texts.sql
│       ├── 002_character_count.sql
│       └── 003_fix_character_unique.sql
└── README.md
```

## Development

### Generate Database Code

This project uses SQLC to generate type-safe Go code from SQL queries:

```bash
sqlc generate
```

### Adding New Migrations

Create new migration files in the `sql/schema/` directory following the naming convention:

```
{number}_{description}.sql
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/new-feature`)
5. Create a Pull Request

## License

This project is open source and available under the [MIT License](LICENSE).

## API Documentation

For detailed API documentation and examples, refer to the individual endpoint handlers in `handlers.go`. Each endpoint includes proper error handling and validation to ensure data integrity and user-friendly responses.
