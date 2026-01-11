-- cards: The Source of Truth (Scryfall Data)
CREATE TABLE IF NOT EXISTS cards (
    scryfall_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    set_code TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    image_uri TEXT
    -- Simple index for autocomplete (LIKE queries)
    -- FTS5 could be an option later, but simple index is fine for now as per PRD
);

CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);

-- inventory: The User's Collection
CREATE TABLE IF NOT EXISTS inventory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scryfall_id TEXT NOT NULL,
    quantity INTEGER DEFAULT 1,
    condition TEXT DEFAULT 'NM',
    is_foil BOOLEAN DEFAULT 0,
    language TEXT DEFAULT 'en',
    location TEXT DEFAULT 'Binder',
    FOREIGN KEY(scryfall_id) REFERENCES cards(scryfall_id)
);

-- jobs: Async Task Tracker
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,               -- 'SYNC_DB', 'CSV_IMPORT'
    status TEXT NOT NULL,             -- 'PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'
    progress_current INTEGER DEFAULT 0,
    progress_total INTEGER DEFAULT 0,
    result_summary TEXT,              -- JSON
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- review_queue: Optimistic Import Buffer
CREATE TABLE IF NOT EXISTS review_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL,
    issue_type TEXT NOT NULL,         -- 'AMBIGUOUS', 'NOT_FOUND'
    raw_data TEXT,                    -- JSON of the imported row
    proposed_values TEXT,             -- JSON of parseable fields
    FOREIGN KEY(job_id) REFERENCES jobs(id)
);

-- system_settings: Key-Value Config
CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
