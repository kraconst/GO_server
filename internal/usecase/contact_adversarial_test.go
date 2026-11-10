package usecase_test

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

func TestContactUsecase_Adversarial_WhitespaceAndEmptyName(t *testing.T) {
	ctx := context.Background()
	uc := usecase.NewContactUsecase(&mockContactRepo{})

	invalidNames := []string{
		"",
		" ",
		"    ",
		"\t",
		"\n",
		"\r\n",
		" \t\r\n ",
		"\u2003\u3000", // Unicode space characters
	}

	for _, name := range invalidNames {
		t.Run("Create with empty or whitespace name", func(t *testing.T) {
			_, err := uc.Create(ctx, domain.Contact{
				CompanyID:      1,
				TelegramUserID: 12345,
				Name:           name,
			})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for name %q, got: %v", name, err)
			}
		})

		t.Run("FindOrCreateByTelegramID with empty or whitespace name", func(t *testing.T) {
			_, err := uc.FindOrCreateByTelegramID(ctx, 1, 12345, name, "username")
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for name %q, got: %v", name, err)
			}
		})
	}
}

func TestContactUsecase_Adversarial_BoundaryIDs(t *testing.T) {
	ctx := context.Background()

	invalidIDs := []int64{0, -1, -500, math.MinInt64}

	for _, badID := range invalidIDs {
		t.Run("Create with invalid company_id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.Create(ctx, domain.Contact{
				CompanyID:      badID,
				TelegramUserID: 123,
				Name:           "Alice",
			})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for company_id %d, got: %v", badID, err)
			}
		})

		t.Run("Create with invalid telegram_user_id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.Create(ctx, domain.Contact{
				CompanyID:      1,
				TelegramUserID: badID,
				Name:           "Alice",
			})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for telegram_user_id %d, got: %v", badID, err)
			}
		})

		t.Run("FindOrCreate with invalid company_id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.FindOrCreateByTelegramID(ctx, badID, 123, "Alice", "")
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for company_id %d, got: %v", badID, err)
			}
		})

		t.Run("FindOrCreate with invalid telegram_user_id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.FindOrCreateByTelegramID(ctx, 1, badID, "Alice", "")
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for telegram_user_id %d, got: %v", badID, err)
			}
		})

		t.Run("GetByID with invalid id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.GetByID(ctx, badID)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for id %d, got: %v", badID, err)
			}
		})

		t.Run("ListByCompany with invalid company_id", func(t *testing.T) {
			uc := usecase.NewContactUsecase(&mockContactRepo{})
			_, err := uc.ListByCompany(ctx, badID)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation for company_id %d, got: %v", badID, err)
			}
		})
	}

	t.Run("MaxInt64 IDs pass validation and are passed to repository intact", func(t *testing.T) {
		repoCalled := false
		repo := &mockContactRepo{
			createFn: func(ctx context.Context, c domain.Contact) (int64, error) {
				if c.CompanyID != math.MaxInt64 || c.TelegramUserID != math.MaxInt64 {
					t.Errorf("expected max int64, got company=%d, tg=%d", c.CompanyID, c.TelegramUserID)
				}
				repoCalled = true
				return 42, nil
			},
			getByIDFn: func(ctx context.Context, id int64) (*domain.Contact, error) {
				return &domain.Contact{
					ID:             id,
					CompanyID:      math.MaxInt64,
					TelegramUserID: math.MaxInt64,
					Name:           "Max",
				}, nil
			},
		}

		uc := usecase.NewContactUsecase(repo)
		c, err := uc.Create(ctx, domain.Contact{
			CompanyID:      math.MaxInt64,
			TelegramUserID: math.MaxInt64,
			Name:           "Max",
		})
		if err != nil || !repoCalled || c.CompanyID != math.MaxInt64 {
			t.Fatalf("failed with MaxInt64: %v, contact=%+v", err, c)
		}
	})

	t.Run("Large input string does not crash and whitespace is trimmed", func(t *testing.T) {
		largeName := "  " + strings.Repeat("A", 1000) + "  "
		largeUsername := "  " + strings.Repeat("B", 1000) + "  "

		repo := &mockContactRepo{
			createFn: func(ctx context.Context, c domain.Contact) (int64, error) {
				if strings.HasPrefix(c.Name, " ") || strings.HasSuffix(c.Name, " ") {
					t.Errorf("name was not trimmed")
				}
				if strings.HasPrefix(c.Username, " ") || strings.HasSuffix(c.Username, " ") {
					t.Errorf("username was not trimmed")
				}
				return 1, nil
			},
		}

		uc := usecase.NewContactUsecase(repo)
		_, err := uc.Create(ctx, domain.Contact{
			CompanyID:      1,
			TelegramUserID: 1,
			Name:           largeName,
			Username:       largeUsername,
		})
		if err != nil {
			t.Fatalf("unexpected error for large input: %v", err)
		}
	})
}
