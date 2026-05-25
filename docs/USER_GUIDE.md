# plsnt ユーザーガイド

plsnt は、プリザンター（Pleasanter）の REST API をコマンドラインから操作するための CLI ツールです。
AI エージェントとの連携を第一に設計されていますが、人間が直接使うことも想定しています。

## 目次

1. [インストール](#インストール)
2. [初期設定](#初期設定)
3. [基本的な使い方](#基本的な使い方)
4. [レコード操作](#レコード操作)
5. [サイト管理](#サイト管理)
6. [スキーマ確認](#スキーマ確認)
7. [ユーザー・グループ・部署管理](#ユーザーグループ部署管理)
8. [CSV インポート](#csv-インポート)
9. [Access データベース連携](#access-データベース連携)
10. [データ移行（migrate）](#データ移行migrate)
11. [バッチ実行](#バッチ実行)
12. [出力フォーマット](#出力フォーマット)
13. [グローバルオプション](#グローバルオプション)
14. [エラーの読み方](#エラーの読み方)
15. [AI エージェントとの連携](#ai-エージェントとの連携)

---

## インストール

### GitHub Releases からダウンロード

[Releases ページ](https://github.com/immmmmmmu/plsnt/releases) から、お使いの OS・アーキテクチャに合ったバイナリをダウンロードしてください。

| OS | アーキテクチャ | ファイル名 |
|----|------------|----------|
| Linux | amd64 | `plsnt_<version>_linux_amd64.tar.gz` |
| Linux | arm64 | `plsnt_<version>_linux_arm64.tar.gz` |
| macOS | amd64 | `plsnt_<version>_darwin_amd64.tar.gz` |
| macOS | arm64 (Apple Silicon) | `plsnt_<version>_darwin_arm64.tar.gz` |
| Windows | amd64 | `plsnt_<version>_windows_amd64.zip` |

ダウンロード後、PATH の通ったディレクトリに配置してください。

```bash
# Linux / macOS の例
tar xzf plsnt_*_linux_amd64.tar.gz
sudo mv plsnt /usr/local/bin/
```

### ソースからビルド

Go 1.26 以上が必要です。

```bash
git clone https://github.com/immmmmmmu/plsnt.git
cd plsnt
go build -ldflags="-s -w" -o plsnt .
```

### バージョン確認

```bash
plsnt version
```

---

## 初期設定

plsnt を使うには、プリザンターサーバーの URL と API キーを設定する必要があります。

> **前提条件**: テナント管理者権限を持つ API キーが必要です。

### プロファイルの作成

```bash
# デフォルトプロファイルを作成
plsnt config set --url https://your-pleasanter.example.com --api-key YOUR_API_KEY

# 名前付きプロファイルを作成（複数環境がある場合）
plsnt config set --name production --url https://prod.example.com --api-key PROD_KEY
plsnt config set --name staging --url https://staging.example.com --api-key STAGING_KEY
```

設定ファイルは `~/.config/plsnt/config.yaml` に保存されます（パーミッション 0600）。

### 接続テスト

```bash
plsnt config test
```

`Connection successful. API key is valid.` と表示されれば成功です。

### プロファイルの切り替え

```bash
# プロファイル一覧を表示（* が現在のプロファイル）
plsnt config list

# アクティブプロファイルを切り替え
plsnt config use production

# コマンド単位で一時的に切り替え
plsnt record list --site-id 100 --profile staging
```

---

## 基本的な使い方

plsnt のコマンドは `plsnt <リソース> <操作>` の形式です。

```
plsnt record list --site-id 100     # レコード一覧
plsnt record get 12345              # レコード取得
plsnt site get 100                  # サイト情報取得
plsnt schema 100                    # カラム定義確認
plsnt user list                     # ユーザー一覧
```

### JSON ペイロードの渡し方

データの作成・更新には `--json` フラグで JSON を渡します。

```bash
# --json フラグで直接指定
plsnt record create --site-id 100 --json '{"Title":"新しいレコード","Body":"本文"}'

# stdin からパイプで渡す
echo '{"Title":"パイプで作成"}' | plsnt record create --site-id 100

# ファイルから渡す
cat payload.json | plsnt record create --site-id 100
```

> stdin 入力のサイズ上限は 10MB です。

---

## レコード操作

### レコード一覧取得

```bash
# 基本的な一覧取得（デフォルト200件）
plsnt record list --site-id 100

# 全件取得（自動ページング）
plsnt record list --site-id 100 --all-pages

# フィルタ・ソートを指定
plsnt record list --site-id 100 \
  --view '{"ColumnFilterHash":{"ClassA":"Red"},"ColumnSorterHash":{"Title":"asc"}}'

# 件数だけ取得
plsnt record list --site-id 100 -o count

# ID一覧だけ取得
plsnt record list --site-id 100 -o ids
```

### レコード取得

```bash
plsnt record get 12345
```

### レコード作成

```bash
plsnt record create --site-id 100 --json '{
  "Title": "タスク名",
  "Body": "タスクの説明",
  "ClassHash": {
    "ClassA": "カテゴリ1"
  },
  "NumHash": {
    "NumA": 100
  },
  "DateHash": {
    "DateA": "2026-03-31"
  },
  "CheckHash": {
    "CheckA": true
  }
}'
```

### レコード更新

```bash
plsnt record update 12345 --json '{
  "Title": "更新後のタイトル",
  "ClassHash": {
    "ClassA": "カテゴリ2"
  }
}'
```

### レコード削除

```bash
plsnt record delete 12345
```

### 一括 Upsert（更新 or 挿入）

キー列を指定して、一致するレコードがあれば更新、なければ新規作成します。

```bash
# 方法1: --keys を使う場合（JSON は配列）
plsnt record upsert --site-id 100 --keys ClassA --json '[
  {"Title":"Item 1","ClassHash":{"ClassA":"key1"}},
  {"Title":"Item 2","ClassHash":{"ClassA":"key2"}}
]'

# 方法2: Keys と Data をまとめて渡す場合
plsnt record upsert --site-id 100 --json '{
  "Keys": ["ClassA"],
  "Data": [
    {"Title":"Item 1","ClassHash":{"ClassA":"key1"}},
    {"Title":"Item 2","ClassHash":{"ClassA":"key2"}}
  ]
}'
```

### 一括削除

```bash
# ID指定で削除（最大1000件）
plsnt record bulk-delete --site-id 100 --ids 101,102,103

# 100件以上は --confirm が必要
plsnt record bulk-delete --site-id 100 --ids 101,102,...,200 --confirm

# View フィルタで削除（--confirm 必須）
plsnt record bulk-delete --site-id 100 \
  --view '{"ColumnFilterHash":{"ClassA":"old"}}' --confirm
```

---

## サイト管理

### サイト情報取得

```bash
plsnt site get 100
```

### サイト作成

```bash
plsnt site create --parent-id 1 --json '{
  "Title": "新しいテーブル",
  "ReferenceType": "Results",
  "SiteSettings": {
    "EditorColumnHash": {
      "General": ["Title","Body","ClassA"]
    }
  }
}'
```

> `--parent-id` にはフォルダ（ReferenceType: Sites）を指定してください。テーブル（Results/Issues）を指定すると警告が表示されます。

### サイト更新

```bash
plsnt site update 100 --json '{"Title":"更新後の名前"}'
```

### サイト削除

```bash
plsnt site delete 100
```

### サイトコピー

```bash
# サイト設定を別の場所にコピー
plsnt site copy 100 --parent-id 200

# コピー時にタイトルを変更
plsnt site copy 100 --parent-id 200 --json '{"Title":"コピーしたテーブル"}'
```

### サイト検索

```bash
plsnt site search --parent-id 1 --keyword "顧客"
```

---

## スキーマ確認

サイトのカラム定義（どんなフィールドがあるか）を確認できます。

```bash
# JSON 形式で表示
plsnt schema 100

# テーブル形式で表示
plsnt schema 100 -o table

# 特定のフィールドだけ表示
plsnt schema 100 --fields ColumnName,LabelText,TypeName
```

テーブル設計を確認してから `--json` ペイロードを組み立てるのに便利です。

---

## ユーザー・グループ・部署管理

### ユーザー

```bash
plsnt user list                                           # 一覧
plsnt user get 1                                          # 取得
plsnt user create --json '{"LoginId":"user1","Name":"山田太郎","Password":"P@ss1234"}'  # 作成
plsnt user update 1 --json '{"Name":"山田次郎"}'           # 更新
plsnt user delete 1                                       # 削除
plsnt user import --file users.csv                        # CSV 一括作成
```

#### ユーザー CSV 一括インポート

CSV ファイルからユーザーを一括作成します。

```csv
LoginId,Name,Password
user1,山田太郎,P@ss1234
user2,佐藤花子,P@ss5678
```

```bash
plsnt user import --file users.csv
```

### グループ

```bash
plsnt group list
plsnt group get 1
plsnt group create --json '{"GroupName":"開発チーム"}'
plsnt group update 1 --json '{"GroupName":"開発1チーム"}'
plsnt group delete 1
```

### 部署

```bash
plsnt dept list
plsnt dept get 1
plsnt dept create --json '{"DeptName":"営業部"}'
plsnt dept update 1 --json '{"DeptName":"営業1部"}'
plsnt dept delete 1
```

---

## CSV インポート

CSV ファイルからレコードを一括登録できます。カラムのマッピングは YAML ファイルで定義します。

### マッピングファイルの例

```yaml
# mapping.yaml
columns:
  - csv_column: "顧客名"
    field: "Title"
  - csv_column: "カテゴリ"
    field: "ClassA"
    type: "ClassHash"
  - csv_column: "金額"
    field: "NumA"
    type: "NumHash"
  - csv_column: "期限"
    field: "DateA"
    type: "DateHash"
```

### インポート実行

```bash
# 新規作成（1件ずつ）
plsnt record import --site-id 100 --file data.csv --mapping mapping.yaml

# Upsert（ClassA をキーにして更新 or 挿入）
plsnt record import --site-id 100 --file data.csv --mapping mapping.yaml --keys ClassA
```

> Shift_JIS の CSV も自動検出して UTF-8 に変換します。

---

## Access データベース連携

Microsoft Access の `.mdb` / `.accdb` ファイルから直接プリザンターにインポートできます。

> **前提条件**: `mdbtools` が必要です。`sudo apt install mdbtools` でインストールしてください。

### テーブル一覧

```bash
plsnt access tables database.mdb
```

### テーブルを CSV としてエクスポート

```bash
plsnt access export database.mdb "顧客マスタ" > customers.csv
```

### テーブルを直接プリザンターにインポート

```bash
plsnt access import database.mdb "顧客マスタ" \
  --site-id 100 --mapping mapping.yaml

# Upsert モード
plsnt access import database.mdb "顧客マスタ" \
  --site-id 100 --mapping mapping.yaml --keys ClassA
```

---

## データ移行（migrate）

CSV とプリザンターのスキーマを突き合わせて、マッピングファイルを自動生成できます。

### マッピング自動生成

```bash
# スキーマを参照してマッピングを自動生成
plsnt migrate generate-mapping --file data.csv --site-id 100

# ファイルに出力
plsnt migrate generate-mapping --file data.csv --site-id 100 --output mapping.yaml
```

カラム名やラベルが一致するフィールドは自動マッピングされます。マッピングできなかったカラムはコメントとして出力されるので、手動で編集してください。

### 移行実行

```bash
# 生成・編集したマッピングで移行を実行
plsnt migrate execute --file data.csv --mapping mapping.yaml --site-id 100

# Upsert モードで実行
plsnt migrate execute --file data.csv --mapping mapping.yaml --site-id 100 --keys ClassA
```

---

## バッチ実行

複数のステップを YAML ファイルに定義して、一括実行できます。

### バッチファイルの例

```yaml
name: "月次レポート作成"
variables:
  site_id: "100"
  month: "2026-03"
steps:
  - name: export-data
    command: record.list
    args:
      site_id: "{{site_id}}"
      view: '{"ColumnFilterHash":{"DateA":"{{month}}"}}'

  - name: create-summary
    command: record.create
    depends_on:
      - export-data
    args:
      site_id: "200"
      json: '{"Title":"月次サマリー {{month}}"}'
```

### 実行

```bash
# 通常実行
plsnt batch run batch.yaml

# ドライラン（実行せずにステップを確認）
plsnt batch run batch.yaml --dry-run

# 変数を上書き
plsnt batch run batch.yaml --var site_id=200 --var month=2026-04

# ログファイルに出力
plsnt batch run batch.yaml --log-file batch.log
```

---

## 出力フォーマット

`-o` / `--output` フラグで出力形式を切り替えられます。

| フォーマット | 説明 | 用途 |
|------------|------|------|
| `json`（デフォルト） | インデント付き JSON | 人間が読む、エージェント連携 |
| `ndjson` | 1行1レコードの JSON | ストリーム処理、パイプ |
| `table` | テーブル形式 | ターミナルでの確認 |
| `csv` | CSV 形式 | Excel、他ツール連携 |
| `count` | レコード件数のみ | 件数確認 |
| `ids` | ID 一覧のみ | スクリプト連携 |

```bash
# テーブル形式で表示
plsnt record list --site-id 100 -o table

# CSV でファイルに出力
plsnt record list --site-id 100 -o csv > records.csv

# 特定のフィールドだけ表示
plsnt record list --site-id 100 -o table --fields Title,ClassA,NumA
```

---

## グローバルオプション

すべてのコマンドで使えるオプションです。

| オプション | 短縮 | 説明 |
|-----------|------|------|
| `--profile` | `-p` | 使用するプロファイル名 |
| `--output` | `-o` | 出力フォーマット（json/table/csv/ndjson/count/ids） |
| `--fields` | | 表示するフィールド名（カンマ区切り） |
| `--json` | | RAW JSON ペイロード |
| `--verbose` | `-v` | 詳細出力 |
| `--silent` | | エラー以外の出力を抑制 |
| `--dry-run` | | リクエスト内容を表示するだけで実行しない |
| `--insecure` | | TLS 証明書の検証をスキップ |
| `--log-file` | | ログの出力先ファイル |

---

## エラーの読み方

plsnt のエラーは構造化されており、エラーコード・メッセージ・対処法が表示されます。

```
Error: [VALIDATION_ERROR] URL and API key are required
Suggestion: Run 'plsnt config set --url <url> --api-key <key>'
```

- **エラーコード**: `INVALID_INPUT`, `VALIDATION_ERROR`, `API_ERROR`, `INTERNAL_ERROR` など
- **メッセージ**: 何が起きたかの説明
- **Suggestion**: 具体的な対処法

### よくあるエラーと対処法

| エラー | 原因 | 対処法 |
|-------|------|-------|
| `URL and API key are required` | プロファイルが未設定 | `plsnt config set` で設定 |
| `Profile "xxx" not found` | 存在しないプロファイルを指定 | `plsnt config list` で確認 |
| `invalid site ID` | サイトIDが不正（正の整数でない） | 正しいサイトIDを確認 |
| `StatusCode: 403` | 権限不足 | テナント管理者権限のAPIキーを使用 |
| `StatusCode: 500` | サーバー側エラー | JSON ペイロードの形式を確認 |
| `WARNING: Using HTTP (not HTTPS)` | 平文通信 | 本番環境では必ず HTTPS を使用 |

---

## AI エージェントとの連携

plsnt はエージェントファースト設計です。AI エージェント（Claude Code、GitHub Copilot など）から効率的に使えるように作られています。

### エージェント向けのポイント

1. **TTY 自動検出**: パイプ経由の場合、自動的にエージェント向けの出力になります
2. **JSON 入出力**: `--json` でペイロードを渡し、`-o json` で結果を受け取れます
3. **スキーマ自己参照**: `plsnt schema <site-id>` でテーブル構造を事前に確認できます
4. **構造化エラー**: エラーは JSON 形式で stderr に出力され、パースしやすくなっています

### エージェントからの使用例

```bash
# 1. まずスキーマを確認
plsnt schema 100 -o json

# 2. スキーマに基づいてレコードを作成
plsnt record create --site-id 100 --json '{"Title":"エージェントが作成","ClassHash":{"ClassA":"auto"}}'

# 3. 結果を確認
plsnt record list --site-id 100 -o json --fields Title,ClassA

# 4. 件数確認
plsnt record list --site-id 100 -o count
```

### パイプラインの例

```bash
# レコード一覧をフィルタして別サイトにコピー
plsnt record list --site-id 100 -o ndjson | \
  jq 'select(.ClassHash.ClassA == "Active")' | \
  while read -r line; do
    plsnt record create --site-id 200 --json "$line"
  done

# ID一覧を取得して一括削除
plsnt record list --site-id 100 -o ids | \
  tr '\n' ',' | \
  sed 's/,$//' | \
  xargs -I {} plsnt record bulk-delete --site-id 100 --ids {} --confirm
```

---

## 設定ファイル

### 設定ファイルの場所

```
~/.config/plsnt/config.yaml
```

`XDG_CONFIG_HOME` が設定されている場合は `$XDG_CONFIG_HOME/plsnt/config.yaml`。

### 設定ファイルの構造

```yaml
current_profile: default
profiles:
  default:
    url: https://pleasanter.example.com
    api_key: YOUR_API_KEY
    api_version: "1.1"
  staging:
    url: https://staging.example.com
    api_key: STAGING_API_KEY
    api_version: "1.1"
```

> 設定ファイルのパーミッションは 0600（オーナーのみ読み書き可）です。パーミッションが緩いと警告が表示されます。

---

## カスタムフィールド（ハッシュ）一覧

プリザンターのカスタムフィールドは、以下の6種類のハッシュで `--json` に指定します。

| ハッシュ名 | カラム例 | 値の型 | 説明 |
|-----------|---------|-------|------|
| `ClassHash` | ClassA〜ClassZ | 文字列 | 分類列（選択肢・ドロップダウン） |
| `NumHash` | NumA〜NumZ | 数値 | 数値列 |
| `DateHash` | DateA〜DateZ | 文字列（DateTime） | 日付列 |
| `DescriptionHash` | DescriptionA〜DescriptionZ | 文字列 | 説明列（長文テキスト） |
| `CheckHash` | CheckA〜CheckZ | 真偽値 | チェック列 |
| `AttachmentsHash` | AttachmentsA〜AttachmentsZ | オブジェクト配列 | 添付ファイル列 |

### 使用例

```json
{
  "Title": "案件名",
  "Body": "詳細説明",
  "ClassHash": { "ClassA": "カテゴリ", "ClassB": "ステータス" },
  "NumHash": { "NumA": 1000000 },
  "DateHash": { "DateA": "2026-03-31T00:00:00" },
  "DescriptionHash": { "DescriptionA": "長い説明文..." },
  "CheckHash": { "CheckA": true },
  "AttachmentsHash": { "AttachmentsA": [] }
}
```

---

## コマンドリファレンス一覧

```
plsnt config set        プロファイル設定
plsnt config list       プロファイル一覧
plsnt config use        アクティブプロファイル切替
plsnt config test       接続テスト

plsnt record get        レコード取得
plsnt record list       レコード一覧
plsnt record create     レコード作成
plsnt record update     レコード更新
plsnt record delete     レコード削除
plsnt record upsert     一括 Upsert
plsnt record import     CSV インポート
plsnt record bulk-delete 一括削除

plsnt site get          サイト情報取得
plsnt site create       サイト作成
plsnt site update       サイト更新
plsnt site delete       サイト削除
plsnt site copy         サイトコピー
plsnt site search       サイト検索

plsnt schema            カラム定義表示

plsnt user list         ユーザー一覧
plsnt user get          ユーザー取得
plsnt user create       ユーザー作成
plsnt user update       ユーザー更新
plsnt user delete       ユーザー削除
plsnt user import       ユーザー CSV 一括作成

plsnt group list        グループ一覧
plsnt group get         グループ取得
plsnt group create      グループ作成
plsnt group update      グループ更新
plsnt group delete      グループ削除

plsnt dept list         部署一覧
plsnt dept get          部署取得
plsnt dept create       部署作成
plsnt dept update       部署更新
plsnt dept delete       部署削除

plsnt access tables     Access テーブル一覧
plsnt access export     Access テーブル CSV エクスポート
plsnt access import     Access テーブル → プリザンター インポート

plsnt migrate generate-mapping  マッピング YAML 自動生成
plsnt migrate execute          マッピングによる CSV 移行実行

plsnt batch run         バッチ YAML 実行

plsnt version           バージョン表示
```
