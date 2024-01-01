DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS transactions;

CREATE TABLE accounts (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    amount REAL NOT NULL,
    currency TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL,
    amount REAL NOT NULL,
    is_expense INTEGER NOT NULL,
    category_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    description TEXT NOT NULL,
    FOREIGN KEY (category_id)
        REFERENCES categories (id),
    FOREIGN KEY (account_id)
        REFERENCES accounts (id)
);

INSERT INTO accounts (name, amount, currency, description)
VALUES
    ('Visa Card RSD', 15000.55, 'RSD', ''),
    ('Visa Card EUR', 410.75, 'EUR' , ''),
    ('Cash', 20000.23,  'RSD' ,'');

INSERT INTO categories (name)
VALUES
    ('🍖 - Food and Drinks'),
    ('🍱 - Restaurants'),
    ('🏠 - Rent and Housing'),
    ('🎮 - Entertainment'),
    ('🧭 - Traveling'),
    ('🧾 - Bills and Taxes'),
    ('💆 - Beauty and Health');

INSERT INTO transactions (account_id, amount, is_expense, category_id, created_at, description)
VALUES
    (1, 2000.189,  1, 1   , '2023-12-31 16:00:00+01:00' ,'' ),
    (1, 50000.245, 0, NULL, '2024-01-01 15:26:45+01:00','Salary');

