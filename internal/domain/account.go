package domain

import "time"

type AccountStatus string

const (
	StatusActive       AccountStatus = "active"
	StatusDisconnected AccountStatus = "disconnected"
)

type Account struct {
	ID           int           `json:"id"`
	Phone        string        `json:"phone"`
	Username     string        `json:"username"`
	EmployeeName string        `json:"employee_name"`
	Status       AccountStatus `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
}
