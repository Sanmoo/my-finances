DROP INDEX idx_categories_account_id;
DROP INDEX idx_categories_account_name;
ALTER TABLE categories DROP COLUMN account_id;
