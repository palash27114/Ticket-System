package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticket-system/internal/ticket"
)

func TestOwnership_MultiUserIsolation(t *testing.T) {
	router, _ := setupTestServer()

	// 1. Create User A and User B
	tokenA := registerAndGetToken(t, router, "User A", "usera@example.com", "password123")
	tokenB := registerAndGetToken(t, router, "User B", "userb@example.com", "password123")

	// 2. User A creates Ticket #1
	createPayload := ticket.CreateTicketRequest{
		Title:       "User A's Secret Ticket",
		Description: "This ticket belongs exclusively to User A.",
	}
	bodyA, _ := json.Marshal(createPayload)
	reqA, _ := http.NewRequest(http.MethodPost, "/tickets", bytes.NewBuffer(bodyA))
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	reqA.Header.Set("Content-Type", "application/json")
	wA := httptest.NewRecorder()
	router.ServeHTTP(wA, reqA)

	if wA.Code != http.StatusCreated {
		t.Fatalf("expected User A ticket creation to succeed with 201, got %d", wA.Code)
	}

	var ticketA ticket.Ticket
	if err := json.Unmarshal(wA.Body.Bytes(), &ticketA); err != nil {
		t.Fatalf("failed to decode User A ticket: %v", err)
	}

	ticketAID := ticketA.ID

	// 3. User B attempts to access User A's Ticket #1 via GET /tickets/:id
	reqBGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%d", ticketAID), nil)
	reqBGet.Header.Set("Authorization", "Bearer "+tokenB)
	wBGet := httptest.NewRecorder()
	router.ServeHTTP(wBGet, reqBGet)

	if wBGet.Code != http.StatusNotFound {
		t.Fatalf("expected User B GET /tickets/%d to return 404 Not Found, got %d", ticketAID, wBGet.Code)
	}

	var errResp map[string]string
	_ = json.Unmarshal(wBGet.Body.Bytes(), &errResp)
	if errResp["error"] != "ticket not found" {
		t.Fatalf("expected error message 'ticket not found', got %q", errResp["error"])
	}

	// 4. User B attempts to modify User A's Ticket #1 via PATCH /tickets/:id/status
	patchPayload := ticket.UpdateStatusRequest{Status: "in_progress"}
	bodyPatch, _ := json.Marshal(patchPayload)
	reqBPatch, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/tickets/%d/status", ticketAID), bytes.NewBuffer(bodyPatch))
	reqBPatch.Header.Set("Authorization", "Bearer "+tokenB)
	reqBPatch.Header.Set("Content-Type", "application/json")
	wBPatch := httptest.NewRecorder()
	router.ServeHTTP(wBPatch, reqBPatch)

	if wBPatch.Code != http.StatusNotFound {
		t.Fatalf("expected User B PATCH /tickets/%d/status to return 404 Not Found, got %d", ticketAID, wBPatch.Code)
	}

	// 5. User B lists tickets via GET /tickets - must not include User A's ticket
	reqBList, _ := http.NewRequest(http.MethodGet, "/tickets", nil)
	reqBList.Header.Set("Authorization", "Bearer "+tokenB)
	wBList := httptest.NewRecorder()
	router.ServeHTTP(wBList, reqBList)

	if wBList.Code != http.StatusOK {
		t.Fatalf("expected User B GET /tickets to return 200 OK, got %d", wBList.Code)
	}

	var ticketsB []ticket.Ticket
	if err := json.Unmarshal(wBList.Body.Bytes(), &ticketsB); err != nil {
		t.Fatalf("failed to decode User B tickets list: %v", err)
	}

	if len(ticketsB) != 0 {
		t.Fatalf("expected User B to have 0 tickets, got %d", len(ticketsB))
	}

	// 6. User A accesses their own Ticket #1 via GET /tickets/:id
	reqAGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/tickets/%d", ticketAID), nil)
	reqAGet.Header.Set("Authorization", "Bearer "+tokenA)
	wAGet := httptest.NewRecorder()
	router.ServeHTTP(wAGet, reqAGet)

	if wAGet.Code != http.StatusOK {
		t.Fatalf("expected User A GET /tickets/%d to return 200 OK, got %d", ticketAID, wAGet.Code)
	}

	// 7. User A updates their own Ticket #1 status via PATCH /tickets/:id/status
	reqAPatch, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/tickets/%d/status", ticketAID), bytes.NewBuffer(bodyPatch))
	reqAPatch.Header.Set("Authorization", "Bearer "+tokenA)
	reqAPatch.Header.Set("Content-Type", "application/json")
	wAPatch := httptest.NewRecorder()
	router.ServeHTTP(wAPatch, reqAPatch)

	if wAPatch.Code != http.StatusOK {
		t.Fatalf("expected User A PATCH /tickets/%d/status to return 200 OK, got %d", ticketAID, wAPatch.Code)
	}

	// 8. User A lists tickets via GET /tickets - must include User A's ticket
	reqAList, _ := http.NewRequest(http.MethodGet, "/tickets", nil)
	reqAList.Header.Set("Authorization", "Bearer "+tokenA)
	wAList := httptest.NewRecorder()
	router.ServeHTTP(wAList, reqAList)

	if wAList.Code != http.StatusOK {
		t.Fatalf("expected User A GET /tickets to return 200 OK, got %d", wAList.Code)
	}

	var ticketsA []ticket.Ticket
	_ = json.Unmarshal(wAList.Body.Bytes(), &ticketsA)
	if len(ticketsA) != 1 {
		t.Fatalf("expected User A to have 1 ticket, got %d", len(ticketsA))
	}
	if ticketsA[0].ID != ticketAID {
		t.Fatalf("expected ticket ID %d, got %d", ticketAID, ticketsA[0].ID)
	}
}
