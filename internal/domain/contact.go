package domain

// Contact — собеседник; TelegramUserID уникален в пределах CompanyID.
type Contact struct {
	ID             int64  `json:"id"`
	CompanyID      int64  `json:"company_id"`
	TelegramUserID int64  `json:"telegram_user_id"`
	Name           string `json:"name"`
	Username       string `json:"username"`
}
