package batch

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectSummary_ExtractsSiteEntries(t *testing.T) {
	stepOutputs := map[string]map[string]string{
		"create-orders": {
			"Id":            "32100",
			"Title":         "注文管理",
			"ReferenceType": "Results",
		},
		"create-products": {
			"Id":            "32101",
			"Title":         "商品マスタ",
			"ReferenceType": "Results",
		},
		"setup-links": {
			"StatusCode": "200",
			"Message":    "updated",
		},
	}

	summary := CollectSummary("shopping-model", stepOutputs)

	require.NotNil(t, summary)
	assert.Equal(t, "shopping-model", summary.TemplateName)
	assert.Len(t, summary.Sites, 2)

	// Verify only steps with "Id" key are extracted
	siteIDs := make(map[string]string)
	for _, site := range summary.Sites {
		siteIDs[site.StepName] = site.SiteID
	}
	assert.Contains(t, siteIDs, "create-orders")
	assert.Contains(t, siteIDs, "create-products")
	assert.NotContains(t, siteIDs, "setup-links")
}

func TestCollectSummary_EmptyStepOutputs(t *testing.T) {
	summary := CollectSummary("empty-template", map[string]map[string]string{})

	require.NotNil(t, summary)
	assert.Equal(t, "empty-template", summary.TemplateName)
	assert.Empty(t, summary.Sites)
}

func TestCollectSummary_NoIdSteps(t *testing.T) {
	stepOutputs := map[string]map[string]string{
		"setup-links": {
			"StatusCode": "200",
		},
		"deploy-scripts": {
			"Message": "deployed",
		},
	}

	summary := CollectSummary("no-sites", stepOutputs)

	require.NotNil(t, summary)
	assert.Empty(t, summary.Sites)
}

func TestScaffoldSummary_WriteTo(t *testing.T) {
	summary := &ScaffoldSummary{
		TemplateName: "shift-management-v3",
		ParentID:     "32085",
		Sites: []SiteEntry{
			{StepName: "create-qualifications", SiteID: "32445", Title: "資格マスタ", ReferenceType: "Results"},
			{StepName: "create-sites", SiteID: "32446", Title: "現場マスタ", ReferenceType: "Results"},
		},
	}

	var buf bytes.Buffer
	n, err := summary.WriteTo(&buf)

	require.NoError(t, err)
	assert.Greater(t, n, int64(0))

	output := buf.String()
	assert.Contains(t, output, "Scaffold Summary")
	assert.Contains(t, output, "shift-management-v3")
	assert.Contains(t, output, "32085")
	assert.Contains(t, output, "資格マスタ")
	assert.Contains(t, output, "32445")
	assert.Contains(t, output, "現場マスタ")
	assert.Contains(t, output, "32446")
}

func TestScaffoldSummary_WriteTo_NoParentID(t *testing.T) {
	summary := &ScaffoldSummary{
		TemplateName: "test-template",
		Sites: []SiteEntry{
			{StepName: "create-table", SiteID: "100", Title: "テスト", ReferenceType: "Results"},
		},
	}

	var buf bytes.Buffer
	_, err := summary.WriteTo(&buf)

	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-template")
	assert.NotContains(t, output, "Folder:")
}

func TestEngine_StepOutputs_ReturnsShallowCopy(t *testing.T) {
	engine := &Engine{
		stepOutputs: map[string]map[string]string{
			"step1": {"Id": "100", "Title": "テスト"},
			"step2": {"Id": "200", "Title": "テスト2"},
		},
	}

	outputs := engine.StepOutputs()

	// Verify it contains the same data
	assert.Len(t, outputs, 2)
	assert.Equal(t, "100", outputs["step1"]["Id"])
	assert.Equal(t, "200", outputs["step2"]["Id"])

	// Verify it's a copy (mutation doesn't affect original)
	outputs["step1"]["Id"] = "MUTATED"
	assert.Equal(t, "100", engine.stepOutputs["step1"]["Id"])

	// Verify adding new key doesn't affect original
	outputs["step3"] = map[string]string{"Id": "300"}
	_, exists := engine.stepOutputs["step3"]
	assert.False(t, exists)
}

func TestEngine_StepOutputs_EmptyOutputs(t *testing.T) {
	engine := &Engine{
		stepOutputs: make(map[string]map[string]string),
	}

	outputs := engine.StepOutputs()
	assert.Empty(t, outputs)
	assert.NotNil(t, outputs)
}

func TestScaffoldSummary_WriteTo_WithParentID(t *testing.T) {
	summary := &ScaffoldSummary{
		TemplateName: "test-template",
		ParentID:     "12345",
		Sites:        []SiteEntry{},
	}

	var buf bytes.Buffer
	_, err := summary.WriteTo(&buf)

	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Folder: 12345")
}
