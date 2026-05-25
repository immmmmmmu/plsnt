package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// FormatVersion returns the formatted version string.
// When Version is "dev", no "v" prefix is added.
// Otherwise, "v" is prepended to the version number.
func FormatVersion(version, commit, date string) string {
	v := version
	if v != "dev" {
		v = "v" + v
	}
	return fmt.Sprintf("plsnt %s (%s, %s)", v, commit, date)
}

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), FormatVersion(Version, Commit, Date))
		},
	}
}
