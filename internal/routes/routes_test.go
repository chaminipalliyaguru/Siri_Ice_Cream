package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"siri_ice_cream/internal/database"
)

const testSecret = "test-secret"

// setupApp creates a fresh app with a brand-new seeded database for each test.
func setupApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := database.Connect(":memory:")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return NewApp(db, testSecret, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// request sends an HTTP request to the app and returns the status code and body.
func request(t *testing.T, app *fiber.App, method, path, token string, body interface{}) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

// login logs in as the given user and returns the token.
func login(t *testing.T, app *fiber.App, username, password string) string {
	t.Helper()
	status, body := request(t, app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": username, "password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("login as %s: status %d, body %s", username, status, body)
	}
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(body, &resp)
	return resp.Token
}

// ---------------- Authentication ----------------

func TestLogin(t *testing.T) {
	app := setupApp(t)

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
	}{
		{"correct password", "owner", "owner123", http.StatusOK},
		{"wrong password", "owner", "wrong", http.StatusUnauthorized},
		{"unknown user", "nobody", "owner123", http.StatusUnauthorized},
		{"empty fields", "", "", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := request(t, app, http.MethodPost, "/api/auth/login", "", map[string]string{
				"username": tc.username, "password": tc.password,
			})
			if status != tc.wantStatus {
				t.Errorf("got status %d, want %d (body: %s)", status, tc.wantStatus, body)
			}
		})
	}
}

func TestProtectedRouteNeedsToken(t *testing.T) {
	app := setupApp(t)

	status, _ := request(t, app, http.MethodGet, "/api/items", "", nil)
	if status != http.StatusUnauthorized {
		t.Errorf("no token: got %d, want 401", status)
	}

	status, _ = request(t, app, http.MethodGet, "/api/items", "not-a-real-token", nil)
	if status != http.StatusUnauthorized {
		t.Errorf("bad token: got %d, want 401", status)
	}
}

func TestLogoutBlocksToken(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "manager", "manager123")

	status, _ := request(t, app, http.MethodPost, "/api/auth/logout", token, nil)
	if status != http.StatusOK {
		t.Fatalf("logout: got %d, want 200", status)
	}

	status, _ = request(t, app, http.MethodGet, "/api/items", token, nil)
	if status != http.StatusUnauthorized {
		t.Errorf("after logout: got %d, want 401", status)
	}
}

// ---------------- Role-based access control ----------------

// TestRolePermissions checks every role against every protected action.
func TestRolePermissions(t *testing.T) {
	newItem := map[string]interface{}{"name": "Mint Cup", "category_id": 1, "price": 300, "quantity": 10}
	editItem := map[string]interface{}{"quantity": 5}
	newCustomer := map[string]interface{}{"name": "Kamal", "email": "kamal@example.com", "phone": "0771111111"}
	editCustomer := map[string]interface{}{"phone": "0772222222"}

	type action struct {
		name, method, path string
		body               interface{}
	}
	actions := []action{
		{"view items", http.MethodGet, "/api/items", nil},
		{"view item", http.MethodGet, "/api/items/1", nil},
		{"add item", http.MethodPost, "/api/items", newItem},
		{"edit item", http.MethodPatch, "/api/items/1", editItem},
		{"delete item", http.MethodDelete, "/api/items/2", nil},
		{"view customers", http.MethodGet, "/api/customers", nil},
		{"add customer", http.MethodPost, "/api/customers", newCustomer},
		{"view customer", http.MethodGet, "/api/customers/1", nil},
		{"edit customer", http.MethodPut, "/api/customers/1", editCustomer},
		{"delete customer", http.MethodDelete, "/api/customers/1", nil},
	}

	// true = allowed, false = must get 403 Forbidden
	permissions := map[string]map[string]bool{
		"owner": {
			"view items": true, "view item": true, "add item": true, "edit item": true, "delete item": true,
			"view customers": true, "add customer": true, "view customer": true, "edit customer": true, "delete customer": true,
		},
		"manager": {
			"view items": true, "view item": true, "add item": true, "edit item": true, "delete item": false,
			"view customers": true, "add customer": true, "view customer": true, "edit customer": true, "delete customer": false,
		},
		"cashier": {
			"view items": true, "view item": true, "add item": false, "edit item": true, "delete item": false,
			"view customers": true, "add customer": true, "view customer": true, "edit customer": true, "delete customer": false,
		},
	}

	for role, allowed := range permissions {
		t.Run(role, func(t *testing.T) {
			app := setupApp(t) // fresh database for each role
			token := login(t, app, role, role+"123")

			for _, a := range actions {
				status, body := request(t, app, a.method, a.path, token, a.body)
				if allowed[a.name] && status >= 300 {
					t.Errorf("%s should be allowed, got %d (body: %s)", a.name, status, body)
				}
				if !allowed[a.name] && status != http.StatusForbidden {
					t.Errorf("%s should be forbidden, got %d (body: %s)", a.name, status, body)
				}
			}
		})
	}
}

// Logging out must only block that one token: signing in again (as the same
// user or another user) must give a token that works.
func TestLoginAgainAfterLogout(t *testing.T) {
	app := setupApp(t)

	for _, user := range []string{"owner", "cashier", "owner", "owner", "manager"} {
		token := login(t, app, user, user+"123")

		status, _ := request(t, app, http.MethodGet, "/api/items", token, nil)
		if status != http.StatusOK {
			t.Fatalf("%s after a previous logout: got %d, want 200", user, status)
		}

		status, _ = request(t, app, http.MethodPost, "/api/auth/logout", token, nil)
		if status != http.StatusOK {
			t.Fatalf("%s logout: got %d, want 200", user, status)
		}
	}
}
