CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    items_per_box INTEGER NOT NULL,
    items_per_shelf INTEGER NOT NULL,
    default_box_price REAL NOT NULL,
    default_shelves_in_store INTEGER NOT NULL
);

