package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Widget", Price: 9.99})
	got, err := s.GetByID(p.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Name != "Widget" {
		t.Errorf("expected 'Widget', got '%s'", got.Name)
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestGetByIDNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetByID(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound for missing product")
	}
}

func TestUpdate(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Widget", Price: 9.99})
	updated, err := s.Update(p.ID, model.Product{Name: "Updated", Price: 19.99})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected 'Updated', got '%s'", updated.Name)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Update(999, model.Product{Name: "X", Price: 1})
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when updating non-existent product")
	}
}

func TestDeleteExisting(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Widget", Price: 9.99})
	if err := s.Delete(p.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err := s.GetByID(p.ID)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound after deletion")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}
