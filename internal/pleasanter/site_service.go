package pleasanter

import (
	"context"
	"fmt"

	"github.com/immmmmmmu/plsnt/internal/api"
)

// SiteService provides site operations.
type SiteService struct {
	client api.Client
}

// NewSiteService creates a new SiteService.
func NewSiteService(client api.Client) *SiteService {
	return &SiteService{client: client}
}

// Get retrieves site metadata using /api/items/{siteId}/getsite.
func (s *SiteService) Get(ctx context.Context, siteID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/getsite", siteID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Create creates a new site under the given parent.
// IMPORTANT: create is non-idempotent, so retry must be disabled.
// Endpoint: /api/items/{parentId}/createsite
func (s *SiteService) Create(ctx context.Context, parentID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/createsite", parentID)
	return s.noRetryClient().PostRaw(ctx, endpoint, payload)
}

// Update updates site settings.
// Endpoint: /api/items/{siteId}/updatesite
func (s *SiteService) Update(ctx context.Context, siteID int64, payload map[string]any) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/updatesite", siteID)
	return s.client.PostRaw(ctx, endpoint, payload)
}

// Delete deletes a site.
// Endpoint: /api/items/{siteId}/deletesite
func (s *SiteService) Delete(ctx context.Context, siteID int64) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/deletesite", siteID)
	return s.client.PostRaw(ctx, endpoint, map[string]any{})
}

// Copy retrieves a source site's settings and creates a new site under destParentID.
// overrides allows the caller to override fields (e.g. Title) before creating.
// The create step uses noRetryClient because it is non-idempotent.
func (s *SiteService) Copy(ctx context.Context, sourceSiteID, destParentID int64, overrides map[string]any) (map[string]any, error) {
	// Step 1: Get source site settings via getsite (not get)
	getEndpoint := fmt.Sprintf("/api/items/%d/getsite", sourceSiteID)
	source, err := s.client.PostRaw(ctx, getEndpoint, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to get source site %d: %w", sourceSiteID, err)
	}

	// Step 2: Extract Response.Data from the source (getsite returns Data as an object)
	response, ok := source["Response"]
	if !ok {
		return nil, fmt.Errorf("source site %d returned no Response field", sourceSiteID)
	}
	responseMap, ok := response.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("source site %d Response field is not an object", sourceSiteID)
	}
	dataMap, ok := responseMap["Data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("source site %d Response.Data is not an object", sourceSiteID)
	}

	// Step 3: Build create payload from source site settings
	payload := make(map[string]any)
	// Copy relevant fields from the site data
	for _, key := range []string{"Title", "Body", "SiteSettings", "ReferenceType"} {
		if v, exists := dataMap[key]; exists {
			payload[key] = v
		}
	}

	// Step 4: Apply overrides
	for k, v := range overrides {
		payload[k] = v
	}

	// Step 5: Create site under destination parent (non-idempotent)
	createEndpoint := fmt.Sprintf("/api/items/%d/createsite", destParentID)
	return s.noRetryClient().PostRaw(ctx, createEndpoint, payload)
}

// Search searches child sites under parentSiteID by title keyword.
// Endpoint: /api/items/{parentSiteID}/get with View.ColumnFilterHash.
func (s *SiteService) Search(ctx context.Context, parentSiteID int64, keyword string) (map[string]any, error) {
	endpoint := fmt.Sprintf("/api/items/%d/get", parentSiteID)
	payload := map[string]any{
		"View": map[string]any{
			"ColumnFilterHash": map[string]any{
				"Title": keyword,
			},
		},
	}
	return s.client.PostRaw(ctx, endpoint, payload)
}

// noRetryClient returns a client with retry disabled for non-idempotent operations.
func (s *SiteService) noRetryClient() api.Client {
	return s.client.NoRetry()
}
