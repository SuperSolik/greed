CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    amount REAL NOT NULL,
    currency TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL,
    amount REAL NOT NULL,
    category_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    description TEXT NOT NULL,
    FOREIGN KEY (category_id)
        REFERENCES categories (id),
    FOREIGN KEY (account_id)
        REFERENCES accounts (id)
);

-- INSERT INTO accounts (name, amount, currency, description)
-- VALUES
--     ('Visa Card RSD', 15000.55, 'RSD', ''),
--     ('Visa Card EUR', 410.75, 'EUR', ''),
--     ('Cash', 20000.23, 'RSD', '');

INSERT INTO categories (name)
VALUES
    ('🗑️ Other'),
    ('💰 Finance'),
    ('🍖 Food and drinks'),
    ('🍱 Eating out'),
    ('🏠 Rent and Housing'),
    ('🎮 Fun'),
    ('🧭 Travel'),
    ('🧾 Bills and Taxes'),
    ('💆 Beauty and Health'),
    ('💱 Exchange');

-- INSERT INTO transactions (account_id, amount, category_id, created_at, description)
-- VALUES
--     (1, -2000.189, 6, '2023-12-31 16:00:00+01:00', ''),
--     (2, 50000.245, 2, '2024-01-01 15:26:45-07:00', 'Salary'),
--     (1, -123.123, 1, '2023-09-10 15:26:45+02:00', 'mcdonalds')
-- ;
--
