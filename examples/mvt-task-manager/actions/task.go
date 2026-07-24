package actions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Task represents a task entity.
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "in_progress", "completed"
	CreatedAt   string `json:"createdAt"`
}

// CreateTaskInput describes input for creating a task.
type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskInput describes input for updating task status.
type UpdateTaskInput struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// DeleteTaskInput describes input for deleting a task.
type DeleteTaskInput struct {
	ID int `json:"id"`
}

// DeleteTaskResult describes response for task deletion.
type DeleteTaskResult struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

var dbConn *sql.DB

// InitDB initializes the SQLite database for MVT Task Manager.
func InitDB(dbPath string) error {
	var err error
	dbConn, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open sqlite database: %w", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = dbConn.Exec(createTableQuery)
	return err
}

// GetDB returns the initialized SQLite database connection.
func GetDB() *sql.DB {
	return dbConn
}

// +zyraaction
func CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	if input.Title == "" {
		return Task{}, fmt.Errorf("task title cannot be empty")
	}

	now := time.Now().Format(time.RFC3339)
	res, err := dbConn.ExecContext(ctx, "INSERT INTO tasks (title, description, status, created_at) VALUES (?, ?, 'pending', ?)", input.Title, input.Description, now)
	if err != nil {
		return Task{}, fmt.Errorf("failed to insert task: %w", err)
	}

	id, _ := res.LastInsertId()
	return Task{
		ID:          int(id),
		Title:       input.Title,
		Description: input.Description,
		Status:      "pending",
		CreatedAt:   now,
	}, nil
}

// +zyraaction
func ListTasks(ctx context.Context, input struct{}) ([]Task, error) {
	rows, err := dbConn.QueryContext(ctx, "SELECT id, title, description, status, created_at FROM tasks ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, nil
}

// +zyraaction
func UpdateTaskStatus(ctx context.Context, input UpdateTaskInput) (Task, error) {
	_, err := dbConn.ExecContext(ctx, "UPDATE tasks SET status = ? WHERE id = ?", input.Status, input.ID)
	if err != nil {
		return Task{}, fmt.Errorf("failed to update task status: %w", err)
	}

	var t Task
	err = dbConn.QueryRowContext(ctx, "SELECT id, title, description, status, created_at FROM tasks WHERE id = ?", input.ID).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("task not found after update: %w", err)
	}
	return t, nil
}

// +zyraaction
func DeleteTask(ctx context.Context, input DeleteTaskInput) (DeleteTaskResult, error) {
	res, err := dbConn.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", input.ID)
	if err != nil {
		return DeleteTaskResult{Success: false, ID: input.ID}, fmt.Errorf("failed to delete task: %w", err)
	}
	affected, _ := res.RowsAffected()
	return DeleteTaskResult{Success: affected > 0, ID: input.ID}, nil
}
