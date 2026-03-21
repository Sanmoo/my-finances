-- Drop indexes
DROP INDEX IF EXISTS idx_entry_tags_tag;
DROP INDEX IF EXISTS idx_entry_tags_entry;
DROP INDEX IF EXISTS idx_entries_credit_card;
DROP INDEX IF EXISTS idx_entries_category;
DROP INDEX IF EXISTS idx_entries_payment_date;
DROP INDEX IF EXISTS idx_entries_realization_date;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS entry_tags;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS credit_cards;
DROP TABLE IF EXISTS accounts;
