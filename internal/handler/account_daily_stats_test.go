package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/handler"
	"github.com/kraconst/GO_server/internal/usecase"
)

type mockDailyStatsRepoForHandler struct {
	getDashboardFunc      func(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error)
	getAccountRatingsFunc func(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error)
}

func (m *mockDailyStatsRepoForHandler) Upsert(ctx context.Context, stats domain.AccountDailyStats) error {
	return nil
}
func (m *mockDailyStatsRepoForHandler) IncrementCounters(ctx context.Context, accountID int64, day time.Time, in, out, contacts int) error {
	return nil
}
func (m *mockDailyStatsRepoForHandler) GetByAccountAndDateRange(ctx context.Context, accountID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	return nil, nil
}
func (m *mockDailyStatsRepoForHandler) GetByCompanyAndDateRange(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountDailyStats, error) {
	return nil, nil
}
func (m *mockDailyStatsRepoForHandler) GetDashboard(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error) {
	if m.getDashboardFunc != nil {
		return m.getDashboardFunc(ctx, companyID, from, to)
	}
	return &domain.DashboardStats{}, nil
}
func (m *mockDailyStatsRepoForHandler) GetAccountRatings(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error) {
	if m.getAccountRatingsFunc != nil {
		return m.getAccountRatingsFunc(ctx, companyID, from, to)
	}
	return nil, nil
}

func setupTestRouter(statsUc *usecase.AccountDailyStatsUsecase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.New(nil, nil, nil, statsUc)

	r.GET("/analytics/dashboard", h.GetDashboard)
	r.GET("/analytics/accounts", h.GetAccountRatings)
	return r
}

func TestHandler_GetDashboard_MockFallback(t *testing.T) {
	r := setupTestRouter(nil)

	req, _ := http.NewRequest(http.MethodGet, "/analytics/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp domain.DashboardStats
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.MessagesIn == 0 {
		t.Errorf("expected non-zero messages_in in mock data")
	}
}

func TestHandler_GetDashboard_WithUsecase(t *testing.T) {
	repo := &mockDailyStatsRepoForHandler{
		getDashboardFunc: func(ctx context.Context, companyID int64, from, to time.Time) (*domain.DashboardStats, error) {
			return &domain.DashboardStats{
				MessagesIn:      999,
				MessagesOut:     888,
				ActiveChats:     50,
				NewContacts:     12,
				AvgResponseTime: 5 * time.Minute,
			}, nil
		},
	}
	uc := usecase.NewAccountDailyStatsUsecase(repo)
	r := setupTestRouter(uc)

	req, _ := http.NewRequest(http.MethodGet, "/analytics/dashboard?from=2026-01-01&to=2026-01-10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp domain.DashboardStats
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.MessagesIn != 999 {
		t.Errorf("expected MessagesIn 999, got %d", resp.MessagesIn)
	}
}

func TestHandler_GetDashboard_InvalidDates(t *testing.T) {
	r := setupTestRouter(nil)

	t.Run("invalid from format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/analytics/dashboard?from=01-01-2026", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", w.Code)
		}
	})

	t.Run("from after to", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/analytics/dashboard?from=2026-05-10&to=2026-05-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", w.Code)
		}
	})
}

func TestHandler_GetAccountRatings(t *testing.T) {
	t.Run("mock fallback", func(t *testing.T) {
		r := setupTestRouter(nil)

		req, _ := http.NewRequest(http.MethodGet, "/analytics/accounts", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var ratings []domain.AccountRating
		if err := json.Unmarshal(w.Body.Bytes(), &ratings); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(ratings) == 0 {
			t.Errorf("expected non-empty ratings in mock fallback")
		}
	})

	t.Run("with usecase", func(t *testing.T) {
		repo := &mockDailyStatsRepoForHandler{
			getAccountRatingsFunc: func(ctx context.Context, companyID int64, from, to time.Time) ([]domain.AccountRating, error) {
				return []domain.AccountRating{
					{
						AccountID:     10,
						Username:      "@alice",
						TotalMessages: 200,
					},
				}, nil
			},
		}
		uc := usecase.NewAccountDailyStatsUsecase(repo)
		r := setupTestRouter(uc)

		req, _ := http.NewRequest(http.MethodGet, "/analytics/accounts?from=2026-02-01&to=2026-02-15", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var ratings []domain.AccountRating
		if err := json.Unmarshal(w.Body.Bytes(), &ratings); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(ratings) != 1 || ratings[0].Username != "@alice" {
			t.Errorf("unexpected ratings response: %+v", ratings)
		}
	})
}
