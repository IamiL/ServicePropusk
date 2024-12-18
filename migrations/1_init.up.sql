CREATE TABLE IF NOT EXISTS buildings (
                                         id uuid PRIMARY KEY,
                                         name TEXT NULL,
                                         description TEXT NULL,
                                         status BOOLEAN NOT NULL,
                                         img_url TEXT NULL
);

CREATE TABLE IF NOT EXISTS passes (
                                      id uuid PRIMARY KEY,
                                      status INTEGER NOT NULL,
                                      created_at TIMESTAMPTZ NOT NULL,
                                      creator uuid NOT NULL,
                                      formed_at TIMESTAMPTZ NULL,
                                      completed_at TIMESTAMPTZ NULL,
                                      moderator INTEGER NULL,
                                      visitor TEXT NULL,
                                      visit_date TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS buildings_passes (
                                                id uuid PRIMARY KEY,
                                                building uuid NOT NULL,
                                                pass uuid NOT NULL,
                                                comment TEXT NULL,
                                                was_used boolean NULL,
                                                passage_time TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS users (
                                     id uuid PRIMARY KEY,
                                     login TEXT,
                                     pass_hash TEXT
)