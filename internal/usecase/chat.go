package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kraconst/GO_server/internal/domain"
)

type ChatUsecase struct {
	chats ChatRepository
}

func NewChatUsecase(chats ChatRepository) *ChatUsecase {
	return &ChatUsecase{
		chats: chats,
	}
}

func isValidChatStatus(status domain.ChatStatus) bool {
	switch status {
	case domain.ChatStatusNewLead,
		domain.ChatStatusNegotiation,
		domain.ChatStatusWaitingReply,
		domain.ChatStatusInterested,
		domain.ChatStatusUnhappy,
		domain.ChatStatusLost,
		domain.ChatStatusActionRequired:
		return true
	default:
		return false
	}
}

func isValidStatusSource(source domain.StatusSource) bool {
	return source == domain.StatusSourceAuto || source == domain.StatusSourceManual
}

func isValidChatPriority(priority domain.ChatPriority) bool {
	return priority == domain.ChatPriorityLow ||
		priority == domain.ChatPriorityNormal ||
		priority == domain.ChatPriorityHigh
}

func (u *ChatUsecase) SetManualStatus(ctx context.Context, chatID int64, status domain.ChatStatus) error {
	if chatID <= 0 {
		return fmt.Errorf("%w: invalid chat_id", domain.ErrValidation)
	}
	if !isValidChatStatus(status) {
		return fmt.Errorf("%w: invalid chat status '%s'", domain.ErrValidation, status)
	}

	return u.chats.UpdateStatus(ctx, chatID, status, domain.StatusSourceManual)
}

func (u *ChatUsecase) ListByCompany(ctx context.Context, companyID int64, status *domain.ChatStatus, employeeID *int64) ([]domain.Chat, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: invalid company_id", domain.ErrValidation)
	}
	if status != nil && !isValidChatStatus(*status) {
		return nil, fmt.Errorf("%w: invalid chat status '%s'", domain.ErrValidation, *status)
	}
	if employeeID != nil && *employeeID <= 0 {
		return nil, fmt.Errorf("%w: invalid employee_id", domain.ErrValidation)
	}

	return u.chats.ListByCompany(ctx, companyID, status, employeeID)
}

func (u *ChatUsecase) GetByID(ctx context.Context, id int64) (*domain.Chat, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	return u.chats.GetByID(ctx, id)
}

func (u *ChatUsecase) Create(ctx context.Context, chat domain.Chat) (*domain.Chat, error) {
	if chat.AccountID <= 0 {
		return nil, fmt.Errorf("%w: invalid account_id", domain.ErrValidation)
	}
	if chat.ContactID <= 0 {
		return nil, fmt.Errorf("%w: invalid contact_id", domain.ErrValidation)
	}

	if chat.Status == "" {
		chat.Status = domain.ChatStatusNewLead
	} else if !isValidChatStatus(chat.Status) {
		return nil, fmt.Errorf("%w: invalid chat status '%s'", domain.ErrValidation, chat.Status)
	}

	if chat.StatusSource == "" {
		chat.StatusSource = domain.StatusSourceAuto
	} else if !isValidStatusSource(chat.StatusSource) {
		return nil, fmt.Errorf("%w: invalid status_source '%s'", domain.ErrValidation, chat.StatusSource)
	}

	if chat.Priority == "" {
		chat.Priority = domain.ChatPriorityNormal
	} else if !isValidChatPriority(chat.Priority) {
		return nil, fmt.Errorf("%w: invalid priority '%s'", domain.ErrValidation, chat.Priority)
	}

	if chat.LastMessageAt.IsZero() {
		chat.LastMessageAt = time.Now().UTC()
	}

	id, err := u.chats.Create(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("usecase create chat: %w", err)
	}

	created, err := u.chats.GetByID(ctx, id)
	if err != nil {
		chat.ID = id
		return &chat, nil
	}
	return created, nil
}
