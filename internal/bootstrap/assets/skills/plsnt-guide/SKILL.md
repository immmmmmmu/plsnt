---
name: plsnt-guide
description: plsnt CLI操作パターン集。site/record/hashフィールド操作、フィルタリング、ソート、出力フォーマット、テーブル間リンクの総合リファレンス。
---

# plsnt CLI 操作ガイド

Pleasanter APIを操作するCLIツール `plsnt` の使い方パターン集。

## 重要な制約事項

### site create について
- `site create` は環境によってIDを返す場合と 302 リダイレクトを返す場合がある
- 返ってきた `Id` フィールドがサイトIDとして使えるとは限らない — `site search` で確認
- **`--parent-id` には必ず Sites（フォルダ）の SiteId を指定する**

### site update の SiteSettings 全体上書き問題（重要）

`site update` は **SiteSettings を全体上書き** する。部分更新はできない。

```bash
# ❌ 危険: Processesのみ送信 → Columns, StatusControls等が全て消失
plsnt site update {id} --json '{"SiteSettings":{"Processes":[...]}}'

# ✅ 安全: site get → マージ → site update
CURRENT=$(plsnt site get {id} -o json 2>/dev/null)
NEW_SS=$(echo "$CURRENT" | python3 -c "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})
ss['Processes'] = [...]  # 追加・変更
print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
")
plsnt site update {id} --json "$NEW_SS"
```

影響を受けるフィールド: Columns, EditorColumnHash, Processes, StatusControls, Scripts, Styles, Links, TitleColumns 等 — すべて SiteSettings 配下

### Issues型テーブル作成の ReferenceType 指定

Issues型テーブルを作成する場合、`"ReferenceType":"Issues"` を **JSONトップレベル** に置く:
```bash
plsnt site create --parent-id {folder_id} --json '{"Title":"名前","ReferenceType":"Issues","SiteSettings":{...}}'
```
SiteSettings 内の ReferenceType だけでは Sites（フォルダ）として作成される。

### EditorColumnHash のキーは常に "General"（重要）

Issues型テーブルでもResults型でも、EditorColumnHash のキーは **常に `"General"`** を使う:
```json
{"EditorColumnHash": {"General": ["ClassA","ClassB","NumA"]}}
```
`"Issues"` キーを使うと、カスタムフィールド（ClassHash, NumHash等）が API 経由で保存されない。

### Hash フィールドの扱い
- `ClassHash`（文字列）、`NumHash`（数値）、`DateHash`（日付）、`DescriptionHash`（長文）、`CheckHash`（チェック）
- テーブル/CSV出力時は自動的にフラット化される（例: `ClassA`）

## コマンド別の引数体系

| コマンド | 構文 | `--site-id` |
|---------|------|-------------|
| `record list` | `record list --site-id <id>` | 必須 |
| `record create` | `record create --site-id <id> --json '...'` | 必須 |
| `record get` | `record get <record-id>` | **不要** |
| `record update` | `record update <record-id> --json '...'` | **不要** |
| `record delete` | `record delete <record-id>` | **不要** |
| `site get` | `site get <site-id>` | - |
| `site create` | `site create --parent-id <id> --json '...'` | - |
| `site update` | `site update <site-id> --json '...'` | - |
| `site search` | `site search --parent-id <id> --keyword "..."` | - |
| `site copy` | `site copy <source-id> --parent-id <id>` | - |

## サイト作成後のカラム設定

```bash
# EditorColumns（配列形式）でカラムを有効化 + Columns でラベル設定
plsnt site update <site-id> --json '{
  "SiteSettings": {
    "EditorColumns": ["ClassA","ClassB","NumA"],
    "Columns": [{"ColumnName":"ClassA","LabelText":"名前"}]
  }
}'
```

`EditorColumnHash` はエラーになる。`EditorColumns`（配列形式）を使用。

## フィルタリング・ソート

```bash
# フィルタ + ソート
plsnt record list --site-id <id> --json '{
  "View": {
    "ColumnFilterHash": {"ClassB": "販売中"},
    "ColumnSorterHash": {"NumA": "asc"}
  }
}'
```

## 出力フォーマット

| フォーマット | フラグ | 用途 |
|------------|--------|------|
| JSON | `-o json` | 詳細確認 |
| テーブル | `-o table` | 人間向け一覧 |
| CSV | `-o csv` | Excel連携 |
| NDJSON | `-o ndjson` | jqストリーム処理 |
| 件数 | `-o count` | 集計 |
| ID一覧 | `-o ids` | スクリプト連携 |

## テーブル間リンク

### リンク設定（Links + ChoicesText が必須）

```bash
# リンク設定: Links だけでなく ChoicesText も必要（UIでドロップダウン表示にするため）
plsnt site update <site-id> --json '{
  "SiteSettings": {
    "Columns": [
      {"ColumnName":"ClassE","LabelText":"リンク先","ChoicesText":"[[<target-site-id>]]"}
    ],
    "Links": [{"ColumnName":"ClassE","SiteId":<target-site-id>,"LabelText":"リンク先"}]
  }
}'
```

### リンクフィールドの値形式

```bash
# レコード作成/更新時: ブラケットなし（プレーンな数字文字列）
plsnt record create --site-id <id> --json '{"ClassHash":{"ClassE":"32149"}}'

# 検索時（ColumnFilterHash）: ブラケット付き
plsnt record list --site-id <id> --json '{"View":{"ColumnFilterHash":{"ClassE":"[32149]"}}}'
```

## エージェント連携コマンド

| コマンド | エージェント | 用途 |
|---------|-------------|------|
| `/site-build` | site-builder | テーブル構築自動化 |
| `/migrate-data` | data-migrator | CSV/Accessデータ移行 |
| `/generate-report` | report-generator | データ集計・レポート |
| `/check-integrity` | integrity-checker | データ整合性検証 |
| `/scaffold-app` | app-scaffolder | テンプレートベースアプリ構築 |

## 実テスト済み注意事項

- フォルダに `record list` → unmarshal エラー。`site get` を使う
- Issues テーブルの Status 取得には `--json '{}'` が必須
- 空Titleで作成すると「タイトル無し」が自動設定される
- NumHash の大きな値はテーブル出力で科学記数法。正確な値は `-o json`
- ColumnFilterHash のステータスフィルタは文字列で指定: `"Status":"100"`

## SiteSettings のスクリプト・スタイル管理

PleasanterのUI改善スクリプト（JavaScript/CSS）はSiteSettings経由でAPIから登録・管理できる。

### Scripts（クライアントサイドJS）

```bash
# 取得→マージ→更新のパターン（既存設定を保持するため必須）
CURRENT=$(plsnt site get $SITE_ID -o json)
SCRIPT_BODY=$(grep -v '^//' script.js | jq -Rs .)

NEW_SETTINGS=$(echo "$CURRENT" | jq "
  .SiteSettings | . + {
    \"Scripts\": [{
      \"Id\": 1,
      \"Title\": \"スクリプト名\",
      \"Body\": $SCRIPT_BODY,
      \"New\": true,
      \"Edit\": true,
      \"Index\": false
    }]
  }
")

plsnt site update $SITE_ID --json "{\"SiteSettings\":$NEW_SETTINGS}"
```

### Styles（クライアントサイドCSS）

Scripts と同じ構造。Body にCSSを記述。

### 必須ルール

- **Id は1以上の整数を必ず指定**（Id: 0 だとUI管理画面に表示されない）
- 既存SiteSettings全体を含めて更新（Scripts追加時もColumns/Links等を保持）
- 実行タイミングフラグ: New（新規）, Edit（編集）, Index（一覧）

お買い物モデルのUIスクリプト例: `scripts/shopping/pleasanter-scripts/`
詳細は `pleasanter-scripts` スキルを参照。

## グローバルフラグ

| フラグ | 説明 |
|--------|------|
| `--insecure` | TLS証明書検証をスキップ |
| `-o, --output` | 出力フォーマット指定 |
| `--fields` | 表示フィールド指定 |
| `--json` | JSONペイロード指定 |
