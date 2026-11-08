package usecase

import (
	"context"

	"github.com/kraconst/GO_server/internal/domain"
)

// AccountRepository определяет контракт доступа к данным аккаунтов.
type AccountRepository interface {
	Create(ctx context.Context, acc domain.Account) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Account, error)
	ListByCompany(ctx context.Context, companyID int64) ([]domain.Account, error)
	ListByEmployee(ctx context.Context, employeeID int64) ([]domain.Account, error)
	UpdateStatus(ctx context.Context, id int64, status domain.AccountStatus) error
	AssignEmployee(ctx context.Context, accountID, employeeID int64) error
}

// EmployeeRepository определяет контракт доступа к данным сотрудников.
type EmployeeRepository interface {
	Create(ctx context.Context, emp domain.Employee) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Employee, error)
	ListByCompany(ctx context.Context, companyID int64) ([]domain.Employee, error)
	Deactivate(ctx context.Context, id int64) error
}

// CompanyRepository определяет контракт доступа к данным компании.
type CompanyRepository interface {
	Create(ctx context.Context, comp domain.Company) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Company, error)
}
