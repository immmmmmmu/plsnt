// Package sitediff provides structural diff between two Pleasanter SitePackage
// JSON files exported from the Web UI.
//
// The package operates entirely locally — no Pleasanter API calls — and is
// designed to absorb cosmetic noise (SiteId churn, array ordering, exporter
// metadata) so that a reviewer sees only meaningful structural changes.
package sitediff

// ChangeKind classifies a single FieldChange.
type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeRemoved  ChangeKind = "removed"
	ChangeModified ChangeKind = "modified"
	ChangeMoved    ChangeKind = "moved"
	ChangeTextDiff ChangeKind = "text_diff"
)

// Options controls Diff behaviour.
type Options struct {
	// IgnoreKeys are leaf key names ignored at any depth (e.g. "SiteId").
	IgnoreKeys []string
	// IgnorePaths are JSONPath-style globs starting with "/" (e.g. "/Sites/*/SiteSettings/Version").
	IgnorePaths []string
	// NoDefaultIgnore disables the built-in DefaultIgnoreKeys list.
	NoDefaultIgnore bool
	// MatchSitesBy controls Sites[] matching: "siteid" (default), "title", or "auto".
	MatchSitesBy string
	// Strict treats duplicate semantic keys (e.g. two Columns with same ColumnName) as errors.
	Strict bool
	// IncludePermissions enables comparison of the Permissions[] array (off by default).
	IncludePermissions bool
	// MaxBytes is the per-file input size limit. Zero means use the package default (50 MiB).
	MaxBytes int64
}

// DefaultIgnoreKeys are leaf keys ignored unless NoDefaultIgnore is set.
//
// Rationale:
//   - SiteId / TenantId        : churn between environments
//   - Creator / Updator         : actor metadata, not structure
//   - CreatedTime / UpdatedTime : timestamps, not structure
//   - HeaderInfo metadata       : exporter machine state, not the package itself
//
// ParentId is intentionally NOT in this list: it is required for Moved
// detection. Moved is computed before leaf-level ignore is applied.
var DefaultIgnoreKeys = []string{
	"SiteId",
	"TenantId",
	"Creator",
	"Updator",
	"CreatedTime",
	"UpdatedTime",
	"BaseSiteId",
	"PackageTime",
	"Server",
	"AssemblyVersion",
	"CreatorName",
}

// SiteRef identifies a site within either side of the diff.
type SiteRef struct {
	SiteID int64  `json:"site_id"`
	Title  string `json:"title"`
	// Path is the breadcrumb of titles from the package root, e.g. "ワークフローv2/部署マスタ".
	Path string `json:"path,omitempty"`
}

// SiteMove represents a site whose ParentId changed between old and new.
type SiteMove struct {
	SiteRef
	OldParentID int64 `json:"old_parent_id"`
	NewParentID int64 `json:"new_parent_id"`
}

// SitePatch represents the per-site list of FieldChanges.
type SitePatch struct {
	SiteRef
	Changes []FieldChange `json:"changes"`
}

// FieldChange is the leaf unit of a diff.
type FieldChange struct {
	// Path is a human-readable JSON-ish locator,
	// e.g. "SiteSettings.Columns[ColumnName=Status].LabelText".
	Path string `json:"path"`
	// Kind discriminates the rest of the fields.
	Kind ChangeKind `json:"kind"`
	// OldValue / NewValue are populated for Modified and TextDiff. Added has only NewValue, Removed only OldValue.
	OldValue any `json:"old_value,omitempty"`
	NewValue any `json:"new_value,omitempty"`
	// UnifiedDiff is set only when Kind == ChangeTextDiff.
	UnifiedDiff string `json:"unified_diff,omitempty"`
}

// SiteDiff is the top-level diff result.
type SiteDiff struct {
	HeaderChanges []FieldChange `json:"header_changes,omitempty"`
	Added         []SiteRef     `json:"added,omitempty"`
	Removed       []SiteRef     `json:"removed,omitempty"`
	Moved         []SiteMove    `json:"moved,omitempty"`
	Modified      []SitePatch   `json:"modified,omitempty"`
}

// HasChanges reports whether the diff contains any change at all.
// Used by --exit-code mode.
func (d *SiteDiff) HasChanges() bool {
	if d == nil {
		return false
	}
	if len(d.HeaderChanges) > 0 || len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Moved) > 0 {
		return true
	}
	for _, p := range d.Modified {
		if len(p.Changes) > 0 {
			return true
		}
	}
	return false
}
