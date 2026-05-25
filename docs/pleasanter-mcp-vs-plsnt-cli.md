# Pleasanter MCP Server vs plsnt CLI 機能比較

> 調査日: 2026-03-14
> 対象: Pleasanter 1.5.2.0 内蔵 MCP Server / plsnt CLI / pleasanter-mcp-server（サードパーティ）

## 概要

Pleasanter 1.5.0.0 以降、本体に MCP（Model Context Protocol）サーバーが内蔵された。
これにより AI アシスタント（Claude、ChatGPT 等）が Pleasanter を直接操作可能になった。

本資料では、Pleasanter 内蔵 MCP Server と plsnt CLI の機能を比較し、
それぞれの強み・棲み分けを整理する。

---

## 1. アーキテクチャ比較

| 項目 | Pleasanter 内蔵 MCP | plsnt CLI |
|------|---------------------|-----------|
| **動作形態** | Pleasanter 本体に統合された HTTP エンドポイント (`/mcp`) | スタンドアロン CLI バイナリ（Go） |
| **プロトコル** | JSON-RPC 2.0（MCP 標準） | REST API → stdout/stderr |
| **主な利用者** | AI アシスタント（Claude Code、Cursor 等） | AI エージェント / 人間 / シェルスクリプト |
| **認証** | `X-API-Key` または `Authorization: Bearer` ヘッダー | プロファイル設定（config.yaml） |
| **言語** | C#（.NET、Pleasanter 本体と同一プロセス） | Go（独立バイナリ） |
| **デプロイ** | Pleasanter 本体の設定で有効化（デフォルト無効） | バイナリ配置のみ |
| **外部依存** | なし（本体に内蔵） | Pleasanter REST API |

---

## 2. 機能マトリクス

### 2.1 レコード操作

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| レコード取得（単一） | `GetItem` | `record get` | 同等 |
| レコード一覧 | `GetItems` + ViewJson | `record list` + `--view` | MCP は CreateViewJson でビュー生成 |
| レコード作成 | `UpdateItem`※ | `record create` | MCP は CreateUpdateItemJson→UpdateItem の2段階 |
| レコード更新 | `UpdateItem` | `record update` | 同等 |
| レコード削除 | **非対応** | `record delete` | MCP にはレコード削除ツールがない |
| 一括アップサート | **非対応** | `record upsert` (bulkupsert API) | plsnt のみ |
| 一括削除 | **非対応** | `record bulk-delete` | plsnt のみ |
| CSV インポート | **非対応** | `record import` + マッピング YAML | plsnt のみ |
| 自動ページング | offset パラメータ | `--all-pages` 自動全件取得 | plsnt の方が便利 |

### 2.2 サイト管理

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| サイト取得 | `GetSite` | `site get` | 同等 |
| サイト名検索 | `GetSiteIdByTitle`（完全/部分/前方一致） | `site search`（キーワード + 親ID） | MCP の方が検索モード豊富 |
| サイト作成 | **非対応** | `site create` | plsnt のみ |
| サイト更新 | `UpdateSite` | `site update` | 同等 |
| サイト削除 | **非対応** | `site delete` | plsnt のみ |
| サイト複製 | **非対応** | `site copy` | plsnt のみ |

### 2.3 ビュー管理

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| ビュー JSON 生成 | `CreateViewJson`（日本語パラメータ対応） | `--view` フラグで JSON 直指定 | MCP は AI 向けに自然言語→JSON 変換 |
| ビュー追加 | `AddView` | **非対応** | MCP のみ |
| ビュー取得 | `GetView` | **非対応** | MCP のみ |
| ビュー名検索 | `GetViewIdByViewName` | **非対応** | MCP のみ |
| ビュー更新 | `UpdateView` | **非対応** | MCP のみ |
| ビューコピー | `CopyView` | **非対応** | MCP のみ |
| ビュー削除 | `DeleteView` | **非対応** | MCP のみ |

### 2.4 ユーザー・組織管理

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| ユーザー一覧 | `GetUsers` | `user list` | 同等 |
| ユーザー名検索 | `GetUserIdByName` | **非対応** | MCP のみ |
| ユーザー作成 | **非対応** | `user create` | plsnt のみ |
| ユーザー更新 | **非対応** | `user update` | plsnt のみ |
| ユーザー削除 | **非対応** | `user delete` | plsnt のみ |
| ユーザー CSV 一括作成 | **非対応** | `user import` | plsnt のみ |
| グループ管理 | **非対応** | `group list/get/create/update/delete` | plsnt のみ |
| 部署管理 | **非対応** | `dept list/get/create/update/delete` | plsnt のみ |

### 2.5 メール・通知

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| メール送信 | `SendEmail`（To/Cc/Bcc 対応） | **非対応** | MCP のみ |

### 2.6 スキーマ・定義

| 操作 | MCP Server | plsnt CLI | 備考 |
|------|:----------:|:---------:|------|
| カラム定義取得 | リソースとして提供（item-fields 等） | `schema` コマンド | plsnt はサイト固有の定義を取得 |

---

## 3. MCP Server 独自機能

### 3.1 Translator（自動変換）

Pleasanter MCP Server の最大の差別化機能。AI が自然言語で操作できるよう、内部で自動変換を行う。

| Translator | 機能 | 例 |
|------------|------|-----|
| **ColumnNameConverter** | 日本語表示名 ↔ 内部列名の相互変換 | 「顧客名」→ `ClassA` |
| **DisplayTranslator** | コード値 → 表示値の変換 | `100` → `未着手` |
| **CodeTranslator** | 表示値 → コード値の逆変換（更新用） | `未着手` → `100` |

plsnt CLI ではこの変換は行わず、ユーザーが内部列名（ClassA 等）を直接指定する設計。

### 3.2 ワークフロープロンプト（9 種）

AI アシスタントに対して、複数ステップの操作手順を提示するプロンプトテンプレート。

| プロンプト名 | 用途 |
|-------------|------|
| `search-records` | サイト名またはIDからレコード検索 |
| `update-records` | レコード更新（確認フロー付き） |
| `assign-user-to-record` | 担当者・管理者の割り当て |
| `save-view` | ビュー作成・保存 |
| `manage-views` | ビュー更新・削除・コピー |
| `search-users` | ユーザー検索 |
| `resolve-choice-values` | 選択肢の表示値→保存値変換 |
| `notify-overdue-records` | 期限超過レコードの抽出・通知 |
| `send-email` | メール送信 |

### 3.3 リソース（6 種）

AI がツールの使い方を参照するための仕様ドキュメント。

| リソース URI | 内容 |
|-------------|------|
| `resource://pleasanter/specs/site-settings` | サイト設定の読み取り仕様 |
| `resource://pleasanter/specs/item-fields` | レコード項目の JSON 仕様 |
| `resource://pleasanter/specs/view-json` | ViewJson の詳細仕様 |
| `resource://pleasanter/specs/choices-pattern` | 選択肢パターンの仕様 |
| `resource://pleasanter/specs/tool-capabilities` | 全ツールの概要 |
| `resource://pleasanter/specs/paging` | ページング仕様 |

### 3.4 ReadOnly モード

`ReadOnlyMode: true` で書き込み操作を完全に制限可能。許可されるのは以下の 10 ツールのみ:

- CreateUpdateItemJson, GetItem, GetItems
- GetSite, GetSiteIdByTitle
- GetUserIdByName, GetUsers
- CreateViewJson, GetView, GetViewIdByViewName

### 3.5 レート制限（4 方式）

APIキー単位でパーティション分離。

| 方式 | 説明 | デフォルト設定 |
|------|------|---------------|
| FixedWindow | 固定時間枠 | 30 リクエスト / 60 秒 |
| SlidingWindow | スライディング時間枠 | 30 リクエスト / 60 秒（5 セグメント） |
| TokenBucket | トークンバケット | 10 トークン、2 秒/1 トークン補充 |
| Concurrency | 同時実行数 | 3 |

### 3.6 MCP ログ

専用テーブル（McpLogs）に操作ログを記録。UI から閲覧・検索可能。

記録項目: セッション ID、リクエスト ID、メソッド名、ターゲット名（ツール/プロンプト）、
開始・終了時刻、経過時間、ステータスコード、API キープレフィックス、クライアント情報、
リクエスト/レスポンスデータ（最大 64KB）

---

## 4. plsnt CLI 独自機能

### 4.1 バッチエンジン

YAML 定義による複数ステップの自動実行。トポロジカルソートで依存関係を解決。

```yaml
name: order-workflow
variables:
  folder_id: "12345"
steps:
  - name: create-orders
    command: site create
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"注文","ReferenceType":"Results"}'
  - name: create-details
    command: site create
    depends_on: [create-orders]
    args:
      parent-id: "{{folder_id}}"
      json: '{"Title":"注文明細","ReferenceType":"Results"}'
```

ステップ出力参照（`{{step_name.Key}}`）でステップ間のデータ受け渡しが可能。

### 4.2 テンプレートライブラリ（11 種）

すぐに使えるドメイン別テンプレート。

| テンプレート | 用途 |
|-------------|------|
| scaffold-shopping / v2 | お買い物モデル |
| scaffold-shift-management / v3 | シフト管理 |
| scaffold-crm | CRM（営業案件） |
| scaffold-employee | 従業員管理 |
| scaffold-task-management | タスク管理 |
| scaffold-inventory | 在庫管理 |
| scaffold-library | 図書館管理 |
| monthly-report | 月間レポート |
| integrity-check | データ整合性チェック |

### 4.3 データマイグレーション

```
plsnt migrate generate-mapping --file data.csv --site-id 12345
plsnt migrate execute --file data.csv --site-id 12345 --mapping mapping.yaml
```

CSV ヘッダーとサイトスキーマから YAML マッピングを自動生成し、
列名一致 → 大文字小文字不問 → 手動マッピングの 3 段階でマッチング。

### 4.4 Access DB インポート

```
plsnt access tables legacy.mdb
plsnt access import legacy.mdb 顧客マスタ --site-id 12345 --mapping mapping.yaml
```

mdbtools を使用して .mdb/.accdb ファイルから直接 Pleasanter にインポート。

### 4.5 出力フォーマット（6 種）

| フォーマット | 用途 |
|-------------|------|
| json | 標準（マシン可読） |
| ndjson | ストリーム処理（1行1レコード） |
| table | 人間向けテーブル表示 |
| csv | スプレッドシート連携 |
| count | レコード数のみ |
| ids | ID のみ（パイプライン連携） |

### 4.6 構造化エラー

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "SiteID must be a positive integer",
    "suggestion": "Specify a valid site ID, e.g. --site-id 1234"
  }
}
```

エラーコード・メッセージ・解決提案を含む JSON エラーを stderr に出力。

### 4.7 マルチプロファイル

複数の Pleasanter 環境（開発・ステージング・本番）を切り替え可能。

```
plsnt config set --name production --url https://pls.example.com --api-key xxx
plsnt config use production
plsnt -p staging record list --site-id 12345
```

### 4.8 Dry Run

`--dry-run` で実際の API 呼び出しなしにリクエスト内容を確認。

---

## 5. 棲み分けと併用パターン

```
┌─────────────────────────────────────────────────────────┐
│                    AI アシスタント                        │
│              (Claude Code, Cursor, etc.)                │
│                                                         │
│  ┌──────────────────┐    ┌──────────────────────────┐   │
│  │  MCP Server      │    │  plsnt CLI               │   │
│  │  (対話的操作)     │    │  (バッチ/自動化)          │   │
│  │                  │    │                          │   │
│  │  - 自然言語検索   │    │  - YAML バッチ実行       │   │
│  │  - 日本語列名     │    │  - CSV インポート        │   │
│  │  - ビュー管理     │    │  - テンプレート展開      │   │
│  │  - メール通知     │    │  - 一括操作             │   │
│  │  - 表示値変換     │    │  - マイグレーション      │   │
│  └────────┬─────────┘    └─────────────┬────────────┘   │
│           │                            │                │
│           │  /mcp (JSON-RPC)           │  REST API      │
│           │                            │                │
│  ┌────────▼────────────────────────────▼────────────┐   │
│  │              Pleasanter Server                    │   │
│  └───────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### MCP Server が適する場面

1. **AI との対話的操作** — 自然言語で「〇〇テーブルの未完了レコードを見せて」
2. **表示名変換が必要な場面** — ClassA ではなく「顧客名」で操作
3. **ビュー管理** — フィルタの作成・保存・共有
4. **通知・メール** — レコードに紐づくメール送信
5. **ReadOnly な閲覧用途** — 安全な参照アクセス

### plsnt CLI が適する場面

1. **バッチ・自動化** — YAML テンプレートでテーブル群を一括構築
2. **データ移行** — CSV/Access DB からの大量データインポート
3. **一括操作** — bulkupsert、bulk-delete による大量レコード処理
4. **サイト管理** — サイト作成・削除・複製の自動化
5. **シェルスクリプト連携** — パイプラインでの加工・集計
6. **マルチ環境運用** — プロファイル切り替えで dev/staging/prod 操作
7. **ユーザー・組織管理** — CRUD + CSV 一括作成

---

## 6. 機能カバレッジ集計

| カテゴリ | MCP Server | plsnt CLI | 備考 |
|---------|:----------:|:---------:|------|
| レコード取得 | 2 | 2 | 同等 |
| レコード変更 | 1 | 3 | plsnt: create/update/delete |
| レコード一括 | 0 | 3 | plsnt: upsert/bulk-delete/import |
| サイト取得 | 2 | 2 | 同等 |
| サイト変更 | 1 | 4 | plsnt: create/update/delete/copy |
| ビュー管理 | 7 | 0 | MCP のみ |
| ユーザー | 2 | 5 | plsnt: CRUD + import |
| グループ | 0 | 5 | plsnt のみ |
| 部署 | 0 | 5 | plsnt のみ |
| メール | 1 | 0 | MCP のみ |
| スキーマ | 6 (リソース) | 1 | MCP はリソースとして豊富 |
| バッチ | 0 | 1 | plsnt のみ |
| マイグレーション | 0 | 2 | plsnt のみ |
| Access DB | 0 | 3 | plsnt のみ |
| Translator | 3 | 0 | MCP のみ（日本語↔内部名変換） |
| プロンプト | 9 | 0 | MCP のみ |
| **合計** | **34** | **36** | |

---

## 7. サードパーティ MCP Server との違い

参考: [Takashi-Matsumura/pleasanter-mcp-server](https://github.com/Takashi-Matsumura/pleasanter-mcp-server)（Node.js 製）

| 項目 | 内蔵 MCP | サードパーティ MCP |
|------|---------|------------------|
| 動作方式 | Pleasanter プロセス内 | 外部 Node.js プロセス |
| 日本語列名変換 | Translator で自動変換 | なし（内部名で操作） |
| ビュー管理 | CRUD 完備（7 ツール） | なし |
| メール送信 | 対応 | なし |
| レート制限 | 4 方式対応 | なし |
| ログ管理 | McpLogs テーブル | なし |
| ReadOnly モード | 対応 | なし |
| 分析機能 | なし | trend_analysis, status_summary |
| 横断検索 | なし | multi_site_search |
| セットアップ | パラメータ有効化のみ | Node.js + 環境変数設定 |

---

## 8. 結論と推奨

### 両方を併用する構成が最強

- **MCP Server**: AI アシスタントの標準接続先として有効化。対話的な操作、ビュー管理、メール通知に活用。
- **plsnt CLI**: 自動化・バッチ処理・データ移行・環境構築に活用。AI エージェントのツールとしても引き続き有効。

### plsnt CLI で検討すべき拡張

1. **ビュー管理コマンド** — MCP にあって plsnt にないギャップ
2. **メール送信コマンド** — 通知ワークフローの自動化
3. **MCP Server としての動作モード** — plsnt 自体を MCP Server として起動（`plsnt mcp serve`）

### MCP Server 有効化手順

`App_Data/Parameters/McpServer.json` の `Enabled` を `true` に変更し、Pleasanter を再起動。

```json
{
  "Enabled": true,
  "ReadOnlyMode": false,
  "RateLimit": {
    "FixedWindow": { "Enabled": true, "PermitLimit": 30, "WindowSeconds": 60 }
  }
}
```
