CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event VARCHAR(1024) NOT NULL,
    fir VARCHAR(1024) DEFAULT 'Неизвестен',
    sec VARCHAR(1024) DEFAULT 'Неизвестен',
    user_id BIGINT,
    username VARCHAR(64));