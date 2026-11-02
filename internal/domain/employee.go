package domain

// Employee — сотрудник компании; деактивация сохраняет историю его работы.
type Employee struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	IsActive  bool   `json:"is_active"`
}
