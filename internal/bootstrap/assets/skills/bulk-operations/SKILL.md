---
name: bulk-operations
description: 一括操作・バッチ処理のパターン集。CSV一括投入、条件付き一括更新、エクスポート、batchコマンドのYAML定義。
---

# 一括操作・バッチ処理

大量データの投入・更新・エクスポートとバッチ機能の使い方。

## CSV からの一括投入

### シェルスクリプトによる一括作成

```bash
while IFS=, read -r code name category price stock; do
  plsnt record create --site-id <site-id> --json "{
    \"Title\":\"$code\",
    \"ClassHash\":{\"ClassA\":\"$code\",\"ClassB\":\"$name\",\"ClassC\":\"$category\"},
    \"NumHash\":{\"NumA\":$price,\"NumB\":$stock}
  }"
done < data.csv
```

注意:
- CSV の値にダブルクォートやカンマが含まれる場合はエスケープが必要
- 大量データ（100件以上）の場合は API レート制限に注意

## 一括更新

### 条件付き一括更新

```bash
# 全レコードの在庫数を +1
for id in $(plsnt record list --site-id <site-id> -o ids); do
  current=$(plsnt record get "$id" --json '{}' -o ndjson | grep -o '"NumB":[0-9.]*' | cut -d: -f2 | cut -d. -f1)
  new=$((current + 1))
  plsnt record update "$id" --json "{\"NumHash\":{\"NumB\":$new}}"
done

# 特定条件のレコードだけ更新
for id in $(plsnt record list --site-id <site-id> --json '{"View":{"ColumnFilterHash":{"ClassC":"東京"}}}' -o ids); do
  plsnt record update "$id" --json '{"ClassHash":{"ClassD":"32149"}}'
done
```

## エクスポート

```bash
# CSV エクスポート
plsnt record list --site-id <id> --json '{}' -o csv --fields "ResultId,ClassA,ClassB,NumA" > export.csv

# 全ページエクスポート（200件以上の場合）
plsnt record list --site-id <id> --all-pages --json '{}' -o csv > full-export.csv

# JSON エクスポート（バックアップ用）
plsnt record list --site-id <id> --all-pages --json '{}' -o json > backup.json
```

## batch コマンド

YAML ファイルで複数ステップを定義して一括実行する。

### YAML 形式

```yaml
name: "バッチ名"
variables:
  site_id: "32108"
  branch_site: "32147"
steps:
  - name: "ステップ1"
    command: "record list"
    args:
      site-id: "{{site_id}}"
      output: "table"
      fields: "ResultId,ClassA,NumA"

  - name: "ステップ2"
    command: "record list"
    args:
      site-id: "{{branch_site}}"
      output: "count"
```

### 実行

```bash
plsnt batch run batch-file.yaml
plsnt batch run batch-file.yaml --dry-run
plsnt batch run batch-file.yaml --var site_id=99999
plsnt batch run batch-file.yaml --log-file batch.log
```

### 利用可能な command

| command | 説明 | 必要な args |
|---------|------|-----------|
| `record list` | レコード一覧 | `site-id` |
| `record create` | レコード作成 | `site-id`, `json` |
| `record get` | レコード取得 | `record-id` |
| `record update` | レコード更新 | `record-id`, `json` |
| `record delete` | レコード削除 | `record-id` |
| `site get` | サイト取得 | `site-id` |
| `site create` | サイト作成 | `parent-id`, `json` |
| `site update` | サイト更新 | `site-id`, `json` |

### 注意事項

- batch は外部プロセスとして `plsnt` コマンドを呼び出す
- `plsnt` が PATH に含まれている必要がある
- 変数は `{{変数名}}` 形式で展開される

## schema コマンド

テーブルのカラム定義を確認する:

```bash
plsnt schema <site-id> -o json
```

table 出力は columns が1行にまとまるため JSON が見やすい。
