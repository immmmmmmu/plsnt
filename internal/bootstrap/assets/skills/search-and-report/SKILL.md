---
name: search-and-report
description: plsnt CLIを使った高度な検索とレポート生成のパターン。部分一致検索、日付範囲フィルタ、複合検索、集計レポート。
---

# 検索・レポート パターン集

plsnt CLI を使った高度な検索とレポート生成のパターン。

## 部分一致検索

`ColumnFilterSearchTypes` に `PartialMatch` を指定:

```bash
plsnt record list --site-id <id> --json '{
  "View": {
    "ColumnFilterHash": {"ClassA": "検索語"},
    "ColumnFilterSearchTypes": {"ClassA": "PartialMatch"}
  }
}' -o table --fields "ResultId,ClassA"
```

検索タイプ一覧:

| 値 | 動作 |
|----|------|
| `ExactMatch` | 完全一致 |
| `PartialMatch` | 部分一致（LIKE %keyword%） |
| `ForwardMatch` | 前方一致（LIKE keyword%） |

## 日付範囲フィルタ

```bash
plsnt record list --site-id <id> --json '{
  "View": {
    "ColumnFilterHash": {
      "CompletionTime": "[\"2026-03-01\",\"2026-03-31\"]"
    }
  }
}'
```

注意: Issues テーブルの場合は `--json` 必須。

## 複合検索

```bash
plsnt record list --site-id <id> --json '{
  "View": {
    "ColumnFilterHash": {"ClassA": "麺", "ClassC": "東京"},
    "ColumnFilterSearchTypes": {"ClassA": "PartialMatch"},
    "ColumnSorterHash": {"NumA": "desc"}
  }
}'
```

## リンクフィールドでのクロステーブル検索

リンクフィールドには ColumnFilterHash でブラケット付きレコードIDを指定してフィルタ:

```bash
plsnt record list --site-id <crm-site-id> --json '{
  "View": {"ColumnFilterHash": {"ClassD": "[<営業所レコードID>]"}}
}'
```

注: レコード作成/更新時はブラケットなし（例: `"ClassD":"32274"`）。ブラケット形式はColumnFilterHash検索時のみ。

## 集計レポート（シェルスクリプト）

### ステータス別案件数

```bash
for status in 100 200 300 900; do
  count=$(plsnt record list --site-id <Issues-SiteId> --json "{\"View\":{\"ColumnFilterHash\":{\"Status\":\"$status\"}}}" -o count)
  echo "Status $status: ${count}件"
done
```

### NDJSON + jq による柔軟な集計

```bash
# NumA の合計
plsnt record list --site-id <id> --json '{}' -o ndjson | jq -s '[.[].NumA] | add'

# Status < 900 の未完了案件
plsnt record list --site-id <id> --json '{}' -o ndjson | jq 'select(.Status < 900)'
```

## ページネーション

デフォルトでは 200 件まで。それ以上の場合は `--all-pages` フラグ:

```bash
plsnt record list --site-id <id> --all-pages -o table
```
