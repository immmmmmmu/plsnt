# `plsnt site diff` — SitePackage JSON Structural Diff

`plsnt site diff <old.json> <new.json>` reports the structural differences
between two **SitePackage JSON files exported from the Pleasanter Web UI**.
The command runs entirely locally — no Pleasanter API calls — so it is
safe to run against archived JSON, in CI, or on machines without network
access.

## Why this exists

Pleasanter's REST API does not expose `sitepackage import` / `sitepackage
export`. The only way to obtain the JSON form of a site tree is the Web
UI's "Export site package" action. That makes day-to-day "what changed?"
review surprisingly painful:

- `diff` and `jq` produce massive false positives because Pleasanter
  re-orders array elements (Columns, Scripts, Views) on every export.
- SiteIds churn between environments, swamping the meaningful changes.
- Multi-line fields (Scripts.Body, HtmlTitleTop, Style.Body) have no
  natural diff in raw JSON.

`site diff` absorbs this noise and reports only what actually changed.

## Quick start

```bash
# default text output, no exit code semantics
plsnt site diff before.json after.json

# review-friendly markdown for a PR comment
plsnt site diff before.json after.json --format markdown > review.md

# CI gate: exit 1 if anything changed
plsnt site diff before.json after.json --exit-code

# pretend that LabelText edits are noise this PR
plsnt site diff before.json after.json --ignore LabelText
```

## How it absorbs noise

### Semantic array matching

Each well-known array name has a matching strategy:

| Array | Primary keys | Fallback |
|---|---|---|
| `Sites[]` | `SiteId`, then `Title`+`ParentId` | physical index |
| `Columns[]` | `ColumnName` | (unique by spec) |
| `Scripts[]` | `Title`, then `Id` | first 80 chars of Body (SHA-256) |
| `Styles[]` | `Title`, then `Id` | first 80 chars of Body (SHA-256) |
| `Views[]` | `Name`, then `Id` | physical index |
| `Processes[]` | `Name`, then `Id` | physical index |
| `Notifications[]` | `Type`+`Address`, then `Id` | physical index |
| `Reminders[]` | `Subject`, then `Id` | physical index |
| `Links[]` | `ColumnName`+`SiteId` | (unique by spec) |
| `StatusControls[]` | `Status`, then `Id` | physical index |

Reordered arrays produce zero diff. Renamed elements (e.g. a Script whose
Title changed) fall through to `Id` or the body hash and are still
matched, so you see one `~ Title: "old" → "new"` instead of one removal
plus one addition.

Some arrays carry meaning in their order (`EditorColumnHash`,
`FilterColumnHash`, `TabSettings`, `GridColumns`). Those are compared by
position so a reorder shows up correctly.

### Default ignore list

The following leaf keys are stripped from comparison unless you pass
`--no-ignore-default`:

```
SiteId, TenantId,
Creator, Updator, CreatedTime, UpdatedTime,
BaseSiteId, PackageTime, Server, AssemblyVersion, CreatorName
```

These are exporter metadata that change every time you re-export. Notice
that `ParentId` is **not** ignored — `Moved` detection runs before leaf
ignore is applied.

### Multi-line fields → unified diff

Fields whose name implies multi-line content are rendered as unified
diff regardless of length:

```
Body, Script,
HtmlTitleTop, HtmlTitleSite, HtmlTitleRecord,
GridGuide, EditorGuide, CalendarGuide, …
ChoicesText, Description, DefaultInput
```

Other strings get the unified diff treatment automatically when they
contain a newline and are longer than 60 characters.

## Output formats

### `--format text` (default)

```
== Header ==
  ~ HeaderInfo.IncludeNotifications: false → true

== Sites ==
  + (added)   [103] 新規追加
  - (removed) [101] 削除予定
  → (moved)   [102] 親変更前   ParentId: 1 → 99

== [102] 親変更前 ==
  ~ Sites[SiteId=102].ParentId: 1 → 99

== [100] 顧客マスタ ==
  ~ Sites[SiteId=100].SiteSettings.Columns[ColumnName=ClassA].LabelText: "顧客コード" → "顧客ID"
  ~ Sites[SiteId=100].SiteSettings.Scripts[Title=amount-helper].Body:
      --- old
      +++ new
      @@ -1,3 +1,3 @@
      -function calc(price, qty) { return price * qty; }
      +function calc(price, qty, discount) { return price * qty * (1 - discount); }
```

### `--format markdown`

GitHub-friendly: heading per site, fenced ` ```diff ` blocks for unified
diffs, table for added/removed/moved sites. Paste straight into a PR.

### `--format json`

Stable, schema'd output for machine consumption:

```json
{
  "modified": [{
    "site_id": 100,
    "title": "顧客マスタ",
    "changes": [{
      "path": "Sites[SiteId=100].SiteSettings.Columns[ColumnName=ClassA].LabelText",
      "kind": "modified",
      "old_value": "顧客コード",
      "new_value": "顧客ID"
    }]
  }]
}
```

## Flag reference

| Flag | Default | Effect |
|---|---|---|
| `--format` | `text` | `text` \| `json` \| `markdown` |
| `--ignore` | _empty_ | comma-separated leaf keys to drop at any depth |
| `--ignore-path` | _empty_ | comma-separated JSONPath globs (`*` = one segment, `**` = many) |
| `--no-ignore-default` | `false` | disable the built-in metadata ignore list |
| `--match-by` | `auto` | `auto` (SiteId then Title) \| `siteid` \| `title` |
| `--exit-code` | `false` | exit 1 when differences exist (POSIX `diff(1)` style) |
| `--strict` | `false` | duplicate semantic keys in either side become hard errors |
| `--include-permissions` | `false` | compare `Permissions[]` arrays (off by default — noisy) |
| `--max-size` | `50MB` | per-file size limit; `1GB` and `1.5GB` are accepted |

## Recipes

### Audit configuration drift on a schedule

```bash
# committed snapshot vs latest export
plsnt site diff configs/baseline.json /tmp/latest-export.json --format markdown
```

### Cross-environment review (SiteIds differ)

```bash
plsnt site diff staging-export.json production-export.json --match-by title
```

### CI gate for prototype JSON

```yaml
# .github/workflows/site-diff.yml
- name: Detect SitePackage drift
  run: |
    plsnt site diff prototype/sitepackage/baseline.json /tmp/latest.json \
      --format markdown --exit-code | tee diff.md
```

The job fails (exit 1) when there are changes, and `diff.md` is ready to
be posted as a PR comment.

### Suppress a noisy PR

A PR that intentionally renames every LabelText:

```bash
plsnt site diff before.json after.json --ignore LabelText
```

A PR that bumps `Version` everywhere:

```bash
plsnt site diff before.json after.json --ignore-path '/Sites/**/Version'
```

## What's *not* diffed

- **`Permissions[]`**: opt-in via `--include-permissions`. Default off
  because individual user IDs flood the diff with noise.
- **`Convertors[]`**: an exporter-side summary that duplicates the
  Sites[] tree.
- **`Data[]`** (record bodies): out of scope. SitePackage JSON does not
  always include records, and record-level diff is a separate concern.

## Limitations

- **API import/export remain Web-UI only.** This command does not paper
  over Pleasanter's missing API; you still need to export by hand. See
  [issue #11](https://github.com/immmmmmmu/plsnt/issues/11) for context.
- **No `apply`/`patch`.** Diff only. Applying a diff back is not safe in
  general because Pleasanter assigns SiteIds server-side.
- **No coloured output yet.** Initial release ships plain text in both
  TTY and non-TTY mode.
