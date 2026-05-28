---
name: display-views
description: CLI出力フォーマット・ソート・フィールド選択・集計表示のパターン集。table/csv/json/ndjson/count/ids。
---

# 表示・ビューパターン

CLI で Pleasanter データを様々な形式で表示するためのパターン集。

## 出力フォーマット一覧

| フォーマット | `-o` | 用途 |
|-----------|------|------|
| テーブル | `table` | 人間が読む一覧表示 |
| CSV | `csv` | Excel/スプレッドシート連携 |
| JSON | `json` | プログラム連携（配列） |
| NDJSON | `ndjson` | パイプライン処理（1行1レコード） |
| 件数 | `count` | 集計スクリプト |
| ID一覧 | `ids` | ループ処理の入力 |

### フィールド選択

```bash
plsnt record list --site-id <id> --json '{}' -o table --fields "ClassA,ClassC,NumA"
```

## ソート

```bash
# 数値降順
plsnt record list --site-id <id> --json '{"View":{"ColumnSorterHash":{"NumA":"desc"}}}' -o table

# 複数カラムソート
plsnt record list --site-id <id> --json '{"View":{"ColumnSorterHash":{"ClassC":"asc","NumA":"desc"}}}' -o table
```

## フィルタリング

```bash
# PartialMatch
plsnt record list --site-id <id> --json '{
  "View":{"ColumnFilterHash":{"ClassC":"東京"},"ColumnFilterSearchTypes":{"ClassC":"PartialMatch"}}
}'

# Issues ステータスフィルタ
plsnt record list --site-id <id> --json '{"View":{"ColumnFilterHash":{"Status":"200"}}}'
```

### 日付範囲フィルタ（注意事項）

サーバー側フィルタが正しく動作しない場合がある。確実にフィルタするにはクライアントサイドで処理する。

## Top N / Bottom N

```bash
plsnt record list --site-id <id> --json '{"View":{"ColumnSorterHash":{"NumA":"desc"}}}' -o ndjson | head -5
```

## Kanban 風ステータスボード

```bash
for status in 100 200 900; do
  echo "=== Status=$status ==="
  plsnt record list --site-id <id> \
    --json "{\"View\":{\"ColumnFilterHash\":{\"Status\":\"$status\"}}}" -o table
done
```

## ピボットテーブル風集計

```bash
plsnt record list --site-id <id> --json '{}' -o ndjson | \
  awk -F'"' '{for(i=1;i<=NF;i++){if($i=="ClassC")region=$(i+2);if($i~/NumA/)val=$(i+1)}} {cnt[region]++;tot[region]+=val} END{for(r in cnt)printf "%-10s %4d件 合計:%d\n",r,cnt[r],tot[r]}'
```

## 注意事項

- ColumnFilterHash は一部のカラムで効かない場合がある（ChoicesText 設定等）
- Issues テーブルの `StartTime` 等は `--json '{}'` 必須
- 大量データは `--all-pages` で全ページ取得（デフォルト200件）
- `--fields` でカラムを絞ると table 表示の幅を制御可能
