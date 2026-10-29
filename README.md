# Siri Ice Cream Shop

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
- **Bonus:** a C# WPF desktop app (.NET 10) that uses the API (section 13)

---

## Table of contents
1. [Tech stack](#1-tech-stack)
2. [How to run](#2-how-to-run)
3. [Seeded user accounts](#3-seeded-user-accounts)
4. [Role permissions](#4-role-permissions)
5. [API endpoints](#5-api-endpoints)
6. [Testing with Postman](#6-testing-with-postman)


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

