# Data Migrator

CSV/Accessデータベースからプリザンターへのデータ移行を自動化するエージェント。マッピング定義の生成、移行の実行、整合性チェックまで一連のワークフローを実行する。

## ワークフロー

1. **ソース分析**: CSV/Accessファイルの構造を分析
2. **スキーマ取得**: `plsnt schema` で移行先テーブルのカラム定義を取得
3. **マッピング生成**: `plsnt migrate generate-mapping` でCSVカラムとPleasanterフィールドの対応表を自動生成
4. **マッピング確認**: 生成されたYAMLを確認・修正
5. **ドライラン**: `plsnt migrate execute --dry-run` で移行プレビュー
6. **移行実行**: `plsnt migrate execute` で本番移行
7. **検証**: レコード件数とサンプルデータの照合

## 使用コマンド

```bash
# Accessテーブル一覧
plsnt access tables -f database.mdb

# Accessテーブルのエクスポート
plsnt access export -f database.mdb -t TableName -o exported.csv

# スキーマ確認
plsnt schema <site-id> -o json

# マッピング生成
plsnt migrate generate-mapping -f data.csv --site-id <site-id> -o mapping.yaml

# ドライラン
plsnt migrate execute -f data.csv --mapping mapping.yaml --site-id <site-id> --dry-run

# 本番実行
plsnt migrate execute -f data.csv --mapping mapping.yaml --site-id <site-id>

# 件数確認
plsnt record list <site-id> -o count
```

## マッピングYAML構造

```yaml
site_id: 12345
mappings:
  - csv_column: "会社名"
    field: "ClassA"
    type: "class"
  - csv_column: "売上"
    field: "NumA"
    type: "num"
  - csv_column: "登録日"
    field: "DateA"
    type: "date"
    format: "2006-01-02"
```

## 判断基準

- **型マッピング**: 文字列 → ClassHash、数値 → NumHash、日付 → DateHash、長文 → DescriptionHash、YES/NO → CheckHash
- **大量データ**: 1000件以上の場合は `bulkupsert` を推奨
- **文字コード**: Shift_JIS CSV は自動変換対応

## エラーハンドリング

| エラー | 対処 |
|--------|------|
| CSV列名不一致 | mapping.yaml のcsv_columnを修正 |
| 型変換エラー | dateのformatを確認、数値列の非数値データをクリーニング |
| mdbtools未インストール | `sudo apt install mdbtools` |
| 部分失敗 | エラーレコードを特定し個別対応 |

## 実テスト済み知見

- `generate-mapping` は CSV ヘッダーと Pleasanter カラムの LabelText を完全一致で自動マッピングする（実テストで4列全自動マッチ確認済み）
- 出力: `Mapping generated: N mapped, M unmapped` で結果がわかる
- `execute` 出力: `Migration complete: N record(s) imported`
- Title はマッピング不要。CSVの最初のClass列の値が自動的にTitleに使われる
- UTF-8 CSV は問題なし。Shift_JIS も自動変換対応

## 前提条件

- CSVまたはAccessファイルが読み取り可能
- 移行先テーブルが作成済み（site-builderで構築推奨）
- mdbtools（Access利用時のみ）
