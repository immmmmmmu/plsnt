# Site Builder

Pleasanterのサイト（テーブル）構築を自動化するエージェント。フォルダ構成の確認からテーブル作成、カラム設定、リンク設定、テストデータ投入まで一連のワークフローを実行する。

## ワークフロー

1. **フォルダ確認**: 対象フォルダのSiteIDを取得し存在確認
2. **テーブル作成**: `plsnt site create` でResults/Issuesテーブルを作成
3. **カラム設定**: `plsnt site update` でEditorColumns、カラム表示名、選択肢を設定
4. **リンク設定**: 関連テーブル間のリンク列を設定
5. **テストデータ投入**: `plsnt record create` または `plsnt record import` でサンプルデータを投入
6. **検証**: `plsnt schema` でカラム定義を確認、`plsnt record list` でデータ確認

## 使用コマンド

```bash
# フォルダ確認
plsnt site get <folder-site-id> -o json

# テーブル作成（Resultsテーブル）
plsnt site create --parent <folder-site-id> --json '{
  "SiteId": 0,
  "Title": "顧客マスタ",
  "ReferenceType": "Results",
  "SiteSettings": {
    "EditorColumnHash": {
      "General": ["ClassA","ClassB","NumA","DateA"]
    }
  }
}'

# カラム表示名設定
plsnt site update <site-id> --json '{
  "SiteSettings": {
    "Columns": [
      {"ColumnName": "ClassA", "LabelText": "会社名"},
      {"ColumnName": "ClassB", "LabelText": "担当者"},
      {"ColumnName": "NumA", "LabelText": "売上金額"}
    ]
  }
}'

# テストデータ投入
plsnt record create <site-id> --json '{
  "Title": "テストレコード",
  "ClassHash": {"ClassA": "サンプル取引先"},
  "NumHash": {"NumA": 100000}
}'

# スキーマ確認
plsnt schema <site-id> -o table
```

## 判断基準

- **ReferenceType選択**: ステータス管理・期限管理が必要 → Issues、不要 → Results
- **カラム割当**: 選択肢 → ClassHash、金額・数量 → NumHash、日付 → DateHash、長文 → DescriptionHash、フラグ → CheckHash
- **EditorColumns**: よく使うカラムを General セクションに配置

## 実テスト済み知見

- フォルダ（ReferenceType: Sites）に `record list` を実行すると unmarshal エラーになる。フォルダ確認は `site get` を使用すること
- 空Titleでレコード作成すると Pleasanter が自動で「タイトル無し」を設定する
- Issues テーブルの Status/StartTime/CompletionTime を取得するには `--json '{}'` フラグが必須
- NumHash の大きな値はテーブル出力で科学記数法（例: `1.2e+06`）になる。正確な値が必要な場合は `-o json` を使用
- `record list` は `--site-id` フラグで SiteID を指定（位置引数ではない）
- `site create` の `--parent-id` でフォルダIDを指定してテーブルを作成

## エラーハンドリング

| エラー | 対処 |
|--------|------|
| 403 Forbidden | テナント管理者APIキーを確認 |
| SiteID not found | 親フォルダのSiteIDを再確認 |
| Invalid ReferenceType | Results または Issues を指定 |
| Column not in EditorColumns | SiteSettings.EditorColumnHash に追加 |
| unmarshal error on record list | フォルダに対しては record list 不可。site get を使用 |

## 前提条件

- テナント管理者権限のAPIキーが設定済み
- 親フォルダのSiteIDが判明していること
