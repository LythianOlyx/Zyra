//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// Product represents an item in the store.
type Product struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	PriceCents  int    `json:"priceCents"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
}

// CartItemInput represents an item sent from the client cart.
type CartItemInput struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// ValidateCartInput input payload for cart validation.
type ValidateCartInput struct {
	Items []CartItemInput `json:"items"`
}

// ValidateCartOutput returns calculated total and validated items.
type ValidateCartOutput struct {
	TotalCents int       `json:"totalCents"`
	Valid      bool      `json:"valid"`
	Items      []Product `json:"items"`
}

// CheckoutSessionInput input for starting a mock checkout session.
type CheckoutSessionInput struct {
	Items []CartItemInput `json:"items"`
}

// CheckoutSessionOutput returns a mock checkout redirect URL.
type CheckoutSessionOutput struct {
	CheckoutURL string `json:"checkoutUrl"`
}

// CreateProductInput holds data to add a new product in admin.
type CreateProductInput struct {
	Name        string `json:"name"`
	PriceCents  int    `json:"priceCents"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
}

var (
	catalogMu sync.RWMutex
	products  = make(map[string]*Product)
	prodSeq   int
)

func init() {
	SeedProduct(&Product{
		ID:          "prod_1",
		Slug:        "zyra-tshirt",
		Name:        "Zyra Developer T-Shirt",
		PriceCents:  2500,
		Description: "100% cotton premium developer tee with Zyra logo.",
		Stock:       50,
	})
	SeedProduct(&Product{
		ID:          "prod_2",
		Slug:        "zyra-sticker-pack",
		Name:        "Zyra Sticker Pack",
		PriceCents:  800,
		Description: "Die-cut vinyl stickers featuring Zero-CGO badge.",
		Stock:       100,
	})
}

// SeedProduct adds or replaces a product.
func SeedProduct(p *Product) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	products[p.ID] = p
}

// ListProducts returns all products in the catalog.
//
// +zyraaction
func ListProducts(ctx context.Context, input struct{}) ([]Product, error) {
	catalogMu.RLock()
	defer catalogMu.RUnlock()

	var result []Product
	for _, p := range products {
		result = append(result, *p)
	}
	return result, nil
}

// ValidateCart validates client cart items and calculates total server-side.
//
// +zyraaction
func ValidateCart(ctx context.Context, input ValidateCartInput) (ValidateCartOutput, error) {
	catalogMu.RLock()
	defer catalogMu.RUnlock()

	total := 0
	var validated []Product

	for _, item := range input.Items {
		p, ok := products[item.ProductID]
		if !ok || item.Quantity <= 0 {
			continue
		}
		if p.Stock < item.Quantity {
			return ValidateCartOutput{}, &zyra.ActionError{
				Code:    zyra.ErrCodeValidationFailed,
				Message: fmt.Sprintf("Insufficient stock for product %s", p.Name),
			}
		}
		total += p.PriceCents * item.Quantity
		validated = append(validated, *p)
	}

	return ValidateCartOutput{
		TotalCents: total,
		Valid:      true,
		Items:      validated,
	}, nil
}

// CreateCheckoutSession creates a mock Stripe Checkout session for cart items.
//
// +zyraaction
func CreateCheckoutSession(ctx context.Context, input CheckoutSessionInput) (CheckoutSessionOutput, error) {
	if len(input.Items) == 0 {
		return CheckoutSessionOutput{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Cart cannot be empty",
		}
	}

	val, err := ValidateCart(ctx, ValidateCartInput{Items: input.Items})
	if err != nil {
		return CheckoutSessionOutput{}, err
	}

	mockURL := fmt.Sprintf("/cart?checkout_success=1&total=%d", val.TotalCents)
	return CheckoutSessionOutput{CheckoutURL: mockURL}, nil
}

// CreateProduct adds a product (admin role required).
//
// +zyraaction
func CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	user, ok := zyra.UserFromContext(ctx)
	if !ok || user == nil {
		return Product{}, &zyra.ActionError{Code: zyra.ErrCodeUnauthorized, Message: "authentication required"}
	}
	isAdmin := false
	for _, r := range user.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return Product{}, &zyra.ActionError{Code: zyra.ErrCodeForbidden, Message: "admin role required"}
	}

	name := strings.TrimSpace(input.Name)
	if name == "" || input.PriceCents <= 0 {
		return Product{}, &zyra.ActionError{Code: zyra.ErrCodeValidationFailed, Message: "valid name and price are required"}
	}

	catalogMu.Lock()
	defer catalogMu.Unlock()

	prodSeq++
	id := fmt.Sprintf("prod_%d", prodSeq+10)
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	p := &Product{
		ID:          id,
		Slug:        slug,
		Name:        name,
		PriceCents:  input.PriceCents,
		Description: input.Description,
		Stock:       input.Stock,
	}
	products[id] = p
	return *p, nil
}
