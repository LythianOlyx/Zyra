package actions_test

import (
	"context"
	"testing"

	"github.com/LythianOlyx/Zyra/examples/saas-starter/actions"
	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

func TestCreateCheckoutSession_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	_, err := actions.CreateCheckoutSession(ctx, actions.CreateCheckoutSessionInput{
		PlanID:     "pro_monthly",
		SuccessURL: "http://localhost:8080/billing/success",
		CancelURL:  "http://localhost:8080/billing/cancel",
	})

	if err == nil {
		t.Fatal("expected unauthorized error for unauthenticated user, got nil")
	}

	actErr, ok := err.(*zyra.ActionError)
	if !ok || actErr.Code != zyra.ErrCodeUnauthorized {
		t.Fatalf("expected ErrCodeUnauthorized, got: %v", err)
	}
}

func TestCreateCheckoutSession_Authenticated(t *testing.T) {
	user := &zyra.User{
		ID:    "usr_test_123",
		Email: "test@example.com",
		Roles: []string{"owner"},
	}
	ctx := zyra.WithUserContext(context.Background(), user)

	resp, err := actions.CreateCheckoutSession(ctx, actions.CreateCheckoutSessionInput{
		PlanID:     "pro_monthly",
		SuccessURL: "http://localhost:8080/billing/success",
		CancelURL:  "http://localhost:8080/billing/cancel",
	})

	if err != nil {
		t.Fatalf("unexpected error creating checkout session: %v", err)
	}

	if resp.SessionID == "" || resp.URL == "" {
		t.Fatalf("expected session ID and URL, got %+v", resp)
	}
}

func TestProjectLifecycle(t *testing.T) {
	user := &zyra.User{
		ID:    "usr_builder_456",
		Email: "builder@example.com",
		Roles: []string{"owner"},
	}
	ctx := zyra.WithUserContext(context.Background(), user)

	// 1. Create Project
	proj, err := actions.CreateProject(ctx, actions.CreateProjectInput{
		Name: "Analytics Platform",
		Slug: "analytics-platform",
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	if proj.ID == "" || proj.OwnerID != user.ID {
		t.Fatalf("unexpected project state: %+v", proj)
	}

	// 2. List Projects
	projects, err := actions.ListProjects(ctx, actions.ListProjectsInput{})
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}
	if len(projects) == 0 {
		t.Fatal("expected at least 1 project in list, got 0")
	}
}
