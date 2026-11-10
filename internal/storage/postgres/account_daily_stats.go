package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
)

type AccountDailyStatsStorage struct {
	db *sqlx.DB
}

func NewAccountDailyStatsStorage(db *sqlx.DB) *AccountDailyStatsStorage {
	return &AccountDailyStatsStorage{db: db}
}

func (s *AccountDailyStatsStorage) Upsert(ctx context.Context, stats domain.AccountDailyStats) error {
	query := `
		INSERT INTO account_daily_stats (account_id, day, messages_in, messages_out, active_chats, new_contacts, avg_response_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (account_id, day)
		DO UPDATE SET
			messages_in = EXCLUDED.messages_in,
			messages_out = EXCLUDED.messages_out,
			active_chats = EXCLUDED.active_chats,
			new_contacts = EXCLUDED.new_contacts,
			avg_response_time = EXCLUDED.avg_response_time
	`
	_, err := s.db.ExecContext(ctx, query,
		stats.AccountID,
		stats.Day.Format("2006-01-02"),
		stats.MessagesIn,
		stats.MessagesOut,
		stats.ActiveChats,
		stats.NewContacts,
		int64(stats.AvgResponseTime),
	)
	if err != nil {
		return fmt.Errorf("upsert account daily stats: %w", err)
	}
	return nil
}

func (s *AccountDailyStatsStorage) IncrementCounters(ctx context.Context, accountID int64, day time.Time, messagesIn, messagesOut, newContacts int) error {
	query := `
		INSERT INTO account_daily_stats (account_id, day, messages_in, messages_out, active_chats, new_contacts, avg_response_time)
		VALUES ($1, $2, $3, $4, 0, $5, 0)
		ON CONFLICT (account_id, day)
		DO UPDATE SET
			messages_in = account_daily_stats.messages_in + EXCLUDED.messages_in,
			messages_out = account_daily_stats.messages_out + EXCLUDED.messages_out,
			new_contacts = account_daily_stats.new_contacts + EXCLUDED.new_contacts
	`
	_, err := s.db.ExecContext(ctx, query,
		accountID,
		day.Format("2006-01-02"),
		messagesIn,
		messagesOut,
		newContacts,
	)
	if err != nil {
		return fmt.Errorf("increment account daily stats counters: %w", err)
	}
	return nil
}

func (s *AccountDailyStatsStorage) GetByAccountAndDateRange(ctx context.Context, accountID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	query := `
		SELECT account_id, day, messages_in, messages_out, active_chats, new_contacts, avg_response_time
		FROM account_daily_stats
		WHERE account_id = $1 AND day >= $2 AND day <= $3
		ORDER BY day ASC
	`
	stats := make([]domain.AccountDailyStats, 0)
	err := s.db.SelectContext(ctx, &stats, query, accountID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("get account daily stats by account and range: %w", err)
	}
	return stats, nil
}

func (s *AccountDailyStatsStorage) GetByCompanyAndDateRange(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	query := `
		SELECT s.account_id, s.day, s.messages_in, s.messages_out, s.active_chats, s.new_contacts, s.avg_response_time
		FROM account_daily_stats s
		JOIN accounts a ON a.id = s.account_id
		WHERE a.company_id = $1 AND s.day >= $2 AND s.day <= $3
		ORDER BY s.day ASC, s.account_id ASC
	`
	stats := make([]domain.AccountDailyStats, 0)
	err := s.db.SelectContext(ctx, &stats, query, companyID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("get account daily stats by company and range: %w", err)
	}
	return stats, nil
}

func (s *AccountDailyStatsStorage) GetDashboard(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error) {
	query := `
		SELECT
			COALESCE(SUM(s.messages_in), 0) AS messages_in,
			COALESCE(SUM(s.messages_out), 0) AS messages_out,
			COALESCE(SUM(s.active_chats), 0) AS active_chats,
			COALESCE(SUM(s.new_contacts), 0) AS new_contacts,
			COALESCE(AVG(s.avg_response_time), 0)::BIGINT AS avg_response_time,
			0 AS unanswered_chats
		FROM account_daily_stats s
		JOIN accounts a ON a.id = s.account_id
		WHERE a.company_id = $1 AND s.day >= $2 AND s.day <= $3
	`
	var res domain.DashboardStats
	err := s.db.GetContext(ctx, &res, query, companyID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.DashboardStats{}, nil
		}
		return nil, fmt.Errorf("get dashboard stats: %w", err)
	}
	return &res, nil
}

func (s *AccountDailyStatsStorage) GetAccountRatings(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error) {
	query := `
		SELECT
			a.id AS account_id,
			a.username,
			COALESCE(e.name, '') AS employee_name,
			COALESCE(SUM(s.messages_in), 0) AS messages_in,
			COALESCE(SUM(s.messages_out), 0) AS messages_out,
			COALESCE(SUM(s.messages_in + s.messages_out), 0) AS total_messages,
			COALESCE(SUM(s.active_chats), 0) AS active_chats,
			COALESCE(SUM(s.new_contacts), 0) AS new_contacts,
			COALESCE(AVG(s.avg_response_time), 0)::BIGINT AS avg_response_time
		FROM accounts a
		LEFT JOIN employees e ON a.employee_id = e.id
		LEFT JOIN account_daily_stats s ON s.account_id = a.id AND s.day >= $2 AND s.day <= $3
		WHERE a.company_id = $1
		GROUP BY a.id, a.username, e.name
		ORDER BY total_messages DESC, a.id ASC
	`
	ratings := make([]domain.AccountRating, 0)
	err := s.db.SelectContext(ctx, &ratings, query, companyID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("get account ratings: %w", err)
	}
	return ratings, nil
}
