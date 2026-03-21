# MyFinances - Architecture Decision Record

## Context

CLI tool for personal finance management. Supports multiple databases for isolated contexts, credit cards with invoice tracking, categories, tags, and math expression parsing for amounts.

## Architecture

Clean Architecture with DDD tactical patterns. Complete isolation of business logic in the domain layer. Ports and Adapters pattern for infrastructure flexibility.

```
myfin/
├── cmd/                    # Entry point (Cobra CLI)
├── internal/
│   ├── domain/             # Pure business logic
│   │   └── entity/         # Expense, Income, Account, CreditCard, Category
│   ├── core/               # Use cases and ports
│   │   ├── usecase/       # Application services
│   │   └── port/           # Repository interfaces
│   └── infrastructure/     # External adapters
│       ├── persistence/     # SQLite implementations
│       │   └── sqlite/      # SQLite repositories
│       ├── database/        # Database manager
│       ├── cli/             # Output formatters
│       └── config/           # User config
├── pkg/
│   └── expr/               # Math expression parser
└── migrations/              # golang-migrate SQL files
```

## Data Model

### SQLite Schema (per database)

```sql
accounts (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
)

credit_cards (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  closing_day INTEGER NOT NULL CHECK(1-31),
  due_day INTEGER NOT NULL CHECK(1-31)
)

categories (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  alias TEXT UNIQUE,
  emoji TEXT,
  type TEXT NOT NULL CHECK('inc', 'exp')
)

entries (
  id INTEGER PRIMARY KEY,
  type TEXT NOT NULL CHECK('income', 'expense'),
  amount REAL NOT NULL,
  currency TEXT NOT NULL,
  description TEXT,
  category_id INTEGER REFERENCES categories(id),
  credit_card_id INTEGER REFERENCES credit_cards(id),
  realization_date TEXT NOT NULL,  -- YYYY-MM-DD, from --date flag
  payment_date TEXT,               -- YYYY-MM-DD, computed for CC
  created_at TEXT NOT NULL
)

entry_tags (
  entry_id INTEGER REFERENCES entries(id),
  tag TEXT NOT NULL,
  PRIMARY KEY (entry_id, tag)
)
```

## Database Management

### Storage Locations

- **Default location**: `~/.myfin/databases/`
- **Custom location**: Configurable via `databases.path` in `~/.myfin.yaml`

### Config File (`~/.myfin.yaml`)

```yaml
databases.path: ~/Dropbox/myfin/
default.currency: BRL
```

### Usage

```bash
# Use default database
myfin add expense 50.00 --category food

# Use specific database
myfin --db work add expense 100.00 --category lunch
```

## Domain Rules

### Credit Card Invoice Logic

Given a credit card with `closing_day` and `due_day`:

- If `realization_date.day <= closing_day`: first installment payment is `due_day` of the **same month**
- If `realization_date.day > closing_day`: first installment payment is `due_day` of the **next month**

Subsequent installments follow monthly intervals.

### Balance Calculation

For account balance purposes, the system uses `payment_date` (not `realization_date`) to determine when expenses affect cash flow.

## CLI Commands

```
myfin add account <name>
myfin add category --type inc|exp --name <name> [--alias <alias>] [--emoji <emoji>]
myfin add credit-card <name> --closing-day <n> --due-day <n>
myfin add expense [amount] [--tags x,y] [--date DD-MM-YY] [--category x] [--description x] [--credit-card x] [--times n]
myfin add income [amount] [--namespace x] [--date x] [--category x] [--description x]
myfin report --format md [--from x] [--until x] [--filter-tags x] [--filter-categories x]
myfin report balances
myfin config default.currency <value>
myfin remove record <id>
```

## Decisions

1. **IDs**: Sequential integers
2. **Categories**: Created on-the-fly (no predefined list), with optional alias and emoji
3. **Tags**: Free-form string labels
4. **Math expressions**: Support +, -, *, / with standard operator precedence
5. **Invoices**: Not stored; computed on-the-fly from realization_date, closing_day, and due_day
6. **Output**: Markdown for reports, plain text for confirmations
7. **Databases**: Multiple SQLite databases for isolation, selected via `--db` flag or config
