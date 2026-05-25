# エージェントの MCP / CLI 使い分けガイド

> 作成日: 2026-03-14
> 対象: Claude Code（AI エージェント）が Pleasanter を操作する際の判断基準

---

## 前提

Claude Code は Pleasanter に対して 2 つの経路を持つ:

1. **plsnt CLI** — Bash ツール経由で実行。スキル/コマンド/バッチと完全連携
2. **Pleasanter MCP** — MCP ツールとして直接呼び出し。日本語変換・ビュー管理対応

両方が使える環境で、エージェントはどちらを選ぶべきか？

---

## 判断フローチャート

```
ユーザーの指示を受け取る
  │
  ├─ スキル/コマンドが呼ばれた？（/build-app-full, /deploy-scripts 等）
  │   → CLI 一択。スキルは plsnt CLI 前提で設計されている。
  │
  ├─ バッチ実行？（テンプレート展開、複数テーブル構築）
  │   → CLI 一択。バッチエンジンは plsnt 固有機能。
  │
  ├─ 一括操作？（CSV インポート、bulk-delete、upsert）
  │   → CLI 一択。MCP にはこれらのツールがない。
  │
  ├─ サイト作成・削除・複製？
  │   → CLI 一択。MCP にはサイト作成ツールがない。
  │
  ├─ ファイルが関係する？（スクリプトデプロイ、CSV 読み込み）
  │   → CLI 一択。MCP はファイル I/O 不可。
  │
  ├─ ループ処理？（N 件を 1 件ずつ処理）
  │   → CLI 一択。MCP にはループ構造がない。
  │
  ├─ 結果をスクリプトに保存・再利用する？
  │   → CLI 一択。MCP の操作はスクリプト化できない。
  │
  ├─ 自然言語で「〇〇を見せて」「探して」？
  │   → MCP が自然。日本語列名変換が活きる。
  │   → ただし CLI でも可能（schema で列名確認後に record list）
  │
  ├─ ビューの作成・保存・共有？
  │   → MCP が適切。CLI にはビュー管理コマンドがない。
  │
  ├─ メール送信？
  │   → MCP 一択。CLI にはメール機能がない。
  │
  └─ 上記のどれにも該当しない単純な CRUD？
      → どちらでもよい。ただし他の操作と組み合わせるなら CLI。
```

**簡潔に言えば:**

| 操作の性質 | 選択 |
|-----------|------|
| **構築・自動化・一括** | CLI |
| **探索・確認・対話** | MCP |
| **迷ったら** | CLI（より確実で再現可能） |

---

## 具体的な使い分けパターン

### パターン 1: 構築ワークフロー（CLI 一択）

```
ユーザー: 「シフト管理アプリを作って」

エージェントの行動:
  1. /scaffold-app → plsnt batch run --dry-run（確認）
  2. plsnt batch run scaffold-shift-management-v3.yaml（構築）
  3. /deploy-scripts → deploy-all.sh（スクリプト）
  4. /seed-data → seed-data.sh（サンプルデータ）

MCP の出番: なし
```

### パターン 2: データ探索（MCP が自然、CLI でも可能）

```
ユーザー: 「未完了のシフトを教えて」

MCP を使う場合:
  1. GetSiteIdByTitle("シフト割当")
  2. CreateViewJson(columnFilterHash: {"ステータス": "予定"})
  3. GetItems → 日本語で表示

CLI を使う場合:
  1. plsnt schema 32450 -o json  → 列名を確認
  2. plsnt record list --site-id 32450 \
       --view '{"ColumnFilterHash":{"Status":"100"}}' -o table

どちらでもよいが、初見のテーブルなら MCP の日本語変換が便利。
既知のテーブルなら CLI の方が高速。
```

### パターン 3: ハイブリッド（MCP で探索 → CLI で実行）

```
ユーザー: 「顧客マスタのデータを整理して、不要なレコードを消して」

エージェントの行動:
  1. [MCP] GetSiteIdByTitle("顧客マスタ") → SiteID 取得
  2. [MCP] CreateViewJson + GetItems → データ確認（日本語表示）
  3. ユーザーに確認: 「これらを削除しますか？」
  4. [CLI] plsnt record bulk-delete --site-id 32445 \
           --ids "123,456,789" --confirm
           → 一括削除は CLI でしかできない

探索は MCP、実行は CLI。
```

### パターン 4: ビュー管理（MCP 一択）

```
ユーザー: 「今日のシフト一覧をビューとして保存して」

エージェントの行動:
  1. [MCP] CreateViewJson(siteId: 32450,
           columnFilterHash: {"日付": "Today"})
  2. [MCP] AddView(siteId: 32450, viewName: "今日のシフト", viewJson: ...)

CLI では不可能（ビュー管理コマンドがない）。
```

### パターン 5: レポート生成（CLI 一択）

```
ユーザー: 「月次レポートを生成して」

エージェントの行動:
  1. /generate-report → plsnt batch run monthly-report.yaml
  2. plsnt record list で集計データ取得
  3. jq で加工
  4. ファイルに出力

MCP の出番: なし。ファイル I/O と集計処理が必要。
```

### パターン 6: メール通知（MCP 一択）

```
ユーザー: 「期限超過のタスクの担当者にメールして」

エージェントの行動:
  1. [MCP] notify-overdue-records プロンプトが起動
  2. [MCP] CreateViewJson(overdue: true) → GetItems
  3. [MCP] SendEmail(to: 担当者, title: "期限超過通知", body: ...)

CLI にはメール機能がない。
```

---

## エージェント設計への組み込み方

### CLAUDE.md への追記案

```markdown
## Pleasanter 操作の使い分け

### plsnt CLI（Bash 経由）を使う場面
- スキル/コマンドの実行（/build-app-full, /deploy-scripts 等）
- バッチ実行（テンプレート展開）
- 一括操作（upsert, bulk-delete, import）
- サイト作成・削除・複製
- スクリプトデプロイ
- ループ処理、ファイル I/O を伴う操作
- 再現可能な操作手順が必要な場合

### Pleasanter MCP を使う場面
- 自然言語での対話的なデータ探索
- ビューの作成・保存・管理
- メール送信
- 初見のテーブル構造の理解（日本語列名変換）

### 判断に迷ったら
CLI を優先する。CLI は再現可能で、エラーハンドリングが確実。
```

### スキル定義での指示

既存のスキル（app-build-workflow, bulk-operations 等）は変更不要。
全て plsnt CLI 前提で正しく設計されている。

新規スキルを作る場合のテンプレート:

```yaml
---
name: pleasanter-explore
description: Pleasanter のデータを対話的に探索する
allowed-tools: Bash, Read, Write, mcp__pleasanter__*
---

# データ探索スキル

## ツール選択ルール

1. テーブル構造の理解: MCP の GetSite + Translator が便利
2. データ検索・フィルタ: MCP の CreateViewJson + GetItems が自然
3. 見つけたデータに対する操作:
   - 単一レコード更新 → MCP の UpdateItem でも CLI でもよい
   - 複数レコード一括操作 → CLI の plsnt record bulk-delete / upsert
   - ファイル出力 → CLI の plsnt record list -o csv > file.csv
```

---

## コンテキスト消費の考慮

エージェントが MCP と CLI を併用する際のコンテキスト管理:

| 操作 | コンテキスト消費 | 注意点 |
|------|-----------------|--------|
| MCP ツール定義のロード | 数千トークン（常時） | Tool Search で遅延ロード可能 |
| MCP ツールのレスポンス | 最大 25,000 トークン/回 | 大量レコードで溢れる |
| plsnt CLI の Bash 実行 | 出力サイズ分 | `-o count` や `-o ids` で最小化可能 |
| plsnt CLI の構造化エラー | 数百トークン | stderr の JSON エラー |

**コンテキスト節約のベストプラクティス**:

```bash
# 悪い例: 全レコードを JSON で取得（コンテキスト大量消費）
plsnt record list --site-id 32450 -o json --all-pages

# 良い例: まず件数を確認
plsnt record list --site-id 32450 -o count

# 良い例: ID だけ取得してからループ
for id in $(plsnt record list --site-id 32450 -o ids); do
  plsnt record get $id -o json | jq '.Title'
done

# 良い例: MCP で探索（表示値変換付きで人間にも見やすい）
# → 件数が少ない対話的探索には MCP が効率的
```

---

## MCP が使えない環境のフォールバック

MCP が有効化されていない場合（Pleasanter 1.4.x、MCP 未設定等）、
全ての操作を plsnt CLI で行う。

| MCP 操作 | CLI フォールバック |
|---------|------------------|
| GetSiteIdByTitle("顧客") | `plsnt site search --parent-id 32085 --keyword 顧客` |
| CreateViewJson(日本語) | `plsnt schema` で列名確認 → `--view` JSON 手動構築 |
| AddView | `plsnt site update` で SiteSettings.Views を更新（要検証） |
| SendEmail | 代替手段なし（Pleasanter UI で手動） |
| GetUserIdByName("田中") | `plsnt user list -o json \| jq '.[] \| select(.Name=="田中")'` |

---

## まとめ: 3 つの原則

### 原則 1: スキル/コマンドは CLI

スキルやコマンド（`/build-app-full`, `/deploy-scripts` 等）から呼ばれる操作は
全て plsnt CLI を使う。これらは CLI 前提で設計されており、
バッチエンジン・ステップ出力参照・シェルスクリプトとの連携が不可欠。

### 原則 2: 対話的探索は MCP

ユーザーが自然言語で「〇〇を見せて」「探して」と言った場合は MCP が自然。
日本語列名変換と表示値変換が、対話体験を大幅に改善する。

### 原則 3: 迷ったら CLI

CLI は常に動く（MCP 未設定でも）、再現可能（スクリプト化できる）、
エラーハンドリングが確実（構造化エラー + exit code）。
MCP でしかできないこと（ビュー管理、メール）以外は CLI を選ぶ。
