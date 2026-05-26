package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mrckurz/CI-CD-MCM/internal/model"
	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

func setupRouter() (*mux.Router, *Handler) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r, h
}

func TestHealthEndpoint(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestGetProductsEmpty(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/products", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCreateAndGetProduct(t *testing.T) {
	r, _ := setupRouter()

	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	req = httptest.NewRequest("GET", "/products/1", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestGetProductNotFound(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/products/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateProduct(t *testing.T) {
	h := NewHandler(store.NewMemoryStore())
	r := mux.NewRouter()
	h.RegisterRoutes(r)

	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", rec.Code)
	}

	updateBody := `{"name":"Updated Widget","price":19.99}`
	req2 := httptest.NewRequest(http.MethodPut, "/products/1", strings.NewReader(updateBody))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec2.Code)
	}
	var updated model.Product
	json.NewDecoder(rec2.Body).Decode(&updated)
	if updated.Name != "Updated Widget" {
		t.Errorf("expected 'Updated Widget', got '%s'", updated.Name)
	}
}

func TestDeleteProduct(t *testing.T) {
	h := NewHandler(store.NewMemoryStore())
	r := mux.NewRouter()
	h.RegisterRoutes(r)

	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected 200 on delete, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec3.Code)
	}
}

func TestCreateInvalidProduct(t *testing.T) {
	h := NewHandler(store.NewMemoryStore())
	r := mux.NewRouter()
	h.RegisterRoutes(r)

	body := `{"name":"","price":9.99}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", rec.Code)
	}
}

func TestCreateBadJSON(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", rr.Code)
	}
}

func TestUpdateProductNotFound(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest(http.MethodPut, "/products/999", strings.NewReader(`{"name":"X","price":1}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateBadJSON(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest(http.MethodPut, "/products/1", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", rr.Code)
	}
}

func TestDeleteProductNotFound(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest(http.MethodDelete, "/products/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
