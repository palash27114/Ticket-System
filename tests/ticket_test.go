package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticket-system/internal/ticket"
	"ticket-system/internal/user"

	"github.com/gin-gonic/gin"
)

func registerAndGetToken(t *testing.T, router *gin.Engine, name, email, password string) string {
	regPayload := user.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(regPayload)
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginPayload := user.LoginRequest{
		Email:    email,
		Password: password,
	}
	loginBody, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	var resp user.LoginResponse
	_ = json.Unmarshal(wLogin.Body.Bytes(), &resp)
	return resp.Token
}

func TestCreateTicket_SuccessAndDefaultOpen(t *testing.T) {
	router, _ := setupTestServer()
	token := registerAndGetToken(t, router, "User One", "user1@example.com", "password123")

	payload := ticket.CreateTicketRequest{
		Title:       "Internet not working",
		Description: "My internet connection has stopped working.",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body: %s", w.Code, w.Body.String())
	}

	var created ticket.Ticket
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created ticket: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected non-zero ticket ID")
	}
	if created.Title != "Internet not working" {
		t.Fatalf("expected title 'Internet not working', got %q", created.Title)
	}
	if created.Description != "My internet connection has stopped working." {
		t.Fatalf("expected description 'My internet connection has stopped working.', got %q", created.Description)
	}
	if created.Status != "open" {
		t.Fatalf("expected default status 'open', got %q", created.Status)
	}
}

func TestCreateTicket_Validation(t *testing.T) {
	router, _ := setupTestServer()
	token := registerAndGetToken(t, router, "User Two", "user2@example.com", "password123")

	// Empty title
	payload := ticket.CreateTicketRequest{
		Title:       "",
		Description: "Valid description",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for empty title, got %d", w.Code)
	}
}

func TestStatusStateMachine_Transitions(t *testing.T) {
	router, _ := setupTestServer()
	token := registerAndGetToken(t, router, "Status Tester", "status@example.com", "password123")

	// Helper to create ticket
	createTicket := func() int64 {
		p := ticket.CreateTicketRequest{
			Title:       "Test State Flow",
			Description: "Flow description",
		}
		b, _ := json.Marshal(p)
		r, _ := http.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(b))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		var created ticket.Ticket
		_ = json.Unmarshal(w.Body.Bytes(), &created)
		return created.ID
	}

	// Helper to patch status
	patchStatus := func(ticketID int64, newStatus string) (int, ticket.Ticket) {
		p := ticket.UpdateStatusRequest{Status: newStatus}
		b, _ := json.Marshal(p)
		r, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/tickets/%d/status", ticketID), bytes.NewBuffer(b))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		var updated ticket.Ticket
		_ = json.Unmarshal(w.Body.Bytes(), &updated)
		return w.Code, updated
	}

	// 1. Success path: open -> in_progress -> closed
	t.Run("open -> in_progress (success)", func(t *testing.T) {
		id := createTicket()
		code, updated := patchStatus(id, "in_progress")
		if code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", code)
		}
		if updated.Status != "in_progress" {
			t.Fatalf("expected status 'in_progress', got %q", updated.Status)
		}

		// and in_progress -> closed
		codeClosed, updatedClosed := patchStatus(id, "closed")
		if codeClosed != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", codeClosed)
		}
		if updatedClosed.Status != "closed" {
			t.Fatalf("expected status 'closed', got %q", updatedClosed.Status)
		}
	})

	// 2. Failure: open -> closed (skipping in_progress)
	t.Run("open -> closed (failure)", func(t *testing.T) {
		id := createTicket()
		code, _ := patchStatus(id, "closed")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 3. Failure: open -> open
	t.Run("open -> open (failure)", func(t *testing.T) {
		id := createTicket()
		code, _ := patchStatus(id, "open")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 4. Failure: in_progress -> open
	t.Run("in_progress -> open (failure)", func(t *testing.T) {
		id := createTicket()
		patchStatus(id, "in_progress")
		code, _ := patchStatus(id, "open")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 5. Failure: in_progress -> in_progress
	t.Run("in_progress -> in_progress (failure)", func(t *testing.T) {
		id := createTicket()
		patchStatus(id, "in_progress")
		code, _ := patchStatus(id, "in_progress")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 6. Failure: closed -> open
	t.Run("closed -> open (failure)", func(t *testing.T) {
		id := createTicket()
		patchStatus(id, "in_progress")
		patchStatus(id, "closed")
		code, _ := patchStatus(id, "open")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 7. Failure: closed -> in_progress
	t.Run("closed -> in_progress (failure)", func(t *testing.T) {
		id := createTicket()
		patchStatus(id, "in_progress")
		patchStatus(id, "closed")
		code, _ := patchStatus(id, "in_progress")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})

	// 8. Failure: closed -> closed
	t.Run("closed -> closed (failure)", func(t *testing.T) {
		id := createTicket()
		patchStatus(id, "in_progress")
		patchStatus(id, "closed")
		code, _ := patchStatus(id, "closed")
		if code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", code)
		}
	})
}
