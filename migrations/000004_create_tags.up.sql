CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_tags_name ON tags(name);

CREATE TABLE IF NOT EXISTS entry_tag_ids (
    entry_id INTEGER NOT NULL REFERENCES entries(id),
    tag_id INTEGER NOT NULL REFERENCES tags(id),
    PRIMARY KEY (entry_id, tag_id)
);

CREATE INDEX idx_entry_tag_ids_entry ON entry_tag_ids(entry_id);
CREATE INDEX idx_entry_tag_ids_tag ON entry_tag_ids(tag_id);
