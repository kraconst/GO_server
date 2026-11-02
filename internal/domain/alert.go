package domain

import "time"

type AlertType string

const (
	AlertTypeNoReply   AlertType = "no_reply"
	AlertTypeComplaint AlertType = "complaint"
	AlertTypeRefusal   AlertType = "refusal"
	AlertTypeNegative  AlertType = "negative"
	AlertTypeUrgent    AlertType = "urgent"
)

// Alert — уведомление о проблеме в диалоге; nil в ResolvedAt означает, что оно не обработано.
type Alert struct {
	ID         int64      `json:"id"`
	CompanyID  int64      `json:"company_id"`
	ChatID     int64      `json:"chat_id"`
	Type       AlertType  `json:"type"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}
