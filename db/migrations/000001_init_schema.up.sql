CREATE TABLE users(
    id                BIGSERIAL primary key,
    email             varchar(50) unique not null,
    password_hash     TEXT NOT NULL,
    is_active         boolean NOT NULL,
    role              TEXT NOT NULL,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);

CREATE TABLE email_tokens (
    id              TEXT PRIMARY KEY,
    token_hash      TEXT NOT NULL,
    user_id         TEXT NOT NULL,
    email           TEXT NOT NULL,
    purpose         TEXT NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    consumed_at     TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);