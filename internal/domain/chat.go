package domain

import "time"

type ChatStatus string

const (
	ChatStatusNewLead        ChatStatus = "new_lead"
	ChatStatusNegotiation    ChatStatus = "negotiation"
	ChatStatusWaitingReply   ChatStatus = "waiting_reply"
	ChatStatusInterested     ChatStatus = "interested"
	ChatStatusUnhappy        ChatStatus = "unhappy"
	ChatStatusLost           ChatStatus = "lost"
	ChatStatusActionRequired ChatStatus = "action_required"
)

type StatusSource string

const (
	StatusSourceAuto   StatusSource = "auto"
	StatusSourceManual StatusSource = "manual"
)

type ChatPriority string

const (
	ChatPriorityLow    ChatPriority = "low"
	ChatPriorityNormal ChatPriority = "normal"
	ChatPriorityHigh   ChatPriority = "high"
)

// Chat — личный диалог аккаунта с контактом; компания определяется через AccountID.
type Chat struct {
	ID                 int64        `json:"id" db:"id"`
	AccountID          int64        `json:"account_id" db:"account_id"`
	ContactID          int64        `json:"contact_id" db:"contact_id"`
	Status             ChatStatus   `json:"status" db:"status"`
	StatusSource       StatusSource `json:"status_source" db:"status_source"`
	Priority           ChatPriority `json:"priority" db:"priority"`
	LastMessageAt      time.Time    `json:"last_message_at" db:"last_message_at"`
	AwaitingReplySince *time.Time   `json:"awaiting_reply_since" db:"awaiting_reply_since"`
	UnreadCount        int          `json:"unread_count" db:"unread_count"`
}

// ChatFilter задаёт параметры фильтрации списка чатов компании.
type ChatFilter struct {
	CompanyID  int64       `json:"company_id"`
	EmployeeID *int64      `json:"employee_id,omitempty"`
	Status     *ChatStatus `json:"status,omitempty"`
}
