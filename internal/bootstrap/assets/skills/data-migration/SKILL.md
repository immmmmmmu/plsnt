---
name: data-migration
description: CSVからのデータ移行とリンクテーブル間の整合性チェック。migrate generate-mapping/executeのパターンガイド。
---

# データ移行・整合性チェック

CSV からのデータ移行と、リンクテーブル間の整合性チェックの方法。

## migrate コマンド

### ワークフロー

1. CSV ファイルを用意
2. `generate-mapping` でマッピング YAML を自動生成
3. 未マッピングカラムを手動修正
4. `execute` で移行実行

### Step 1: マッピング生成

```bash
plsnt migrate generate-mapping --file data.csv --site-id <site-id> --output mapping.yaml
```

自動マッピングの仕組み:
- CSV ヘッダーとサイトのカラムラベル（`LabelText`）を比較
- 完全一致 → 大文字小文字無視 → カラム名一致の順で検索
- マッチしないカラムはコメントとして出力

### Step 2: マッピング修正

未マッピングのカラムを手動で修正:
```yaml
columns:
    営業所コード: ClassB
    営業所名: ClassA
    月間売上目標: NumB    # コメントを外して対応カラムを指定
```

### Step 3: 移行実行

```bash
# 新規作成（全行を insert）
plsnt migrate execute --file data.csv --mapping mapping.yaml --site-id <site-id>

# upsert（キーで既存レコードを更新 or 新規作成）
plsnt migrate execute --file data.csv --mapping mapping.yaml --site-id <site-id> --keys "ClassB"
```

- `--keys` なし: 全行を新規レコードとして作成
- `--keys` あり: キーが一致するレコードを更新、なければ作成

## データ整合性チェック

### リンク切れ検出

```bash
valid_ids=$(plsnt record list --site-id $BRANCH_SITE -o ids)
for emp_id in $(plsnt record list --site-id $EMPLOYEE_SITE -o ids); do
  ref=$(plsnt record get "$emp_id" --json '{}' -o ndjson | \
    grep -o "\"$LINK_FIELD\":\"[^\"]*\"" | cut -d'"' -f4)
  if echo "$valid_ids" | grep -q "^${ref}$"; then
    echo "OK"
  else
    echo "NG: リンク切れ ($ref)"
  fi
done
```

### 重複チェック

```bash
plsnt record list --site-id <id> --json '{}' -o ndjson | jq -r '.ClassB' | sort | uniq -d
```

## 注意事項

- CSV はヘッダー行が必須、文字コードは UTF-8
- マッピング YAML の値は Pleasanter のカラム名（ClassA, NumA 等）
- 大量データの場合は API レート制限に注意
