package master

// MasterResult holds the result of a master data import operation.
type MasterResult struct {
	SiteID  int64 `json:"site_id"`
	Added   int   `json:"added"`
	Updated int   `json:"updated"`
	Errors  int   `json:"errors"`
}
