package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmd_HasMasterSubcommand(t *testing.T) {
	cmd := NewCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "master" {
			found = true
			break
		}
	}
	assert.True(t, found, "workflow should have master subcommand")
}

func TestMasterCmd_Flags(t *testing.T) {
	cmd := newMasterCmd()

	// Verify all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("site-id"))
	assert.NotNil(t, cmd.Flags().Lookup("file"))
	assert.NotNil(t, cmd.Flags().Lookup("key"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))

	// Verify shorthand flags
	assert.Equal(t, "f", cmd.Flags().Lookup("file").Shorthand)
	assert.Equal(t, "k", cmd.Flags().Lookup("key").Shorthand)

	// Verify default values
	assert.Equal(t, "ClassA", cmd.Flags().Lookup("key").DefValue)
}

func TestMasterCmd_RequiredFlags_MissingSiteID(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"master", "--file", "test.csv"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestMasterCmd_RequiredFlags_MissingFile(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"master", "--site-id", "12345"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestMasterCmd_SiteIDValidation(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"master", "--site-id", "0", "--file", "test.csv"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "site-id must be a positive integer")
}

func TestMasterCmd_FileNotFound(t *testing.T) {
	cmd := NewCmd()
	cmd.SetArgs([]string{"master", "--site-id", "12345", "--file", "/nonexistent/path.csv"})
	err := cmd.Execute()
	assert.Error(t, err)
}
