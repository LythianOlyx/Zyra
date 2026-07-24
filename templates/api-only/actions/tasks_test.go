//go:build zyratemplate

package actions

import (
	"context"
	"testing"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

func TestTasks_CRUD(t *testing.T) {
	ctx := context.Background()

	// 1. List
	items, err := ListTasks(ctx, struct{}{})
	if err != nil || len(items) == 0 {
		t.Fatalf("expected seeded tasks, got %v, err %v", items, err)
	}

	// 2. Unauthenticated Create fails
	_, err = CreateTask(ctx, CreateTaskInput{Title: "Unauth Task"})
	if err == nil {
		t.Error("expected error for unauthenticated CreateTask")
	}

	// 3. Authenticated Create succeeds
	authCtx := zyra.WithUserContext(ctx, &zyra.User{ID: "usr_api_client"})
	created, err := CreateTask(authCtx, CreateTaskInput{Title: "Auth API Task", Description: "Via Bearer token"})
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}
	if created.Title != "Auth API Task" {
		t.Errorf("unexpected created task: %+v", created)
	}

	// 4. Update
	updated, err := UpdateTaskStatus(ctx, UpdateTaskInput{ID: created.ID, Status: "completed"})
	if err != nil || updated.Status != "completed" {
		t.Errorf("expected updated status 'completed', got %+v, err %v", updated, err)
	}

	// 5. Delete
	deleted, err := DeleteTask(ctx, DeleteTaskInput{ID: created.ID})
	if err != nil || !deleted {
		t.Errorf("expected delete to succeed, got %v, err %v", deleted, err)
	}
}
