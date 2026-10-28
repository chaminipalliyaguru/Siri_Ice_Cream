package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"siri_ice_cream/internal/models"
)

func TestMeReturnsLoggedInUser(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "cashier", "cashier123")

	status, body := request(t, app, http.MethodGet, "/api/auth/me", token, nil)
	if status != http.StatusOK {
		t.Fatalf("me: got %d, body %s", status, body)
	}

	var user models.User
	json.Unmarshal(body, &user)
	if user.Username != "cashier" || user.Role != models.RoleCashier {
		t.Errorf("me returned the wrong user: %+v", user)
	}
}

func TestLoginNeverReturnsPasswordHash(t *testing.T) {
	app := setupApp(t)

	_, body := request(t, app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "owner", "password": "owner123",
	})

	var raw map[string]map[string]interface{}
	json.Unmarshal(body, &raw)
	if _, found := raw["user"]["password_hash"]; found {
		t.Error("login response must not contain the password hash")
	}
	if _, found := raw["user"]["PasswordHash"]; found {
		t.Error("login response must not contain the password hash")
	}
}

func TestListCategories(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "cashier", "cashier123")

	status, body := request(t, app, http.MethodGet, "/api/categories", token, nil)
	if status != http.StatusOK {
		t.Fatalf("categories: got %d", status)
	}

	var categories []models.Category
	json.Unmarshal(body, &categories)
	if len(categories) != 3 {
		t.Errorf("got %d categories, want 3", len(categories))
	}
}

func TestInvalidJSONBody(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	req := httptest.NewRequest(http.MethodPost, "/api/items", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("got %d, want 400", resp.StatusCode)
	}
}

func TestErrorsUseJSONFormat(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "cashier", "cashier123")

	// Cashier is not allowed to delete items.
	_, body := request(t, app, http.MethodDelete, "/api/items/1", token, nil)

	var resp map[string]string
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("error body is not JSON: %s", body)
	}
	if resp["error"] == "" {
		t.Errorf(`error body should be {"error": "..."}, got %s`, body)
	}
}

func TestEveryResponseHasRequestID(t *testing.T) {
	app := setupApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/items", nil) // no token: 401
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.Header.Get("X-Request-Id") == "" {
		t.Error("response is missing the X-Request-Id header")
	}
}

func TestCustomerSearchByName(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "manager", "manager123")

	request(t, app, http.MethodPost, "/api/customers", token, map[string]string{"name": "Nimal Perera", "email": "nimal@example.com"})
	request(t, app, http.MethodPost, "/api/customers", token, map[string]string{"name": "Kamal Silva", "email": "kamal@example.com"})

	_, body := request(t, app, http.MethodGet, "/api/customers?name=NIM", token, nil)
	var customers []models.Customer
	json.Unmarshal(body, &customers)
	if len(customers) != 1 || customers[0].Name != "Nimal Perera" {
		t.Errorf("search returned %+v, want only Nimal Perera", customers)
	}
}

func TestUpdateCustomerKeepsOwnEmail(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	request(t, app, http.MethodPost, "/api/customers", token, map[string]string{"name": "Nimal", "email": "nimal@example.com"})

	// Sending the customer's own email again must not count as a duplicate.
	status, body := request(t, app, http.MethodPut, "/api/customers/1", token, map[string]string{
		"name": "Nimal Perera", "email": "nimal@example.com",
	})
	if status != http.StatusOK {
		t.Errorf("got %d, want 200 (body: %s)", status, body)
	}
}
