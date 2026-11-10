package domain

import "time"

// AccountDailyStats — метрики аккаунта за календарный день в часовом поясе компании.
// Ключ — пара AccountID и Day. Day представляет дату (без времени суток).
// AvgResponseTime задаётся как time.Duration, при JSON-кодировании — в наносекундах.
type AccountDailyStats struct {
	AccountID       int64         `json:"account_id" db:"account_id"`
	Day             time.Time     `json:"day" db:"day"`
	MessagesIn      int           `json:"messages_in" db:"messages_in"`
	MessagesOut     int           `json:"messages_out" db:"messages_out"`
	ActiveChats     int           `json:"active_chats" db:"active_chats"`
	NewContacts     int           `json:"new_contacts" db:"new_contacts"`
	AvgResponseTime time.Duration `json:"avg_response_time" db:"avg_response_time"`
}

// DashboardStats — сводные агрегированные метрики компании за выбранный период.
type DashboardStats struct {
	MessagesIn      int           `json:"messages_in" db:"messages_in"`
	MessagesOut     int           `json:"messages_out" db:"messages_out"`
	ActiveChats     int           `json:"active_chats" db:"active_chats"`
	NewContacts     int           `json:"new_contacts" db:"new_contacts"`
	AvgResponseTime time.Duration `json:"avg_response_time" db:"avg_response_time"`
	UnansweredChats int           `json:"unanswered_chats" db:"unanswered_chats"`
}

// AccountRating — рейтинг активности аккаунта за выбранный период.
type AccountRating struct {
	AccountID       int64         `json:"account_id" db:"account_id"`
	Username        string        `json:"username" db:"username"`
	EmployeeName    string        `json:"employee_name" db:"employee_name"`
	MessagesIn      int           `json:"messages_in" db:"messages_in"`
	MessagesOut     int           `json:"messages_out" db:"messages_out"`
	TotalMessages   int           `json:"total_messages" db:"total_messages"`
	ActiveChats     int           `json:"active_chats" db:"active_chats"`
	NewContacts     int           `json:"new_contacts" db:"new_contacts"`
	AvgResponseTime time.Duration `json:"avg_response_time" db:"avg_response_time"`
}
