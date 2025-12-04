CREATE TABLE IF NOT EXISTS "cards_master" (
    "id" TEXT PRIMARY KEY,
    "name" TEXT,
    -- "mana_cost" TEXT,
    -- "cmc" NUMERIC,
    -- "type_line" TEXT,
    -- "oracle_text" TEXT,
    "set" TEXT,
    "set_name" TEXT,
    "collector_number" INT,
    "rarity" TEXT
);

CREATE TABLE IF NOT EXISTS "cards_owned" (
    "id" TEXT PRIMARY KEY,
    "quantity" INT,
    "location" TEXT,
    FOREIGN KEY ("id") REFERENCES "cards_master"("id")
);

-- CREATE TABLE IF NOT EXISTS image_uris (
--     "id" TEXT PRIMARY KEY,
--     "small" TEXT,
--     "normal" TEXT,
--     "large" TEXT,
--     FOREIGN KEY ("id") REFERENCES "cards_master"("id")
-- );

CREATE TABLE IF NOT EXISTS "scryfall_default_cards_log" (
    "default_cards_id" TEXT PRIMARY KEY
);
