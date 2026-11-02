package domain

import "time"

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// ChatSummary — актуальная AI-выжимка, одна на ChatID.
// LastMessageID ссылается на внутренний Message.ID, до которого выполнен анализ.
type ChatSummary struct {
	ChatID        int64     `json:"chat_id"`
	Topic         string    `json:"topic"`
	Agreements    string    `json:"agreements"`
	NextStep      string    `json:"next_step"`
	OpenQuestions string    `json:"open_questions"`
	Risk          RiskLevel `json:"risk"`
	LastMessageID int64     `json:"last_message_id"`
	GeneratedAt   time.Time `json:"generated_at"`
}
