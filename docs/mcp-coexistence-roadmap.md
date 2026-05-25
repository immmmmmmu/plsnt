# plsnt CLI 拡張方針 — MCP 共存時代のロードマップ

> 作成日: 2026-03-14
> 前提: Pleasanter 1.5.0+ 内蔵 MCP Server との併用

## 基本方針

plsnt の強みは **バッチ処理・一括操作・自動化・再現性** にある。
MCP Server は **対話的操作・日本語変換・ビュー管理** に強い。

「MCP があるから plsnt に追加する」のではなく、
「plsnt の強みを MCP が存在する世界でより活かすにはどうするか」で判断する。

---

## 棲み分けの明確化

```
MCP Server の領域              plsnt CLI の領域
─────────────────              ─────────────────
AI との対話的操作               バッチ・自動化
日本語列名 ↔ 内部名変換        YAML テンプレート展開
ビュー CRUD                    一括操作（upsert, bulk-delete）
メール送信                     CSV/Access DB マイグレーション
ReadOnly モード                サイト作成・削除・複製
ワークフロープロンプト           ユーザー/グループ/部署 CRUD
レート制限・ログ管理            マルチプロファイル（dev/stg/prod）
                               シェルスクリプト連携
                               構造化エラー
```

**重複を避け、接点での摩擦を減らす拡張に集中する。**

---

## 候補評価

### 候補1: ビュー管理コマンド (`plsnt view`)

| 項目 | 評価 |
|------|------|
| 実装コスト | Medium |
| 利用頻度 | Medium |
| MCP 相乗効果 | Medium |
| **推奨度** | **Should（ただし専用コマンド不要の可能性あり）** |

**分析**: テンプレートで scaffold する際にビュー定義も含めたいケースは実在する
（例: シフト管理テーブルにカレンダービューと日別一覧ビューを自動作成）。

ただし、**既存の `site update --json` で SiteSettings.Views を操作可能な可能性がある**。
専用の `view` コマンドを追加する前に、`site update` のペイロードで
ビュー設定を流し込めるかを検証すべき。可能なら、バッチテンプレートの中で
`site update` ステップとしてビューを定義する方が実装コストが低い。

**次のアクション**: SiteSettings.Views の API 読み書き可否を実機で検証

---

### 候補2: MCP Server 設定管理 (`plsnt mcp`)

| 項目 | 評価 |
|------|------|
| 実装コスト | Medium |
| 利用頻度 | Low |
| MCP 相乗効果 | Low |
| **推奨度** | **Won't** |

**理由**: McpServer.json は Pleasanter サーバー側のファイルシステムにある。
plsnt は REST API クライアントであり、サーバー側ファイルを操作する手段を持たない。
SSH/RDP でサーバーにログインするか、管理画面で行う操作であり、plsnt の責務範囲外。

---

### 候補3: スキーマの日本語名解決

| 項目 | 評価 |
|------|------|
| 実装コスト | **実質ゼロ（既に実装済み）** |
| 利用頻度 | High |
| MCP 相乗効果 | **High** |
| **推奨度** | **Must（完了済み）** |

**分析**: `plsnt schema <site-id>` は既に `LabelText`（日本語名）と
`ColumnName`（内部名 ClassA 等）の両方を出力している。

```
$ plsnt schema 32445 -o json
```

で `{"ColumnName":"ClassA", "LabelText":"顧客名", "Type":"Class"}` が取得可能。
AI エージェントが plsnt を使う際に、MCP の Translator 相当の情報を
このコマンドで取得できる。**追加実装は不要。**

---

### 候補4: MCP ログ分析

| 項目 | 評価 |
|------|------|
| 実装コスト | Low（テーブル存在が前提） |
| 利用頻度 | Low |
| MCP 相乗効果 | Medium |
| **推奨度** | **Won't（現時点）** |

**理由**: MCP ログは McpLogs テーブルに保存されるが、このテーブルが
通常の record API でアクセス可能かは未検証。仮にアクセスできても、
Pleasanter UI の McpLogs 画面（Controller/View が存在）で十分な可能性が高い。
ニーズが明確になるまで保留。

---

### 候補5: scaffold 完了時の MCP 向けサマリー

| 項目 | 評価 |
|------|------|
| 実装コスト | Low |
| 利用頻度 | Medium |
| MCP 相乗効果 | **High** |
| **推奨度** | **Should** |

**分析**: `plsnt batch run scaffold-xxx.yaml` の完了時に、
作成されたサイト一覧とフィールドマッピングのサマリーを出力する。

```
## scaffold 完了サマリー
- フォルダ SiteID: 32444
- テーブル一覧:
  - 顧客マスタ (SiteID: 32445): ClassA=顧客名, ClassB=電話番号, NumA=与信限度額
  - 注文 (SiteID: 32446): ClassA=顧客(Link:32445), DateA=注文日, NumA=合計金額

## MCP で操作する場合
GetSiteIdByTitle("顧客マスタ") → SiteID 32445
CreateViewJson(siteId: 32445, columnFilterHash: {"顧客名": "xxx"})
```

これにより、plsnt で scaffold したアプリを MCP から即座に操作開始できる。
plsnt → MCP のハンドオフがスムーズになる。

---

### 候補6: plsnt を MCP Server として動作 (`plsnt mcp serve`)

| 項目 | 評価 |
|------|------|
| 実装コスト | **High** |
| 利用頻度 | Low |
| MCP 相乗効果 | **Negative（競合）** |
| **推奨度** | **Won't** |

**理由**: Pleasanter 本体に MCP Server が内蔵されている以上、
plsnt で同じものを作るのは二重開発。本家の方が Pleasanter 内部に直接アクセスでき、
Translator も使えるため、どうやっても本家が優位。リソースの無駄遣い。

---

### 候補7: 相互補完コマンド

| 項目 | 評価 |
|------|------|
| 実装コスト | High |
| 利用頻度 | Low |
| MCP 相乗効果 | Medium |
| **推奨度** | **Won't** |

**理由**: plsnt のコマンドと MCP のツール呼び出しは構造が異なり、
変換レイヤーの保守コストが高い。MCP のリソース仕様は Pleasanter 側で
提供されるべきであり、plsnt が二重管理する意味がない。

---

## 決定ロードマップ

### Phase A: 即時（対応不要）

| # | 項目 | 状態 |
|---|------|------|
| 1 | **schema の日本語名解決** | **既に実装済み**。LabelText + ColumnName の両方を出力中 |

### Phase B: 短期（次の機能追加時に併せて）

| # | 項目 | 内容 | 見積もり |
|---|------|------|---------|
| 2 | **ビュー定義の検証** | `site update` でSiteSettings.Viewsが操作可能か実機検証 | 1日 |
| 3 | **ビューのテンプレート統合** | 検証結果に応じて、テンプレートYAMLにビュー設定ステップを追加 | 2-3日 |

### Phase C: 中期（ユーザーからの要望があれば）

| # | 項目 | 内容 | トリガー |
|---|------|------|---------|
| 4 | **scaffold サマリー出力** | batch run 完了時の MCP 向けサマリー | MCP を有効化して実際に併用し始めた時 |
| 5 | **schema diff** | 2サイトのスキーマ差分表示 | 環境間のスキーマ同期が課題になった時 |

### 実装しないもの

| 項目 | 理由 |
|------|------|
| plsnt mcp serve | 本家内蔵 MCP と競合。二重開発 |
| plsnt mcp enable/disable | サーバー側ファイル操作は plsnt の責務外 |
| MCP ログ分析 | Pleasanter UI の McpLogs 画面で十分 |
| 相互補完コマンド | 変換レイヤーの保守コスト > 利用価値 |

---

## 結論

**plsnt は MCP の代替ではなく、MCP の補完である。**

- MCP は「AI が Pleasanter を操作する窓口」
- plsnt は「自動化・バッチ・移行・構築の道具」

両者の接点（schema 情報、scaffold → MCP ハンドオフ）での摩擦を減らす拡張だけに
集中し、MCP と機能が被る領域には手を出さない。
