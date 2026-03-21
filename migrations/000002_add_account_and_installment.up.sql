ALTER TABLE entries ADD COLUMN account_id INTEGER NOT NULL REFERENCES accounts(id);
ALTER TABLE entries ADD COLUMN installment INTEGER DEFAULT 0;
ALTER TABLE entries ADD COLUMN parent_entry_id INTEGER REFERENCES entries(id);

CREATE INDEX idx_entries_account ON entries(account_id);
CREATE INDEX idx_entries_parent ON entries(parent_entry_id);
