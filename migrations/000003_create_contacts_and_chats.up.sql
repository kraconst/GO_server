-- Contacts
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL DEFAULT '',
    UNIQUE (company_id, telegram_user_id)
);

CREATE INDEX IF NOT EXISTS idx_contacts_company_id ON contacts(company_id);

-- Chats
CREATE TABLE IF NOT EXISTS chats (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    contact_id BIGINT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'new_lead',
    status_source VARCHAR(16) NOT NULL DEFAULT 'auto' CHECK (status_source IN ('auto', 'manual')),
    priority VARCHAR(16) NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high')),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    awaiting_reply_since TIMESTAMPTZ,
    unread_count INT NOT NULL DEFAULT 0,
    UNIQUE (account_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_chats_account_id ON chats(account_id);
CREATE INDEX IF NOT EXISTS idx_chats_contact_id ON chats(contact_id);
CREATE INDEX IF NOT EXISTS idx_chats_status ON chats(status);
CREATE INDEX IF NOT EXISTS idx_chats_last_message_at ON chats(last_message_at DESC);
