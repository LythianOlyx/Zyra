package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zyra-framework/zyra/pkg/zyra"
)

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	projectsMu sync.RWMutex
	projectsDB = make(map[string]*Project)
)

// +zyraaction
type CreateProjectInput struct {
	Name string `json:"name" validate:"required,min=3,max=50"`
	Slug string `json:"slug" validate:"required,slug"`
}

// +zyraaction
func CreateProject(ctx context.Context, input CreateProjectInput) (*Project, error) {
	user, ok := zyra.UserFromContext(ctx)
	if !ok {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Authentication required to create a project",
		}
	}

	if len(input.Name) < 3 {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeValidationFailed,
			Message: "Project name must be at least 3 characters long",
		}
	}

	projectsMu.Lock()
	defer projectsMu.Unlock()

	id := fmt.Sprintf("proj_%d", time.Now().UnixNano())
	proj := &Project{
		ID:        id,
		Name:      input.Name,
		Slug:      input.Slug,
		OwnerID:   user.ID,
		CreatedAt: time.Now(),
	}
	projectsDB[id] = proj

	// Broadcast realtime stream event
	zyra.Broadcast("projects", map[string]any{
		"event":     "project.created",
		"projectId": id,
		"name":      proj.Name,
		"ownerId":   user.ID,
	})

	return proj, nil
}

// +zyraaction
type ListProjectsInput struct{}

// +zyraaction
func ListProjects(ctx context.Context, _ ListProjectsInput) ([]*Project, error) {
	user, ok := zyra.UserFromContext(ctx)
	if !ok {
		return nil, &zyra.ActionError{
			Code:    zyra.ErrCodeUnauthorized,
			Message: "Authentication required to list projects",
		}
	}

	projectsMu.RLock()
	defer projectsMu.RUnlock()

	var userProjects []*Project
	for _, p := range projectsDB {
		if p.OwnerID == user.ID {
			userProjects = append(userProjects, p)
		}
	}
	return userProjects, nil
}
