package pleasanter

import (
	"context"
	"fmt"

	"github.com/immmmmmmu/plsnt/internal/api"
)

// UserService provides user operations.
type UserService struct {
	client api.Client
}

// NewUserService creates a new UserService.
func NewUserService(client api.Client) *UserService {
	return &UserService{client: client}
}

// List retrieves all users.
func (s *UserService) List(ctx context.Context) (map[string]any, error) {
	return s.client.PostRaw(ctx, "/api/users/get", map[string]any{})
}

// Get retrieves a single user by ID.
func (s *UserService) Get(ctx context.Context, userID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/users/%d/get", userID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Create creates a new user.
// IMPORTANT: create is non-idempotent, so retry must be disabled.
func (s *UserService) Create(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return s.noRetryClient().PostRaw(ctx, "/api/users/create", payload)
}

// Update updates an existing user.
func (s *UserService) Update(ctx context.Context, userID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/users/%d/update", userID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Delete deletes an existing user.
func (s *UserService) Delete(ctx context.Context, userID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/users/%d/delete", userID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// BulkCreate creates multiple users sequentially.
// It returns results collected so far and the first error encountered.
func (s *UserService) BulkCreate(ctx context.Context, users []map[string]any) ([]map[string]any, error) {
	var results []map[string]any
	for _, user := range users {
		result, err := s.Create(ctx, user)
		if err != nil {
			return results, fmt.Errorf("failed to create user %v: %w", user["LoginId"], err)
		}
		results = append(results, result)
	}
	return results, nil
}

// noRetryClient returns a client with retry disabled for non-idempotent operations.
func (s *UserService) noRetryClient() api.Client {
	return s.client.NoRetry()
}
