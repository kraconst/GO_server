package domain

// Employee — сотрудник компании; деактивация сохраняет историю его работы.
type Employee struct {
	ID        int64  `json:"id" db:"id"`
	CompanyID int64  `json:"company_id" db:"company_id"`
	Name      string `json:"name" db:"name"`
	Position  string `json:"position" db:"position"`
	IsActive  bool   `json:"is_active" db:"is_active"`
}
