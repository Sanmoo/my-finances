# MyFinances - Architecture Decision Record

## Context

CLI tool for personal finance management. Supports multiple namespaces (isolated contexts), credit cards with invoice tracking, categories, tags, and math expression parsing for amounts.

## Architecture

Clean Architecture with DDD tactical patterns. Complete isolation of business logic in the domain layer. Ports and Adapters pattern for infrastructure flexibility.

```
myfin/
├── cmd/                    # Entry point (Cobra CLI)
├── internal/
│   ├── domain/             # Pure business logic
│   │   └── entity/         # Expense, Income, Account, CreditCard, Category, Namespace
│   ├── core/               # Use cases and ports
│   │   ├── usecase/        # Application services
│   │   └── port/           # Repository interfaces
│   ├── infrastructure/     # External adapters
│   │   ├── persistence/    # SQLite implementations
│   │   ├── cli/            # Output formatters
│   │   └── config/         # User config
│   └── pkg/expr/           # Math expression parser
└── myfin.db
```

## Data Model

### SQLite Schema

```sql
namespaces (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  default_credit_card_id INTEGER REFERENCES credit_cards(id),
  default_currency TEXT DEFAULT 'BRL'
)

accounts (
  id INTEGER PRIMARY KEY,
  namespace_id INTEGER REFERENCES namespaces(id),
  name TEXT NOT NULL
)

credit_cards (
  id INTEGER PRIMARY KEY,
  namespace_id INTEGER REFERENCES namespaces(id),
  name TEXT NOT NULL,
  closing_day INTEGER NOT NULL,
  due_day INTEGER NOT NULL
)

categories (
  id INTEGER PRIMARY KEY,
  namespace_id INTEGER REFERENCES namespaces(id),
  name TEXT NOT NULL,
  alias TEXT,
  emoji TEXT,
  type TEXT NOT NULL CHECK(type IN ('inc', 'exp'))
)

entries (
  id INTEGER PRIMARY KEY,
  namespace_id INTEGER REFERENCES namespaces(id),
  type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
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
myfin report --format md [--namespace x] [--from x] [--until x] [--filter-tags x] [--filter-categories x]
myfin report balances [--namespace x]
myfin config default.namespace|default.currency|default.credit-card <value>
myfin remove record <id> [--namespace x]
```

## Decisions

1. **IDs**: Sequential integers
2. **Categories**: Created on-the-fly (no predefined list), with optional alias and emoji
3. **Tags**: Free-form string labels
4. **Math expressions**: Support +, -, *, / with standard operator precedence
5. **Invoices**: Not stored; computed on-the-fly from realization_date, closing_day, and due_day
6. **Output**: Markdown for reports, plain text for confirmations
