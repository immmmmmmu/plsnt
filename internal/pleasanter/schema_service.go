package pleasanter

import (
	"context"
	"fmt"
	"strings"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

// SchemaService provides schema operations.
type SchemaService struct {
	client api.Client
}

// NewSchemaService creates a new SchemaService.
func NewSchemaService(client api.Client) *SchemaService {
	return &SchemaService{client: client}
}

// GetSchema retrieves column definitions for a site.
func (s *SchemaService) GetSchema(ctx context.Context, siteID int64) (*SchemaInfo, error) {
	endpoint := fmt.Sprintf("/api/items/%d/getsite", siteID)

	var resp SiteResponse
	if err := s.client.Post(ctx, endpoint, &struct{}{}, &resp); err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, errs.New(errs.CodeSiteNotFound,
			fmt.Sprintf("site %d not found", siteID)).
			WithSuggestion("Check the site ID and ensure your API key has access")
	}
	if resp.StatusCode != 200 {
		return nil, errs.New(errs.CodeServerError,
			fmt.Sprintf("getsite returned status %d", resp.StatusCode))
	}

	site := resp.Response.Data
	if site.SiteId == 0 {
		return nil, errs.New(errs.CodeSiteNotFound,
			fmt.Sprintf("site %d returned no data", siteID))
	}
	info := &SchemaInfo{
		SiteID:    site.SiteId,
		Title:     site.Title,
		TableType: site.SiteSettings.ReferenceType,
	}

	for _, col := range site.SiteSettings.Columns {
		sc := SchemaColumn{
			ColumnName: col.ColumnName,
			LabelText:  col.LabelText,
			Type:       classifyColumnType(col.ColumnName),
			Required:   col.Required,
		}
		if col.ChoicesText != "" {
			sc.Choices = parseChoices(col.ChoicesText)
		}
		info.Columns = append(info.Columns, sc)
	}

	return info, nil
}

func classifyColumnType(name string) string {
	switch {
	case strings.HasPrefix(name, "Class"):
		return "Class"
	case strings.HasPrefix(name, "Num"):
		return "Num"
	case strings.HasPrefix(name, "Date"):
		return "Date"
	case strings.HasPrefix(name, "Description"):
		return "Description"
	case strings.HasPrefix(name, "Check"):
		return "Check"
	case strings.HasPrefix(name, "Attachments"):
		return "Attachments"
	default:
		return "Standard"
	}
}

func parseChoices(text string) []string {
	lines := strings.Split(text, "\n")
	choices := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			choices = append(choices, trimmed)
		}
	}
	return choices
}
