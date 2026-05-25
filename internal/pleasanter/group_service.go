package pleasanter

import (
	"context"
	"fmt"

	"github.com/immmmmmmu/plsnt/internal/api"
)

// GroupService provides group operations.
type GroupService struct {
	client api.Client
}

// NewGroupService creates a new GroupService.
func NewGroupService(client api.Client) *GroupService {
	return &GroupService{client: client}
}

// List retrieves all groups.
func (s *GroupService) List(ctx context.Context) (map[string]any, error) {
	endpoint := "/api/groups/get"
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Get retrieves a single group by ID.
func (s *GroupService) Get(ctx context.Context, groupID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/groups/%d/get", groupID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Create creates a new group.
// IMPORTANT: create is non-idempotent, so retry must be disabled.
func (s *GroupService) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	endpoint := "/api/groups/create"
	return s.noRetryClient().PostRaw(ctx, endpoint, payload)
}

// Update updates an existing group.
func (s *GroupService) Update(ctx context.Context, groupID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/groups/%d/update", groupID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Delete deletes an existing group.
func (s *GroupService) Delete(ctx context.Context, groupID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/groups/%d/delete", groupID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// noRetryClient returns a client with retry disabled for non-idempotent operations.
func (s *GroupService) noRetryClient() api.Client {
	return s.client.NoRetry()
}
