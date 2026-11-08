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
	ID         int64  `json:"id" db:"id"`
	CompanyID  int64  `json:"company_id" db:"company_id"`
	EmployeeID int64  `json:"employee_id" db:"employee_id"`
	Phone      string `json:"phone" db:"phone"`
	Username   string `json:"username" db:"username"`
	// EmployeeName оставлено для совместимости с текущими моками API.
	// Ответственный в модели определяется через EmployeeID.
	EmployeeName   string        `json:"employee_name" db:"employee_name"`
	TelegramUserID int64         `json:"telegram_user_id" db:"telegram_user_id"`
	Status         AccountStatus `json:"status" db:"status"`
	// Session содержит зашифрованную сессию MTProto.
	Session        []byte     `json:"-" db:"session"`
	LastActivityAt *time.Time `json:"last_activity_at" db:"last_activity_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}
