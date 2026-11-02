package domain

import "time"

// DefaultNoReplyThreshold — порог ожидания ответа по умолчанию из ТЗ.
const DefaultNoReplyThreshold = 2 * time.Hour

// Company — компания-клиент; Timezone хранит имя зоны IANA, например Europe/Moscow.
// NoReplyThreshold задаётся как time.Duration, при JSON-кодировании — в наносекундах.
type Company struct {
	ID               int64         `json:"id"`
	Name             string        `json:"name"`
	Timezone         string        `json:"timezone"`
	NoReplyThreshold time.Duration `json:"no_reply_threshold"`
	CreatedAt        time.Time     `json:"created_at"`
}
