package actions

import (
	"context"
	"testing"
)

func TestMVTTaskActions_CRUD(t *testing.T) {
	if err := InitDB(":memory:"); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if GetDB() == nil {
		t.Fatal("GetDB returned nil")
	}

	ctx := context.Background()

	// 1. Validation check
	_, err := CreateTask(ctx, CreateTaskInput{Title: "", Description: "empty title"})
	if err == nil {
		t.Error("expected error for empty task title")
	}

	// 2. Create Task
	created, err := CreateTask(ctx, CreateTaskInput{Title: "Unit Test Task", Description: "Testing task action CRUD"})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if created.ID <= 0 || created.Status != "pending" {
		t.Errorf("unexpected created task: %+v", created)
	}

	// 3. List Tasks
	list, err := ListTasks(ctx, struct{}{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Errorf("unexpected task list: %+v", list)
	}

	// 4. Update Task Status
	updated, err := UpdateTaskStatus(ctx, UpdateTaskInput{ID: created.ID, Status: "completed"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}
	if updated.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", updated.Status)
	}

	// 5. Delete Task
	delRes, err := DeleteTask(ctx, DeleteTaskInput{ID: created.ID})
	if err != nil || !delRes.Success {
		t.Fatalf("DeleteTask failed: %v, res: %+v", err, delRes)
	}

	// 6. List Tasks after deletion
	afterDel, err := ListTasks(ctx, struct{}{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(afterDel) != 0 {
		t.Errorf("expected empty list after deletion, got %d items", len(afterDel))
	}
}
