# isdict-commons

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test](https://github.com/simp-lee/isdict-commons/workflows/Test/badge.svg)](https://github.com/simp-lee/isdict-commons/actions)

Shared data models and utilities for English-Chinese dictionary applications.

## Features

- **Database Models**: Complete schema for words, pronunciations, senses, examples, and variants
- **API Response Structures**: Unified response format with error handling
- **Text Normalization**: Smart case-folding that preserves apostrophes and slashes
- **Standardized Enumerations**: POS (23 types), accents (10 types), CEFR levels, etc.
- **Multi-source Support**: CEFR, Oxford, CET, word frequency, Collins ratings
- **Database Migration**: Enterprise-grade PostgreSQL schema migration tools
- **Full Test Coverage**: 100% coverage for all testable packages

## Installation

```bash
go get github.com/simp-lee/isdict-commons@latest
```

**Requirements:** Go 1.24+

## Quick Start

```go
import (
    "github.com/simp-lee/isdict-commons/model"
    "github.com/simp-lee/isdict-commons/textutil"
)
```

### Create Word Entry

```go
word := model.Word{
    Headword:           "example",
    HeadwordNormalized: textutil.ToNormalized("example"),
    CEFRLevel:          3,  // B1 (1-6)
    CETLevel:           1,  // 1=CET-4, 2=CET-6
    TranslationZH:      "例子",
}

pronunciation := model.Pronunciation{
    WordID:    word.ID,
    Accent:    2,  // American
    IPA:       "/ɪɡˈzæmpəl/",
    IsPrimary: true,
}

sense := model.Sense{
    WordID:       word.ID,
    POS:          1,  // noun
    DefinitionEN: "a thing characteristic of its kind",
    DefinitionZH: "例子；典型",
}
```

### Text Normalization

```go
textutil.ToNormalized("Example")  // "example"
textutil.ToNormalized("I'm")      // "i'm" (preserves apostrophes)
textutil.ToNormalized("and/or")   // "and/or" (preserves slashes)
```

### API Responses

```go
response := model.WordResponse{
    ID:       1,
    Headword: "example",
    // ... fields
}

successResponse := model.NewSuccessResponse(response)
errorResponse := model.NewErrorResponse("NOT_FOUND", "Word not found", nil)
```

## Package Structure

```
isdict-commons/
├── model/          # Database and response models, enumerations
├── textutil/       # Text normalization utilities
└── migration/      # Database schema migration tools
```

## Core Models

### Word

```go
type Word struct {
    ID                 uint
    Headword           string  // Original case-preserved form
    HeadwordNormalized string  // Lowercase for lookups
    CEFRLevel          int
    CEFRSource         string  // CEFR data source
    CETLevel           int
    OxfordLevel        int
    SchoolLevel        int
    FrequencyRank      int
    FrequencyCount     int
    CollinsStars       int
    TranslationZH      string

    Pronunciations []Pronunciation
    Senses         []Sense
    WordVariants   []WordVariant
}
```

### Pronunciation

```go
type Pronunciation struct {
    ID        uint
    WordID    uint
    Accent    int    // 1=British, 2=American, etc.
    IPA       string
    IsPrimary bool
}
```

### Sense

```go
type Sense struct {
    ID           uint
    WordID       uint
    POS          int    // 1=noun, 2=verb, 3=adjective, etc.
    CEFRLevel    int
    CEFRSource   string // CEFR data source
    OxfordLevel  int    // Sense-level Oxford annotation
    DefinitionEN string
    DefinitionZH string
    SenseOrder   int

    Examples []Example
}
```

### Example

```go
type Example struct {
    ID           uint
    SenseID      uint
    SentenceEN   string
    SentenceZH   string
    ExampleOrder int
}
```

### WordVariant

```go
type WordVariant struct {
    ID                 uint
    WordID             uint
    VariantText        string
    HeadwordNormalized string
    Kind               VariantKind  // 1=form, 2=alias
    FormType           *int         // 1=past, 2=past_participle, 5=plural, etc.
    Tags               []string     // Additional tags (PostgreSQL array)
    FrequencyRank      int
    FrequencyCount     int
}
```

## Enumerations

### Part of Speech

| Code | Name | Code | Name |
|------|------|------|------|
| 1 | noun | 2 | verb |
| 3 | adjective | 4 | adverb |
| 5 | pronoun | 6 | preposition |
| 7 | conjunction | 8 | article |
| 9 | interjection | 10 | determiner |

See `PosCodeToName` for all 23 types (codes 0-22).

```go
model.GetPOSName(1)        // "noun"
model.ParsePOS("verb")     // 2, true
```

### Accents

| Code | Name | Code | Name |
|------|------|------|------|
| 1 | british | 2 | american |
| 3 | australian | 4 | newzealand |
| 5 | canadian | 6 | irish |
| 7 | scottish | 8 | indian |
| 9 | southafrican | 10 | other |

### Form Types

| Code | Name | Code | Name |
|------|------|------|------|
| 1 | past | 2 | past_participle |
| 3 | present_3rd | 4 | gerund |
| 5 | plural | 6 | comparative |
| 7 | superlative | 8 | possessive |
| 9 | infinitive | | |

### CEFR Levels

| Code | Level | Code | Level |
|------|-------|------|-------|
| 1 | A1 | 2 | A2 |
| 3 | B1 | 4 | B2 |
| 5 | C1 | 6 | C2 |

### Other Classifications

- **Oxford**: 0=none, 1=Oxford 3000, 2=Oxford 5000
- **CET**: 0=none, 1=CET-4, 2=CET-6
- **Collins**: 0-5 stars (5 = most frequent)

## Model Relationships

```
Word
├── Pronunciations
├── Senses
│   └── Examples
└── WordVariants
```

## Text Utilities

### ToNormalized

Converts to lowercase, preserves apostrophes and slashes, removes other punctuation.

```go
textutil.ToNormalized("Example")  // "example"
textutil.ToNormalized("I'm")      // "i'm"
textutil.ToNormalized("and/or")   // "and/or"
```

## Testing

```bash
go test ./...           # Run all tests
go test -cover ./...    # With coverage
```

## Database Migration

The `migration` package provides tools for PostgreSQL schema migration:

```go
import (
    "github.com/simp-lee/isdict-commons/migration"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Connect to database
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Create migrator
migrator := migration.NewMigrator(db)

// Run migration with options
err = migrator.Migrate(&migration.MigrateOptions{
    DropTables:     false,  // Don't drop existing tables
    SkipExtensions: false,  // Enable PostgreSQL extensions (pg_trgm)
    SkipIndexes:    false,  // Create all indexes
    Verbose:        true,   // Enable detailed logging
})
```

**Migration includes:**
- Table creation with constraints
- PostgreSQL `pg_trgm` extension for fuzzy search
- Custom unique constraints with NULL handling
- GIN trigram indexes for performance
- Migration verification and integrity checks

## API Response Structure

All API endpoints use a unified response format:

```go
// Success response
{
    "success": true,
    "data": { ... },
    "error": null,
    "meta": {
        "page": 1,
        "page_size": 20,
        "total": 100
    }
}

// Error response
{
    "success": false,
    "data": null,
    "error": {
        "code": "NOT_FOUND",
        "message": "Word not found",
        "details": null
    },
    "meta": null
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- [isdict-api](https://github.com/simp-lee/isdict-api) - RESTful API service
