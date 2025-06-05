CREATE TABLE IF NOT EXISTS subscriptions (
     id SERIAL PRIMARY KEY,
     email TEXT NOT NULL,
     city TEXT NOT NULL,
     frequency TEXT NOT NULL,
     token TEXT NOT NULL UNIQUE,
     is_confirmed BOOLEAN NOT NULL DEFAULT FALSE
);