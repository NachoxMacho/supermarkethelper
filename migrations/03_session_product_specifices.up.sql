CREATE TABLE IF NOT EXISTS session_product_specifics (
    product_id INTEGER NOT NULL,
    session_id TEXT NOT NULL,
    box_price REAL NOT NULL,
    shelves_in_store INTEGER NOT NULL
);
