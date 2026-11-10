package usecase_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

func TestChatUsecase_Adversarial_InvalidStatusStrings(t *testing.T) {
	ctx := context.Background()
	uc := usecase.NewChatUsecase(&mockChatRepo{})

	invalidStatuses := []string{
		"",
		"unknown",
		"DROP TABLE chats;",
		"NEW_LEAD",
		"New_Lead",
		"new_lead\n",
		" new_lead ",
		"SELECT * FROM chats",
		"<script>alert(1)</script>",
		"negotiation\x00",
		"action_required_extra",
	}

	for _, bad := range invalidStatuses {
		t.Run("SetManualStatus with status "+bad, func(t *testing.T) {
			err := uc.SetManualStatus(ctx, 1, domain.ChatStatus(bad))
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for status %q, got: %v", bad, err)
			}
		})

		t.Run("ListByCompany with status "+bad, func(t *testing.T) {
			st := domain.ChatStatus(bad)
			_, err := uc.ListByCompany(ctx, 1, &st, nil)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for status %q, got: %v", bad, err)
			}
		})

		if bad != "" { // empty status is allowed in Create because it defaults to new_lead
			t.Run("Create with status "+bad, func(t *testing.T) {
				_, err := uc.Create(ctx, domain.Chat{
					AccountID: 1,
					ContactID: 1,
					Status:    domain.ChatStatus(bad),
				})
				if !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("expected ErrValidation for status %q in Create, got: %v", bad, err)
				}
			})
		}
	}
}

func TestChatUsecase_Adversarial_BoundaryIDs(t *testing.T) {
	ctx := context.Background()

	invalidIDs := []int64{0, -1, -100, math.MinInt64}

	for _, badID := range invalidIDs {
		t.Run("SetManualStatus boundary id", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			err := uc.SetManualStatus(ctx, badID, domain.ChatStatusNewLead)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for id %d, got: %v", badID, err)
			}
		})

		t.Run("GetByID boundary id", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			_, err := uc.GetByID(ctx, badID)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for id %d, got: %v", badID, err)
			}
		})

		t.Run("ListByCompany boundary companyID", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			_, err := uc.ListByCompany(ctx, badID, nil, nil)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for companyID %d, got: %v", badID, err)
			}
		})

		t.Run("ListByCompany boundary employeeID", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			empID := badID
			_, err := uc.ListByCompany(ctx, 1, nil, &empID)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for employeeID %d, got: %v", badID, err)
			}
		})

		t.Run("Create boundary accountID", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			_, err := uc.Create(ctx, domain.Chat{AccountID: badID, ContactID: 1})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for accountID %d, got: %v", badID, err)
			}
		})

		t.Run("Create boundary contactID", func(t *testing.T) {
			uc := usecase.NewChatUsecase(&mockChatRepo{})
			_, err := uc.Create(ctx, domain.Chat{AccountID: 1, ContactID: badID})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for contactID %d, got: %v", badID, err)
			}
		})
	}

	t.Run("MaxInt64 ID does not cause integer overflow or panic", func(t *testing.T) {
		repoCalled := false
		repo := &mockChatRepo{
			getByIDFn: func(ctx context.Context, id int64) (*domain.Chat, error) {
				if id != math.MaxInt64 {
					t.Errorf("expected max int64 %d, got %d", int64(math.MaxInt64), id)
				}
				repoCalled = true
				return &domain.Chat{ID: id}, nil
			},
		}
		uc := usecase.NewChatUsecase(repo)
		chat, err := uc.GetByID(ctx, math.MaxInt64)
		if err != nil || !repoCalled || chat.ID != math.MaxInt64 {
			t.Fatalf("expected successful call with MaxInt64, got err=%v, chat=%+v", err, chat)
		}
	})
}
