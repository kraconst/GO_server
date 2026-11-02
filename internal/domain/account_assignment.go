package domain

import "time"

// AccountAssignment — период ответственности сотрудника за аккаунт.
// To == nil обозначает текущее назначение. Компания определяется через AccountID.
type AccountAssignment struct {
	AccountID  int64      `json:"account_id"`
	EmployeeID int64      `json:"employee_id"`
	From       time.Time  `json:"from"`
	To         *time.Time `json:"to"`
}
