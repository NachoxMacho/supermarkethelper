CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    box_price FLOAT NOT NULL,
    items_per_box INTEGER NOT NULL,
    shelves_in_store INTEGER NOT NULL,
    items_per_shelf INTEGER NOT NULL
);
