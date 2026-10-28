package routes

import (
	"encoding/json"
	"net/http"
	"testing"

	"siri_ice_cream/internal/models"
)

func TestCustomerCRUD(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	// Create
	status, body := request(t, app, http.MethodPost, "/api/customers", token, map[string]string{
		"name": "Nimal Perera", "email": "Nimal@Example.com", "phone": "0771234567",
	})
	if status != http.StatusCreated {
		t.Fatalf("create: got %d, body %s", status, body)
	}
	var created models.Customer
	json.Unmarshal(body, &created)
	if created.Email != "nimal@example.com" {
		t.Errorf("email should be saved in lower case, got %q", created.Email)
	}

	// List
	status, body = request(t, app, http.MethodGet, "/api/customers", token, nil)
	var list []models.Customer
	json.Unmarshal(body, &list)
	if status != http.StatusOK || len(list) != 1 {
		t.Fatalf("list: got %d with %d customers", status, len(list))
	}

	// Update (PUT)
	status, body = request(t, app, http.MethodPut, "/api/customers/1", token, map[string]string{"name": "Nimal P."})
	var updated models.Customer
	json.Unmarshal(body, &updated)
	if status != http.StatusOK || updated.Name != "Nimal P." || updated.Phone != "0771234567" {
		t.Fatalf("update: got %d, customer %+v", status, updated)
	}

	// Delete
	if status, _ := request(t, app, http.MethodDelete, "/api/customers/1", token, nil); status != http.StatusOK {
		t.Fatalf("delete: got %d", status)
	}
	if status, _ := request(t, app, http.MethodGet, "/api/customers/1", token, nil); status != http.StatusNotFound {
		t.Fatalf("get after delete: got %d, want 404", status)
	}
}

func TestCustomerValidation(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "manager", "manager123")

	request(t, app, http.MethodPost, "/api/customers", token, map[string]string{
		"name": "First", "email": "same@example.com",
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing name", map[string]string{"email": "a@example.com"}, http.StatusBadRequest},
		{"missing email", map[string]string{"name": "A"}, http.StatusBadRequest},
		{"invalid email", map[string]string{"name": "A", "email": "not-an-email"}, http.StatusBadRequest},
		{"duplicate email", map[string]string{"name": "B", "email": "SAME@example.com"}, http.StatusConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := request(t, app, http.MethodPost, "/api/customers", token, tc.body)
			if status != tc.wantStatus {
				t.Errorf("got %d, want %d (body: %s)", status, tc.wantStatus, body)
			}
		})
	}
}
