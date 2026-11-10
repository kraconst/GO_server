package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kraconst/GO_server/internal/domain"
)

type MessageUsecase struct {
	messages MessageRepository
}

func NewMessageUsecase(messages MessageRepository) *MessageUsecase {
	return &MessageUsecase{
		messages: messages,
	}
}

func (u *MessageUsecase) Create(ctx context.Context, msg domain.Message) (int64, error) {
	if msg.ChatID <= 0 {
		return 0, fmt.Errorf("%w: invalid chat_id", domain.ErrValidation)
	}
	if msg.TelegramMessageID <= 0 {
		return 0, fmt.Errorf("%w: invalid telegram_message_id", domain.ErrValidation)
	}
	if msg.Direction != domain.MessageDirectionIn && msg.Direction != domain.MessageDirectionOut {
		return 0, fmt.Errorf("%w: direction must be 'in' or 'out'", domain.ErrValidation)
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}

	id, err := u.messages.Create(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("usecase create message: %w", err)
	}
	return id, nil
}

func (u *MessageUsecase) ListByChat(ctx context.Context, chatID int64, limit, offset int) ([]domain.Message, error) {
	if chatID <= 0 {
		return nil, fmt.Errorf("%w: invalid chat_id", domain.ErrValidation)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return u.messages.ListByChat(ctx, chatID, limit, offset)
}
