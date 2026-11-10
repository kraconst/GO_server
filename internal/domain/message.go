package domain

import "time"

type MessageDirection string

const (
	MessageDirectionIn  MessageDirection = "in"
	MessageDirectionOut MessageDirection = "out"
)

// MediaType описывает вложение без хранения самого файла.
// ТЗ задаёт открытый список типов; здесь перечислены явно указанные в нём.
type MediaType string

const (
	MediaTypePhoto    MediaType = "photo"
	MediaTypeDocument MediaType = "document"
	MediaTypeVoice    MediaType = "voice"
)

// Message — сообщение диалога. Пара ChatID и TelegramMessageID должна быть уникальна.
type Message struct {
	ID                int64            `json:"id" db:"id"`
	ChatID            int64            `json:"chat_id" db:"chat_id"`
	TelegramMessageID int64            `json:"telegram_message_id" db:"telegram_message_id"`
	Direction         MessageDirection `json:"direction" db:"direction"`
	Text              string           `json:"text" db:"text"`
	MediaType         *MediaType       `json:"media_type" db:"media_type"`
	SentAt            time.Time        `json:"sent_at" db:"sent_at"`
}
