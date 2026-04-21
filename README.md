# isdict-commons

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Shared Go commons layer for the isdict pipeline and downstream services. This module owns the current 15-table PostgreSQL schema, shared GORM models, controlled enums, deterministic normalization helpers, and the PostgreSQL migration entrypoint used across repos.

## Installation

```bash
go get github.com/simp-lee/isdict-commons@latest
```

Requirements: Go 1.25 language features with a patched Go toolchain at Go 1.26.2 or newer. The module pins this via go.mod so local and CI runs do not build against the vulnerable Go 1.26.0 standard library. PostgreSQL is required for schema migration and opt-in integration tests.

## Package Layout

```text
isdict-commons/
├── model/      # 15 GORM models and shared controlled values
├── norm/       # Deterministic normalization helpers and frozen alias maps
└── migration/  # PostgreSQL schema migration and verification entrypoint
```

## Current Schema

The model package maps to these 15 tables:

| Table | Model | Purpose |
| --- | --- | --- |
| `import_runs` | `ImportRun` | Import and enrichment provenance |
| `entries` | `Entry` | Top-level dictionary entries |
| `senses` | `Sense` | Ordered senses per entry |
| `sense_glosses_en` | `SenseGlossEN` | Ordered English glosses |
| `sense_glosses_zh` | `SenseGlossZH` | Ordered zh-Hans glosses with provenance |
| `sense_labels` | `SenseLabel` | Controlled sense labels |
| `sense_examples` | `SenseExample` | Ordered example sentences |
| `pronunciation_ipas` | `PronunciationIPA` | IPA pronunciations |
| `pronunciation_audios` | `PronunciationAudio` | Audio filename records |
| `entry_forms` | `EntryForm` | Forms and aliases |
| `lexical_relations` | `LexicalRelation` | Synonym, antonym, and derived links |
| `entry_summaries_zh` | `EntrySummaryZH` | Chinese summaries |
| `entry_learning_signals` | `EntryLearningSignal` | Entry-level learner annotations |
| `sense_learning_signals` | `SenseLearningSignal` | Sense-level learner annotations |
| `entry_etymologies` | `EntryEtymology` | Etymology text |

The schema is PostgreSQL-first: primary keys are `int64`, `ImportRun` carries provenance, `entries.pos` and pronunciation accent codes are stored as text codes, and embedded SQL adds partial indexes, expression indexes, a GIN trigram index, and identity-column behavior that GORM does not express by itself.

## Shared Enums And Normalization

`model` exports the shared controlled values used by importer, API, and web code:

- POS: 23 text codes via `POS*`, `POSCodeToName`, `POSNameToCode`, and `ValidPOSCodes`
- Accent: 11 text codes via `Accent*`, `AccentCodeToName`, and `AccentNameToCode`
- Controlled sense labels: 62 label codes across `grammar`, `register`, `region`, `temporal`, `domain`, and `attitude`, exposed through `LabelType*`, `LabelCodeToNameByType`, `LabelNameToCodeByType`, and `ValidLabelCodesByType`
- Relation and provenance enums: `RelationType*`, `RelationKind*`, `ImportRunStatus*`, and `CEFRSource*`
- Learning scales: `CEFRLevel*`, `OxfordLevel*`, `CETLevel*`, `SchoolLevel*`, and `CollinsStars*`, plus code/name maps

`norm` exports the deterministic helpers and frozen lookup maps:

- `NormalizeHeadword` and `IsMultiword`
- `CanonicalizePOS` and `RawPOSToLearnerPOS`
- `NormalizeLabelAlias` and `LabelAliasMap`
- `NormalizeTopic` and `TopicToDomainMap`
- `NormalizeAccentCode` and `AccentTagMap`
- `NewZhNormalizer` and `(*ZhNormalizer).Normalize`
- `AudioLocalFilename`

Commons intentionally keeps normalization deterministic and single-input. Importer-specific heuristics that need wider context stay outside this module.

## Migration Usage

Use the `migration` package for PostgreSQL schema creation and verification:

```go
package main

import (
    "github.com/simp-lee/isdict-commons/migration"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func run(dsn string) error {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    return migration.RunMigration(db, migration.MigrateOptions{
        DropTables: false,
        Verbose:    true,
    })
}
```

`RunMigration` runs `AutoMigrate` for all 15 models, executes the embedded SQL files under `migration/sql/`, runs `ANALYZE` as best-effort, and verifies that required tables, `pg_trgm`, indexes, and the managed `id` identity columns/sequences are present and aligned.

## Testing

Run the default test suite with:

```bash
go test ./...
```

The PostgreSQL integration test is opt-in and lives in the migration package:

```bash
ISDICT_TEST_POSTGRES_DSN='postgres:///isdict_test_db?sslmode=disable&search_path=public' \
ISDICT_TEST_POSTGRES_ALLOW_DESTRUCTIVE_RESET='drop-public-migration-tables' \
go test ./migration
```

The integration test skips when `-short` is enabled, when `ISDICT_TEST_POSTGRES_DSN` is unset, or when `ISDICT_TEST_POSTGRES_ALLOW_DESTRUCTIVE_RESET=drop-public-migration-tables` is not set. The default-safe path is a hostless local DSN with `search_path=public`; if your shell exports `PGSERVICE`, leave it unset, and if it exports `PGHOST`, leave it unset or point it at an approved local Unix socket directory such as `/var/run/postgresql`. If you intentionally need an explicit host or `service`/`servicefile` indirection for a disposable remote CI target, also set `ISDICT_TEST_POSTGRES_ALLOW_REMOTE_DSN=allow-remote-disposable-instance`; otherwise those connections are rejected by design. Its cleanup path is intentionally strict: it refuses to run unless the database name is one of `isdict_test`, `isdict_test_db`, `isdict_integration_test`, `isdict_integration_test_db`, `isdict_migration_test`, or `isdict_migration_test_db`, `current_schema()` is `public`, and the normalized `search_path` is exactly `public`. Point it at one of those dedicated disposable test databases only.

## v1.0.0 Breaking Changes

This repository now reflects the v1.0.0 schema reset.

- Removed the legacy `Word`, legacy `Sense`, `Example`, `Pronunciation`, and `WordVariant` model set. Use the current 15-table schema instead.
- Removed API response structs from commons. Response DTOs belong in downstream repos.
- Removed `textutil.ToNormalized`. Use `norm.NormalizeHeadword` and the other `norm` helpers.
- Removed `migration.NewMigrator` and `Migrator.Migrate`. Use `migration.RunMigration(db, migration.MigrateOptions{...})`.
- POS and accent storage moved from integer codes to text codes. Update persisted data, enum handling, and raw SQL accordingly.
