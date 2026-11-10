package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/handler"
)

func TestHandler_Adversarial_QueryParsing(t *testing.T) {
	h := handler.New(nil, nil, nil, nil, nil)
	r := setupTestRouter(h)

	t.Run("malformed status query params return 400", func(t *testing.T) {
		malformedStatuses := []string{
			"invalid",
			"NEW_LEAD",
			"New_Lead",
			"new_lead ",
			" new_lead",
			"DROP TABLE chats;",
			"<script>alert(1)</script>",
			"new_lead\n",
			"negotiation\x00",
		}

		for _, bad := range malformedStatuses {
			req, err := http.NewRequest(http.MethodGet, "/chats?status="+url.QueryEscape(bad), nil)
			if err != nil {
				t.Fatalf("unexpected NewRequest error: %v", err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for status %q, got %d: %s", bad, w.Code, w.Body.String())
			}
		}
	})

	t.Run("empty status query param is ignored and returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/chats?status=", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for empty status query, got %d", w.Code)
		}
	})

	t.Run("malformed employee_id query params return 400", func(t *testing.T) {
		malformedEmployeeIDs := []string{
			"abc",
			"-1",
			"-500",
			"0",
			"1.5",
			"1e5",
			" 1 ",
			"9223372036854775808", // MaxInt64 + 1 (overflow)
			"null",
			"NaN",
		}

		for _, bad := range malformedEmployeeIDs {
			req, _ := http.NewRequest(http.MethodGet, "/chats?employee_id="+bad, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for employee_id %q, got %d", bad, w.Code)
			}
		}
	})

	t.Run("empty employee_id query param is ignored and returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/chats?employee_id=", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for empty employee_id query, got %d", w.Code)
		}
	})

	t.Run("multiple query params combinations", func(t *testing.T) {
		// Mock setup has:
		// Chat 1: Account 1 -> Employee 1
		// Chat 2: Account 2 -> Employee 2

		// 1. Employee 2 has negotiation chat
		req, _ := http.NewRequest(http.MethodGet, "/chats?status=negotiation&employee_id=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var chats []domain.Chat
		if err := json.Unmarshal(w.Body.Bytes(), &chats); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if len(chats) != 1 || chats[0].ID != 2 {
			t.Fatalf("expected 1 chat with ID 2, got %+v", chats)
		}

		// 2. Mismatch combination: employee 1 has no negotiation chat -> returns empty list [] (not 404 or null)
		req, _ = http.NewRequest(http.MethodGet, "/chats?status=negotiation&employee_id=1", nil)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var emptyChats []domain.Chat
		if err := json.Unmarshal(w.Body.Bytes(), &emptyChats); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if len(emptyChats) != 0 {
			t.Fatalf("expected 0 chats, got %d", len(emptyChats))
		}
		if string(w.Body.Bytes()) != "[]" {
			t.Fatalf("expected JSON '[]', got: %s", w.Body.String())
		}
	})
}

func TestHandler_Adversarial_MockStateMutation(t *testing.T) {
	h := handler.New(nil, nil, nil, nil, nil)
	r := setupTestRouter(h)

	t.Run("modifying status updates mockChats and consecutive GET /chats/:id returns updated manual status", func(t *testing.T) {
		targetChatID := int64(2)
		newStatus := domain.ChatStatusActionRequired

		body := fmt.Sprintf(`{"status": "%s"}`, newStatus)
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/chats/%d/status", targetChatID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Consecutive GET /chats/:id
		reqGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/chats/%d", targetChatID), nil)
		wGet := httptest.NewRecorder()
		r.ServeHTTP(wGet, reqGet)

		if wGet.Code != http.StatusOK {
			t.Fatalf("expected 200 on GET, got %d", wGet.Code)
		}

		var chat domain.Chat
		if err := json.Unmarshal(wGet.Body.Bytes(), &chat); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		if chat.ID != targetChatID {
			t.Fatalf("expected chat ID %d, got %d", targetChatID, chat.ID)
		}
		if chat.Status != newStatus {
			t.Fatalf("expected status %s, got %s", newStatus, chat.Status)
		}
		if chat.StatusSource != domain.StatusSourceManual {
			t.Fatalf("expected status_source manual, got %s", chat.StatusSource)
		}

		// Second consecutive update to another status
		newStatus2 := domain.ChatStatusUnhappy
		body2 := fmt.Sprintf(`{"status": "%s"}`, newStatus2)
		req2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/chats/%d/status", targetChatID), bytes.NewBufferString(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 on second update, got %d", w2.Code)
		}

		reqGet2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/chats/%d", targetChatID), nil)
		wGet2 := httptest.NewRecorder()
		r.ServeHTTP(wGet2, reqGet2)

		var chat2 domain.Chat
		if err := json.Unmarshal(wGet2.Body.Bytes(), &chat2); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if chat2.Status != newStatus2 {
			t.Fatalf("expected updated status %s, got %s", newStatus2, chat2.Status)
		}
	})

	t.Run("modifying non-existent chat ID returns 404 without panicking", func(t *testing.T) {
		nonExistentIDs := []string{
			"999999",
			fmt.Sprintf("%d", int64(math.MaxInt64)),
			"12345678",
		}

		for _, nonExistentID := range nonExistentIDs {
			body := `{"status": "lost"}`
			req, _ := http.NewRequest(http.MethodPost, "/chats/"+nonExistentID+"/status", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Must not panic
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404 for non-existent ID %s, got %d: %s", nonExistentID, w.Code, w.Body.String())
			}
		}
	})

	t.Run("modifying with invalid ID boundary values returns 400", func(t *testing.T) {
		badIDs := []string{
			"0",
			"-1",
			"-9999",
			"not-a-number",
			"9223372036854775808", // MaxInt64 + 1
		}

		for _, badID := range badIDs {
			body := `{"status": "lost"}`
			req, _ := http.NewRequest(http.MethodPost, "/chats/"+badID+"/status", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for bad ID %s, got %d", badID, w.Code)
			}
		}
	})

	t.Run("modifying with invalid status payloads returns 400", func(t *testing.T) {
		badPayloads := []string{
			`{"status": ""}`,
			`{"status": "unknown"}`,
			`{"status": "NEW_LEAD"}`,
			`{"status": "DROP TABLE chats;"}`,
			`{"status": 123}`,
			`{"status": null}`,
			`{}`,
			`not a json`,
		}

		for _, payload := range badPayloads {
			req, _ := http.NewRequest(http.MethodPost, "/chats/1/status", bytes.NewBufferString(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for payload %s, got %d", payload, w.Code)
			}
		}
	})
}
