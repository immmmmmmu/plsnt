package workflow

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/immmmmmmu/plsnt/internal/api"
	"github.com/immmmmmmu/plsnt/internal/config"
	"github.com/immmmmmmu/plsnt/internal/errs"
	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/immmmmmmu/plsnt/internal/workflow/export"
)

const dateFormat = "2006-01-02"

func newExportCmd() *cobra.Command {
	var (
		headerSiteID int64
		detailSiteID int64
		deptSiteID   int64
		typeSiteID   int64
		from         string
		to           string
		statuses     []int
		bom          bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export approved application details as CSV",
		Long:  "Export application detail records filtered by date range and status as CSV to stdout",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate site IDs
			if headerSiteID <= 0 {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("header-site-id must be a positive integer, got: %d", headerSiteID)).
					WithSuggestion("Specify a valid site ID, e.g. --header-site-id 32205")
			}
			if detailSiteID <= 0 {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("detail-site-id must be a positive integer, got: %d", detailSiteID)).
					WithSuggestion("Specify a valid site ID, e.g. --detail-site-id 32206")
			}

			// Validate date format
			if _, err := time.Parse(dateFormat, from); err != nil {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("--from must be YYYY-MM-DD format, got: %q", from)).
					WithSuggestion("Example: --from 2026-01-01")
			}
			if _, err := time.Parse(dateFormat, to); err != nil {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("--to must be YYYY-MM-DD format, got: %q", to)).
					WithSuggestion("Example: --to 2026-03-31")
			}

			// Validate from <= to
			fromDate, _ := time.Parse(dateFormat, from)
			toDate, _ := time.Parse(dateFormat, to)
			if fromDate.After(toDate) {
				return errs.New(errs.CodeValidationError,
					fmt.Sprintf("--from (%s) must not be after --to (%s)", from, to)).
					WithSuggestion("Swap the dates or correct the range")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Initialize Pleasanter API client
			client, err := newExportClient(cmd)
			if err != nil {
				return err
			}

			svc := pleasanter.NewRecordService(client)
			ctx := context.Background()

			// 2. Fetch header records with status + date filter
			headerView := &pleasanter.View{
				ColumnFilterHash: map[string]string{
					"Status":          buildStatusFilter(statuses),
					"CreatedTime": buildDateFilter(from, to),
				},
			}

			headerResp, err := svc.ListAll(ctx, pleasanter.ListOptions{
				SiteID: headerSiteID,
				View:   headerView,
			})
			if err != nil {
				return err
			}

			headers := headerResp.Response.Data
			if len(headers) == 0 {
				// No matching headers; output CSV header only
				w := cmd.OutOrStdout()
				if bom {
					if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
						return err
					}
				}
				return export.GenerateCSV(nil, nil, w, nil)
			}

			// 3. Collect header IDs for detail filter
			headerIDs := make([]string, 0, len(headers))
			for _, h := range headers {
				headerIDs = append(headerIDs, recordID(h))
			}

			// 4. Fetch detail records linked to headers via ClassA
			detailView := &pleasanter.View{
				ColumnFilterHash: map[string]string{
					"ClassA": buildDetailFilter(headerIDs),
				},
			}

			detailResp, err := svc.ListAll(ctx, pleasanter.ListOptions{
				SiteID: detailSiteID,
				View:   detailView,
			})
			if err != nil {
				return err
			}

			// 5. Resolve master data if site IDs provided
			var opts *export.ResolveOptions
			if deptSiteID > 0 || typeSiteID > 0 {
				opts = &export.ResolveOptions{}
			}
			if deptSiteID > 0 {
				deptMap, err := fetchMasterMap(ctx, svc, deptSiteID)
				if err != nil {
					return err
				}
				opts.Departments = deptMap
			}
			if typeSiteID > 0 {
				typeMap, err := fetchMasterMap(ctx, svc, typeSiteID)
				if err != nil {
					return err
				}
				opts.AppTypes = typeMap
			}

			// 6. Output CSV
			w := cmd.OutOrStdout()
			if bom {
				if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
					return err
				}
			}
			return export.GenerateCSV(headers, detailResp.Response.Data, w, opts)
		},
	}

	cmd.Flags().Int64Var(&headerSiteID, "header-site-id", 0, "application header table site ID (required)")
	cmd.Flags().Int64Var(&detailSiteID, "detail-site-id", 0, "application detail table site ID (required)")
	cmd.Flags().Int64Var(&deptSiteID, "dept-site-id", 0, "department master table site ID (for name resolution)")
	cmd.Flags().Int64Var(&typeSiteID, "type-site-id", 0, "application type master table site ID (for name resolution)")
	cmd.Flags().StringVar(&from, "from", "", "start date YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&to, "to", "", "end date YYYY-MM-DD (required)")
	cmd.Flags().IntSliceVar(&statuses, "status", []int{400, 900}, "target statuses (comma-separated)")
	cmd.Flags().BoolVar(&bom, "bom", false, "output BOM-prefixed UTF-8 (for Excel)")
	_ = cmd.MarkFlagRequired("header-site-id")
	_ = cmd.MarkFlagRequired("detail-site-id")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

// recordID returns the record ID as a string (IssueId takes priority).
func recordID(r pleasanter.Record) string {
	if r.IssueId != 0 {
		return strconv.FormatInt(r.IssueId, 10)
	}
	return strconv.FormatInt(r.ResultId, 10)
}

// buildStatusFilter converts a slice of status codes to a JSON array string
// for Pleasanter's ColumnFilterHash.
// Pleasanter expects the format: ["400","900"] (not comma-separated).
func buildStatusFilter(statuses []int) string {
	parts := make([]string, len(statuses))
	for i, s := range statuses {
		parts[i] = fmt.Sprintf("%q", strconv.Itoa(s))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// buildDateFilter creates a Pleasanter date range filter string.
// Format: ["YYYY-MM-DD","YYYY-MM-DD"]
func buildDateFilter(from, to string) string {
	return fmt.Sprintf("[%q,%q]", from, to)
}

// buildDetailFilter creates a ColumnFilterHash value for ClassA
// to match multiple header record IDs.
// Format: [id1],[id2],...
func buildDetailFilter(headerIDs []string) string {
	parts := make([]string, len(headerIDs))
	for i, id := range headerIDs {
		parts[i] = "[" + id + "]"
	}
	return strings.Join(parts, ",")
}

// fetchMasterMap はマスタテーブルの全レコードを取得し、レコードID → Title のマップを返す。
// リンクフィールドの名前解決に使用する。
func fetchMasterMap(ctx context.Context, svc *pleasanter.RecordService, siteID int64) (map[string]string, error) {
	resp, err := svc.ListAll(ctx, pleasanter.ListOptions{
		SiteID: siteID,
	})
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, len(resp.Response.Data))
	for _, r := range resp.Response.Data {
		id := recordID(r)
		m[id] = r.Title
	}
	return m, nil
}

// newExportClient initializes a Pleasanter API client from config.
// Follows the same pattern as cmd/record/record.go newClient().
func newExportClient(cmd *cobra.Command) (api.Client, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, err
	}

	profileFlag, _ := cmd.Flags().GetString("profile")
	profile, _, err := cfg.ActiveProfileWithOverride(profileFlag)
	if err != nil {
		return nil, errs.New(errs.CodeValidationError, err.Error()).
			WithSuggestion("Run 'plsnt config set' to configure a profile")
	}

	url, apiKey, apiVersion := profile.Resolve()
	if url == "" || apiKey == "" {
		return nil, errs.New(errs.CodeValidationError, "URL and API key are required").
			WithSuggestion("Run 'plsnt config set --url <url> --api-key <key>'")
	}

	var opts []api.Option
	if insecure, _ := cmd.Flags().GetBool("insecure"); insecure {
		opts = append(opts, api.WithInsecure())
	}
	return api.New(url, apiKey, apiVersion, opts...), nil
}
