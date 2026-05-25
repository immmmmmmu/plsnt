package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	batchpkg "github.com/immmmmmmu/plsnt/internal/batch"
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/validate"
)

// DeployResult holds the structured output of a workflow deploy execution.
type DeployResult struct {
	TemplateName string       `json:"template_name"`
	FolderID     int64        `json:"folder_id"`
	DryRun       bool         `json:"dry_run,omitempty"`
	Steps        []StepResult `json:"steps"`
	Errors       []string     `json:"errors"`
}

// StepResult holds the result of a single deploy step.
type StepResult struct {
	Name    string `json:"name"`
	SiteID  string `json:"site_id,omitempty"`
	Title   string `json:"title,omitempty"`
	Type    string `json:"type,omitempty"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
}

// resolveTemplatePath resolves a template name to a file path.
// It searches in templates/workflow/{name}.yaml relative to the current directory.
func resolveTemplatePath(name string) (string, error) {
	tmplPath := filepath.Join("templates", "workflow", name+".yaml")
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return "", errs.New(errs.CodeValidationError,
			fmt.Sprintf("template %q not found at %s", name, tmplPath)).
			WithSuggestion("Available templates are in templates/workflow/. Use --template <name> without .yaml extension")
	}
	return tmplPath, nil
}

func newDeployCmd() *cobra.Command {
	var (
		template string
		folderID int64
		dryRun   bool
		setVars  []string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy workflow tables from template",
		Long:  "Create workflow tables (masters + application tables) from a YAML template",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate template exists
			if _, err := resolveTemplatePath(template); err != nil {
				return err
			}

			// Validate folder-id is positive (rejects 0 with tenant root hint)
			if err := validate.ParentIDInt(folderID, "folder-id"); err != nil {
				return err
			}

			// Validate --set key=value format
			for _, s := range setVars {
				if !strings.Contains(s, "=") {
					return errs.New(errs.CodeInvalidInput,
						fmt.Sprintf("invalid --set format: %q (expected key=value)", s)).
						WithSuggestion("Use --set key=value, e.g. --set dept_site_id=32100")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Resolve template path
			tmplPath, err := resolveTemplatePath(template)
			if err != nil {
				return err
			}

			// 2. Read and parse raw YAML (without expanding variables yet)
			raw, err := os.ReadFile(tmplPath)
			if err != nil {
				return errs.New(errs.CodeInternalError, fmt.Sprintf("failed to read template: %v", err))
			}

			batch, err := batchpkg.ParseRaw(raw)
			if err != nil {
				return err
			}

			// 3. Inject folder_id from --folder-id flag
			if batch.Variables == nil {
				batch.Variables = make(map[string]string)
			}
			batch.Variables["folder_id"] = fmt.Sprintf("%d", folderID)

			// 4. Apply --set overrides
			for _, s := range setVars {
				parts := strings.SplitN(s, "=", 2)
				if len(parts) == 2 {
					batch.Variables[parts[0]] = parts[1]
				}
			}

			// 5. Expand variables after all overrides
			batch.ExpandVariables()

			// 6. Create and run batch engine
			writer := cmd.OutOrStdout()
			engine, err := batchpkg.NewEngine(batchpkg.EngineOptions{
				DryRun: dryRun,
				Writer: writer,
			})
			if err != nil {
				return errs.Wrap(err, errs.CodeInternalError)
			}
			defer engine.Close()

			results, execErr := engine.Execute(context.Background(), batch)

			// 7. Build deploy result
			deployResult := buildDeployResult(batch.Name, folderID, dryRun, results, engine.StepOutputs())

			if execErr != nil {
				deployResult.Errors = append(deployResult.Errors, execErr.Error())
			}

			// 8. Output JSON result
			jsonBytes, err := json.Marshal(deployResult)
			if err != nil {
				return errs.New(errs.CodeInternalError, fmt.Sprintf("failed to marshal result: %v", err))
			}
			fmt.Fprintln(writer, string(jsonBytes))

			if execErr != nil {
				return execErr
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "", "template name (required)")
	cmd.Flags().Int64VarP(&folderID, "folder-id", "f", 0, "parent folder site ID (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show execution plan without creating tables")
	cmd.Flags().StringArrayVarP(&setVars, "set", "s", nil, "override template variables (key=value)")
	_ = cmd.MarkFlagRequired("template")
	_ = cmd.MarkFlagRequired("folder-id")

	return cmd
}

// buildDeployResult constructs a DeployResult from execution results and step outputs.
func buildDeployResult(
	templateName string,
	folderID int64,
	isDryRun bool,
	results []batchpkg.ExecutionResult,
	stepOutputs map[string]map[string]string,
) DeployResult {
	dr := DeployResult{
		TemplateName: templateName,
		FolderID:     folderID,
		DryRun:       isDryRun,
		Steps:        make([]StepResult, 0, len(results)),
		Errors:       []string{},
	}

	for _, r := range results {
		sr := StepResult{
			Name:    r.StepName,
			Success: r.Success,
		}

		// Enrich with step output data (SiteID, Title, ReferenceType)
		if outputs, ok := stepOutputs[r.StepName]; ok {
			sr.SiteID = outputs["Id"]
			sr.Title = outputs["Title"]
			sr.Type = outputs["ReferenceType"]
		}

		if !r.Success && r.Error != nil {
			sr.Output = r.Error.Error()
		}

		dr.Steps = append(dr.Steps, sr)
	}

	return dr
}
