DROP INDEX IF EXISTS idx_entries_account;
DROP INDEX IF EXISTS idx_entries_parent;
ALTER TABLE entries DROP COLUMN parent_entry_id;
ALTER TABLE entries DROP COLUMN installment;
ALTER TABLE entries DROP COLUMN account_id;
