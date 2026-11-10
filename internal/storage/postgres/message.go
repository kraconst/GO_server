package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
)

type MessageStorage struct {
	db *sqlx.DB
}

func NewMessageStorage(db *sqlx.DB) *MessageStorage {
	return &MessageStorage{db: db}
}

func (s *MessageStorage) Create(ctx context.Context, msg domain.Message) (int64, error) {
	query := `
		INSERT INTO messages (chat_id, telegram_message_id, direction, text, media_type, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (chat_id, telegram_message_id) DO NOTHING
		RETURNING id
	`
	var id int64
	err := s.db.QueryRowContext(ctx, query,
		msg.ChatID,
		msg.TelegramMessageID,
		msg.Direction,
		msg.Text,
		msg.MediaType,
		msg.SentAt,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("insert message: %w", err)
	}
	return id, nil
}

func (s *MessageStorage) ListByChat(ctx context.Context, chatID int64, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT id, chat_id, telegram_message_id, direction, text, media_type, sent_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY sent_at ASC
		LIMIT $2 OFFSET $3
	`
	messages := make([]domain.Message, 0)
	err := s.db.SelectContext(ctx, &messages, query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages by chat: %w", err)
	}
	return messages, nil
}
