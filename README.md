# Ticket System Backend

A simple, clean, production-ready, Dockerized Go REST API for a ticket management system. Built with idiomatic Go, PostgreSQL, JWT authentication, and a strict ticket status state machine.

---

## Tech Stack

- **Language:** Go (1.24 / 1.27)
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin)
- **API Documentation:** [Swagger / OpenAPI](https://github.com/swaggo/gin-swagger) via Swagger UI
- **Database:** PostgreSQL (v16)
- **Database Driver / Connection Pool:** [pgx/v5](https://github.com/jackc/pgx/v5) (`pgxpool`)
- **Authentication / Authorization:** [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5)
- **Password Hashing:** [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Containerization:** Docker & Docker Compose
- **Migrations:** Embedded SQL migrations (`migrations/001_init.sql`)

---

## Features

- **Interactive Swagger UI:** Test all endpoints (`/health`, `/auth/*`, `/tickets/*`) directly in the browser with authorization support at `/swagger`.
- **User Registration & Login:** Secure authentication with bcrypt hashed passwords and JWT token issuance.
- **Strict Ownership Isolation:** Authenticated users can only create, list, retrieve, and update their own tickets. Accessing or updating another user's ticket returns `404 Not Found` to prevent information leakage.
- **Ticket Status State Machine:** Tickets strictly follow the linear flow:
  ```text
  open -> in_progress -> closed
  ```
  - `open -> in_progress` is allowed.
  - `in_progress -> closed` is allowed.
  - Any other transition (including reopening a closed ticket or transitioning to the same status) is strictly rejected with `400 Bad Request`.
- **Automatic Migrations:** Embedded SQL schema migrations automatically run on startup.
- **Dockerized & Cloud-Ready:** Multi-stage Docker build producing a minimal runtime image listening on `0.0.0.0:8080`.

---

## Project Structure

```text
ticket-system/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point & dependency wiring
├── internal/
│   ├── config/
│   │   └── config.go            # Environment variable configuration
│   ├── database/
│   │   └── postgres.go          # pgxpool connection & migration runner
│   ├── middleware/
│   │   └── auth.go              # JWT authentication & context injection middleware
│   ├── auth/
│   │   ├── handler.go           # Register & Login HTTP handlers
│   │   ├── service.go           # Registration, bcrypt hashing, and login service
│   │   └── jwt.go               # JWT issuance and token validation
│   ├── user/
│   │   ├── model.go             # User entity and DTO definitions
│   │   └── repository.go        # User repository interface and Postgres implementation
│   ├── ticket/
│   │   ├── model.go             # Ticket entity, DTOs, and state machine validation
│   │   ├── repository.go        # Ticket repository interface and Postgres implementation
│   │   ├── service.go           # Ticket business logic & ownership validation
│   │   └── handler.go           # Ticket HTTP handlers (CRUD & status update)
│   └── server/
│       └── router.go            # Router setup and endpoint registration
├── migrations/
│   ├── 001_init.sql             # Database schema DDL, indexes, and constraints
│   └── migrations.go            # Embedded SQL migration asset
├── tests/
│   ├── mock_repo_test.go        # In-memory test fixtures and mock repositories
│   ├── auth_test.go             # Auth unit and integration tests
│   ├── ticket_test.go           # Ticket CRUD & state machine transition tests
│   └── ownership_test.go        # Multi-user authorization & isolation tests
├── docs/                        # Auto-generated Swagger / OpenAPI specs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── Dockerfile                   # Multi-stage production container build
├── docker-compose.yml           # Docker Compose dev environment (API + PostgreSQL)
├── .env.example                 # Sample environment variables
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Swagger UI (Interactive API Testing)

Once the server is running, open your browser and navigate to:
```text
http://localhost:8080/swagger
```
or
```text
http://localhost:8080/swagger/index.html
```

### How to test protected endpoints in Swagger:
1. Use `POST /auth/register` to create an account.
2. Use `POST /auth/login` to authenticate and copy the returned `token`.
3. Click the green **Authorize** button at the top right of the Swagger UI page.
4. In the value input, enter:
   ```text
   Bearer <your_token>
   ```
5. Click **Authorize** and then **Close**.
6. You can now execute and test all protected `/tickets` endpoints directly from Swagger!

---

## API Documentation

All request and response bodies use JSON. Protected endpoints require the `Authorization` header formatted as:
```http
Authorization: Bearer <token>
```

### 1. Health Check
- **Method:** `GET`
- **URL:** `/health`
- **Authentication:** None
- **Response Example (200 OK):**
  ```json
  {
    "status": "ok"
  }
  ```
- **Status Codes:** `200 OK`

---

### 2. User Registration
- **Method:** `POST`
- **URL:** `/auth/register`
- **Authentication:** None
- **Request Example:**
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }
  ```
- **Response Example (201 Created):**
  ```json
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
  ```
- **Status Codes:**
  - `201 Created`: User successfully registered.
  - `400 Bad Request`: Validation failure (empty fields, invalid email format, password under 6 characters).
  - `409 Conflict`: Email already exists.

---

### 3. User Login
- **Method:** `POST`
- **URL:** `/auth/login`
- **Authentication:** None
- **Request Example:**
  ```json
  {
    "email": "john@example.com",
    "password": "password123"
  }
  ```
- **Response Example (200 OK):**
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Status Codes:**
  - `200 OK`: Successful login with valid JWT returned.
  - `400 Bad Request`: Invalid request payload.
  - `401 Unauthorized`: Invalid email or password.

---

### 4. Create Ticket
- **Method:** `POST`
- **URL:** `/tickets`
- **Authentication:** Bearer Token required
- **Request Example:**
  ```json
  {
    "title": "Internet not working",
    "description": "My internet connection has stopped working."
  }
  ```
- **Response Example (201 Created):**
  ```json
  {
    "id": 1,
    "title": "Internet not working",
    "description": "My internet connection has stopped working.",
    "status": "open",
    "created_at": "2026-09-16T10:00:00Z",
    "updated_at": "2026-09-16T10:00:00Z"
  }
  ```
- **Status Codes:**
  - `201 Created`: Ticket successfully created with initial status `open`.
  - `400 Bad Request`: Empty title or description.
  - `401 Unauthorized`: Missing or invalid JWT.

---

### 5. List Tickets
- **Method:** `GET`
- **URL:** `/tickets`
- **Authentication:** Bearer Token required
- **Response Example (200 OK):**
  ```json
  [
    {
      "id": 2,
      "title": "Login issue",
      "description": "Cannot login",
      "status": "open",
      "created_at": "2026-09-16T10:05:00Z",
      "updated_at": "2026-09-16T10:05:00Z"
    },
    {
      "id": 1,
      "title": "Internet problem",
      "description": "Internet is down",
      "status": "closed",
      "created_at": "2026-09-16T10:00:00Z",
      "updated_at": "2026-09-16T10:02:00Z"
    }
  ]
  ```
- **Status Codes:**
  - `200 OK`: Returns only tickets belonging to the authenticated user, ordered by `created_at DESC`.
  - `401 Unauthorized`: Missing or invalid JWT.

---

### 6. Get Own Ticket
- **Method:** `GET`
- **URL:** `/tickets/:id`
- **Authentication:** Bearer Token required
- **Response Example (200 OK):**
  ```json
  {
    "id": 1,
    "title": "Internet not working",
    "description": "My internet connection has stopped working.",
    "status": "open",
    "created_at": "2026-09-16T10:00:00Z",
    "updated_at": "2026-09-16T10:00:00Z"
  }
  ```
- **Status Codes:**
  - `200 OK`: Ticket details returned.
  - `401 Unauthorized`: Missing or invalid JWT.
  - `404 Not Found`: Ticket does not exist or does not belong to the authenticated user (`{"error": "ticket not found"}`).

---

### 7. Update Ticket Status
- **Method:** `PATCH`
- **URL:** `/tickets/:id/status`
- **Authentication:** Bearer Token required
- **Request Example:**
  ```json
  {
    "status": "in_progress"
  }
  ```
- **Response Example (200 OK):**
  ```json
  {
    "id": 1,
    "title": "Internet not working",
    "description": "My internet connection has stopped working.",
    "status": "in_progress",
    "created_at": "2026-09-16T10:00:00Z",
    "updated_at": "2026-09-16T10:04:00Z"
  }
  ```
- **Status Codes:**
  - `200 OK`: Status successfully updated.
  - `400 Bad Request`: Invalid transition attempted (e.g. `open -> closed`, `closed -> open`, or same state).
  - `401 Unauthorized`: Missing or invalid JWT.
  - `404 Not Found`: Ticket does not exist or does not belong to the authenticated user (`{"error": "ticket not found"}`).

---

## Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port on which the application listens (`0.0.0.0:PORT`) |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/ticketdb?sslmode=disable` | PostgreSQL connection connection string |
| `JWT_SECRET` | `change-me` | Secret key used for signing and verifying JWTs |
| `JWT_EXPIRATION_HOURS` | `24` | Token expiration duration in hours |

---

## Local Setup

### 1. Download Dependencies
```bash
go mod download
```

### 2. Configure Environment
Copy `.env.example` to `.env` and configure your database connection string and JWT secret:
```bash
cp .env.example .env
```

### 3. Run PostgreSQL Locally
Ensure a PostgreSQL server is running and a database named `ticketdb` exists:
```bash
createdb ticketdb
```

### 4. Run the Application
The application automatically runs migrations on startup:
```bash
go run cmd/server/main.go
```

### 5. Run Tests
Execute the comprehensive test suite covering authentication, ticket CRUD, state machine transitions, and ownership isolation:
```bash
go test -v ./tests/...
```

---

## Docker Setup

### Option 1: Docker Compose (Recommended for local development)
Runs the API server and a PostgreSQL container with persistent volume storage:
```bash
docker compose up --build
```
To run in background:
```bash
docker compose up -d --build
```
To stop and remove containers:
```bash
docker compose down
```

### Option 2: Docker Build & Run (Single Container)
Build the image:
```bash
docker build -t ticket-system .
```
Run with environment variables:
```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:secret@host.docker.internal:5432/ticketdb?sslmode=disable" \
  -e JWT_SECRET="your-production-secret-key" \
  ticket-system
```

---

## Health Check

Verify application availability by pinging the health check endpoint:
```bash
curl http://localhost:8080/health
```
**Expected Response:**
```json
{
  "status": "ok"
}
```

---

## Deployment

The application is fully containerized and deployable to any Docker-compatible platform (Render, Railway, Fly.io, AWS ECS, GCP Cloud Run, DigitalOcean App Platform, Kubernetes).

- **Deployed URL:** `<URL>`
- **Health URL:** `<URL>/health`
