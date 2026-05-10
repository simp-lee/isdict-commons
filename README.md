# isdict-commons

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Shared Go commons layer for the isdict pipeline and downstream services. This module owns the current 21-table PostgreSQL schema, shared GORM models, controlled enums, deterministic normalization helpers, and the PostgreSQL migration entrypoint used across repos.

## Installation

```bash
go get github.com/simp-lee/isdict-commons@latest
```

Requirements: Go 1.25 language features with a patched Go toolchain at Go 1.26.2 or newer. The module pins this via go.mod so local and CI runs do not build against the vulnerable Go 1.26.0 standard library. PostgreSQL is required for schema migration and opt-in integration tests.

## Package Layout

```text
isdict-commons/
├── model/      # 21-table GORM schema and shared controlled values
├── norm/       # Deterministic normalization helpers and frozen alias maps
└── migration/  # PostgreSQL schema migration and verification entrypoint
```

## Current Schema

The model package maps to these 21 tables:

| Table | Model | Purpose |
| --- | --- | --- |
| `import_runs` | `ImportRun` | Import and enrichment provenance |
| `entries` | `Entry` | Top-level dictionary entries |
| `senses` | `Sense` | Ordered senses per entry |
| `sense_glosses_en` | `SenseGlossEN` | Ordered English glosses |
| `sense_glosses_zh` | `SenseGlossZH` | Ordered zh-Hans glosses with provenance |
| `sense_labels` | `SenseLabel` | Controlled sense labels |
| `sense_examples` | `SenseExample` | Ordered example sentences |
| `entry_definitions` | `EntryDefinition` | Structured entry/sense-optional zh-Hans definitions |
| `entry_examples` | `EntryExample` | Canonical entry-level examples with optional sense alignment |
| `pronunciation_ipas` | `PronunciationIPA` | IPA pronunciations |
| `pronunciation_audios` | `PronunciationAudio` | Audio filename records |
| `entry_forms` | `EntryForm` | Forms and aliases |
| `headword_relation_edges` | `HeadwordRelationEdge` | OEWN headword/POS lexical relation edges |
| `entry_summaries_zh` | `EntrySummaryZH` | Chinese summaries |
| `entry_learning_signals` | `EntryLearningSignal` | Entry-level learner annotations |
| `entry_cefr_source_signals` | `EntryCEFRSourceSignal` | Entry-level CEFR source evidence |
| `sense_learning_signals` | `SenseLearningSignal` | Sense-level learner annotations |
| `sense_cefr_source_signals` | `SenseCEFRSourceSignal` | Sense-level CEFR source evidence |
| `entry_etymologies` | `EntryEtymology` | Etymology text |
| `entry_search_terms` | `EntrySearchTerm` | Derived search, suggestion, and phrase-search read model |
| `featured_candidates` | `FeaturedCandidate` | Derived featured recommendation candidate read model |

The schema is PostgreSQL-first: primary keys are `int64`, `ImportRun` carries provenance, `entries.pos` and pronunciation accent codes are stored as text codes, and embedded SQL adds partial indexes, expression indexes, GIN trigram indexes, and identity-column behavior that GORM does not express by itself.

### Derived Read Models

`entry_search_terms` is the derived read model for search, suggestion, and phrase search. It materializes one `headword` row per `entries` row and one `form` or `alias` row per `entry_forms` row, using the same normalized values already stored on `entries.normalized_headword` and `entry_forms.normalized_form`. Entry-level learning signals are denormalized onto every search term; missing learning signals are stored as `0`.

Search consumers should use `entry_search_terms.normalized_term`, not trigram-search `entries.normalized_headword` or `entry_forms.normalized_form` directly. The source-table btree indexes remain for exact lookup, but their old trigram indexes are dropped. `entry_search_terms` carries the trigram GIN index, prefix btree index, `entry_id`, `pos`, `is_multiword + normalized_term`, and positive learning-signal partial indexes.

`featured_candidates` is the derived read model for featured recommendation candidates. It materializes entries that have `entry_learning_signals.frequency_rank > 0 OR entry_learning_signals.cefr_level > 0 OR entry_learning_signals.school_level > 0`, carries the fields needed for ranking and word/phrase splitting, and avoids downstream runtime joins between `entries` and `entry_learning_signals`.

Both read models are maintained by migration/import refresh code; they are not business source data. `RunMigration` refreshes them once after schema/index migration, and full importers should call `migration.RefreshReadModels(db)` after `entries`, `entry_forms`, and `entry_learning_signals` are loaded.

### Lexical Relations

`headword_relation_edges` is the product relation table. Each row is an OEWN-owned evidence edge from a source normalized headword plus POS code to a target normalized headword plus POS code, with required `import_run_id` provenance. Product relation types cover the OEWN 2025 synset and sense relation payloads: `synonym`, `antonym`, `hypernym`, `hyponym`, `meronym`, `holonym`, `similar_to`, `also_see`, `derivation`, `pertainym`, `domain_topic`, `domain_region`, `exemplifies`, `attribute`, `entails`, `causes`, `event`, `agent`, `result`, `by_means_of`, `undergoer`, `instrument`, `uses`, `state`, `property`, `location`, `material`, `vehicle`, `participle`, `body_part`, and `destination`.

`source_relation_type` stores the exact OEWN JSON field that produced the edge. It is limited to the real 2025 relation fields used by synsets or entry senses: `members`, `antonym`, `derivation`, `pertainym`, `hypernym`, `mero_part`, `mero_member`, `mero_substance`, `similar`, `also`, `domain_topic`, `domain_region`, `exemplifies`, `attribute`, `entails`, `causes`, `event`, `agent`, `result`, `by_means_of`, `undergoer`, `instrument`, `uses`, `state`, `property`, `location`, `material`, `vehicle`, `participle`, `body_part`, and `destination`. OEWN metadata fields such as definitions, examples, ILI, Wikidata IDs, verb frames, subcategories, and adjective positions are not relation edges.

OEWN raw POS values are collapsed to four product POS codes for relation lookup: `n -> Noun`, `v -> Verb`, `a/s -> Adjective`, and `r -> Adverb`. Numbered entry POS keys such as `n-1` and `v-2` use the same base-code mapping.

Open English WordNet 2025 Edition is the only lexical relation source. Wiktionary/Kaikki still supplies entries, senses, glosses, examples, pronunciations, forms, aliases, phrase entries, and etymology text, but Wiktionary/Kaikki relation data is not imported or displayed. Wiktionary `derived` data is intentionally outside the lexical relation module.

### Learning And CEFR Signals

`entry_learning_signals.cefr_level` and `sense_learning_signals.cefr_level` store the final aggregated CEFR level used by downstream queries. CEFR levels use `0..6`: `unknown=0`, `A1=1`, `A2=2`, `B1=3`, `B2=4`, `C1=5`, `C2=6`. The aggregate `cefr_source` is either unset (`''`) or the real adopted source: `oxford`, `cefrj`, or `octanove`.

`entry_cefr_source_signals` and `sense_cefr_source_signals` store the per-source raw CEFR evidence before aggregation. Their composite primary keys are `(entry_id, cefr_source)` and `(sense_id, cefr_source)`, `cefr_source` only allows `oxford`, `cefrj`, and `octanove`, and `cefr_level` uses the same `0..6` scale. `cefr_run_id` points at `import_runs.id` for provenance.

`OxfordLevel` / `oxford_level` is Oxford 3000/5000 list membership (`0..2`), not an Oxford CEFR A1-C2 level. Oxford CEFR evidence belongs in the source evidence tables with `cefr_source='oxford'`.

`SchoolLevel` / `school_level` remains the aggregated learning stage (`0..3`) used by products and read models. It does not represent a specific textbook source, textbook appearance count, or learner-facing source label. `entry_learning_signals.school_run_id` stores internal import-run provenance for idempotent school imports and validation only. School vocabulary sources can contain word-only rows without definitions; those rows belong in learning signals, not in `entry_definitions`.

### Entry Content

`entry_definitions` stores structured dictionary definitions for entry-level content and optional sense-aligned content. It supports source families such as `wiktionary`, `ecdict`, and `school` without limiting the allowed source names. `sense_id` is nullable because school or ECDICT definitions should only be attached to a Wiktionary sense when the importer has high-confidence alignment. Chinese definition text is required, and `normalized_zh_hans_key` is importer-provided for duplicate prevention within the same entry and POS.

`entry_examples` is the canonical entry-level bilingual examples table. It supports Wiktionary and school examples, nullable `sense_id`, required English sentence text, optional zh-Hans translation, and importer-provided normalized sentence keys for duplicate prevention under each entry.

For both tables, `source` and `source_run_id` are internal provenance used for replay, replacement, and validation. They do not mean downstream products must display source labels to learners.

## Shared Enums And Normalization

`model` exports the shared controlled values used by importer, API, and web code:

- POS: 23 text codes via `POS*`, `POSCodeToName`, `POSNameToCode`, and `ValidPOSCodes`
- Accent: 11 text codes via `Accent*`, `AccentCodeToName`, and `AccentNameToCode`
- Controlled sense labels: 85 label codes across `grammar`, `register`, `region`, `temporal`, `domain`, `attitude`, and `variety`, exposed through `LabelType*`, `LabelCodeToNameByType`, `LabelNameToCodeByType`, and `ValidLabelCodesByType`
- Relation and provenance enums: `RelationType*`, `HeadwordRelationPOS*`, `OEWNPartOfSpeechCode*`, `OEWNSenseType*`, `OEWNSourceRelation*`, `RelationKind*`, `ImportRunStatus*`, and `CEFRSource*`
- Learning scales: `CEFRLevel*`, `OxfordLevel*`, `CETLevel*`, `SchoolLevel*`, and `CollinsStars*`, plus code/name maps

`norm` exports the deterministic helpers and frozen lookup maps:

- `NormalizeHeadword` and `IsMultiword`
- `CanonicalizePOS` and `RawPOSToLearnerPOS`
- `NormalizeLabelAlias`, `NormalizeLabelText`, and `LabelAliasMap` for conservative Wiktionary qualifier/raw_gloss label mapping, including phrase labels and comma/semicolon combinations
- `NormalizeTopic` and `TopicToDomainMap` for controlled domain mapping with case-insensitive space/hyphen/underscore variants
- `NormalizeAccentCode` and `AccentTagMap`
- `NewZhNormalizer` and `(*ZhNormalizer).Normalize`
- `AudioLocalFilename` and `NormalizeAudioAccentCode`, including high-confidence Wiktionary audio filename prefix detection for common English accent codes

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

`RunMigration` runs `AutoMigrate` for all 21 tables, executes the embedded SQL files under `migration/sql/`, refreshes `entry_search_terms` and `featured_candidates`, runs `ANALYZE` as best-effort, and verifies that required tables, `pg_trgm`, indexes, and the managed `id` identity columns/sequences are present and aligned.

After a full importer load, refresh the derived read models explicitly:

```go
if err := migration.RefreshReadModels(db); err != nil {
    return err
}
```

Recommended downstream query direction:

- Search/suggestion/fuzzy phrase search: query `entry_search_terms`, filter on `normalized_term`, `term_kind`, `pos`, `is_multiword`, and the denormalized learning-signal fields, then order by `term_rank`, `frequency_rank`, and stable tie-breakers.
- Featured words/phrases: query `featured_candidates`, filter by `is_multiword`, and order by `quality_rank`, then optional learning-signal tie-breakers such as `cefr_level DESC`, `collins_stars DESC`, `school_level ASC`, and `entry_id`.

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

The real-data PostgreSQL test is a separate opt-in path for an existing imported database. It runs `RunMigration` with `DropTables=false`, verifies source-table row counts stay unchanged, refreshes only the derived read models, and checks that `entry_search_terms` and `featured_candidates` exactly match the source joins:

```bash
ISDICT_REALDATA_POSTGRES_DSN='host=localhost port=5432 user=isdict password=... dbname=isdict_db sslmode=disable TimeZone=Asia/Shanghai' \
ISDICT_REALDATA_POSTGRES_ALLOW_REFRESH='refresh-derived-read-models' \
go test ./migration -run TestRunMigration_PostgresRealData -count=1 -timeout=30m
```

## v1.0.0 Breaking Changes

This repository now reflects the v1.0.0 schema reset.

- Removed the legacy `Word`, legacy `Sense`, `Example`, `Pronunciation`, and `WordVariant` model set. Use the current 21-table schema instead.
- Removed API response structs from commons. Response DTOs belong in downstream repos.
- Removed `textutil.ToNormalized`. Use `norm.NormalizeHeadword` and the other `norm` helpers.
- Removed `migration.NewMigrator` and `Migrator.Migrate`. Use `migration.RunMigration(db, migration.MigrateOptions{...})`.
- POS and accent storage moved from integer codes to text codes. Update persisted data, enum handling, and raw SQL accordingly.

## v1.0.7 Schema Changes

- Added `entry_search_terms` as the canonical search/suggestion/phrase-search read model.
- Added `featured_candidates` as the canonical featured recommendation candidate read model.
- Added `migration.RefreshReadModels(db)` for deterministic `TRUNCATE + INSERT SELECT` refreshes.
- Dropped source-table trigram indexes `idx_entries_normalized_headword_trgm` and `idx_entry_forms_normalized_form_trgm`; exact lookup btree indexes remain.
