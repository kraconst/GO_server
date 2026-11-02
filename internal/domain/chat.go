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
	ID                 int64        `json:"id"`
	AccountID          int64        `json:"account_id"`
	ContactID          int64        `json:"contact_id"`
	Status             ChatStatus   `json:"status"`
	StatusSource       StatusSource `json:"status_source"`
	Priority           ChatPriority `json:"priority"`
	LastMessageAt      time.Time    `json:"last_message_at"`
	AwaitingReplySince *time.Time   `json:"awaiting_reply_since"`
	UnreadCount        int          `json:"unread_count"`
}
