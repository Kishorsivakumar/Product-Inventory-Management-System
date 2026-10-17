package service

import (
	"context"
	"errors"
	"strings"

	"inventory/internal/model"
	"inventory/internal/repository"
)

// ValidationError carries per-field messages for the client to display.
// Conflict marks uniqueness violations (HTTP 409 instead of 422).
type ValidationError struct {
	Fields   map[string]string
	Conflict bool
}

func (e *ValidationError) Error() string { return "validation failed" }

type ProductService struct {
	repo              *repository.ProductRepository
	lowStockThreshold int
}

func NewProductService(repo *repository.ProductRepository, threshold int) *ProductService {
	return &ProductService{repo: repo, lowStockThreshold: threshold}
}

func (s *ProductService) withFlags(p *model.Product) *model.Product {
	p.LowStock = p.Quantity <= s.lowStockThreshold
	return p
}

func validate(in *model.ProductInput) error {
	f := map[string]string{}
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	if in.Status == "" {
		in.Status = model.StatusActive
	}
	if in.Name == "" {
		f["name"] = "Name is required"
	}
	if in.SKU == "" {
		f["sku"] = "SKU is required"
	}
	if in.CategoryID <= 0 {
		f["category_id"] = "Category is required"
	}
	if in.Price <= 0 {
		f["price"] = "Price must be greater than zero"
	}
	if in.Quantity < 0 {
		f["quantity"] = "Quantity cannot be negative"
	}
	if !model.ValidStatus(in.Status) {
		f["status"] = "Status must be active, inactive or discontinued"
	}
	if len(f) > 0 {
		return &ValidationError{Fields: f}
	}
	return nil
}

func (s *ProductService) List(ctx context.Context, f model.ListFilter) ([]model.Product, error) {
	f.Search = strings.TrimSpace(f.Search)
	products, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	for i := range products {
		s.withFlags(&products[i])
	}
	return products, nil
}

func (s *ProductService) Get(ctx context.Context, id int64) (*model.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.withFlags(p), nil
}

func (s *ProductService) Create(ctx context.Context, in model.ProductInput) (*model.Product, error) {
	if err := validate(&in); err != nil {
		return nil, err
	}
	p, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, wrap(err)
	}
	return s.withFlags(p), nil
}

func (s *ProductService) Update(ctx context.Context, id int64, in model.ProductInput) (*model.Product, error) {
	if err := validate(&in); err != nil {
		return nil, err
	}
	p, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return nil, wrap(err)
	}
	return s.withFlags(p), nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) UpdateStock(ctx context.Context, id int64, u model.StockUpdate) (*model.Product, error) {
	f := map[string]string{}
	if u.Quantity == nil {
		f["quantity"] = "Quantity is required"
	} else if *u.Quantity < 0 {
		f["quantity"] = "Quantity cannot be negative"
	}
	if u.Mode != "set" && u.Mode != "add" && u.Mode != "remove" {
		f["mode"] = "Mode must be set, add or remove"
	}
	if len(f) > 0 {
		return nil, &ValidationError{Fields: f}
	}

	qty := *u.Quantity
	var p *model.Product
	var err error
	switch u.Mode {
	case "set":
		p, err = s.repo.SetQuantity(ctx, id, qty)
	case "add":
		p, err = s.repo.AdjustQuantity(ctx, id, qty)
	case "remove":
		p, err = s.repo.AdjustQuantity(ctx, id, -qty)
	}
	if err != nil {
		return nil, wrap(err)
	}
	return s.withFlags(p), nil
}

// wrap turns repository conflicts into field-level validation errors.
func wrap(err error) error {
	switch {
	case errors.Is(err, repository.ErrDuplicateSKU):
		return &ValidationError{Fields: map[string]string{"sku": "SKU already exists"}, Conflict: true}
	case errors.Is(err, repository.ErrCategoryNotFound):
		return &ValidationError{Fields: map[string]string{"category_id": "Category does not exist"}}
	}
	return err
}
