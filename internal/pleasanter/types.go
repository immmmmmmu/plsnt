package pleasanter

import "encoding/json"

// APIRequest is the base request structure for all Pleasanter API calls.
type APIRequest struct {
	ApiVersion string `json:"ApiVersion,omitempty"`
	ApiKey     string `json:"ApiKey,omitempty"`
	Offset     int    `json:"Offset,omitempty"`
}

// APIResponse is the common response structure from Pleasanter API.
type APIResponse struct {
	StatusCode int          `json:"StatusCode"`
	Response   ResponseBody `json:"Response"`
}

// ResponseBody is the response body.
type ResponseBody struct {
	Offset     int      `json:"Offset"`
	PageSize   int      `json:"PageSize"`
	TotalCount int      `json:"TotalCount"`
	Data       []Record `json:"Data"`
}

// Record represents a Pleasanter record (API v1.1 nested structure).
type Record struct {
	ResultId        int64                  `json:"ResultId,omitempty"`
	IssueId         int64                  `json:"IssueId,omitempty"`
	SiteId          int64                  `json:"SiteId"`
	Title           string                 `json:"Title"`
	Body            string                 `json:"Body,omitempty"`
	Status          int                    `json:"Status,omitempty"`
	ClassHash       map[string]string      `json:"ClassHash"`
	NumHash         map[string]json.Number `json:"NumHash"`
	DateHash        map[string]string      `json:"DateHash"`
	DescriptionHash map[string]string      `json:"DescriptionHash"`
	CheckHash       map[string]bool        `json:"CheckHash"`
	AttachmentsHash map[string]any         `json:"AttachmentsHash"`
	CreatedTime     string                 `json:"CreatedTime,omitempty"`
	UpdatedTime     string                 `json:"UpdatedTime,omitempty"`
	Creator         int64                  `json:"Creator,omitempty"`
	Updator         int64                  `json:"Updator,omitempty"`
}

// SiteResponse is the getsite API response.
type SiteResponse struct {
	StatusCode int              `json:"StatusCode"`
	Response   SiteResponseBody `json:"Response"`
}

// SiteResponseBody is the site response body.
// Note: Pleasanter getsite returns Data as a single object, not an array.
type SiteResponseBody struct {
	Data SiteData `json:"Data"`
}

// SiteData is site detail data.
type SiteData struct {
	SiteId       int64        `json:"SiteId"`
	Title        string       `json:"Title"`
	Body         string       `json:"Body,omitempty"`
	SiteSettings SiteSettings `json:"SiteSettings"`
}

// SiteSettings is site configuration.
type SiteSettings struct {
	ReferenceType string      `json:"ReferenceType"`
	Columns       []ColumnDef `json:"Columns"`
}

// ColumnDef is a column definition in site settings.
type ColumnDef struct {
	ColumnName  string `json:"ColumnName"`
	LabelText   string `json:"LabelText"`
	TypeName    string `json:"TypeName,omitempty"`
	ChoicesText string `json:"ChoicesText,omitempty"`
	Required    bool   `json:"Required,omitempty"`
}

// SchemaInfo is the schema command output type.
type SchemaInfo struct {
	SiteID    int64          `json:"site_id"`
	Title     string         `json:"title"`
	TableType string         `json:"table_type"`
	Columns   []SchemaColumn `json:"columns"`
}

// SchemaColumn is a column in schema output.
type SchemaColumn struct {
	ColumnName string   `json:"column_name"`
	LabelText  string   `json:"label_text"`
	Type       string   `json:"type"`
	Choices    []string `json:"choices,omitempty"`
	Required   bool     `json:"required"`
}

// View is filter/sort parameters embedded in API requests.
type View struct {
	ColumnFilterHash       map[string]string `json:"ColumnFilterHash,omitempty"`
	ColumnFilterSearchTypes map[string]string `json:"ColumnFilterSearchTypes,omitempty"`
	ColumnSorterHash       map[string]string `json:"ColumnSorterHash,omitempty"`
}
