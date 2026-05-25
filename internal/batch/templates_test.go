package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateYAML_ParseAndValidate(t *testing.T) {
	templates := []struct {
		file      string
		name      string
		stepCount int
	}{
		{"scaffold-crm.yaml", "scaffold-crm", 3},
		{"scaffold-task-management.yaml", "scaffold-task-management", 2},
		{"scaffold-inventory.yaml", "scaffold-inventory", 2},
		{"scaffold-employee.yaml", "scaffold-employee", 2},
		{"integrity-check.yaml", "integrity-check", 3},
		{"monthly-report.yaml", "monthly-report", 4},
		{"scaffold-shopping.yaml", "scaffold-shopping", 7},
		{"scaffold-shopping-v2.yaml", "scaffold-shopping-v2", 14},
	}

	templatesDir := filepath.Join("..", "..", "templates")

	for _, tt := range templates {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join(templatesDir, tt.file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read template %s: %v", tt.file, err)
			}

			batch, err := ParseRaw(data)
			if err != nil {
				t.Fatalf("failed to parse template %s: %v", tt.file, err)
			}

			if batch.Name != tt.name {
				t.Errorf("expected name %q, got %q", tt.name, batch.Name)
			}

			if len(batch.Steps) != tt.stepCount {
				t.Errorf("expected %d steps, got %d", tt.stepCount, len(batch.Steps))
			}

			// Validate topological sort works
			sorted, err := TopologicalSort(batch.Steps)
			if err != nil {
				t.Fatalf("topological sort failed: %v", err)
			}
			if len(sorted) != len(batch.Steps) {
				t.Errorf("sorted length %d != steps length %d", len(sorted), len(batch.Steps))
			}
		})
	}
}

func TestTemplateYAML_VariableExpansion(t *testing.T) {
	templates := []struct {
		file     string
		varName  string
		varValue string
	}{
		{"scaffold-crm.yaml", "folder_id", "12345"},
		{"scaffold-task-management.yaml", "folder_id", "99999"},
		{"scaffold-inventory.yaml", "folder_id", "55555"},
		{"scaffold-employee.yaml", "folder_id", "77777"},
		{"scaffold-shopping.yaml", "folder_id", "88888"},
		{"scaffold-shopping-v2.yaml", "folder_id", "44444"},
	}

	templatesDir := filepath.Join("..", "..", "templates")

	for _, tt := range templates {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join(templatesDir, tt.file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read template %s: %v", tt.file, err)
			}

			batch, err := ParseRaw(data)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			// Override variable
			batch.Variables[tt.varName] = tt.varValue
			batch.ExpandVariables()

			// Check that variable was expanded in first step's args
			firstStep := batch.Steps[0]
			parentVal, ok := firstStep.Args["parent-id"]
			if !ok {
				t.Fatal("expected 'parent-id' arg in first step")
			}
			if parentVal != tt.varValue {
				t.Errorf("expected parent=%q after expansion, got %q", tt.varValue, parentVal)
			}
		})
	}
}

func TestTemplateYAML_DependencyOrder(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")

	// CRM template: deal and activity depend on customer master
	data, err := os.ReadFile(filepath.Join(templatesDir, "scaffold-crm.yaml"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	batch, err := ParseRaw(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	sorted, err := TopologicalSort(batch.Steps)
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}

	// First step must be create-customer-master (no dependencies)
	if sorted[0].Name != "create-customer-master" {
		t.Errorf("expected first sorted step to be 'create-customer-master', got %q", sorted[0].Name)
	}

	// create-deal-table and create-activity-log should come after
	for _, s := range sorted[1:] {
		if s.Name != "create-deal-table" && s.Name != "create-activity-log" {
			t.Errorf("unexpected step after customer master: %q", s.Name)
		}
	}
}

func TestTemplateYAML_StepNamesUnique(t *testing.T) {
	templates := []string{
		"scaffold-crm.yaml",
		"scaffold-task-management.yaml",
		"scaffold-inventory.yaml",
		"scaffold-employee.yaml",
		"integrity-check.yaml",
		"monthly-report.yaml",
		"scaffold-shopping.yaml",
		"scaffold-shopping-v2.yaml",
	}

	templatesDir := filepath.Join("..", "..", "templates")

	for _, file := range templates {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(templatesDir, file))
			if err != nil {
				t.Fatalf("failed to read: %v", err)
			}

			batch, err := ParseRaw(data)
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			seen := make(map[string]bool)
			for _, step := range batch.Steps {
				if step.Name == "" {
					t.Error("empty step name found")
				}
				if seen[step.Name] {
					t.Errorf("duplicate step name: %q", step.Name)
				}
				seen[step.Name] = true

				if step.Command == "" {
					t.Errorf("step %q has empty command", step.Name)
				}
			}
		})
	}
}
