package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportCmd_Flags(t *testing.T) {
	cmd := newExportCmd()

	// Verify all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("header-site-id"))
	assert.NotNil(t, cmd.Flags().Lookup("detail-site-id"))
	assert.NotNil(t, cmd.Flags().Lookup("dept-site-id"))
	assert.NotNil(t, cmd.Flags().Lookup("type-site-id"))
	assert.NotNil(t, cmd.Flags().Lookup("from"))
	assert.NotNil(t, cmd.Flags().Lookup("to"))
	assert.NotNil(t, cmd.Flags().Lookup("status"))
	assert.NotNil(t, cmd.Flags().Lookup("bom"))
}

func TestExportCmd_RequiredFlags_HeaderSiteID(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export", "--detail-site-id", "200", "--from", "2026-01-01", "--to", "2026-03-31"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header-site-id")
}

func TestExportCmd_RequiredFlags_DetailSiteID(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export", "--header-site-id", "100", "--from", "2026-01-01", "--to", "2026-03-31"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detail-site-id")
}

func TestExportCmd_RequiredFlags_From(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export", "--header-site-id", "100", "--detail-site-id", "200", "--to", "2026-03-31"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "from")
}

func TestExportCmd_RequiredFlags_To(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export", "--header-site-id", "100", "--detail-site-id", "200", "--from", "2026-01-01"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to")
}

func TestExportCmd_DateValidation_InvalidFormat(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export",
		"--header-site-id", "100",
		"--detail-site-id", "200",
		"--from", "2026/01/01",
		"--to", "2026-03-31",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestExportCmd_DateValidation_FromAfterTo(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export",
		"--header-site-id", "100",
		"--detail-site-id", "200",
		"--from", "2026-04-01",
		"--to", "2026-03-01",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--from")
}

func TestExportCmd_DateValidation_InvalidToFormat(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export",
		"--header-site-id", "100",
		"--detail-site-id", "200",
		"--from", "2026-01-01",
		"--to", "not-a-date",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestExportCmd_SiteIDValidation_Zero(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export",
		"--header-site-id", "0",
		"--detail-site-id", "200",
		"--from", "2026-01-01",
		"--to", "2026-03-31",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header-site-id")
}

func TestExportCmd_SiteIDValidation_DetailZero(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"export",
		"--header-site-id", "100",
		"--detail-site-id", "0",
		"--from", "2026-01-01",
		"--to", "2026-03-31",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detail-site-id")
}

func TestExportCmd_DefaultStatus(t *testing.T) {
	cmd := newExportCmd()
	// Default status should be [400, 900]
	statusFlag := cmd.Flags().Lookup("status")
	require.NotNil(t, statusFlag)
	assert.Equal(t, "[400,900]", statusFlag.DefValue)
}

func TestExportCmd_BuildStatusFilter(t *testing.T) {
	tests := []struct {
		name     string
		statuses []int
		expected string
	}{
		{"single status", []int{400}, `["400"]`},
		{"multiple statuses", []int{400, 900}, `["400","900"]`},
		{"three statuses", []int{200, 400, 900}, `["200","400","900"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildStatusFilter(tt.statuses)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExportCmd_BuildDateFilter(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		expected string
	}{
		{
			"normal range",
			"2026-01-01",
			"2026-03-31",
			"[\"2026-01-01\",\"2026-03-31\"]",
		},
		{
			"same day",
			"2026-04-01",
			"2026-04-01",
			"[\"2026-04-01\",\"2026-04-01\"]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDateFilter(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExportCmd_BuildDetailFilter(t *testing.T) {
	tests := []struct {
		name      string
		headerIDs []string
		expected  string
	}{
		{"single ID", []string{"1001"}, "[1001]"},
		{"multiple IDs", []string{"1001", "1002", "1003"}, "[1001],[1002],[1003]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDetailFilter(tt.headerIDs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCmd_HasExportSubcommand(t *testing.T) {
	cmd := NewCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "export" {
			found = true
			break
		}
	}
	assert.True(t, found, "workflow should have export subcommand")
}
