package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/usecase"
)

type mockAccountDailyStatsRepo struct {
	upsertFunc                   func(ctx context.Context, stats domain.AccountDailyStats) error
	incrementCountersFunc        func(ctx context.Context, accountID int64, day time.Time, messagesIn, messagesOut, newContacts int) error
	getByAccountAndDateRangeFunc func(ctx context.Context, accountID int64, from, to time.Time) ([]domain.AccountDailyStats, error)
	getByCompanyAndDateRangeFunc func(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountDailyStats, error)
	getDashboardFunc             func(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error)
	getAccountRatingsFunc        func(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error)
}

func (m *mockAccountDailyStatsRepo) Upsert(ctx context.Context, stats domain.AccountDailyStats) error {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, stats)
	}
	return nil
}

func (m *mockAccountDailyStatsRepo) IncrementCounters(ctx context.Context, accountID int64, day time.Time, messagesIn, messagesOut, newContacts int) error {
	if m.incrementCountersFunc != nil {
		return m.incrementCountersFunc(ctx, accountID, day, messagesIn, messagesOut, newContacts)
	}
	return nil
}

func (m *mockAccountDailyStatsRepo) GetByAccountAndDateRange(ctx context.Context, accountID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	if m.getByAccountAndDateRangeFunc != nil {
		return m.getByAccountAndDateRangeFunc(ctx, accountID, from, to)
	}
	return nil, nil
}

func (m *mockAccountDailyStatsRepo) GetByCompanyAndDateRange(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	if m.getByCompanyAndDateRangeFunc != nil {
		return m.getByCompanyAndDateRangeFunc(ctx, companyID, from, to)
	}
	return nil, nil
}

func (m *mockAccountDailyStatsRepo) GetDashboard(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error) {
	if m.getDashboardFunc != nil {
		return m.getDashboardFunc(ctx, companyID, from, to)
	}
	return &domain.DashboardStats{}, nil
}

func (m *mockAccountDailyStatsRepo) GetAccountRatings(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error) {
	if m.getAccountRatingsFunc != nil {
		return m.getAccountRatingsFunc(ctx, companyID, from, to)
	}
	return nil, nil
}

func TestAccountDailyStatsUsecase_Upsert(t *testing.T) {
	ctx := context.Background()

	t.Run("valid stats", func(t *testing.T) {
		called := false
		repo := &mockAccountDailyStatsRepo{
			upsertFunc: func(ctx context.Context, stats domain.AccountDailyStats) error {
				called = true
				if stats.AccountID != 1 {
					t.Errorf("expected AccountID 1, got %d", stats.AccountID)
				}
				return nil
			},
		}
		uc := usecase.NewAccountDailyStatsUsecase(repo)

		err := uc.Upsert(ctx, domain.AccountDailyStats{
			AccountID:       1,
			Day:             time.Now(),
			MessagesIn:      10,
			MessagesOut:     5,
			ActiveChats:     3,
			NewContacts:     2,
			AvgResponseTime: 30 * time.Second,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected Upsert to be called on repo")
		}
	})

	t.Run("invalid account_id", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.Upsert(ctx, domain.AccountDailyStats{
			AccountID: 0,
			Day:       time.Now(),
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("zero day", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.Upsert(ctx, domain.AccountDailyStats{
			AccountID: 1,
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("negative metric values", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.Upsert(ctx, domain.AccountDailyStats{
			AccountID:  1,
			Day:        time.Now(),
			MessagesIn: -1,
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestAccountDailyStatsUsecase_IncrementCounters(t *testing.T) {
	ctx := context.Background()

	t.Run("valid increment", func(t *testing.T) {
		called := false
		repo := &mockAccountDailyStatsRepo{
			incrementCountersFunc: func(ctx context.Context, accountID int64, day time.Time, in, out, contacts int) error {
				called = true
				if accountID != 2 || in != 5 || out != 3 || contacts != 1 {
					t.Errorf("unexpected parameters: accountID=%d, in=%d, out=%d, contacts=%d", accountID, in, out, contacts)
				}
				return nil
			},
		}
		uc := usecase.NewAccountDailyStatsUsecase(repo)

		err := uc.IncrementCounters(ctx, 2, time.Now(), 5, 3, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Errorf("expected IncrementCounters to be called on repo")
		}
	})

	t.Run("invalid account_id", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.IncrementCounters(ctx, -1, time.Now(), 1, 0, 0)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("zero day", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.IncrementCounters(ctx, 1, time.Time{}, 1, 0, 0)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("negative increments", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		err := uc.IncrementCounters(ctx, 1, time.Now(), -1, 0, 0)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestAccountDailyStatsUsecase_GetDashboard(t *testing.T) {
	ctx := context.Background()

	t.Run("valid company_id with custom date range", func(t *testing.T) {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		repo := &mockAccountDailyStatsRepo{
			getDashboardFunc: func(ctx context.Context, companyID int64, f, toDate time.Time) (*domain.DashboardStats, error) {
				return &domain.DashboardStats{
					MessagesIn:      100,
					MessagesOut:     80,
					ActiveChats:     20,
					NewContacts:     5,
					AvgResponseTime: 2 * time.Minute,
				}, nil
			},
		}
		uc := usecase.NewAccountDailyStatsUsecase(repo)

		res, err := uc.GetDashboard(ctx, 1, from, to)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.MessagesIn != 100 {
			t.Errorf("expected 100 MessagesIn, got %d", res.MessagesIn)
		}
	})

	t.Run("invalid company_id", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		_, err := uc.GetDashboard(ctx, 0, time.Time{}, time.Time{})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("from after to error", func(t *testing.T) {
		from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		_, err := uc.GetDashboard(ctx, 1, from, to)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestAccountDailyStatsUsecase_GetAccountRatings(t *testing.T) {
	ctx := context.Background()

	t.Run("valid company_id", func(t *testing.T) {
		repo := &mockAccountDailyStatsRepo{
			getAccountRatingsFunc: func(ctx context.Context, companyID int64, f, toDate time.Time) ([]domain.AccountRating, error) {
				return []domain.AccountRating{
					{
						AccountID:     1,
						Username:      "@anna",
						TotalMessages: 50,
					},
				}, nil
			},
		}
		uc := usecase.NewAccountDailyStatsUsecase(repo)

		ratings, err := uc.GetAccountRatings(ctx, 1, time.Time{}, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ratings) != 1 || ratings[0].TotalMessages != 50 {
			t.Errorf("unexpected ratings result: %+v", ratings)
		}
	})

	t.Run("invalid company_id", func(t *testing.T) {
		uc := usecase.NewAccountDailyStatsUsecase(&mockAccountDailyStatsRepo{})
		_, err := uc.GetAccountRatings(ctx, -5, time.Time{}, time.Time{})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}
