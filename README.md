# Siri Ice Cream API

A REST API for an ice cream shop, built with **Go + Fiber v2 + SQLite (GORM)**.

It manages **inventory items** and **customer records**, with **JWT authentication** and
**role-based access control** for three roles: **Owner**, **Manager** and **Cashier**.

**Features at a glance**
- Login / logout with JWT, passwords hashed with bcrypt
- Role permissions enforced on the backend (401 / 403)
- CRUD for items and customers, **search by name** for both
- Input validation with clear error messages, one JSON error format
- Settings from a `.env` file, validated at startup
- Structured logging (`log/slog`) with a request ID on every request
- 32 automated tests and a Postman collection with 74 automatic checks
- ER diagram of the database

---

## Table of contents
1. [Tech stack](#1-tech-stack)
2. [How to run](#2-how-to-run)
3. [Seeded user accounts](#3-seeded-user-accounts)
4. [Role permissions](#4-role-permissions)
5. [API endpoints](#5-api-endpoints)
6. [Testing with Postman](#6-testing-with-postman)
7. [Database design (ER diagram)](#7-database-design-er-diagram)
8. [Unit tests](#8-unit-tests)
9. [Logging](#9-logging)
10. [Project structure](#10-project-structure)
11. [Design decisions](#11-design-decisions)
12. [Troubleshooting](#12-troubleshooting)

---

## 1. Tech stack

| Part | Technology |
|---|---|
| Language | Go (version in `go.mod`) |
| Web framework | [Fiber v2](https://github.com/gofiber/fiber) |
| Database | SQLite |
| Database access | [GORM](https://gorm.io) with the pure-Go driver [`ncruces/go-sqlite3`](https://github.com/ncruces/go-sqlite3) (no C compiler needed) |
| Authentication | JWT ([`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt)) |
| Password hashing | bcrypt (`golang.org/x/crypto/bcrypt`) |
| Configuration | `.env` file ([`joho/godotenv`](https://github.com/joho/godotenv)) |
| Logging | `log/slog` (Go standard library) |

## 2. How to run

### Option A: on your computer

**Requirement:** Go installed (`go version` to check; the minimum version is in `go.mod`).

```cmd
git clone https://github.com/chaminipalliyaguru/Siri_Ice_Cream.git
cd Siri_Ice_Cream
copy .env.example .env
go mod tidy
go run ./cmd/server
```
(On macOS / Linux use `cp .env.example .env`.)

### Option B: GitHub Codespaces (nothing to install)

1. On the GitHub repository page: **Code → Codespaces → Create codespace on master**.
2. In the terminal:
   ```bash
   cp .env.example .env
   go run ./cmd/server
   ```
3. Open the **Ports** tab. To call the API from Postman, right-click port **3000** →
   **Port Visibility → Public**, and use the forwarded address as `baseUrl`.

### What happens on start

The API starts on **http://localhost:3000** and you will see:
```
level=INFO msg="server starting" env=development port=3000 db_path=siri_icecream_shop.db
```
On the first run, the app automatically:
- creates the SQLite database file `siri_icecream_shop.db`
- creates all tables from the Go models (GORM AutoMigrate)
- adds the starting data (seeder): 3 users, 3 categories and 4 items

To start again with a fresh database, stop the server and delete `siri_icecream_shop.db`.

### Settings (`.env`)

All settings have defaults, so the app also runs without a `.env` file.
Real environment variables always win over the `.env` file.

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` or `production` |
| `PORT` | `3000` | HTTP port (1–65535) |
| `DB_PATH` | `siri_icecream_shop.db` | SQLite database file |
| `JWT_SECRET` | `dev-secret-change-me` | Key used to sign tokens |

The settings are **checked when the app starts**. If something is wrong, the app does not start
and prints every problem at once, for example:
```
could not load config: invalid config: JWT_SECRET must be at least 32 characters in production
```
In **production** (`APP_ENV=production`), `JWT_SECRET` is required and must be at least 32 characters,
so the public default secret can never be used on a real server.

`.env` is ignored by git, so secrets are never committed. `.env.example` shows which settings exist.

## 3. Seeded user accounts

User accounts are created by the database seeder (`internal/database/seed.go`), as the
assignment allows. Passwords are stored as **bcrypt hashes**, never as plain text.
The seeder only runs when the `users` table is empty, so restarting never duplicates data.

| Username | Password | Role |
|---|---|---|
| `owner` | `owner123` | Owner |
| `manager` | `manager123` | Manager |
| `cashier` | `cashier123` | Cashier |

Also seeded:
- **Categories (flavors):** Chocolate, Vanilla, Strawberry
- **Items:** Chocolate Cone, Double Chocolate Cup, Vanilla Cone, Strawberry Sundae

## 4. Role permissions

Permissions are enforced on the **backend** by middleware. All rules are defined in one place:
`internal/routes/routes.go`.

| Action | Owner | Manager | Cashier |
|---|:-:|:-:|:-:|
| View items | ✅ | ✅ | ✅ |
| Add item | ✅ | ✅ | ❌ |
| Edit item | ✅ | ✅ | ✅ |
| Delete item | ✅ | ❌ | ❌ |
| View customers | ✅ | ✅ | ✅ |
| Add customer | ✅ | ✅ | ✅ |
| Edit customer | ✅ | ✅ | ✅ |
| Delete customer | ✅ | ❌ | ❌ |

| Situation | Response |
|---|---|
| No token, invalid token, expired token or logged-out token | **401 Unauthorized** |
| Valid token, but the role is not allowed | **403 Forbidden** |

## 5. API endpoints

Base URL: `http://localhost:3000`

Every endpoint **except login** needs this header:
```
Authorization: Bearer <token>
```

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/auth/login` | Log in and receive a JWT token (valid for 24 hours) |
| POST | `/api/auth/logout` | Log out. The token can no longer be used |
| GET | `/api/auth/me` | Get the currently logged-in user |

### Inventory items

| Method | Endpoint | Description | Roles |
|---|---|---|---|
| GET | `/api/items` | List all items (each with its category) | All |
| GET | `/api/items?name={text}` | **Search items by name** (case-insensitive, partial match) | All |
| GET | `/api/items/{id}` | Get one item | All |
| POST | `/api/items` | Create an item | Owner, Manager |
| PUT / PATCH | `/api/items/{id}` | Update an item (only the fields sent are changed) | All |
| DELETE | `/api/items/{id}` | Delete an item | Owner |

### Customers

| Method | Endpoint | Description | Roles |
|---|---|---|---|
| GET | `/api/customers` | List all customers | All |
| GET | `/api/customers?name={text}` | Search customers by name | All |
| GET | `/api/customers/{id}` | Get one customer | All |
| POST | `/api/customers` | Create a customer | All |
| PUT / PATCH | `/api/customers/{id}` | Update a customer (only the fields sent are changed) | All |
| DELETE | `/api/customers/{id}` | Delete a customer | Owner |

### Categories

| Method | Endpoint | Description | Roles |
|---|---|---|---|
| GET | `/api/categories` | List flavors (to know which `category_id` to use) | All |

### Request and response examples

**Login** – `POST /api/auth/login`
```json
{ "username": "owner", "password": "owner123" }
```
Response `200 OK`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { "id": 1, "name": "Shop Owner", "username": "owner", "role": "owner", "created_at": "...", "updated_at": "..." }
}
```
The password hash is never included in any response.

**Create item** – `POST /api/items`
```json
{ "name": "Mint Cup", "category_id": 2, "price": 300, "quantity": 10 }
```
Response `201 Created`:
```json
{
  "id": 5, "name": "Mint Cup", "category_id": 2,
  "category": { "id": 2, "name": "Vanilla", "created_at": "...", "updated_at": "..." },
  "price": 300, "quantity": 10, "created_at": "...", "updated_at": "..."
}
```

**Update only the quantity** – `PATCH /api/items/5`
```json
{ "quantity": 0 }
```
The other fields (name, price, category) stay the same.

**Search items** – `GET /api/items?name=choc` returns *Chocolate Cone* and *Double Chocolate Cup*.

**Create customer** – `POST /api/customers`
```json
{ "name": "Nimal Perera", "email": "nimal@example.com", "phone": "0771234567" }
```

### Validation rules

| Resource | Rule | Status |
|---|---|---|
| Login | `username` and `password` are required | 400 |
| Item | `name` is required | 400 |
| Item | `price` must be greater than 0 | 400 |
| Item | `quantity` cannot be negative | 400 |
| Item | `category_id` must be an existing category | 400 |
| Customer | `name` is required | 400 |
| Customer | `email` is required and must be a valid email address | 400 |
| Customer | `email` must be unique (case-insensitive) | 409 |
| Any | the request body must be valid JSON | 400 |
| Any | `{id}` in the URL must be a positive number | 400 |
| Any | a record with that `{id}` must exist | 404 |

### Error format

Every error uses the same JSON shape:
```json
{ "error": "you do not have permission to do this" }
```
Unexpected server errors are logged with full details, but the client only receives
`{"error": "internal server error"}`, so internal details are never exposed.

### Status codes used

| Code | Meaning |
|---|---|
| 200 | OK |
| 201 | Created |
| 400 | Bad request (invalid input) |
| 401 | Unauthorized (not logged in / invalid token) |
| 403 | Forbidden (role not allowed) |
| 404 | Not found |
| 409 | Conflict (duplicate email) |
| 500 | Server error |

## 6. Testing with Postman

A ready-made collection is included: **`docs/SiriIceCream.postman_collection.json`**
(42 requests, each with automatic checks).

1. Open Postman → **Import** → select the file.
2. Start the server (`go run ./cmd/server`).
3. Open folder **1. Auth** and send **Login as Owner**. The token is saved automatically,
   so every other request works straight away.

**Run everything at once:** right-click the collection → **Run collection** → **Run**.
The requests run in order and every one checks its own status code and response
(74 checks in total). You can run it as many times as you like.

| Folder | What it shows |
|---|---|
| 1. Auth | Login as each role (tokens saved), wrong password (401), empty fields (400), `/me` |
| 2. Categories | List the flavors |
| 3. Items | List, search, get, create, PUT, PATCH, validation errors (400), delete, 404 |
| 4. Customers | Create, list, search, get, PUT, PATCH, duplicate email (409), invalid email (400), delete, 404 |
| 5. Role permissions | Cashier and Manager trying actions they are **not** allowed (403), Owner deleting (200), no token (401) |
| 6. Logout | Logout, then the same token is rejected (401) |

Collection variables (filled in automatically): `baseUrl` (default `http://localhost:3000`),
`token`, `ownerToken`, `managerToken`, `cashierToken`, `itemId`, `customerId`, `customerEmail`.
To use another address (for example a Codespaces URL), change `baseUrl`.

## 7. Database design (ER diagram)

![ER diagram](docs/er-diagram.png)

| Table | Purpose | Key points |
|---|---|---|
| **users** | Staff who log in | `username` is unique; `password_hash` stores a bcrypt hash; `role` is `owner`, `manager` or `cashier` |
| **categories** | Ice cream flavors (Chocolate, Vanilla, Strawberry) | `name` is unique |
| **items** | Inventory | `category_id` is a **foreign key** to `categories`; `name` has an index for faster search |
| **customers** | Shop customers | `email` is unique and stored in lower case |

**Relationship:** one category has many items (**1 : N**). Each item belongs to exactly one category.
A category that still has items cannot be deleted (`ON DELETE RESTRICT`).

**Users and customers are separate tables:** *users* are staff who log in, while *customers* are
people the shop sells to and never log in.

**Where rules are enforced:**
- **In the database:** primary keys, `NOT NULL`, unique `username` / `email` / category `name`,
  the foreign key, and `quantity` defaulting to 0.
- **In the application:** price > 0, quantity ≥ 0 and a valid email format.
  Roles are fixed constants in `models/user.go`; users are only created by the seeder.

Tables are created automatically from the Go structs in `internal/models` (GORM AutoMigrate).
The diagram source is `docs/er-diagram.mmd` ([Mermaid](https://mermaid.js.org) format, which GitHub can display).

## 8. Unit tests

```cmd
go test ./...
```
Add `-v` to see every test by name. There are **32 tests** in 7 files.

Each test uses a fresh **in-memory** SQLite database, so tests never touch the real
`siri_icecream_shop.db` and never affect each other. API tests send real HTTP requests
into the app with Fiber's `app.Test()`, without starting a server.

| File | What it checks |
|---|---|
| `internal/auth/jwt_test.go` | Token creation and reading, 24-hour lifetime, wrong secret, expired token, malformed token, logout blacklist |
| `internal/config/config_test.go` | Default settings, invalid settings, production rules for `JWT_SECRET` |
| `internal/database/database_test.go` | Tables are created, the seeder adds the right data, seeding twice does not duplicate, passwords are hashed |
| `internal/routes/routes_test.go` | Login cases, 401 without a token, logout, and **every role against every action** |
| `internal/routes/item_routes_test.go` | Item create / read / update / delete, search by name, validation, 404 and invalid id |
| `internal/routes/customer_routes_test.go` | Customer create / read / update / delete, email normalisation, validation, duplicate email (409) |
| `internal/routes/extra_routes_test.go` | `/me`, login never returns the password hash, categories, invalid JSON, JSON error format, `X-Request-Id` header, customer search, updating a customer with their own email |

## 9. Logging

The app uses Go's structured logger `log/slog`:
- **development:** readable text
- **production:** JSON, which log tools can search and filter

Every request writes one line with a **request ID**, method, path, status, duration, IP and (when
logged in) the user ID. The level follows the status: 2xx → INFO, 4xx → WARN, 5xx → ERROR.
```
level=WARN msg=request request_id=5dbb46d5-... method=DELETE path=/api/items/1 status=403 duration_ms=0 ip=127.0.0.1 user_id=3
```
The same request ID is returned to the client in the `X-Request-Id` response header, so a reported
problem can be found in the logs. Passwords, tokens and the JWT secret are never logged.

## 10. Project structure

```
Siri_Ice_Cream/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point: config → logger → database → seed → routes → start
├── internal/
│   ├── config/
│   │   ├── config.go               # Loads .env and environment variables, validates them
│   │   └── config_test.go
│   ├── logging/
│   │   └── logging.go              # Structured logger (text in development, JSON in production)
│   ├── database/
│   │   ├── database.go             # Opens SQLite and creates the tables
│   │   ├── seed.go                 # Starting users, categories and items
│   │   └── database_test.go
│   ├── models/                     # Database tables as Go structs
│   │   ├── user.go
│   │   ├── category.go
│   │   ├── item.go
│   │   └── customer.go
│   ├── auth/
│   │   ├── jwt.go                  # Create and verify JWT tokens
│   │   ├── blacklist.go            # Logged-out tokens
│   │   └── jwt_test.go
│   ├── middleware/
│   │   ├── auth.go                 # RequireAuth (401) and RequireRoles (403)
│   │   └── request_logger.go       # One structured log line per request
│   ├── handlers/                   # Request handling, one file per resource
│   │   ├── auth_handler.go
│   │   ├── category_handler.go
│   │   ├── item_handler.go
│   │   ├── customer_handler.go
│   │   └── helpers.go              # Shared helpers (parse body, parse id, errors)
│   └── routes/
│       ├── routes.go               # URLs, middleware, role permissions, error handler
│       └── *_test.go               # API tests
├── docs/
│   ├── er-diagram.png
│   ├── er-diagram.mmd
│   └── SiriIceCream.postman_collection.json
├── .env.example                    # Template for the .env file
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

### How a request flows

```
HTTP request
   │
   ▼
Request ID + request logger   (every request gets an ID and one log line)
   │
   ▼
RequireAuth    → is there a valid, not-logged-out token?   (no → 401)
   │
   ▼
RequireRoles   → is this user's role allowed here?         (no → 403)
   │
   ▼
Handler        → validate input → database (GORM)          (400 / 404 / 409)
   │
   ▼
JSON response  ← any error goes to one error handler → status code + {"error": "..."}
```

## 11. Design decisions

- **Clear package structure.** Each package has one job: `config` (settings), `models` (data),
  `database` (connection and seed), `auth` (tokens), `middleware` (security and logging),
  `handlers` (request logic), `routes` (wiring and permissions).
- **`internal/` folder.** Go does not allow other projects to import packages inside `internal/`,
  which keeps the app's code private.
- **Dependency injection.** The database connection, secret and logger are passed in through
  constructors (`NewItemHandler(db)`, `NewApp(db, secret, logger)`), not global variables.
  This is what makes the tests easy: they create the app with an in-memory database.
- **Permissions in one place.** All role rules are in `routes.go`, so they are easy to review
  and change. Handlers do not check roles themselves.
- **Fail fast on bad configuration.** Invalid settings stop the app at startup with a clear message,
  instead of causing problems later.
- **Request structs.** Handlers read the JSON body into small request structs, not directly into
  the models, so clients cannot set fields like `id` or `created_at`.
- **Partial updates.** Update requests use pointer fields, so a field that is not sent (`nil`) is
  left unchanged, while `0` or `""` can still be set on purpose.
- **Logout with JWT.** A JWT cannot be deleted, so logout adds the token to a blacklist until it
  expires. The blacklist is kept in memory, so it resets when the server restarts. A production
  system could store it in Redis or the database instead.
- **Security.** Passwords are hashed with bcrypt; the login error is the same for an unknown user
  and a wrong password (so usernames cannot be guessed); only the HS256 signing method is accepted;
  all SQL uses parameter placeholders (`?`) to prevent SQL injection; secrets are never logged.
- **Consistent errors.** One error handler returns every error as `{"error": "..."}` and hides
  internal details of unexpected errors.
- **Email normalisation.** Customer emails are trimmed and lower-cased before saving, so
  `Nimal@Example.com` and `nimal@example.com` count as the same email.
- **Price** is stored as a number (`REAL`). A real shop might store cents as an integer to avoid
  rounding issues.

### Possible next steps
- Split the handlers into service and repository layers
- Versioned SQL migrations instead of AutoMigrate
- Graceful shutdown and a `/health` endpoint
- Pagination for the list endpoints

## 12. Troubleshooting

| Problem | Solution |
|---|---|
| `An Application Control policy has blocked this file` or `blocked by your organization's Device Guard policy` (Windows) | The computer does not allow new programs to run. Use **GitHub Codespaces** (section 2, option B) or another computer |
| `could not load config: invalid config: ...` | Fix the setting named in the message in your `.env` file |
| `address already in use` | Another program uses port 3000. Stop it, or set another `PORT` |
| Postman gets **401** from a Codespaces URL, and no log line appears | The port is private. In the **Ports** tab set port 3000 to **Public** |
| `token has been logged out` after a restart | Log in again to get a new token |
| `checksum mismatch` or missing modules | Delete `go.sum` and run `go mod tidy` |
