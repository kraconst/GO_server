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
	ID                int64            `json:"id"`
	ChatID            int64            `json:"chat_id"`
	TelegramMessageID int64            `json:"telegram_message_id"`
	Direction         MessageDirection `json:"direction"`
	Text              string           `json:"text"`
	MediaType         *MediaType       `json:"media_type"`
	SentAt            time.Time        `json:"sent_at"`
}
