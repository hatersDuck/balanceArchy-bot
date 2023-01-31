CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event VARCHAR(1024) NOT NULL,
    fir VARCHAR(1024) DEFAULT 'Неизвестно',
    sec VARCHAR(1024) DEFAULT 'Неизвестно',
    user_id BIGINT,
    username VARCHAR(64));