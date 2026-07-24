//go:build zyratemplate

package actions

import (
	"context"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func TestCatalog_ListAndValidate(t *testing.T) {
	ctx := context.Background()

	prods, err := ListProducts(ctx, struct{}{})
	if err != nil || len(prods) < 2 {
		t.Fatalf("expected at least 2 seeded products, got %v, err %v", prods, err)
	}

	val, err := ValidateCart(ctx, ValidateCartInput{
		Items: []CartItemInput{
			{ProductID: "prod_1", Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if val.TotalCents != 5000 { // 2500 * 2
		t.Errorf("expected total 5000 cents, got %d", val.TotalCents)
	}
}

func TestCheckoutSession_MockRedirect(t *testing.T) {
	ctx := context.Background()

	res, err := CreateCheckoutSession(ctx, CheckoutSessionInput{
		Items: []CartItemInput{
			{ProductID: "prod_2", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected checkout error: %v", err)
	}
	if res.CheckoutURL == "" {
		t.Error("expected non-empty checkout URL")
	}
}

func TestCreateProduct_AdminRoleGated(t *testing.T) {
	ctx := context.Background()

	_, err := CreateProduct(ctx, CreateProductInput{Name: "Cap", PriceCents: 1500})
	if err == nil {
		t.Error("expected unauthorized error for anonymous user")
	}

	adminCtx := zyra.WithUserContext(ctx, &zyra.User{ID: "usr_admin", Roles: []string{"admin"}})
	prod, err := CreateProduct(adminCtx, CreateProductInput{Name: "Zyra Cap", PriceCents: 1500, Stock: 20})
	if err != nil {
		t.Fatalf("unexpected error creating product: %v", err)
	}
	if prod.Name != "Zyra Cap" {
		t.Errorf("unexpected product created: %+v", prod)
	}
}
