package domain

import "time"

// AccountDailyStats — метрики аккаунта за календарный день в часовом поясе компании.
// Ключ — пара AccountID и Day. Day представляет дату (без времени суток).
// AvgResponseTime задаётся как time.Duration, при JSON-кодировании — в наносекундах.
type AccountDailyStats struct {
	AccountID       int64         `json:"account_id"`
	Day             time.Time     `json:"day"`
	MessagesIn      int           `json:"messages_in"`
	MessagesOut     int           `json:"messages_out"`
	ActiveChats     int           `json:"active_chats"`
	NewContacts     int           `json:"new_contacts"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}
