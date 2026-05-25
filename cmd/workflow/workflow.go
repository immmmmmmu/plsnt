package workflow

import "github.com/spf13/cobra"

// NewCmd creates the "workflow" parent command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow management commands",
		Long:  "Manage workflow applications (deploy, master data, approval, export)",
	}
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newMasterCmd())
	cmd.AddCommand(newExportCmd())
	return cmd
}
