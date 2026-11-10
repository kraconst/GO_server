CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    telegram_message_id BIGINT NOT NULL,
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('in', 'out')),
    text TEXT NOT NULL DEFAULT '',
    media_type VARCHAR(50),
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chat_id, telegram_message_id)
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id_sent_at ON messages(chat_id, sent_at ASC);