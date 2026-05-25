package batch

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	batchpkg "github.com/immmmmmmu/plsnt/internal/batch"
	"github.com/immmmmmmu/plsnt/internal/errs"
)

// NewCmd creates the batch command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Execute batch operations from YAML definition",
	}
	cmd.AddCommand(newRunCmd())
	return cmd
}

func newRunCmd() *cobra.Command {
	var (
		dryRun  bool
		logFile string
		vars    []string
		silent  bool
	)

	cmd := &cobra.Command{
		Use:   "run <batch-file.yaml>",
		Short: "Execute a batch file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			batch, err := batchpkg.ParseFile(args[0])
			if err != nil {
				return err
			}

			// Apply --var overrides
			for _, v := range vars {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) != 2 {
					return errs.New(errs.CodeInvalidInput, fmt.Sprintf("invalid --var format: %q (expected key=value)", v))
				}
				if batch.Variables == nil {
					batch.Variables = make(map[string]string)
				}
				batch.Variables[parts[0]] = parts[1]
			}

			// Re-expand variables after overrides
			if len(vars) > 0 {
				// Re-parse to get un-expanded args, then expand with updated variables
				raw, readErr := os.ReadFile(args[0])
				if readErr != nil {
					return errs.New(errs.CodeInternalError, fmt.Sprintf("failed to re-read batch file: %v", readErr))
				}
				reparsed, parseErr := batchpkg.ParseRaw(raw)
				if parseErr != nil {
					return parseErr
				}
				reparsed.Variables = batch.Variables
				reparsed.ExpandVariables()
				batch = reparsed
			}

			engine, err := batchpkg.NewEngine(batchpkg.EngineOptions{
				DryRun:  dryRun,
				LogFile: logFile,
				Writer:  os.Stdout,
				Silent:  silent,
			})
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError)
			}
			defer engine.Close()

			results, err := engine.Execute(context.Background(), batch)
			if err != nil {
				return err
			}

			// Print step completion summary
			fmt.Fprintf(os.Stdout, "\n--- Batch %q complete: %d/%d steps succeeded ---\n",
				batch.Name, countSuccess(results), len(results))

			// Print scaffold summary (sites created) to stderr
			if !silent && !dryRun {
				summary := batchpkg.CollectSummary(batch.Name, engine.StepOutputs())
				if len(summary.Sites) > 0 {
					_, _ = summary.WriteTo(os.Stderr)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show steps without executing")
	cmd.Flags().StringVar(&logFile, "log-file", "", "write execution log to file")
	cmd.Flags().StringArrayVar(&vars, "var", nil, "override variable (key=value)")
	cmd.Flags().BoolVar(&silent, "silent", false, "suppress scaffold summary output")
	return cmd
}

func countSuccess(results []batchpkg.ExecutionResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}
