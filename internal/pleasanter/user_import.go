package pleasanter

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

// ParseUserCSV reads a CSV and returns user payloads for Pleasanter /api/users/create.
// Expected CSV headers: LoginId, Name, Password, DeptId (optional), FirstAndLastNameOrder (optional)
func ParseUserCSV(r io.Reader) ([]map[string]any, error) {
	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, errs.New(errs.CodeInvalidInput, "failed to read CSV header")
	}

	// Build header index
	headerIdx := make(map[string]int)
	for i, h := range headers {
		headerIdx[strings.TrimSpace(h)] = i
	}

	// Required fields
	requiredFields := []string{"LoginId", "Name", "Password"}
	for _, f := range requiredFields {
		if _, ok := headerIdx[f]; !ok {
			return nil, errs.New(errs.CodeInvalidInput,
				"CSV is missing required column: "+f).
				WithSuggestion("CSV must have columns: LoginId, Name, Password")
		}
	}

	var users []map[string]any
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errs.New(errs.CodeInvalidInput, "failed to read CSV row")
		}

		user := map[string]any{
			"LoginId":  row[headerIdx["LoginId"]],
			"Name":     row[headerIdx["Name"]],
			"Password": row[headerIdx["Password"]],
		}

		if idx, ok := headerIdx["DeptId"]; ok && idx < len(row) && row[idx] != "" {
			if v, err := strconv.ParseInt(row[idx], 10, 64); err == nil {
				user["DeptId"] = v
			}
		}
		if idx, ok := headerIdx["FirstAndLastNameOrder"]; ok && idx < len(row) && row[idx] != "" {
			if v, err := strconv.ParseInt(row[idx], 10, 64); err == nil {
				user["FirstAndLastNameOrder"] = v
			}
		}

		users = append(users, user)
	}

	return users, nil
}
