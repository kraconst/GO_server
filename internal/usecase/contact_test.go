package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

type mockContactRepo struct {
	createFn          func(ctx context.Context, c domain.Contact) (int64, error)
	getByIDFn         func(ctx context.Context, id int64) (*domain.Contact, error)
	getByTelegramIDFn func(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error)
	listByCompanyFn   func(ctx context.Context, companyID int64) ([]domain.Contact, error)
}

func (m *mockContactRepo) Create(ctx context.Context, c domain.Contact) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	return 0, nil
}

func (m *mockContactRepo) GetByID(ctx context.Context, id int64) (*domain.Contact, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockContactRepo) GetByTelegramID(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error) {
	if m.getByTelegramIDFn != nil {
		return m.getByTelegramIDFn(ctx, companyID, tgUserID)
	}
	return nil, domain.ErrNotFound
}

func (m *mockContactRepo) ListByCompany(ctx context.Context, companyID int64) ([]domain.Contact, error) {
	if m.listByCompanyFn != nil {
		return m.listByCompanyFn(ctx, companyID)
	}
	return []domain.Contact{}, nil
}

func TestContactUsecase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid company_id", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		_, err := uc.Create(ctx, domain.Contact{CompanyID: 0, TelegramUserID: 123, Name: "Test"})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid telegram_user_id", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		_, err := uc.Create(ctx, domain.Contact{CompanyID: 1, TelegramUserID: 0, Name: "Test"})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		_, err := uc.Create(ctx, domain.Contact{CompanyID: 1, TelegramUserID: 123, Name: "   "})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("successful creation", func(t *testing.T) {
		repo := &mockContactRepo{
			createFn: func(ctx context.Context, c domain.Contact) (int64, error) {
				if c.Name != "John Doe" {
					t.Errorf("expected trimmed name 'John Doe', got %s", c.Name)
				}
				if c.Username != "johndoe" {
					t.Errorf("expected trimmed username 'johndoe', got %s", c.Username)
				}
				return 42, nil
			},
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return &domain.Contact{ID: id, CompanyID: 1, TelegramUserID: 123, Name: "John Doe", Username: "johndoe"}, nil
			},
		}

		uc := usecase.NewContactUsecase(repo)
		c, err := uc.Create(ctx, domain.Contact{CompanyID: 1, TelegramUserID: 123, Name: "  John Doe  ", Username: "  johndoe  "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 42 || c.Name != "John Doe" {
			t.Fatalf("unexpected contact: %+v", c)
		}
	})

	t.Run("fallback when GetByID fails after create", func(t *testing.T) {
		repo := &mockContactRepo{
			createFn: func(ctx context.Context, c domain.Contact) (int64, error) {
				return 42, nil
			},
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return nil, errors.New("db disconnect")
			},
		}

		uc := usecase.NewContactUsecase(repo)
		c, err := uc.Create(ctx, domain.Contact{CompanyID: 1, TelegramUserID: 123, Name: "John Doe"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 42 {
			t.Fatalf("expected ID 42, got %d", c.ID)
		}
	})
}

func TestContactUsecase_FindOrCreateByTelegramID(t *testing.T) {
	ctx := context.Background()

	t.Run("validation error on invalid args", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		if _, err := uc.FindOrCreateByTelegramID(ctx, 0, 10, "Name", ""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
		if _, err := uc.FindOrCreateByTelegramID(ctx, 1, 0, "Name", ""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
		if _, err := uc.FindOrCreateByTelegramID(ctx, 1, 10, "  ", ""); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("existing contact found", func(t *testing.T) {
		repo := &mockContactRepo{
			getByTelegramIDFn: func(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error) {
				return &domain.Contact{ID: 10, CompanyID: companyID, TelegramUserID: tgUserID, Name: "Existing"}, nil
			},
		}

		uc := usecase.NewContactUsecase(repo)
		c, err := uc.FindOrCreateByTelegramID(ctx, 1, 12345, "Existing", "user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 10 {
			t.Fatalf("expected ID 10, got %d", c.ID)
		}
	})

	t.Run("not found creates new contact", func(t *testing.T) {
		created := false
		repo := &mockContactRepo{
			getByTelegramIDFn: func(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error) {
				return nil, domain.ErrNotFound
			},
			createFn: func(ctx context.Context, c domain.Contact) (int64, error) {
				created = true
				return 99, nil
			},
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return &domain.Contact{ID: id, CompanyID: 1, TelegramUserID: 12345, Name: "New Contact"}, nil
			},
		}

		uc := usecase.NewContactUsecase(repo)
		c, err := uc.FindOrCreateByTelegramID(ctx, 1, 12345, "New Contact", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created || c.ID != 99 {
			t.Fatalf("expected creation of contact 99, got %+v", c)
		}
	})

	t.Run("db error propagated", func(t *testing.T) {
		repo := &mockContactRepo{
			getByTelegramIDFn: func(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error) {
				return nil, errors.New("db timeout")
			},
		}

		uc := usecase.NewContactUsecase(repo)
		_, err := uc.FindOrCreateByTelegramID(ctx, 1, 12345, "New Contact", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestContactUsecase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		if _, err := uc.GetByID(ctx, 0); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockContactRepo{
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return nil, domain.ErrNotFound
			},
		}
		uc := usecase.NewContactUsecase(repo)
		_, err := uc.GetByID(ctx, 123)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &mockContactRepo{
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return &domain.Contact{ID: id, Name: "Alice"}, nil
			},
		}
		uc := usecase.NewContactUsecase(repo)
		c, err := uc.GetByID(ctx, 5)
		if err != nil || c.Name != "Alice" {
			t.Fatalf("unexpected result: %v, %+v", err, c)
		}
	})
}

func TestContactUsecase_ListByCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid company_id", func(t *testing.T) {
		uc := usecase.NewContactUsecase(&mockContactRepo{})
		if _, err := uc.ListByCompany(ctx, 0); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &mockContactRepo{
			listByCompanyFn: func(ctx context.Context, companyID int64) ([]domain.Contact, error) {
				return []domain.Contact{{ID: 1, Name: "Alice"}}, nil
			},
		}
		uc := usecase.NewContactUsecase(repo)
		list, err := uc.ListByCompany(ctx, 1)
		if err != nil || len(list) != 1 {
			t.Fatalf("unexpected result: %v, len=%d", err, len(list))
		}
	})
}
