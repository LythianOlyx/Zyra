//go:build zyratemplate

package actions

import (
	"context"
	"testing"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

func adminCtx() context.Context {
	return zyra.WithUserContext(context.Background(), &zyra.User{
		ID:    "usr_admin",
		Email: "admin@example.com",
		Roles: []string{"admin"},
	})
}

func memberCtx() context.Context {
	return zyra.WithUserContext(context.Background(), &zyra.User{
		ID:    "usr_member",
		Email: "member@example.com",
		Roles: []string{"member"},
	})
}

func TestListUsers_RequiresAdminRole(t *testing.T) {
	_, err := ListUsers(context.Background(), ListUsersInput{})
	if err == nil {
		t.Fatal("expected error for unauthenticated call")
	}

	_, err = ListUsers(memberCtx(), ListUsersInput{})
	if err == nil {
		t.Fatal("expected error for non-admin call")
	}
}

func TestListUsers_AdminCanListAndPaginate(t *testing.T) {
	res, err := ListUsers(adminCtx(), ListUsersInput{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total < 3 {
		t.Errorf("expected at least 3 seeded users, got %d", res.Total)
	}
}

func TestCreateUser_ValidationAndDuplicateCheck(t *testing.T) {
	ctx := adminCtx()

	_, err := CreateUser(ctx, CreateUserInput{Email: "invalid", Name: "Test"})
	if err == nil {
		t.Error("expected validation error for invalid email")
	}

	_, err = CreateUser(ctx, CreateUserInput{Email: "admin@example.com", Name: "Duplicate"})
	if err == nil {
		t.Error("expected duplicate email error")
	}

	created, err := CreateUser(ctx, CreateUserInput{Email: "newuser@example.com", Name: "New User", Role: "manager"})
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}
	if created.ID == "" || created.Email != "newuser@example.com" {
		t.Errorf("unexpected created user: %+v", created)
	}
}
