# plsnt ユーザーガイド v2 — MCP + CLI + Claude Code 統合環境

> 前提: このガイドは Pleasanter 1.5 以降 + plsnt CLI + Claude Code の3ツール統合環境を対象としています。
> CLI 単体の詳細は [USER_GUIDE.md](USER_GUIDE.md) を参照してください。

## 目次

1. [3つのツールの役割](#3つのツールの役割)
2. [環境構築](#環境構築)
3. [使い分けの判断基準](#使い分けの判断基準)
4. [Claude Code から使う](#claude-code-から使う)
5. [MCP で対話的に操作する](#mcp-で対話的に操作する)
6. [CLI で自動化する](#cli-で自動化する)
7. [ハイブリッドパターン](#ハイブリッドパターン)
8. [実践レシピ](#実践レシピ)
9. [トラブルシューティング](#トラブルシューティング)
10. [用語集](#用語集)

---

## 3つのツールの役割

Pleasanter を操作する手段は3つあります。それぞれ得意領域が異なり、**補完関係** にあります。

```
┌──────────────────────────────────────────────────────────┐
│                    Claude Code（指揮者）                    │
│    自然言語で指示 → 適切なツールを自動選択して実行           │
│                                                          │
│  ┌─────────────────────┐    ┌─────────────────────────┐  │
│  │   Pleasanter MCP    │    │      plsnt CLI          │  │
│  │   （対話・探索）      │    │   （自動化・構築）        │  │
│  │                     │    │                         │  │
│  │ ・テーブル構造確認    │    │ ・バッチ実行             │  │
│  │ ・データ検索          │    │ ・一括操作               │  │
│  │ ・ビュー管理          │    │ ・サイト構築             │  │
│  │ ・メール送信          │    │ ・スクリプトデプロイ      │  │
│  │ ・日本語列名変換      │    │ ・CSV/ファイル操作       │  │
│  └─────────────────────┘    └─────────────────────────┘  │
│                          │                               │
│                   Pleasanter Server                       │
└──────────────────────────────────────────────────────────┘
```

### 一言で言うと

| ツール | 役割 | 強み |
|--------|------|------|
| **Claude Code** | 指揮者 | 自然言語で指示するだけで、MCP と CLI を適切に使い分ける |
| **Pleasanter MCP** | 対話窓口 | 「〇〇を見せて」「ビューを作って」が自然に通じる |
| **plsnt CLI** | 自動化エンジン | バッチ・一括操作・再現性のある構築が得意 |

---

## 環境構築

### 前提条件

- Pleasanter 1.5.0 以降（MCP Server 内蔵）
- plsnt CLI（インストール済み。未導入なら [USER_GUIDE.md](USER_GUIDE.md) 参照）
- Claude Code（Anthropic CLI）

### Step 1: plsnt CLI のプロファイル設定

```bash
plsnt config set --url https://your-pleasanter.example.com --api-key YOUR_API_KEY
plsnt config test
```

### Step 2: Pleasanter MCP Server の有効化

Pleasanter サーバー側で MCP を有効化します。

```
# サーバー上の設定ファイルを編集
App_Data/Parameters/McpServer.json

{
    "Enabled": true,     ← false から true に変更
    "ReadOnlyMode": false,
    ...
}
```

変更後、サーバーを再起動（IIS の場合は `iisreset`）。

### Step 3: .mcp.json の生成

plsnt CLI のプロファイルから `.mcp.json` を自動生成します。

```bash
plsnt config mcp-setup --output .mcp.json
```

生成される `.mcp.json`:

```json
{
  "mcpServers": {
    "pleasanter": {
      "type": "http",
      "url": "https://your-pleasanter.example.com/mcp",
      "headers": {
        "X-Api-Key": "${PLEASANTER_API_KEY}"
      }
    }
  }
}
```

> **セキュリティ**: API キーは環境変数 `PLEASANTER_API_KEY` で渡すことを推奨。
> `.mcp.json` に直接記載する場合は `.gitignore` に追加してください。

```bash
# 環境変数を設定（~/.bashrc 等に追加）
export PLEASANTER_API_KEY="your-api-key-here"

# .gitignore に追加
echo ".mcp.json" >> .gitignore
```

### Step 4: Claude Code の再起動

```bash
claude   # Claude Code を起動（.mcp.json が自動読み込みされる）
```

### 接続確認

Claude Code から以下のように話しかけて、MCP 接続を確認します。

```
> ユーザー一覧を取得して
```

ユーザー情報が表示されれば、MCP + CLI + Claude Code の統合環境が完成です。

---

## 使い分けの判断基準

### フローチャート

```
やりたいことは？
  │
  ├─ 「〇〇を見せて」「〇〇ってどんなテーブル？」
  │    → MCP（対話的探索）
  │
  ├─ ビューの作成・管理
  │    → MCP（AddView / CreateViewJson）
  │
  ├─ メールを送りたい
  │    → MCP（SendEmail）
  │
  ├─ テーブルを構築したい、テンプレートを展開したい
  │    → CLI（batch run / site create）
  │
  ├─ 100件以上の一括操作（作成・更新・削除）
  │    → CLI（upsert / bulk-delete / import）
  │
  ├─ スクリプト（JS/CSS）をデプロイしたい
  │    → CLI（deploy-all.sh）
  │
  ├─ CSVファイルからインポート、ファイルにエクスポート
  │    → CLI（record import / record list -o csv）
  │
  ├─ 繰り返し実行する操作手順を定義したい
  │    → CLI（batch YAML）
  │
  └─ 迷った
       → CLI（デフォルト）
```

### 比較表

| 操作 | MCP | CLI | どちらが適切か |
|------|:---:|:---:|--------------|
| テーブル構造の理解 | ○ | ○ | MCP — 日本語列名で表示してくれる |
| データ検索（少量） | ○ | ○ | MCP — 自然言語でフィルタできる |
| データ検索（大量・集計） | △ | ○ | CLI — jq でパイプ処理、件数制限なし |
| レコード CRUD（単一） | ○ | ○ | どちらでもOK |
| レコード一括操作 | × | ○ | CLI 一択 |
| サイト作成・削除 | × | ○ | CLI 一択 |
| ビュー作成・管理 | ○ | △ | MCP — 専用ツールあり |
| メール送信 | ○ | × | MCP 一択 |
| CSV インポート | × | ○ | CLI 一択 |
| ファイル出力 | × | ○ | CLI 一択 |
| バッチ実行（YAML） | × | ○ | CLI 一択 |
| スクリプトデプロイ | × | ○ | CLI 一択 |

> **○** = 対応  **△** = 一部対応  **×** = 非対応

---

## Claude Code から使う

Claude Code は **指揮者** として、MCP と CLI を自動的に使い分けます。
ユーザーは自然言語で指示するだけです。

### 基本的な会話例

```
> シフト割当テーブルの構造を教えて
  → Claude Code が MCP の GetSite を使い、日本語列名付きで説明

> 今月のシフトを一覧で見せて
  → Claude Code が MCP の GetItems を使い、結果をフォーマットして表示

> 警備員マスタに5人分のサンプルデータを入れて
  → Claude Code が plsnt CLI の record create を使い、5件作成

> shift-management-v3 テンプレートでテーブルを構築して
  → Claude Code が plsnt CLI の batch run を使い、テンプレート展開
```

### Claude Code 専用のスキル

Claude Code には Pleasanter 専用のスキルが組み込まれています。

| スキル | 説明 | 呼び出し方 |
|--------|------|-----------|
| `/site-build` | テンプレートからテーブル構築 | `/site-build shift-management-v3` |
| `/deploy-scripts` | UIスクリプトのデプロイ | `/deploy-scripts shopping` |
| `/seed-data` | サンプルデータ投入 | `/seed-data 32450 10` |
| `/build-app-full` | 構築→スクリプト→データの全フロー | `/build-app-full shopping` |
| `/scaffold-app` | 対話的なアプリ設計＋構築 | `/scaffold-app` |
| `/generate-report` | レポート生成 | `/generate-report 32450` |
| `/check-integrity` | テーブル間の整合性チェック | `/check-integrity 32444` |
| `/pleasanter-explore` | 対話的データ探索（MCP活用） | 自然言語で質問するだけ |

### コンテキストの活用

Claude Code は会話の文脈を記憶しています。連続した操作が自然にできます。

```
> シフト割当テーブルで、欠勤ステータスのレコードを見せて
  （MCP で検索して表示）

> その中で代替が必要なものはどれ？
  （前の結果を踏まえて分析）

> 代替要員を田中さんに変更して
  （CLI で record update を実行）
```

---

## MCP で対話的に操作する

MCP（Model Context Protocol）は Pleasanter 1.5 で内蔵された対話インターフェースです。
Claude Code を介して使います。

### 利用可能な MCP ツール

| ツール | 説明 |
|--------|------|
| `GetSite` | サイト設定を取得（列定義、リンク、スクリプト等） |
| `GetSiteIdByTitle` | サイト名からSiteIDを検索 |
| `GetItems` | レコード一覧を取得（フィルタ・ソート可） |
| `GetItem` | 単一レコードを取得 |
| `GetUsers` | ユーザー一覧を取得 |
| `GetUserIdByName` | ユーザー名からIDを検索 |
| `CreateViewJson` | View の JSON 定義を生成 |
| `AddView` | ビューを作成 |
| `GetView` | ビューを取得 |
| `GetViewIdByViewName` | ビュー名からIDを検索 |
| `UpdateView` | ビューを更新 |
| `CopyView` | ビューを複製 |
| `DeleteView` | ビューを削除 |
| `CreateUpdateItemJson` | レコード作成/更新用JSONを生成 |
| `UpdateItem` | レコードを更新 |
| `SendEmail` | メールを送信 |

### テーブル構造の確認

```
> 「シフト割当」テーブルの構造を教えて
```

Claude Code が MCP の `GetSiteIdByTitle` → `GetSite` を順に呼び出し、
以下のような情報を整理して表示します:

- テーブル名、種別（Issues/Results）
- カラム一覧（日本語ラベル + カラム名 + 型）
- リンク関係（どのテーブルと接続しているか）
- 適用されているスクリプト・スタイル

### ビューの作成

```
> シフト割当テーブルに「今週の確定シフト」ビューを作って。
> StartTime が今週で、Status が 200（確定）のフィルタ。
```

Claude Code が MCP の `CreateViewJson` → `AddView` を使って、
フィルタ付きビューを作成します。

### メール送信

```
> SiteID 32450 の Record 12345 についてメールを送って。
> 宛先は admin@example.com、件名は「シフト確認依頼」。
```

MCP の `SendEmail` でメール送信します。

### MCP の制限事項

MCP には以下の制限があります。これらの操作には CLI を使ってください。

| 制限 | 代替手段 |
|------|---------|
| レコード削除ができない | `plsnt record delete` / `plsnt record bulk-delete` |
| サイト作成・削除ができない | `plsnt site create` / `plsnt site delete` |
| 一括操作（upsert, bulk-delete）ができない | `plsnt record upsert` / `plsnt record bulk-delete` |
| CSV インポート・エクスポートができない | `plsnt record import` / `plsnt record list -o csv` |
| ファイル I/O ができない | CLI のリダイレクト（`> file.csv`） |
| ループ処理ができない | CLI のシェルスクリプト / バッチ YAML |

---

## CLI で自動化する

plsnt CLI は **再現性のある自動化** が得意です。
CLI 単体の詳細は [USER_GUIDE.md](USER_GUIDE.md) を参照してください。

### CLI が最適な場面

#### 1. テンプレートからのテーブル構築

```bash
# YAML テンプレートで複数テーブルを一括構築
plsnt batch run templates/scaffold-shift-management-v3.yaml \
  --var parent_id=32085

# 構築結果のサマリー表示（SiteID、カラム、リンク関係）
# → stderr にサマリーが自動出力される
```

#### 2. 大量データの一括操作

```bash
# CSV から一括インポート
plsnt record import --site-id 32447 --file guards.csv --mapping mapping.yaml

# 条件指定で一括削除
plsnt record bulk-delete --site-id 32450 \
  --view '{"ColumnFilterHash":{"Status":"910"}}' --confirm

# 一括 Upsert（キー列で更新 or 挿入）
plsnt record upsert --site-id 32447 --keys ClassA --json '[...]'
```

#### 3. スクリプトデプロイ

```bash
# UI スクリプト（JS/CSS/ServerScript）を一括デプロイ
bash scripts/shift-management-v3/pleasanter-scripts/deploy-all.sh
```

#### 4. データのエクスポート

```bash
# JSON で全件取得
plsnt record list --site-id 32450 --all-pages -o json > shifts.json

# CSV でエクスポート
plsnt record list --site-id 32450 -o csv --fields Title,ClassA,NumA > shifts.csv

# NDJSON でストリーム処理
plsnt record list --site-id 32450 -o ndjson | jq '.ClassHash.ClassA' | sort | uniq -c
```

#### 5. バッチ処理（YAML 定義）

```yaml
# monthly-report.yaml
name: "月次シフトレポート"
variables:
  site_id: "32450"
  month: "2026-03"
steps:
  - name: count-shifts
    command: record.list
    args:
      site_id: "{{site_id}}"
      view: '{"ColumnFilterHash":{"DateA":"{{month}}"}}'
      output: count

  - name: export-csv
    command: record.list
    args:
      site_id: "{{site_id}}"
      view: '{"ColumnFilterHash":{"DateA":"{{month}}"}}'
      output: csv
```

```bash
plsnt batch run monthly-report.yaml --var month=2026-03
```

### MCP 接続設定の生成

```bash
# 現在のプロファイルから .mcp.json を生成
plsnt config mcp-setup

# ファイルに出力
plsnt config mcp-setup --output .mcp.json

# 特定プロファイルから生成
plsnt config mcp-setup -p production --output .mcp.json
```

---

## ハイブリッドパターン

MCP と CLI の組み合わせで、単独では難しい操作を実現します。

### パターン 1: MCP で探索 → CLI で一括処理

```
1. [MCP] 「売上が0の商品を見せて」
   → GetItems でフィルタして対象レコードを表示

2. [確認] 「これらを削除してよいか？」
   → ユーザーが確認

3. [CLI] plsnt record bulk-delete --site-id 32100 --ids 101,102,103 --confirm
   → 確認済みレコードを一括削除
```

### パターン 2: CLI で構築 → MCP でビュー設定

```
1. [CLI] plsnt batch run templates/scaffold-shopping.yaml
   → テーブル群を構築

2. [MCP] 「注文テーブルにカレンダービューを追加して」
   → AddView でカレンダービューを作成

3. [MCP] 「未出荷の注文だけ表示するビューも作って」
   → CreateViewJson + AddView でフィルタビューを作成
```

### パターン 3: MCP で構造理解 → CLI でデータ移行

```
1. [MCP] 「顧客マスタテーブルの構造を教えて」
   → GetSite で列定義を日本語名付きで表示

2. [理解] ClassA=顧客名, ClassB=住所, NumA=与信枠 と把握

3. [CLI] plsnt migrate generate-mapping --file customers.csv --site-id 32200
   → マッピング YAML を自動生成

4. [CLI] plsnt migrate execute --file customers.csv --mapping mapping.yaml --site-id 32200
   → データ移行を実行
```

### パターン 4: MCP で確認 → CLI でスクリプト修正 → CLI でデプロイ

```
1. [MCP] 「シフト割当テーブルのスクリプト一覧を見せて」
   → GetSite でスクリプト名と概要を表示

2. [Claude Code] 「勤務時間計算のスクリプトに休憩時間の控除を追加して」
   → ローカルファイルを編集

3. [CLI] bash scripts/shift-management-v3/pleasanter-scripts/deploy-all.sh
   → 更新されたスクリプトをデプロイ
```

---

## 実践レシピ

### レシピ 1: 新しい業務アプリを0から構築する

最も一般的なワークフローです。

```
Step 1: 要件を Claude Code に伝える
> 「警備会社のシフト管理アプリを作りたい。
>  現場マスタ、警備員マスタ、シフト割当の3テーブルが必要。」

Step 2: Claude Code がテンプレート YAML を作成
→ templates/scaffold-xxx.yaml を自動生成

Step 3: テーブル構築（CLI）
> 「構築して」
→ plsnt batch run templates/scaffold-xxx.yaml

Step 4: サンプルデータ投入（CLI）
> 「サンプルデータを10件ずつ入れて」
→ plsnt record create を繰り返し実行

Step 5: ビュー作成（MCP）
> 「カレンダービューと、ステータス別フィルタビューを作って」
→ MCP の AddView を使用

Step 6: UI スクリプト作成・デプロイ（CLI）
> 「ステータスに応じて行を色分けするスクリプトを作って」
→ スクリプトファイルを生成 → deploy-all.sh でデプロイ
```

### レシピ 2: 既存テーブルのデータを調査する

```
> 「SiteID 32450 のデータの傾向を分析して」

Claude Code の動き:
1. [MCP] GetSite でテーブル構造を確認
2. [CLI] plsnt record list -o json --all-pages で全件取得
3. [分析] jq で集計・グルーピング
4. [表示] 結果をテーブル形式で表示
```

### レシピ 3: 定期レポートを自動生成する

```
> 「毎月のシフト実績レポートを生成するバッチを作って」

Claude Code の動き:
1. バッチ YAML ファイルを生成（monthly-report.yaml）
2. 変数で月を指定できるようにする

使い方:
$ plsnt batch run monthly-report.yaml --var month=2026-03
```

### レシピ 4: テーブル間の整合性をチェックする

```
> 「シフト管理v3 フォルダ配下のテーブル間リンクに矛盾がないか確認して」

Claude Code の動き:
1. [MCP] GetSite で各テーブルのリンク定義を取得
2. [CLI] plsnt record list で参照先に存在しない ID がないかチェック
3. [表示] 問題があれば一覧表示
```

---

## トラブルシューティング

### MCP 接続エラー

| 症状 | 原因 | 対処法 |
|------|------|--------|
| `401 認証できませんでした` | API キーが無効 or 環境変数未設定 | `.mcp.json` の API キーを確認。`PLEASANTER_API_KEY` 環境変数を設定 |
| `302 /errors/notfound` | MCP Server が無効 | `McpServer.json` で `"Enabled": true` に変更し、サーバー再起動 |
| MCP ツールが表示されない | `.mcp.json` が読み込まれていない | Claude Code を再起動。`.mcp.json` の JSON 構文を確認 |
| `500 Internal Server Error` | サーバー側エラー | Pleasanter のログを確認。MCP のバージョン互換性を確認 |

### CLI エラー

| 症状 | 原因 | 対処法 |
|------|------|--------|
| `URL and API key are required` | プロファイル未設定 | `plsnt config set --url <url> --api-key <key>` |
| `StatusCode: 403` | 権限不足 | テナント管理者権限の API キーを使用 |
| `WARNING: Using HTTP` | 平文通信 | 開発環境以外では HTTPS を使用 |

### Claude Code の動作

| 症状 | 原因 | 対処法 |
|------|------|--------|
| MCP を使わず CLI だけ使う | MCP 接続が確立していない | Claude Code を再起動し、MCP ツール一覧が表示されるか確認 |
| CLI を使わず MCP だけ使う | ルーティングルールの優先度 | CLAUDE.md のルーティングルールを確認 |
| 操作が途中で止まる | コンテキストウィンドウの圧迫 | `/compact` で圧縮、または新しいセッションを開始 |

---

## 用語集

| 用語 | 説明 |
|------|------|
| **MCP** | Model Context Protocol。AI モデルが外部ツールと通信するための標準プロトコル。Pleasanter 1.5 で本体に内蔵された |
| **plsnt CLI** | Pleasanter REST API を操作するコマンドラインツール。Go 製。バッチ処理・一括操作が得意 |
| **Claude Code** | Anthropic の AI アシスタント CLI。MCP と CLI を自動的に使い分けて Pleasanter を操作する |
| **SiteID** | Pleasanter のテーブル（サイト）の一意識別子 |
| **SiteSettings** | テーブルの設定（カラム定義、リンク、スクリプト、ビュー等）を保持する JSON オブジェクト |
| **View** | テーブルのフィルタ・ソート・表示列を定義した設定。カレンダー表示やフィルタ済み一覧を保存できる |
| **ColumnFilterHash** | View 内のフィルタ条件（列名→値のハッシュ） |
| **ColumnSorterHash** | View 内のソート条件（列名→asc/desc） |
| **ClassHash〜AttachmentsHash** | Pleasanter の6種カスタムフィールドハッシュ（[詳細](USER_GUIDE.md#カスタムフィールドハッシュ一覧)） |
| **batch YAML** | plsnt CLI の複数ステップを定義する YAML ファイル。テンプレート変数、依存関係、ステップ出力参照に対応 |
| **scaffold テンプレート** | 複数テーブル + リンク + カラム設定を一括構築する batch YAML |
| **プロファイル** | plsnt CLI の接続先サーバー・API キーのセット。複数環境の切り替えに使う |
| **`.mcp.json`** | Claude Code が MCP Server に接続するための設定ファイル。プロジェクトルートに配置 |

---

## 付録: ツール選択の早見表

### 「やりたいこと」から逆引き

| やりたいこと | 使うツール | コマンド/操作 |
|------------|-----------|-------------|
| テーブルの列名を知りたい | MCP | `GetSite` → 日本語列名で表示 |
| テーブルの列名を知りたい（MCP無し） | CLI | `plsnt schema <site-id> -o table` |
| 「〇〇のデータを見せて」 | MCP | `GetItems` + フィルタ |
| 特定条件のデータを CSV で出力 | CLI | `plsnt record list --view '...' -o csv > file.csv` |
| レコードを1件作成 | MCP or CLI | どちらでもOK |
| レコードを100件一括作成 | CLI | `plsnt record upsert` or `plsnt record import` |
| レコードを削除 | CLI | `plsnt record delete` / `plsnt record bulk-delete` |
| 新しいテーブルを作る | CLI | `plsnt site create` |
| カレンダービューを追加 | MCP | `AddView` |
| フィルタビューを追加 | MCP | `CreateViewJson` → `AddView` |
| メールを送る | MCP | `SendEmail` |
| テンプレートからアプリ構築 | CLI | `plsnt batch run templates/scaffold-xxx.yaml` |
| JS スクリプトをデプロイ | CLI | `bash deploy-all.sh` |
| データの傾向を分析 | Claude Code | 自然言語で依頼 → MCP + CLI を自動選択 |
| 月次レポートを自動生成 | CLI | `plsnt batch run report.yaml` |

---

> **v1 との違い**: v1（USER_GUIDE.md）は CLI 単体のリファレンスです。
> v2 は MCP と Claude Code を加えた3ツール統合環境での使い方に焦点を当てています。
> CLI の詳細コマンドリファレンスは引き続き [USER_GUIDE.md](USER_GUIDE.md) を参照してください。
