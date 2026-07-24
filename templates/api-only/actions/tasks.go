//go:build zyratemplate

package actions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LythianOlyx/Zyra/pkg/zyra"
)

// TaskItem represents an API task entity.
type TaskItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending" | "completed"
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateTaskInput holds payload for creating a task.
type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskInput holds payload for updating task status.
type UpdateTaskInput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// DeleteTaskInput holds payload for deleting a task.
type DeleteTaskInput struct {
	ID string `json:"id"`
}

var (
	tasksMu  sync.RWMutex
	taskMap  = make(map[string]*TaskItem)
	taskSeq  int
)

func init() {
	SeedTask(&TaskItem{
		ID:          "task_1",
		Title:       "Build Native Mobile Frontend",
		Description: "Consume Zyra Headless API via OpenAPI client generator",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	})
}

// SeedTask adds a task record into memory.
func SeedTask(t *TaskItem) {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	taskMap[t.ID] = t
}

// ListTasks returns all task items.
//
// +zyraaction
func ListTasks(ctx context.Context, input struct{}) ([]TaskItem, error) {
	tasksMu.RLock()
	defer tasksMu.RUnlock()

	var result []TaskItem
	for _, t := range taskMap {
		result = append(result, *t)
	}
	return result, nil
}

// CreateTask creates a new task (Requires authentication).
//
// +zyraaction
func CreateTask(ctx context.Context, input CreateTaskInput) (TaskItem, error) {
	_, ok := zyra.UserFromContext(ctx)
	if !ok {
		return TaskItem{}, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Authentication required (Bearer token)",
		}
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return TaskItem{}, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Title cannot be empty",
		}
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	taskSeq++
	id := fmt.Sprintf("task_%d", taskSeq+10)
	t := &TaskItem{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	taskMap[id] = t
	return *t, nil
}

// UpdateTaskStatus updates task status.
//
// +zyraaction
func UpdateTaskStatus(ctx context.Context, input UpdateTaskInput) (TaskItem, error) {
	tasksMu.Lock()
	defer tasksMu.Unlock()

	t, ok := taskMap[input.ID]
	if !ok {
		return TaskItem{}, &zyra.ActionError{
			Code:    zyra.ErrCodeNotFound,
			Message: "Task not found",
		}
	}

	if input.Status != "" {
		t.Status = input.Status
	}
	return *t, nil
}

// DeleteTask removes a task record.
//
// +zyraaction
func DeleteTask(ctx context.Context, input DeleteTaskInput) (bool, error) {
	tasksMu.Lock()
	defer tasksMu.Unlock()

	if _, ok := taskMap[input.ID]; !ok {
		return false, &zyra.ActionError{
			Code:    zyra.ErrCodeNotFound,
			Message: "Task not found",
		}
	}

	delete(taskMap, input.ID)
	return true, nil
}
