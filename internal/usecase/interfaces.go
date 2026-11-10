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
	Update(ctx context.Context, emp domain.Employee) error
	Deactivate(ctx context.Context, id int64) error
}

// CompanyRepository определяет контракт доступа к данным компании.
type CompanyRepository interface {
	Create(ctx context.Context, comp domain.Company) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Company, error)
}

// MessageRepository определяет контракт доступа к сообщениям диалогов.
type MessageRepository interface {
	Create(ctx context.Context, msg domain.Message) (int64, error)
	ListByChat(ctx context.Context, chatID int64, limit, offset int) ([]domain.Message, error)
}

// ContactRepository определяет контракт доступа к данным контактов.
type ContactRepository interface {
	Create(ctx context.Context, c domain.Contact) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Contact, error)
	GetByTelegramID(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error)
	ListByCompany(ctx context.Context, companyID int64) ([]domain.Contact, error)
}

// ChatRepository определяет контракт доступа к данным чатов.
type ChatRepository interface {
	Create(ctx context.Context, chat domain.Chat) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Chat, error)
	ListByCompany(ctx context.Context, companyID int64, status *domain.ChatStatus, employeeID *int64) ([]domain.Chat, error)
	UpdateStatus(ctx context.Context, id int64, status domain.ChatStatus, source domain.StatusSource) error
}
