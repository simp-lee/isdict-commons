# isdict-commons

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Shared data models and utilities for English-Chinese dictionary applications.

## Features

- Database models for words, pronunciations, senses, examples, and variants
- API response structures
- Text normalization utilities
- Standardized enumerations (POS, accents, CEFR levels, etc.)
- Multi-source support (CEFR, Oxford, CET, word frequency, Collins)

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
errorResponse := model.NewErrorResponse("Word not found")
```

## Package Structure

```
isdict-commons/
├── model/          # Database and response models, enumerations
└── textutil/       # Text normalization utilities
```

## Core Models

### Word

```go
type Word struct {
    ID                 uint
    Headword           string  // Original case-preserved form
    HeadwordNormalized string  // Lowercase for lookups
    CEFRLevel          int
    CEFRSource         string
    CETLevel           int
    OxfordLevel        int
    FrequencyRank      int
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
    FrequencyRank      int
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



## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- [isdict-api](https://github.com/simp-lee/isdict-api) - RESTful API service
