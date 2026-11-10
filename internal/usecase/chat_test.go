package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

type mockChatRepo struct {
	createFn        func(ctx context.Context, chat domain.Chat) (int64, error)
	getByIDFn       func(ctx context.Context, id int64) (*domain.Chat, error)
	listByCompanyFn func(ctx context.Context, companyID int64, status *domain.ChatStatus, employeeID *int64) ([]domain.Chat, error)
	updateStatusFn  func(ctx context.Context, id int64, status domain.ChatStatus, source domain.StatusSource) error
}

func (m *mockChatRepo) Create(ctx context.Context, chat domain.Chat) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, chat)
	}
	return 0, nil
}

func (m *mockChatRepo) GetByID(ctx context.Context, id int64) (*domain.Chat, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrNotFound
}

func (m *mockChatRepo) ListByCompany(ctx context.Context, companyID int64, status *domain.ChatStatus, employeeID *int64) ([]domain.Chat, error) {
	if m.listByCompanyFn != nil {
		return m.listByCompanyFn(ctx, companyID, status, employeeID)
	}
	return []domain.Chat{}, nil
}

func (m *mockChatRepo) UpdateStatus(ctx context.Context, id int64, status domain.ChatStatus, source domain.StatusSource) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, source)
	}
	return nil
}

func TestChatUsecase_SetManualStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid chat id", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		err := uc.SetManualStatus(ctx, 0, domain.ChatStatusNegotiation)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid chat status", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		err := uc.SetManualStatus(ctx, 1, domain.ChatStatus("unknown_status"))
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("not found in repository", func(t *testing.T) {
		repo := &mockChatRepo{
			updateStatusFn: func(ctx context.Context, id int64, status domain.ChatStatus, source domain.StatusSource) error {
				return domain.ErrNotFound
			},
		}
		uc := usecase.NewChatUsecase(repo)
		err := uc.SetManualStatus(ctx, 999, domain.ChatStatusInterested)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("successful manual status update with all valid statuses", func(t *testing.T) {
		validStatuses := []domain.ChatStatus{
			domain.ChatStatusNewLead,
			domain.ChatStatusNegotiation,
			domain.ChatStatusWaitingReply,
			domain.ChatStatusInterested,
			domain.ChatStatusUnhappy,
			domain.ChatStatusLost,
			domain.ChatStatusActionRequired,
		}

		for _, st := range validStatuses {
			called := false
			repo := &mockChatRepo{
				updateStatusFn: func(ctx context.Context, id int64, status domain.ChatStatus, source domain.StatusSource) error {
					called = true
					if id != 10 {
						t.Errorf("expected id 10, got %d", id)
					}
					if status != st {
						t.Errorf("expected status %s, got %s", st, status)
					}
					if source != domain.StatusSourceManual {
						t.Errorf("expected source %s, got %s", domain.StatusSourceManual, source)
					}
					return nil
				},
			}
			uc := usecase.NewChatUsecase(repo)
			err := uc.SetManualStatus(ctx, 10, st)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", st, err)
			}
			if !called {
				t.Fatalf("updateStatusFn not called for %s", st)
			}
		}
	})
}

func TestChatUsecase_ListByCompany(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid company_id", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		_, err := uc.ListByCompany(ctx, 0, nil, nil)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		st := domain.ChatStatus("bogus")
		_, err := uc.ListByCompany(ctx, 1, &st, nil)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid employee_id", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		empID := int64(-5)
		_, err := uc.ListByCompany(ctx, 1, nil, &empID)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("success with filters passed through", func(t *testing.T) {
		st := domain.ChatStatusInterested
		empID := int64(3)
		repo := &mockChatRepo{
			listByCompanyFn: func(ctx context.Context, companyID int64, status *domain.ChatStatus, employeeID *int64) ([]domain.Chat, error) {
				if companyID != 1 {
					t.Errorf("expected companyID 1, got %d", companyID)
				}
				if status == nil || *status != st {
					t.Errorf("expected status %v, got %v", st, status)
				}
				if employeeID == nil || *employeeID != empID {
					t.Errorf("expected employeeID %v, got %v", empID, employeeID)
				}
				return []domain.Chat{{ID: 101, Status: st}}, nil
			},
		}

		uc := usecase.NewChatUsecase(repo)
		chats, err := uc.ListByCompany(ctx, 1, &st, &empID)
		if err != nil || len(chats) != 1 {
			t.Fatalf("unexpected result: %v, len=%d", err, len(chats))
		}
	})
}

func TestChatUsecase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		_, err := uc.GetByID(ctx, 0)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockChatRepo{
			getByIDFn: func(ctx context.Context, id int64) (*domain.Chat, error) {
				return nil, domain.ErrNotFound
			},
		}
		uc := usecase.NewChatUsecase(repo)
		_, err := uc.GetByID(ctx, 42)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &mockChatRepo{
			getByIDFn: func(ctx context.Context, id int64) (*domain.Chat, error) {
				return &domain.Chat{ID: id, Status: domain.ChatStatusNewLead}, nil
			},
		}
		uc := usecase.NewChatUsecase(repo)
		chat, err := uc.GetByID(ctx, 42)
		if err != nil || chat.ID != 42 {
			t.Fatalf("unexpected result: %v, %+v", err, chat)
		}
	})
}

func TestChatUsecase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid account_id or contact_id", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		if _, err := uc.Create(ctx, domain.Chat{AccountID: 0, ContactID: 1}); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
		if _, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: 0}); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid enum values", func(t *testing.T) {
		uc := usecase.NewChatUsecase(&mockChatRepo{})
		if _, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: 1, Status: "bad_status"}); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for status, got %v", err)
		}
		if _, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: 1, StatusSource: "bad_source"}); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for source, got %v", err)
		}
		if _, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: 1, Priority: "bad_priority"}); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for priority, got %v", err)
		}
	})

	t.Run("success with default values", func(t *testing.T) {
		repo := &mockChatRepo{
			createFn: func(ctx context.Context, chat domain.Chat) (int64, error) {
				if chat.Status != domain.ChatStatusNewLead {
					t.Errorf("expected default status %s, got %s", domain.ChatStatusNewLead, chat.Status)
				}
				if chat.StatusSource != domain.StatusSourceAuto {
					t.Errorf("expected default source %s, got %s", domain.StatusSourceAuto, chat.StatusSource)
				}
				if chat.Priority != domain.ChatPriorityNormal {
					t.Errorf("expected default priority %s, got %s", domain.ChatPriorityNormal, chat.Priority)
				}
				if chat.LastMessageAt.IsZero() {
					t.Errorf("expected LastMessageAt to be set")
				}
				return 77, nil
			},
			getByIDFn: func(ctx context.Context, id int64) (*domain.Chat, error) {
				return &domain.Chat{
					ID:            id,
					AccountID:     1,
					ContactID:     2,
					Status:        domain.ChatStatusNewLead,
					StatusSource:  domain.StatusSourceAuto,
					Priority:      domain.ChatPriorityNormal,
					LastMessageAt: time.Now().UTC(),
				}, nil
			},
		}

		uc := usecase.NewChatUsecase(repo)
		chat, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: 2})
		if err != nil || chat.ID != 77 {
			t.Fatalf("unexpected result: %v, %+v", err, chat)
		}
	})
}
