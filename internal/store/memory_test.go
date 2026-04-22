package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	tests := []struct {
		name  string
		input model.Product
	}{
		{"simple product", model.Product{Name: "Leberkassemmal", Price: 4.20}},
		{"zero price", model.Product{Name: "Schundblatt Heute", Price: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemoryStore()
			created := s.Create(tt.input)

			got, err := s.GetByID(created.ID)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got.Name != tt.input.Name {
				t.Errorf("expected name %q, got %q", tt.input.Name, got.Name)
			}
			if got.Price != tt.input.Price {
				t.Errorf("expected price %v, got %v", tt.input.Price, got.Price)
			}
		})
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}

func TestUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Alt", Price: 1.00})

	updated, err := s.Update(created.ID, model.Product{Name: "Neu", Price: 9.99})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "Neu" {
		t.Errorf("expected name %q, got %q", "Neu", updated.Name)
	}
}

func TestDeleteProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Wegwerf", Price: 5.00})

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("expected no error on delete, got %v", err)
	}
	if _, err := s.GetByID(created.ID); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetByID(9999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
