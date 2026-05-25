package pleasanter

import (
	"context"
	"fmt"

	"github.com/immmmmmmu/plsnt/internal/api"
)

// DeptService provides department operations.
type DeptService struct {
	client api.Client
}

// NewDeptService creates a new DeptService.
func NewDeptService(client api.Client) *DeptService {
	return &DeptService{client: client}
}

// List retrieves all departments.
func (s *DeptService) List(ctx context.Context) (map[string]any, error) {
	return s.client.PostRaw(ctx, "/api/depts/get", map[string]any{})
}

// Get retrieves a single department by ID.
func (s *DeptService) Get(ctx context.Context, deptID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/depts/%d/get", deptID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Create creates a new department.
// IMPORTANT: create is non-idempotent, so retry must be disabled.
func (s *DeptService) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return s.noRetryClient().PostRaw(ctx, "/api/depts/create", payload)
}

// Update updates an existing department.
func (s *DeptService) Update(ctx context.Context, deptID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/depts/%d/update", deptID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Delete deletes an existing department.
func (s *DeptService) Delete(ctx context.Context, deptID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/depts/%d/delete", deptID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// noRetryClient returns a client with retry disabled for non-idempotent operations.
func (s *DeptService) noRetryClient() api.Client {
	return s.client.NoRetry()
}
