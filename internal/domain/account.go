package domain

import "time"

type AccountStatus string

const (
	StatusConnecting   AccountStatus = "connecting"
	StatusActive       AccountStatus = "active"
	StatusDisconnected AccountStatus = "disconnected"
	StatusBanned       AccountStatus = "banned"
)

// Account — рабочий Telegram-аккаунт компании.
type Account struct {
	ID         int64  `json:"id"`
	CompanyID  int64  `json:"company_id"`
	EmployeeID int64  `json:"employee_id"`
	Phone      string `json:"phone"`
	Username   string `json:"username"`
	// EmployeeName оставлено для совместимости с текущими моками API.
	// Ответственный в модели определяется через EmployeeID.
	EmployeeName   string        `json:"employee_name"`
	TelegramUserID int64         `json:"telegram_user_id"`
	Status         AccountStatus `json:"status"`
	// Session содержит зашифрованную сессию MTProto.
	Session        []byte     `json:"-"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
