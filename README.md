# plsnt — Pleasanter CLI Tool

**エージェントファースト**の Pleasanter REST API クライアント。
AI エージェント（Claude Code 等）との連携を第一に設計された CLI ツールです。

```
plsnt record list --site-id 100 -o json    # レコード一覧（JSON）
plsnt site create --parent-id 1 --json '{...}'  # サイト作成
plsnt batch run template.yaml              # バッチ実行
plsnt workflow deploy --template full-deploy --folder-id 100  # ワークフロー構築
```

## 特徴

- **エージェントファースト**: TTY 自動検出、`--json` RAW ペイロード、スキーマ自己参照
- **Pleasanter MCP 共存**: MCP で対話的探索、CLI で自動化の [ハイブリッド運用](docs/USER_GUIDE_V2.md)
- **バッチエンジン**: YAML テンプレートで複数テーブルの一括構築
- **ワークフロー**: 申請・承認フローアプリを `workflow deploy` 一発で構築
- **構成差分監査**: Web UI でエクスポートした SitePackage JSON 同士を [`plsnt site diff`](docs/site-diff.md) で意味的に比較（API 不要）

## インストール

### どちらを選ぶか

| 利用形態 | 推奨 |
|---------|-----|
| **Claude Code / Claude Desktop でリポジトリを開いて使う**（スキル・サンプル込みで使いたい開発者向け） | **A. ソースから** |
| MCP サーバーとして Claude Desktop の通常チャットから使うだけ（CLI 単体利用） | **B. バイナリ** |
| CI / 本番サーバで非対話的に実行 | **B. バイナリ** |
| Go 未インストール・非エンジニア | **B. バイナリ** |

`.claude/skills/` 配下の 28 スキル、`scripts/`、`templates/`、`sitepackage/` を使うなら **A が必須**。
バイナリだけ持っていても「CLI として動く」だけで、スキルや業務テンプレートは手に入らない。

### A. ソースからビルド（推奨）

```bash
git clone https://github.com/immmmmmmu/plsnt.git
cd plsnt
go build -ldflags="-s -w" -o plsnt .
```

Windows 向けクロスビルド（WSL / Linux から）:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o plsnt.exe .
```

Windows ネイティブビルド（PowerShell、Go 1.25+ が必要）:
```powershell
git clone https://github.com/immmmmmmu/plsnt.git C:\dev\plsnt
cd C:\dev\plsnt
go build -ldflags="-s -w" -o plsnt.exe .
# 任意で PATH に追加
Move-Item plsnt.exe C:\tools\plsnt\plsnt.exe
```

> Go 1.25 以上が必要です。`winget install GoLang.Go` で 5 分。

### B. ビルド済みバイナリ

[Releases](https://github.com/immmmmmmu/plsnt/releases) から OS/アーキテクチャに合ったバイナリをダウンロード。

**Linux / macOS:**
```bash
tar xzf plsnt_*_linux_amd64.tar.gz
sudo mv plsnt /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
Expand-Archive plsnt_*_windows_amd64.zip -DestinationPath C:\tools\plsnt
$env:Path += ";C:\tools\plsnt"
```

## 初期設定

```bash
# プロファイル作成
plsnt config set --url https://your-pleasanter.example.com --api-key YOUR_API_KEY

# 接続テスト
plsnt config test

# MCP 接続設定の自動生成（Pleasanter 1.5+）
plsnt config mcp-setup --output .mcp.json
```

設定ファイルは次の場所に保存されます（APIキーは平文、`chmod 600` 相当で保護）:

| OS | パス |
|----|------|
| Linux / macOS | `~/.config/plsnt/config.yaml` |
| Windows | `%USERPROFILE%\.config\plsnt\config.yaml` |

## Windows での Claude Desktop 連携

Claude Desktop から plsnt を MCP サーバーとして利用する場合:

1. `plsnt.exe` を `C:\tools\plsnt\` などに配置
2. PowerShell で接続設定:
   ```powershell
   C:\tools\plsnt\plsnt.exe config set --url http://your-pleasanter --api-key YOUR_KEY
   C:\tools\plsnt\plsnt.exe config test
   ```
3. `%APPDATA%\Claude\claude_desktop_config.json` に追加:
   ```json
   {
     "mcpServers": {
       "plsnt": {
         "command": "C:\\tools\\plsnt\\plsnt.exe",
         "args": ["mcp", "serve"]
       }
     }
   }
   ```
4. Claude Desktop を再起動

### Claude Code でプロジェクトを開く（スキル活用）

本リポジトリには `.claude/skills/` 配下に Pleasanter 操作の知見集（28 スキル）が含まれています。
Claude Desktop 内蔵の Claude Code、または Claude Code CLI で本リポジトリのフォルダを開くと、
これらのスキルが自動で認識されます。

```powershell
git clone https://github.com/immmmmmmu/plsnt.git C:\dev\plsnt
# Claude Desktop → Claude Code → C:\dev\plsnt を開く
```

> **注意**: シェルスクリプト（`scripts/*.sh`）を実行するには、Windows では **Git Bash**（Git for Windows に付属）
> または **WSL** が必要です。

## クイックスタート

### レコード操作

```bash
# 一覧取得
plsnt record list --site-id 100

# フィルタ・ソート
plsnt record list --site-id 100 \
  --view '{"ColumnFilterHash":{"ClassA":"Red"},"ColumnSorterHash":{"Title":"asc"}}'

# 作成
plsnt record create --site-id 100 --json '{"Title":"新規","ClassHash":{"ClassA":"A"}}'

# 更新
plsnt record update 12345 --json '{"Title":"更新後"}'

# 一括 Upsert
plsnt record upsert --site-id 100 --keys ClassA --json '[...]'
```

### サイト管理

```bash
plsnt site get 100                          # サイト情報
plsnt site create --parent-id 1 --json '{...}'  # 作成
plsnt schema 100 -o table                   # カラム定義確認
```

### バッチ実行

```bash
plsnt batch run template.yaml               # テンプレート展開
plsnt batch run template.yaml --dry-run     # ドライラン
plsnt batch run template.yaml --var key=val # 変数上書き
```

### ワークフロー（申請・承認アプリ）

```bash
# テーブル一括構築（8テーブル + リンク）
plsnt workflow deploy --template full-deploy --folder-id 12345

# マスタデータ投入
plsnt workflow master --site-id 32630 --file departments.csv

# 承認済み申請の CSV エクスポート
plsnt workflow export --header-site-id 32635 --detail-site-id 32636 \
  --from 2026-04-01 --to 2026-04-30
```

## 出力フォーマット

| フォーマット | 用途 |
|------------|------|
| `json`（デフォルト） | エージェント連携、プログラム処理 |
| `table` | ターミナル確認 |
| `csv` | Excel、他ツール連携 |
| `ndjson` | ストリーム処理、パイプ |
| `count` | 件数確認 |
| `ids` | スクリプト連携 |

```bash
plsnt record list --site-id 100 -o csv > records.csv
plsnt record list --site-id 100 -o ndjson | jq '.ClassHash.ClassA'
```

## コマンド一覧

```
plsnt config set/list/use/test/mcp-setup   プロファイル管理
plsnt record get/list/create/update/delete  レコード CRUD
plsnt record upsert/import/bulk-delete      一括操作
plsnt site get/create/update/delete/copy    サイト管理
plsnt schema                                カラム定義表示
plsnt user/group/dept list/get/create/...   ユーザー・組織管理
plsnt access tables/export/import           Access DB 連携
plsnt migrate generate-mapping/execute      データ移行
plsnt batch run                             バッチ YAML 実行
plsnt workflow deploy/master/export         ワークフロー管理
```

## MCP 共存（Pleasanter 1.5+）

3 つの経路があり、用途別にルーティングします:

| 用途 | 経路 |
|---|---|
| スキル / エージェント連携の **既定**（CRUD・バッチ・ワークフロー・サイト構築） | **plsnt MCP**（`plsnt mcp serve`、20 ツール） |
| 対話的探索・ビュー CRUD・メール送信・日本語列名自動変換 | **Pleasanter MCP**（本家 1.5+、`/mcp`） |
| シェルスクリプト / CI / CD・大量データ・ユーザー/グループ/部署 CRUD | **plsnt CLI** |

設計判断の根拠と 20 ツール選定基準は [ADR-0001](docs/adr/0001-pleasanter-mcp-vs-plsnt-mcp.md) を、運用ガイドは [USER_GUIDE_V2.md](docs/USER_GUIDE_V2.md) を参照してください。

## ドキュメント

| ドキュメント | 対象 |
|------------|------|
| [USER_GUIDE.md](docs/USER_GUIDE.md) | CLI 詳細リファレンス |
| [USER_GUIDE_V2.md](docs/USER_GUIDE_V2.md) | MCP + CLI + Claude Code 統合ガイド |
| [ADR-0001](docs/adr/0001-pleasanter-mcp-vs-plsnt-mcp.md) | Pleasanter MCP と plsnt MCP の棲み分け（設計判断） |
| [CONTEXT.md](CONTEXT.md) | ワークフローアプリの構造説明（エージェント向け） |
| [CLAUDE.md](CLAUDE.md) | 開発者向けガイド |

## 開発

```bash
go test ./...                    # 全テスト
go test -tags=integration ./...  # 統合テスト（実 Pleasanter 必要）
go test -cover ./...             # カバレッジ付き
golangci-lint run                # リント
```

## 開発元

[株式会社HereNow](https://here-now.jp)

## ライセンス

[AGPL v3](LICENSE) - Pleasanter 本体と同じライセンス。

Copyright (C) 2026 [株式会社HereNow](https://here-now.jp)
