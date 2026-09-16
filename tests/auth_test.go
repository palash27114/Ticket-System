package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ticket-system/internal/auth"
	"ticket-system/internal/health"
	"ticket-system/internal/user"

	"github.com/golang-jwt/jwt/v5"
)

func TestHealthEndpoint_AllServices(t *testing.T) {
	router, _ := setupTestServer()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// When DB pool is nil, we get 200 with status=ok for api and status=unknown for database
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d body: %s", w.Code, w.Body.String())
	}

	var body health.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected overall status 'ok', got %q", body.Status)
	}

	apiSvc, ok := body.Services["api"]
	if !ok {
		t.Fatal("expected 'api' key in services")
	}
	if apiSvc.Status != "ok" {
		t.Fatalf("expected api status 'ok', got %q", apiSvc.Status)
	}

	dbSvc, ok := body.Services["database"]
	if !ok {
		t.Fatal("expected 'database' key in services")
	}
	// With nil pool, database should be "unknown" (no connection configured)
	if dbSvc.Status != "unknown" {
		t.Fatalf("expected database status 'unknown' when pool is nil, got %q", dbSvc.Status)
	}
}

func TestHealthEndpoint_StructureValid(t *testing.T) {
	router, _ := setupTestServer()

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := raw["status"]; !ok {
		t.Fatal("response must contain 'status' field")
	}
	if _, ok := raw["services"]; !ok {
		t.Fatal("response must contain 'services' field")
	}

	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		t.Fatal("'services' must be an object")
	}
	if _, ok := services["api"]; !ok {
		t.Fatal("'services' must contain 'api' key")
	}
	if _, ok := services["database"]; !ok {
		t.Fatal("'services' must contain 'database' key")
	}
}

func TestRegistration_Success(t *testing.T) {
	router, _ := setupTestServer()

	payload := user.RegisterRequest{
		Name:     "Alice Wonderland",
		Email:    "alice@example.com",
		Password: "password123",
	}
	bodyBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp user.RegisterResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == 0 {
		t.Fatalf("expected non-zero user ID")
	}
	if resp.Email != "alice@example.com" {
		t.Fatalf("expected email 'alice@example.com', got %q", resp.Email)
	}
	if resp.Name != "Alice Wonderland" {
		t.Fatalf("expected name 'Alice Wonderland', got %q", resp.Name)
	}

	// Verify no password or password_hash leaked in JSON
	var rawMap map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &rawMap)
	if _, exists := rawMap["password"]; exists {
		t.Fatalf("password should not be returned in register response")
	}
	if _, exists := rawMap["password_hash"]; exists {
		t.Fatalf("password_hash should not be returned in register response")
	}
}

func TestRegistration_DuplicateEmail(t *testing.T) {
	router, _ := setupTestServer()

	payload := user.RegisterRequest{
		Name:     "Bob Builder",
		Email:    "bob@example.com",
		Password: "secretpassword",
	}
	bodyBytes, _ := json.Marshal(payload)

	// First registration
	req1, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first registration expected 201, got %d", w1.Code)
	}

	// Duplicate registration with same email
	req2, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict for duplicate email, got %d", w2.Code)
	}
}

func TestRegistration_Validation(t *testing.T) {
	router, _ := setupTestServer()

	// Short password (< 6 chars)
	invalidPayload := user.RegisterRequest{
		Name:     "Charlie",
		Email:    "charlie@example.com",
		Password: "123",
	}
	bodyBytes, _ := json.Marshal(invalidPayload)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for short password, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	router, _ := setupTestServer()

	// Register
	regPayload := user.RegisterRequest{
		Name:     "Dave",
		Email:    "dave@example.com",
		Password: "securepassword",
	}
	regBytes, _ := json.Marshal(regPayload)
	reqReg, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(regBytes))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	// Login
	loginPayload := user.LoginRequest{
		Email:    "dave@example.com",
		Password: "securepassword",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBytes))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", wLogin.Code)
	}

	var resp user.LoginResponse
	if err := json.Unmarshal(wLogin.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	if resp.Token == "" {
		t.Fatalf("expected non-empty JWT token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	router, _ := setupTestServer()

	regPayload := user.RegisterRequest{
		Name:     "Eve",
		Email:    "eve@example.com",
		Password: "correctpassword",
	}
	regBytes, _ := json.Marshal(regPayload)
	reqReg, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(regBytes))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	// Login with incorrect password
	loginPayload := user.LoginRequest{
		Email:    "eve@example.com",
		Password: "wrongpassword",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBytes))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", wLogin.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(wLogin.Body.Bytes(), &resp)
	if resp["error"] != "password incorrect" {
		t.Fatalf("expected password error, got %q", resp["error"])
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	router, _ := setupTestServer()

	loginPayload := user.LoginRequest{
		Email:    "nobody@example.com",
		Password: "anypassword",
	}
	loginBytes, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBytes))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for nonexistent user, got %d", wLogin.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(wLogin.Body.Bytes(), &resp)
	if resp["error"] != "please register first" {
		t.Fatalf("expected registration prompt, got %q", resp["error"])
	}
}

func TestAuthMiddleware_MissingJWT(t *testing.T) {
	router, _ := setupTestServer()

	req, _ := http.NewRequest(http.MethodGet, "/tickets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized when missing token, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidJWT(t *testing.T) {
	router, _ := setupTestServer()

	req, _ := http.NewRequest(http.MethodGet, "/tickets", nil)
	req.Header.Set("Authorization", "Bearer bad.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for invalid token, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredJWT(t *testing.T) {
	router, cfg := setupTestServer()

	// Create a token expired 1 hour ago
	expiredClaims := &auth.Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, _ := expiredToken.SignedString([]byte(cfg.JWTSecret))

	req, _ := http.NewRequest(http.MethodGet, "/tickets", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for expired token, got %d", w.Code)
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	router, _ := setupTestServer()

	req, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req.RequestURI = "/swagger/index.html"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for /swagger/index.html, got %d", w.Code)
	}
}
