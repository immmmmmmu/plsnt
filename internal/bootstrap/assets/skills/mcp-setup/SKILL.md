---
name: mcp-setup
description: Pleasanter MCP Serverのセットアップスキル。サーバー側有効化、plsnt config mcp-setup、環境変数設定、疎通確認、トラブルシューティング。
---

# MCP セットアップスキル

Pleasanter MCP Server を有効化し、Claude Code から利用可能にする手順。

## 前提条件

- Pleasanter 1.5.0 以降
- テナント管理者権限の API キー
- plsnt CLI がインストール・設定済み

## セットアップ手順

### Step 1: MCP Server 有効化（サーバー側）

Pleasanter の `App_Data/Parameters/McpServer.json` を編集:

```json
{
  "Enabled": true
}
```

ファイルが存在しない場合は新規作成する。

### Step 2: サーバー再起動

```bash
# IIS の場合
iisreset

# Kestrel の場合
sudo systemctl restart pleasanter
```

### Step 3: plsnt config mcp-setup

```bash
# .mcp.json を生成（API キーは環境変数で参照）
plsnt config mcp-setup --output .mcp.json

# API キーを直接埋め込む場合（非推奨: Git にコミットしないこと）
plsnt config mcp-setup --output .mcp.json --embed-key
```

生成される `.mcp.json` の例（環境変数参照パターン）:

```json
{
  "mcpServers": {
    "pleasanter": {
      "url": "http://localhost/mcp",
      "headers": {
        "X-Api-Key": "${PLEASANTER_API_KEY}"
      }
    }
  }
}
```

### Step 4: 環境変数の設定

`--embed-key` 未使用時は環境変数を設定する:

```bash
# .bashrc / .zshrc に追加
export PLEASANTER_API_KEY="your-api-key-here"
```

### Step 5: .gitignore に追加

```bash
echo ".mcp.json" >> .gitignore
```

### Step 6: Claude Code 再起動

MCP サーバー設定は Claude Code 起動時に読み込まれるため、再起動が必要。

## 疎通確認

```bash
# MCP エンドポイントに POST → 200 OK なら成功
curl -X POST http://localhost/mcp \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: $PLEASANTER_API_KEY" \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
```

成功時: HTTP 200 + JSON レスポンス
失敗時: 302（リダイレクト）または 401

## トラブルシューティング

| 症状 | 原因 | 対処 |
|------|------|------|
| 302 リダイレクト | MCP Server が有効化されていない | McpServer.json の Enabled: true を確認 → サーバー再起動 |
| 401 Unauthorized | API キーが不正 | API キーの値を確認。X-Api-Key ヘッダーの大文字小文字に注意 |
| Connection refused | Pleasanter が起動していない | サーバーの稼働状態を確認 |
| MCP ツールが表示されない | .mcp.json が読み込まれていない | Claude Code を再起動。.mcp.json のパスを確認 |
| 環境変数が展開されない | シェルの設定が反映されていない | `source ~/.bashrc` 後に Claude Code を再起動 |

## MCP 認証ヘッダー

ヘッダー名は `X-Api-Key`（ハイフン区切り、各単語先頭大文字）。
`x-api-key` や `X-API-KEY` でも動作するが、公式ドキュメントの表記に合わせる。

## MCP と CLI の使い分け

セットアップ後の使い分けは pleasanter-explore スキルを参照。基本方針:
- MCP: 対話的なデータ探索、ビュー作成・管理、メール送信
- CLI: バッチ処理、一括操作、サイト作成、スクリプトデプロイ、再現性が必要な操作
