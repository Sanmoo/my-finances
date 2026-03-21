-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Create credit_cards table
CREATE TABLE IF NOT EXISTS credit_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    closing_day INTEGER NOT NULL CHECK(closing_day >= 1 AND closing_day <= 31),
    due_day INTEGER NOT NULL CHECK(due_day >= 1 AND due_day <= 31)
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    alias TEXT UNIQUE,
    emoji TEXT,
    type TEXT NOT NULL CHECK(type IN ('inc', 'exp'))
);

-- Create entries table
CREATE TABLE IF NOT EXISTS entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    amount REAL NOT NULL CHECK(amount > 0),
    currency TEXT NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES categories(id),
    credit_card_id INTEGER REFERENCES credit_cards(id),
    realization_date TEXT NOT NULL,
    payment_date TEXT,
    created_at TEXT NOT NULL
);

-- Create entry_tags table
CREATE TABLE IF NOT EXISTS entry_tags (
    entry_id INTEGER NOT NULL REFERENCES entries(id),
    tag TEXT NOT NULL,
    PRIMARY KEY (entry_id, tag)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_entries_realization_date ON entries(realization_date);
CREATE INDEX IF NOT EXISTS idx_entries_payment_date ON entries(payment_date);
CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category_id);
CREATE INDEX IF NOT EXISTS idx_entries_credit_card ON entries(credit_card_id);
CREATE INDEX IF NOT EXISTS idx_entry_tags_entry ON entry_tags(entry_id);
CREATE INDEX IF NOT EXISTS idx_entry_tags_tag ON entry_tags(tag);
