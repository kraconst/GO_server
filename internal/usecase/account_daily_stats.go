package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/kraconst/GO_server/internal/domain"
)

type AccountDailyStatsUsecase struct {
	repo AccountDailyStatsRepository
}

func NewAccountDailyStatsUsecase(repo AccountDailyStatsRepository) *AccountDailyStatsUsecase {
	return &AccountDailyStatsUsecase{repo: repo}
}

func (u *AccountDailyStatsUsecase) Upsert(ctx context.Context, stats domain.AccountDailyStats) error {
	if stats.AccountID <= 0 {
		return fmt.Errorf("%w: account_id must be a positive number", domain.ErrValidation)
	}
	if stats.Day.IsZero() {
		return fmt.Errorf("%w: day is required", domain.ErrValidation)
	}
	if stats.MessagesIn < 0 || stats.MessagesOut < 0 || stats.ActiveChats < 0 || stats.NewContacts < 0 || stats.AvgResponseTime < 0 {
		return fmt.Errorf("%w: metric values cannot be negative", domain.ErrValidation)
	}
	return u.repo.Upsert(ctx, stats)
}

func (u *AccountDailyStatsUsecase) IncrementCounters(ctx context.Context, accountID int64, day time.Time, messagesIn, messagesOut, newContacts int) error {
	if accountID <= 0 {
		return fmt.Errorf("%w: account_id must be a positive number", domain.ErrValidation)
	}
	if day.IsZero() {
		return fmt.Errorf("%w: day is required", domain.ErrValidation)
	}
	if messagesIn < 0 || messagesOut < 0 || newContacts < 0 {
		return fmt.Errorf("%w: increment values cannot be negative", domain.ErrValidation)
	}
	return u.repo.IncrementCounters(ctx, accountID, day, messagesIn, messagesOut, newContacts)
}

func (u *AccountDailyStatsUsecase) GetByAccountAndDateRange(ctx context.Context, accountID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	if accountID <= 0 {
		return nil, fmt.Errorf("%w: account_id must be a positive number", domain.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		to = time.Now().UTC()
		from = to.AddDate(0, 0, -30)
	}
	if from.After(to) {
		return nil, fmt.Errorf("%w: 'from' date must be before or equal to 'to' date", domain.ErrValidation)
	}
	return u.repo.GetByAccountAndDateRange(ctx, accountID, from, to)
}

func (u *AccountDailyStatsUsecase) GetByCompanyAndDateRange(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: company_id must be a positive number", domain.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		to = time.Now().UTC()
		from = to.AddDate(0, 0, -30)
	}
	if from.After(to) {
		return nil, fmt.Errorf("%w: 'from' date must be before or equal to 'to' date", domain.ErrValidation)
	}
	return u.repo.GetByCompanyAndDateRange(ctx, companyID, from, to)
}

func (u *AccountDailyStatsUsecase) GetDashboard(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: company_id must be a positive number", domain.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		to = time.Now().UTC()
		from = to.AddDate(0, 0, -30)
	}
	if from.After(to) {
		return nil, fmt.Errorf("%w: 'from' date must be before or equal to 'to' date", domain.ErrValidation)
	}
	return u.repo.GetDashboard(ctx, companyID, from, to)
}

func (u *AccountDailyStatsUsecase) GetAccountRatings(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: company_id must be a positive number", domain.ErrValidation)
	}
	if from.IsZero() || to.IsZero() {
		to = time.Now().UTC()
		from = to.AddDate(0, 0, -30)
	}
	if from.After(to) {
		return nil, fmt.Errorf("%w: 'from' date must be before or equal to 'to' date", domain.ErrValidation)
	}
	return u.repo.GetAccountRatings(ctx, companyID, from, to)
}
