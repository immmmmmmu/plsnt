---
name: setup-issues
description: Pleasanter Issues テーブル（期限・ステータス管理）の使い方ガイド。ResultsとIssuesの違い、組み込みフィールド、--json必須の注意点。
---

# Issues テーブル活用ガイド

Pleasanter の Issues テーブル（期限・ステータス管理）の使い方。

## Results と Issues の違い

| 項目 | Results | Issues |
|------|---------|--------|
| ReferenceType | `Results` | `Issues` |
| IDフィールド名 | `ResultId` | `IssueId` |
| ステータス管理 | なし | 組み込み `Status` フィールド |
| 期間管理 | なし | 組み込み `StartTime`/`CompletionTime` |
| 進捗率 | なし | `ProgressRate` (0-100) |
| 用途 | マスタデータ、台帳 | タスク管理、案件管理 |

## 重要な注意点

### `--json '{}'` を必ず指定する

Issues テーブルの組み込みフィールドは、**`--json '{}'` を付けないと表示されない**。

```bash
# NG: StartTime/CompletionTime が空になる
plsnt record list --site-id <id> -o table --fields "IssueId,ClassA,Status"

# OK: --json '{}' を付ける
plsnt record list --site-id <id> --json '{}' -o table --fields "IssueId,ClassA,Status"
```

## Status 値の対応

| 値 | 意味 |
|----|------|
| 100 | 新規 |
| 150 | 準備中 |
| 200 | 実施中 |
| 300 | レビュー中 |
| 900 | 完了 |
| 910 | 保留 |

## レコード作成

```bash
plsnt record create --site-id <site-id> --json '{
  "ClassHash": {"ClassA": "案件名"},
  "Status": 100,
  "StartTime": "2026-03-01T00:00:00",
  "CompletionTime": "2026-03-31T00:00:00",
  "NumHash": {"NumA": 500}
}'
```

`Status`, `StartTime`, `CompletionTime` はトップレベル（Hash の外）に指定。

## フィルタリング

```bash
# ステータスでフィルタ（文字列で指定）
plsnt record list --site-id <id> --json '{"View":{"ColumnFilterHash":{"Status":"100"}}}'

# 未完了は範囲指定不可のため jq でフィルタ
plsnt record list --site-id <id> --json '{}' -o ndjson | jq 'select(.Status < 900)'
```
