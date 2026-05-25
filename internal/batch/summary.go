package batch

import (
	"fmt"
	"io"
	"strings"
)

// SiteEntry represents a created site in the scaffold summary.
type SiteEntry struct {
	StepName      string
	SiteID        string
	Title         string
	ReferenceType string
}

// ScaffoldSummary holds the summary of a scaffold batch execution.
type ScaffoldSummary struct {
	TemplateName string
	ParentID     string
	Sites        []SiteEntry
}

// CollectSummary builds a ScaffoldSummary from engine step outputs.
// It identifies site create steps by checking for "Id" key
// in the captured output.
func CollectSummary(templateName string, stepOutputs map[string]map[string]string) *ScaffoldSummary {
	summary := &ScaffoldSummary{
		TemplateName: templateName,
	}

	for stepName, values := range stepOutputs {
		id, hasID := values["Id"]
		if !hasID {
			continue
		}

		entry := SiteEntry{
			StepName:      stepName,
			SiteID:        id,
			Title:         values["Title"],
			ReferenceType: values["ReferenceType"],
		}
		summary.Sites = append(summary.Sites, entry)
	}

	return summary
}

// WriteTo writes the summary to the given writer in human-readable format.
func (s *ScaffoldSummary) WriteTo(w io.Writer) (int64, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "\n=== Scaffold Summary ===\n")
	fmt.Fprintf(&b, "Template: %s\n", s.TemplateName)
	if s.ParentID != "" {
		fmt.Fprintf(&b, "Folder: %s\n", s.ParentID)
	}
	fmt.Fprintf(&b, "\nSites created:\n")

	for _, site := range s.Sites {
		fmt.Fprintf(&b, "  %-20s (SiteID: %s) %s\n",
			site.Title, site.SiteID, site.ReferenceType)
	}

	n, err := fmt.Fprint(w, b.String())
	return int64(n), err
}
