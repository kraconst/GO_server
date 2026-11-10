CREATE TABLE IF NOT EXISTS account_daily_stats (
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    messages_in INT NOT NULL DEFAULT 0,
    messages_out INT NOT NULL DEFAULT 0,
    active_chats INT NOT NULL DEFAULT 0,
    new_contacts INT NOT NULL DEFAULT 0,
    avg_response_time BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (account_id, day)
);

CREATE INDEX IF NOT EXISTS idx_account_daily_stats_day ON account_daily_stats(day);
