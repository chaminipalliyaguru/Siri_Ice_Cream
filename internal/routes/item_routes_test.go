package routes

import (
	"encoding/json"
	"net/http"
	"testing"

	"siri_ice_cream/internal/models"
)

func TestItemCRUD(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	// Create
	status, body := request(t, app, http.MethodPost, "/api/items", token, map[string]interface{}{
		"name": "Mint Cup", "category_id": 2, "price": 300, "quantity": 10,
	})
	if status != http.StatusCreated {
		t.Fatalf("create: got %d, body %s", status, body)
	}
	var created models.Item
	json.Unmarshal(body, &created)
	if created.ID == 0 || created.Name != "Mint Cup" || created.Category == nil || created.Category.Name != "Vanilla" {
		t.Fatalf("create: unexpected item %+v", created)
	}

	// Read one
	status, body = request(t, app, http.MethodGet, "/api/items/5", token, nil)
	if status != http.StatusOK {
		t.Fatalf("get: got %d, body %s", status, body)
	}

	// Update only the price (PATCH)
	status, body = request(t, app, http.MethodPatch, "/api/items/5", token, map[string]interface{}{"price": 320})
	var updated models.Item
	json.Unmarshal(body, &updated)
	if status != http.StatusOK || updated.Price != 320 || updated.Name != "Mint Cup" || updated.Quantity != 10 {
		t.Fatalf("update: got %d, item %+v", status, updated)
	}

	// Delete
	status, _ = request(t, app, http.MethodDelete, "/api/items/5", token, nil)
	if status != http.StatusOK {
		t.Fatalf("delete: got %d", status)
	}

	// Now it should be gone
	status, _ = request(t, app, http.MethodGet, "/api/items/5", token, nil)
	if status != http.StatusNotFound {
		t.Fatalf("get after delete: got %d, want 404", status)
	}
}

func TestItemSearchByName(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "cashier", "cashier123")

	status, body := request(t, app, http.MethodGet, "/api/items?name=CHOC", token, nil)
	if status != http.StatusOK {
		t.Fatalf("search: got %d", status)
	}
	var items []models.Item
	json.Unmarshal(body, &items)

	// The seed data has 2 chocolate items; search should ignore upper/lower case.
	if len(items) != 2 {
		t.Fatalf("search: got %d items, want 2", len(items))
	}
}

func TestItemValidation(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"missing name", map[string]interface{}{"category_id": 1, "price": 100}},
		{"zero price", map[string]interface{}{"name": "A", "category_id": 1, "price": 0}},
		{"negative quantity", map[string]interface{}{"name": "A", "category_id": 1, "price": 100, "quantity": -1}},
		{"unknown category", map[string]interface{}{"name": "A", "category_id": 99, "price": 100}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := request(t, app, http.MethodPost, "/api/items", token, tc.body)
			if status != http.StatusBadRequest {
				t.Errorf("got %d, want 400 (body: %s)", status, body)
			}
		})
	}
}

func TestItemNotFoundAndBadID(t *testing.T) {
	app := setupApp(t)
	token := login(t, app, "owner", "owner123")

	if status, _ := request(t, app, http.MethodGet, "/api/items/999", token, nil); status != http.StatusNotFound {
		t.Errorf("unknown id: got %d, want 404", status)
	}
	if status, _ := request(t, app, http.MethodGet, "/api/items/abc", token, nil); status != http.StatusBadRequest {
		t.Errorf("bad id: got %d, want 400", status)
	}
	if status, _ := request(t, app, http.MethodDelete, "/api/items/999", token, nil); status != http.StatusNotFound {
		t.Errorf("delete unknown id: got %d, want 404", status)
	}
}
