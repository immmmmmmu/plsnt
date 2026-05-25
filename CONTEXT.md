# plsnt-workflow: Pleasanter x AI Agent Workflow Application

Pleasanter を UI 基盤として、申請・承認・支出管理の共通事務フローを実現するオープンソースアプリケーション。AI エージェント（Claude Code 等）が plsnt CLI 経由でアプリの構築・運用を行い、人間は Pleasanter の画面だけを使う。

## Architecture

```
+-----------------------------------------------------------+
|  Human Layer (Pleasanter UI + Client Scripts JS)          |
|  申請入力 / 承認操作 / 一覧閲覧 / 検索                    |
+-----------------------------------------------------------+
|  AI Agent Layer (plsnt CLI + Go)                          |
|  plsnt workflow deploy / master / approve / export        |
+-----------------------------------------------------------+
|  Pleasanter (Data / UI / Permissions / Process)           |
|  REST API / SiteSettings / StatusControl                  |
+-----------------------------------------------------------+
```

- **Human Layer**: 申請者・承認者は Pleasanter UI のみ操作。クライアントスクリプト（JS/CSS）が UI を補助する
- **AI Agent Layer**: plsnt CLI が Pleasanter REST API を操作。テーブル構築、マスタ投入、承認設定、CSV 出力を自動化する
- **Pleasanter**: データ格納・UI 表示・権限管理・プロセス機能（承認フロー）を提供。サーバースクリプトは使用しない

## Table Schema

8 テーブルで構成。マスタ 5 テーブル（Results 型）+ トランザクション 3 テーブル。

### Master Tables (Results)

| Table | Key Columns | Description |
|---|---|---|
| Department Master | ClassA: dept code, Title: dept name, ClassB: parent dept code | Organization hierarchy. Parent resolved by Go code (not self-referencing link) |
| Position Master | ClassA: position code, Title: position name, NumA: approval limit amount, NumB: hierarchy level | Position definitions with approval authority limits |
| Approval Rule Master | Title: rule name, NumA/NumB: amount range, NumC: approval stages, ClassC-E: approver specs (`role:X` or `user:ID`) | Rules for multi-stage approval. Converted to Pleasanter Process settings by `plsnt workflow deploy` |
| Application Type Master | ClassA: type code, Title: type name, CheckA: attachment required, DescriptionA: payment type choices | Application categories (e.g., expense, advance payment) |
| Vendor Master | ClassA: vendor code, Title: vendor name, DescriptionA: address, DescriptionB: bank account | Vendor/payee information |

### Transaction Tables

| Table | Type | Key Columns | Description |
|---|---|---|---|
| Application Header | **Issues** | Title: app title (auto-numbered), Status: 100-900, ClassA: app type (link), ClassB: payment type, ClassC: dept (link), ClassD: vendor (link), ClassE: current approver, NumA: approval stage, NumB: total amount | Main application record. Issues type for built-in Status + CompletionTime + Process integration |
| Application Detail | Results | Title: item name, ClassA: header (link), NumA: amount, DateA: usage date, ClassD: category | Line items for each application |
| Approval History | Results | Title: action (approve/reject/return/route-set), ClassA: header (link), ClassB: approver name, NumB: approval stage, DateA: action datetime | Audit trail of approval actions |

### Table Relations

```
Department Master ---+
                     |
App Type Master -----+--> Application Header --+--> Application Detail
                     |                          |
Vendor Master -------+                          +--> Approval History

Approval Rule Master ---> App Type Master (ClassB link)
```

### Status Definitions (Application Header)

| Code | Name | Description |
|---|---|---|
| 100 | Draft | Applicant is editing |
| 200 | Submitted | Awaiting approval |
| 300 | In Approval | Multi-stage approval in progress (L2) |
| 400 | Approved | All stages approved |
| 500 | Returned | Returned by approver |
| 600 | Rejected | Rejected by approver |
| 900 | Settled | Accounting process complete |

### Approval Flow (L1/L2 Coexistence)

- **L1 (Fixed route)**: Pleasanter Process buttons handle Status transitions (100->200->900). For simple single-stage approvals
- **L2 (Conditional route)**: `plsnt workflow approve` evaluates Approval Rule Master, resolves approvers, sets ClassE/NumA, transitions Status 200->300->...->900. For multi-stage approvals based on amount/department/type conditions

## Commands

### plsnt workflow deploy

Build workflow tables from template YAML and apply Process/StatusControl settings.

```bash
# Build all 8 tables + links + processes + status controls
plsnt workflow deploy --template full-deploy --folder-id 12345

# Build common master tables only (5 tables)
plsnt workflow deploy --template common-masters --folder-id 12345

# Build expense app tables only (3 tables, requires masters already exist)
plsnt workflow deploy --template expense-app --folder-id 12345

# Dry run
plsnt workflow deploy --template full-deploy --folder-id 12345 --dry-run

# Override template variables
plsnt workflow deploy --template common-masters --folder-id 12345 --set dept_site_id=32100
```

| Flag | Short | Type | Required | Description |
|---|---|---|---|---|
| `--template` | `-t` | string | Yes | Template name: common-masters / expense-app / full-deploy |
| `--folder-id` | `-f` | int64 | Yes | Destination folder SiteID |
| `--dry-run` | | bool | No | Show execution plan only |
| `--set` | `-s` | string[] | No | Override template variables (key=value) |

### plsnt workflow master

Bulk import master data from CSV/JSON via upsert.

```bash
# Import department master
plsnt workflow master --site-id 32200 --file departments.csv

# Specify upsert key column
plsnt workflow master --site-id 32200 --file departments.csv --key ClassA

# Dry run
plsnt workflow master --site-id 32200 --file departments.csv --dry-run
```

| Flag | Short | Type | Required | Description |
|---|---|---|---|---|
| `--site-id` | | int64 | Yes | Target table SiteID |
| `--file` | `-f` | string | Yes | CSV or JSON file path |
| `--key` | `-k` | string | No | Upsert key column (default: ClassA) |
| `--dry-run` | | bool | No | Show import plan only |

### plsnt workflow approve

Regenerate Pleasanter Process settings from Approval Rule Master and apply to Application Header table.

```bash
# Regenerate and apply process settings
plsnt workflow approve --header-site-id 32205 --rule-site-id 32202

# Dry run (show generated processes)
plsnt workflow approve --header-site-id 32205 --rule-site-id 32202 --dry-run
```

| Flag | Short | Type | Required | Description |
|---|---|---|---|---|
| `--header-site-id` | | int64 | Yes | Application Header table SiteID |
| `--rule-site-id` | | int64 | Yes | Approval Rule Master SiteID |
| `--dry-run` | | bool | No | Show generated process settings only |

### plsnt workflow export

Export approved application details as CSV for a given date range.

```bash
# Monthly CSV export
plsnt workflow export --header-site-id 32205 --detail-site-id 32206 \
  --from 2026-04-01 --to 2026-04-30

# File output
plsnt workflow export --header-site-id 32205 --detail-site-id 32206 \
  --from 2026-04-01 --to 2026-04-30 > 2026-04-expense.csv

# Filter by specific status
plsnt workflow export --header-site-id 32205 --detail-site-id 32206 \
  --from 2026-04-01 --to 2026-04-30 --status 400
```

| Flag | Short | Type | Required | Description |
|---|---|---|---|---|
| `--header-site-id` | | int64 | Yes | Application Header table SiteID |
| `--detail-site-id` | | int64 | Yes | Application Detail table SiteID |
| `--from` | | string | Yes | Start date (YYYY-MM-DD) |
| `--to` | | string | Yes | End date (YYYY-MM-DD) |
| `--status` | | int[] | No | Target statuses (default: [400, 900]) |

CSV columns: Date, Application No., Applicant, Department, Usage, Amount, Payment Type, Status.

## Templates

Template YAML files used by `plsnt workflow deploy` and `plsnt batch run`.

| File | Description |
|---|---|
| `templates/workflow/common-masters.yaml` | Create 5 common master tables (Department, Position, App Type, Approval Rule, Vendor) |
| `templates/workflow/expense-app.yaml` | Create 3 expense app tables (Header, Detail, Approval History). Requires master tables already exist |
| `templates/workflow/full-deploy.yaml` | Create all 8 tables in one batch (masters + expense app + link setup) |
| `templates/workflow/link-setup.yaml` | Reference definition for table-to-table link settings. Not directly executable by batch engine; use `scripts/workflow/setup-links.sh` |
| `templates/workflow/process-setup.yaml` | Process settings (approval flow) definition for Application Header. 13 processes covering multi-stage approval, return, and rejection. Requires site get -> merge -> site update pattern |
| `templates/workflow/status-controls.yaml` | StatusControl settings for field visibility/editability per status. Requires site get -> merge -> site update pattern |

## Scripts

Shell scripts and client-side scripts for deployment and UI enhancement.

| File | Description |
|---|---|
| `scripts/workflow/full-deploy.sh` | Full deployment script: table creation + link setup + process settings + status controls + sample data |
| `scripts/workflow/setup-links.sh` | Table-to-table link setup using safe site get -> merge -> site update pattern (avoids SiteSettings overwrite) |
| `scripts/workflow/pleasanter-scripts/amount-helper.js` | Client-side JS: amount input assistance and auto-calculation for Application Detail |
| `scripts/workflow/seed-data/departments.csv` | Sample department master data |
| `scripts/workflow/seed-data/positions.csv` | Sample position master data |

## Design Principles

1. **Pleasanter Process maximization**: Approval flows (buttons, status transitions, conditions, notifications, auto-numbering, access control) are implemented using Pleasanter's built-in Process feature, not custom code
2. **No server scripts**: All business logic runs in plsnt Go code. Only client scripts (JS/CSS) are used for UI enhancement
3. **SiteSettings full overwrite**: `site update` overwrites the entire SiteSettings. Always follow the pattern: `site get` -> merge new settings into existing -> `site update`
4. **Issues type for EditorColumnHash**: Issues tables use `"General"` key in EditorColumnHash
5. **Link field triple**: Table-to-table links require all three: Links definition + ChoicesText `"[[SiteId]]"` on source column + TitleColumns on target table
6. **Environment-independent templates**: Template YAMLs use variables (`{{folder_id}}`, `{{dept_site_id}}`) instead of hardcoded SiteIDs
7. **Agent-first design**: CLI output is JSON by default for machine consumption. TTY auto-detection switches to human-readable table format

## Quick Start

Deploy a complete workflow application to a new environment in 5 steps.

```bash
# 1. Configure plsnt profile
plsnt config set --url http://your-pleasanter-server --api-key YOUR_API_KEY

# 2. Deploy all 8 tables with full-deploy template
plsnt workflow deploy --template full-deploy --folder-id FOLDER_SITE_ID

# 3. Import sample master data
./scripts/workflow/seed-data.sh

# 4. Apply process settings (approval flow) to Application Header
plsnt workflow approve --header-site-id HEADER_SITE_ID --rule-site-id RULE_SITE_ID

# 5. Open Pleasanter UI and test: create application -> submit -> approve
```

Alternatively, use `scripts/workflow/full-deploy.sh FOLDER_SITE_ID` to run steps 2-4 in one command.

## License

AGPL v3

## Related Documents

| Document | Path | Description |
|---|---|---|
| Architecture | `docs/design/plsnt-workflow/architecture.md` | System architecture, component design, L1/L2 approval coexistence |
| CLI Spec | `docs/design/plsnt-workflow/cli-spec.md` | Detailed CLI interface specification with flag tables and output formats |
| Table Schema | `docs/design/plsnt-workflow/table-schema.md` | Full column definitions, link relations, status/process definitions |
| Requirements | `docs/spec/plsnt-workflow/requirements.md` | Functional and non-functional requirements (REQ-001 to REQ-073) |
| Project Spec | `docs/pleasanter-ai-agent-workflow-v05.md` | Original project proposal (draft v0.5) |
| Task Overview | `docs/tasks/plsnt-workflow/overview.md` | Implementation task breakdown and phases |
| CLI Dev Guide | `CLAUDE.md` | plsnt CLI development guide (build, test, project conventions) |
