package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRequiredInt64(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		param   string
		want    int64
		wantErr bool
	}{
		{
			name:  "float64 value",
			args:  map[string]any{"site_id": float64(12345)},
			param: "site_id",
			want:  12345,
		},
		{
			name:    "missing value",
			args:    map[string]any{},
			param:   "site_id",
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    map[string]any{"site_id": "abc"},
			param:   "site_id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			got, err := requiredInt64(request, tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestOptionalString(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"name": "hello",
	}

	assert.Equal(t, "hello", optionalString(request, "name"))
	assert.Equal(t, "", optionalString(request, "missing"))
}
