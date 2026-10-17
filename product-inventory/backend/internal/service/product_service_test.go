package service

import (
	"context"
	"errors"
	"testing"

	"inventory/internal/model"
)

func intPtr(v int) *int { return &v }

func validInput() model.ProductInput {
	return model.ProductInput{
		Name: "Notebook", SKU: "nb-1", CategoryID: 1,
		Price: 45, Quantity: 10, Status: model.StatusActive,
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*model.ProductInput)
		wantField string // empty means the input must be valid
	}{
		{"valid input", func(in *model.ProductInput) {}, ""},
		{"zero quantity is allowed", func(in *model.ProductInput) { in.Quantity = 0 }, ""},
		{"empty status defaults to active", func(in *model.ProductInput) { in.Status = "" }, ""},
		{"price zero", func(in *model.ProductInput) { in.Price = 0 }, "price"},
		{"price negative", func(in *model.ProductInput) { in.Price = -5 }, "price"},
		{"quantity negative", func(in *model.ProductInput) { in.Quantity = -1 }, "quantity"},
		{"blank name", func(in *model.ProductInput) { in.Name = "   " }, "name"},
		{"blank sku", func(in *model.ProductInput) { in.SKU = "" }, "sku"},
		{"missing category", func(in *model.ProductInput) { in.CategoryID = 0 }, "category_id"},
		{"unknown status", func(in *model.ProductInput) { in.Status = "archived" }, "status"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mutate(&in)
			err := validate(&in)
			if tc.wantField == "" {
				if err != nil {
					t.Fatalf("expected valid input, got %v", err)
				}
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if _, ok := ve.Fields[tc.wantField]; !ok {
				t.Fatalf("expected error on %q, got %v", tc.wantField, ve.Fields)
			}
			if ve.Conflict {
				t.Fatal("validation errors must not be marked as conflicts")
			}
		})
	}
}

func TestValidateNormalisesInput(t *testing.T) {
	in := validInput()
	in.Name = "  Notebook  "
	in.SKU = "  nb-a4-001 "
	if err := validate(&in); err != nil {
		t.Fatal(err)
	}
	if in.Name != "Notebook" || in.SKU != "NB-A4-001" {
		t.Fatalf("got name=%q sku=%q", in.Name, in.SKU)
	}
}

// Stock validation runs before any database call, so a nil repository is safe here.
func TestUpdateStockValidation(t *testing.T) {
	svc := NewProductService(nil, 10)
	tests := []struct {
		name      string
		in        model.StockUpdate
		wantField string
	}{
		{"missing quantity", model.StockUpdate{Mode: "set"}, "quantity"},
		{"negative quantity", model.StockUpdate{Mode: "set", Quantity: intPtr(-5)}, "quantity"},
		{"missing mode", model.StockUpdate{Quantity: intPtr(5)}, "mode"},
		{"unknown mode", model.StockUpdate{Mode: "multiply", Quantity: intPtr(5)}, "mode"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.UpdateStock(context.Background(), 1, tc.in)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if _, ok := ve.Fields[tc.wantField]; !ok {
				t.Fatalf("expected error on %q, got %v", tc.wantField, ve.Fields)
			}
		})
	}
}

func TestLowStockFlag(t *testing.T) {
	svc := NewProductService(nil, 10)
	tests := []struct {
		qty  int
		want bool
	}{{0, true}, {8, true}, {10, true}, {11, false}, {150, false}}
	for _, tc := range tests {
		p := svc.withFlags(&model.Product{Quantity: tc.qty})
		if p.LowStock != tc.want {
			t.Errorf("quantity %d: low_stock=%v, want %v", tc.qty, p.LowStock, tc.want)
		}
	}
}
