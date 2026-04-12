# AGENTS.md

## Build & Run

```bash
go build -o myfin ./cmd/
```

## Test

```bash
go test ./...
```

No linter, formatter, or CI is configured.

## Architecture

Clean Architecture / Ports & Adapters. Dependency direction: `cmd` → `infrastructure` → `core` → `domain`.

```
cmd/                        # Cobra CLI entry point (all in main.go)
internal/domain/entity/     # Pure business logic, no external deps
internal/core/port/         # Repository interfaces
internal/core/usecase/      # Application services
internal/infrastructure/    # Adapters (persistence, config, CLI, i18n)
pkg/expr/                   # Public math expression parser
```

## Key Context

**Storage is YAML, not SQLite.** All repositories write YAML files under the data path.

Data files: `<data.path>/<account>/<year>/<year>-<month>-<account>-entries.yaml`

Config: `~/.myfin.yaml` with `data.path`, `default.currency`, `locale` (defaults: `~/.myfin/data`, `BRL`, `pt-BR`).

## Business Rules

- Tags must be registered before use (`myfin add tag <name>`)
- `--times` is required when `--credit-card` is specified
- Categories are scoped per account (`--account` required on `add category`)
- Credit card payment date: if realization day ≤ closing_day → due same month; otherwise → due next month
- Amount values accept math expressions via `pkg/expr` (e.g. `1000/3`, `5000+1000`)

## Conventions

- Domain entities use functional options pattern (e.g. `WithDescription`, `WithCategoryAlias`)
- Tests use `stretchr/testify`
- CLI date flags accept flexible formats: `DD`, `MM-DD`, `YY-MM-DD`, or `YYYY-MM-DD`