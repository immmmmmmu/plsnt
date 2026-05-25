# MCP Server vs plsnt CLI：Claude Code スキル/コマンド連携の実態

> 調査日: 2026-03-14
> 核心的な問い: MCP ツールは Claude Code のスキル/コマンド/バッチと連携できるのか？

---

## 結論（先に）

**plsnt CLI（Bash 経由）の方が、スキル/コマンド/バッチとの連携において圧倒的に有利。**

MCP ツールは「対話的な1回きりの操作」には向くが、
「スキルで定義された複数ステップの自動化ワークフロー」には構造的に向かない。

---

## 1. 両者の連携メカニズムの違い

### plsnt CLI の連携方式

```
スキル定義（SKILL.md）
  │
  │ プロンプトで指示: 「plsnt record list を実行せよ」
  │
  ▼
Claude Code が Bash ツールを呼び出す
  │
  │ plsnt record list --site-id 123 -o json
  │
  ▼
stdout に JSON が出力される
  │
  │ jq で加工、変数に格納、次のコマンドに渡す
  │
  ▼
次の plsnt コマンドへ（確実なパイプライン）
```

**特長**:
- **Bash のフル機能が使える**: ループ、条件分岐、パイプ、jq、変数展開
- **出力サイズ制限なし**: CLI の stdout に制限はない
- **スクリプト化可能**: deploy-all.sh、seed-data.sh のような複合スクリプト
- **バッチエンジン統合**: YAML テンプレートでステップ出力参照 `{{step.Key}}`

### MCP ツールの連携方式

```
スキル定義（SKILL.md）
  │
  │ allowed-tools: mcp__pleasanter__*
  │ プロンプトで指示: 「GetItems を使え」
  │
  ▼
Claude Code が MCP ツールを呼び出す
  │
  │ mcp__pleasanter__GetItems(siteId: 123)
  │
  ▼
MCP レスポンスがコンテキストに入る（最大 25,000 トークン）
  │
  │ Claude が結果を「読んで」次のアクションを「判断」する
  │
  ▼
次の MCP ツール呼び出し（Claude の推論に依存）
```

**特長**:
- **日本語列名変換が自動**: Translator が内部名を解決
- **対話的操作に最適**: 人間が「〇〇を見せて」と言えば動く
- **ビュー管理が充実**: 保存ビューの CRUD が可能

---

## 2. スキル/コマンドとの連携比較

### 2.1 複数ステップの自動化

```
シナリオ: 5テーブルを作成し、リンクを設定し、スクリプトをデプロイする
```

**plsnt CLI の場合**:
```bash
# バッチエンジンが依存関係を解決して実行
plsnt batch run templates/scaffold-shopping.yaml --var folder_id=32085

# ステップ出力参照が自動で動く
# create-orders の SiteID → create-details の Links に自動展開
```

**MCP の場合**:
```
Claude に「5テーブル作ってリンク設定して」と依頼
  → Claude が GetSiteIdByTitle で既存確認
  → Claude が... あれ、CreateSite ツールがない
  → MCP にはサイト作成ツールがない
  → 行き詰まり
```

**判定: plsnt の圧勝。** MCP にはサイト作成・削除すらない。

### 2.2 一括データ操作

```
シナリオ: CSV から 500 件のレコードをインポート
```

**plsnt CLI**:
```bash
plsnt migrate generate-mapping --file data.csv --site-id 123 --output mapping.yaml
plsnt migrate execute --file data.csv --site-id 123 --mapping mapping.yaml
# 1コマンドで完了
```

**MCP**:
```
UpdateItem を 500 回呼ぶ？
  → 25,000 トークン制限でレスポンスが切れる
  → レート制限（デフォルト 30 req/60s）に引っかかる
  → 現実的ではない
```

**判定: plsnt の圧勝。** MCP は大量操作に向かない。

### 2.3 スクリプトデプロイ

```
シナリオ: 12 個の JS/CSS ファイルを Pleasanter にデプロイ
```

**plsnt CLI + deploy-all.sh**:
```bash
# 既存 SiteSettings を取得
current=$(plsnt site get $SITE_ID -o json)

# JS ファイルを読み込んで Scripts 配列を構築
for file in *.js; do
  body=$(jq -Rs . < "$file")
  # 配列に追加
done

# 全体をマージして更新
plsnt site update $SITE_ID --json "{\"SiteSettings\":$merged}"
```

**MCP**:
```
UpdateSite ツールはある
  → しかし JS ファイルをどう渡す？
  → MCP にはファイル読み込みツールがない
  → Claude Code の Read ツールで読んで、MCP の UpdateSite に渡す？
  → 12 ファイル分のコンテキスト消費が膨大
  → SiteSettings 全体上書き問題も自力で解決する必要がある
```

**判定: plsnt + シェルスクリプトの圧勝。**

### 2.4 対話的なレコード検索

```
シナリオ: 「未完了のシフトを教えて」
```

**plsnt CLI**:
```bash
plsnt record list --site-id 32450 \
  --view '{"ColumnFilterHash":{"Status":"100"}}' -o table
# ユーザーが ClassA = 現場ID、ClassB = 警備員ID と知っている必要がある
```

**MCP**:
```
search-records プロンプトが起動
  → CreateViewJson(siteId: 32450, columnFilterHash: {"ステータス": "予定"})
  → 日本語「予定」→ コード値「100」に自動変換
  → GetItems で取得
  → DisplayTranslator が ClassA の ID → 「渋谷」に変換して表示
```

**判定: MCP の勝利。** 日本語名変換と表示値変換は対話的操作で真価を発揮。

### 2.5 ビュー管理

```
シナリオ: カレンダービューを保存して共有
```

**plsnt CLI**: ビュー管理コマンドなし（site update で可能かもしれないが未検証）

**MCP**: CreateViewJson → AddView → 完了

**判定: MCP の勝利。**

---

## 3. 構造的な制約の整理

### MCP がスキル/コマンドと連携しにくい理由

| 制約 | 説明 | plsnt CLI では |
|------|------|----------------|
| **パイプライン不可** | MCP の出力を次の MCP 入力に直接渡せない。Claude の推論を経由する | Bash パイプで `\|` や `$()` で直接渡せる |
| **出力トークン制限** | 25,000 トークン上限。大量レコードの一覧取得で切れる | stdout に制限なし |
| **ループ処理不可** | MCP ツールを for ループで回す仕組みがない | `for id in $(plsnt ... -o ids)` |
| **ファイルI/O不可** | MCP ツールはファイル読み書きできない | `jq -Rs . < file.js` |
| **エラーハンドリング** | MCP エラーは Claude の推論に依存 | `$?` や `2>/dev/null` で制御 |
| **冪等性の保証** | MCP の再実行は Claude が判断 | バッチエンジンが管理 |
| **バッチ連携** | MCP ツールは YAML バッチに組み込めない | `plsnt batch run` で実行 |

### plsnt CLI がスキル/コマンドと連携しやすい理由

| 強み | 説明 |
|------|------|
| **Bash ツール経由の呼び出し** | スキルの `allowed-tools: Bash` で全 plsnt コマンドが使える |
| **JSON 出力の機械可読性** | `-o json` の出力を jq で確実にパースできる |
| **構造化エラー** | stderr に JSON エラーが出るため、Claude が原因を正確に把握 |
| **スクリプト化** | deploy-all.sh 等の複合スクリプトをスキルから呼べる |
| **バッチエンジン** | YAML テンプレートでステップ間のデータ受け渡しが確実 |
| **マルチプロファイル** | `-p staging` でスキル内から環境切り替え可能 |

---

## 4. 現在の plsnt スキル/コマンド連携の実績

### 確立されたワークフロー

```
/build-app-full shift-management-v3 --folder-id 32085
  │
  ├─ Phase 1: テーブル設計（スキル: column-design, relational-modeling）
  │    → plsnt schema でカラム確認
  │
  ├─ Phase 2: テンプレート YAML 作成
  │    → plsnt batch run --dry-run で確認
  │
  ├─ Phase 3: テーブル構築
  │    → plsnt batch run scaffold-xxx.yaml
  │    → ステップ出力で SiteID 自動取得
  │
  ├─ Phase 4: env.sh 生成
  │    → SiteID を環境変数として記録
  │
  ├─ Phase 5: UI スクリプト作成（スキル: pleasanter-scripts）
  │    → Claude Code が JS/CSS を生成
  │
  ├─ Phase 6: デプロイ（コマンド: /deploy-scripts）
  │    → deploy-all.sh が plsnt site get → jq マージ → plsnt site update
  │
  ├─ Phase 7: サンプルデータ（コマンド: /seed-data）
  │    → seed-data.sh が plsnt record create を連続実行
  │
  └─ Phase 8: レビュー（エージェント: script-reviewer）
       → plsnt site get でスクリプト確認
```

**このワークフロー全体が、MCP ツールでは実現不可能。**

理由:
1. サイト作成ツールがない（Phase 3 が不可能）
2. ファイルI/Oがない（Phase 5-6 が不可能）
3. バッチエンジンがない（Phase 3 のステップ出力参照が不可能）
4. 一括レコード作成がない（Phase 7 が非効率）

---

## 5. MCP が活きる場面（plsnt では難しいこと）

| 場面 | MCP の優位性 | plsnt の代替手段 |
|------|------------|----------------|
| 「顧客名で検索して」 | 日本語名 → ClassA 自動変換 | `plsnt schema` で LabelText 確認後に手動指定 |
| ビューの保存・共有 | AddView / UpdateView | site update の SiteSettings.Views で可能（未検証） |
| メール通知 | SendEmail ツール | なし |
| 初見のテーブル探索 | 日本語で「何があるか見せて」 | `plsnt schema` + `plsnt record list` |
| ReadOnly アクセス制御 | ReadOnlyMode パラメータ | なし（API キーの権限で制御） |

---

## 6. 推奨する併用アーキテクチャ

```
┌──────────────────────────────────────────────────────────────┐
│                     Claude Code セッション                     │
│                                                               │
│  ┌─────────────────────┐     ┌─────────────────────────────┐ │
│  │    対話的操作         │     │    自動化・構築              │ │
│  │    (MCP Server)      │     │    (plsnt CLI)              │ │
│  │                      │     │                             │ │
│  │  「何がある？」       │     │  /build-app-full            │ │
│  │  「未完了を見せて」   │     │  /deploy-scripts            │ │
│  │  「田中さんに割当」   │     │  /seed-data                 │ │
│  │  「ビューを保存」     │     │  /migrate-data              │ │
│  │  「メール送って」     │     │  /check-integrity           │ │
│  │                      │     │  /generate-report           │ │
│  │  スキル連携: 弱い     │     │  スキル連携: 強い            │ │
│  │  バッチ連携: 不可     │     │  バッチ連携: 完全            │ │
│  │  ループ処理: 不可     │     │  ループ処理: 完全            │ │
│  └──────────┬───────────┘     └──────────────┬──────────────┘ │
│             │                                │               │
│             │ JSON-RPC 2.0                    │ REST API      │
│             │ /mcp                            │ /api/items/   │
│  ┌──────────▼────────────────────────────────▼──────────────┐ │
│  │                  Pleasanter Server                        │ │
│  └───────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘

MCP の役割: 探索・確認・対話（人間の「目」）
plsnt の役割: 構築・自動化・一括操作（人間の「手」）
```

---

## 7. plsnt 側で検討すべき拡張（再評価）

前回の検討（mcp-coexistence-roadmap.md）を踏まえ、
スキル/コマンド連携の観点で再評価する。

### 7.1 ビュー管理（優先度: Should → Must に格上げ）

**理由**: バッチテンプレートでテーブル scaffold する際、ビュー定義も含めたい。
現状はテーブル作成後に UI で手動設定しており、自動化の穴になっている。

**実装方針**: `site update` で SiteSettings.Views が操作可能か検証。
可能なら専用コマンド不要で、テンプレート YAML のステップとして追加するだけ。

### 7.2 scaffold サマリーの MCP 向け出力（優先度: Should）

**理由**: plsnt でテーブルを構築した後、MCP で対話的に操作したい場面がある。
その際に「SiteID いくつ？列名なに？」の情報が必要。

**実装方針**: `plsnt batch run` 完了時に以下を出力するオプション:
```
作成されたサイト:
  顧客マスタ (32445): ClassA=顧客名, ClassB=電話番号
  注文 (32446): ClassA=顧客(Link:32445), DateA=注文日
```

### 7.3 MCP Server の接続設定ヘルパー（優先度: Could）

**理由**: Pleasanter MCP Server を Claude Code に接続する設定を生成。

**実装方針**:
```bash
plsnt config mcp-setup
# → .mcp.json に以下を出力:
# {
#   "mcpServers": {
#     "pleasanter": {
#       "type": "http",
#       "url": "http://localhost/mcp",
#       "headers": { "X-API-Key": "***" }
#     }
#   }
# }
```

既存のプロファイル設定（URL + API キー）から自動生成できる。

---

## 8. 最終的な判断

| 問い | 回答 |
|------|------|
| MCP は plsnt を置き換えるか？ | **No.** 自動化・構築・一括操作では plsnt が不可欠 |
| plsnt は MCP を置き換えるか？ | **No.** 対話的操作・日本語変換・ビュー管理では MCP が優位 |
| スキル/コマンドとの連携はどちらが強い？ | **plsnt CLI が圧倒的に強い。** Bash パイプ、ループ、ファイル I/O が使える |
| MCP をスキルから使えるか？ | 使えるが制約が多い。パイプライン不可、出力制限、ループ不可 |
| 両者の接点は？ | plsnt で構築 → MCP で対話的に操作。schema 情報が接点 |

---

## 9. MCP の制約は Pleasanter 固有か？ — 一般論としての検証

### 結論: MCP プロトコル自体の構造的制約であり、全 MCP サーバーに共通する

前節までの分析で挙げた MCP の制約（パイプライン不可、ループ不可、出力トークン制限、
ファイル I/O 不可、推論依存のチェイニング）は、Pleasanter MCP 固有ではなく、
**MCP プロトコルと Claude Code の統合アーキテクチャに起因する**。

ただし、ドメインによって「痛み」の度合いが大きく異なる。

### MCP プロトコル共通の制約

| 制約 | 原因 | CLI（Bash）なら |
|------|------|----------------|
| **チェイニングが推論依存** | ツール A → B の受け渡しは Claude の判断を経由 | `$(cmd_a) \| cmd_b` で確実 |
| **ループ処理がない** | MCP プロトコルにループ構造がない | `for id in $(...)` |
| **出力トークン制限** | Claude Code のデフォルト 25,000 トークン上限 | stdout に制限なし |
| **コンテキスト消費** | ツール定義だけで数千〜数万トークン消費 | `--help` は必要時のみ |
| **高 effort が必要** | 中 effort では MCP 出力を読み飛ばす・誤読する報告あり | Bash の `$?` で確実 |
| **ファイル I/O 不可** | MCP ツール単体ではローカルファイル操作不可 | 当然可能 |
| **スクリプト化不可** | MCP の操作手順を `.sh` に保存して再実行できない | deploy-all.sh 等 |

> 参考: [MCP Tools Effort Levels](https://docs.bswen.com/blog/2026-03-13-mcp-tools-effort/)
> — 「Claude would claim a tool returned data that it didn't」
> 「complex nested JSON from MCP tools got partially processed」

### ドメイン別の影響度

**MCP の制約が顕在化しないドメイン（MCP が有利）**:

| MCP サーバー | なぜ痛くないか |
|-------------|--------------|
| **Database（PostgreSQL 等）** | SQL 自体がパイプラインでありループ。1 回の `query` 呼び出しで JOIN + WHERE + GROUP BY が完結 |
| **GitHub** | PR 取得・コメント追加は 1 回の呼び出しで完結。`gh` CLI と大差ない |
| **Web Search / Fetch** | 検索して結果を返す。チェイニング不要 |
| **Slack / メール** | メッセージ送信は 1 回の呼び出しで完結 |

**共通点: 1 回のツール呼び出しで操作が完結するドメイン。**

**MCP の制約が痛いドメイン（CLI が有利）**:

| MCP サーバー | なぜ痛いか |
|-------------|----------|
| **Pleasanter** | CRUD の組み合わせ + リンク設定 + スクリプトデプロイ = 多段ステップ |
| **AWS / GCP / Azure** | インフラ構築は 10-50 ステップのオーケストレーション。Terraform / CLI の方が確実 |
| **CI/CD** | パイプライン定義は YAML ファイルであり、MCP で操作する意味が薄い |
| **ファイル操作中心** | リファクタリング、コード生成。Read/Write/Edit ツールの方が直接的 |
| **データ移行** | CSV 読み込み → 変換 → 一括投入。ファイル I/O + ループが必須 |

**共通点: 複数ステップの連鎖・ループ・ファイル I/O が必要なドメイン。**

### 判断基準: MCP vs CLI の選択フレームワーク

```
あなたの操作は…

  1 回の呼び出しで完結する？
    → Yes: MCP で十分。CLI と大差なし。
    → No: 次へ

  複数ステップの連鎖が必要？
    → Yes: CLI が有利。Bash パイプ/変数/バッチエンジン。
    → No: 次へ

  ループ処理（N 件を 1 件ずつ）が必要？
    → Yes: CLI が圧倒的に有利。
    → No: 次へ

  ローカルファイルの読み書きが必要？
    → Yes: CLI 一択。MCP は不可能。
    → No: 次へ

  操作手順の再現・共有が必要？
    → Yes: CLI（スクリプト化）。MCP は再現不可能。
    → No: MCP でも CLI でもどちらでもよい。
```

### plsnt にとっての意味

Pleasanter の操作は上記フレームワークの「CLI が有利」に全て該当する：
- テーブル構築 = 多段ステップ
- データ移行 = ファイル I/O + ループ
- スクリプトデプロイ = ファイル I/O + マージ
- テンプレート展開 = バッチエンジン + ステップ参照

**plsnt CLI + スキル/コマンドの設計判断は正しかった。**
MCP が登場しても、この優位性は MCP プロトコルの構造的制約により維持される。

MCP は「対話的な探索・確認」という補完的な役割で価値を発揮する。
両者は競合ではなく、異なるレイヤーの道具である。

```
MCP = 対話的操作（1 回完結、日本語変換、ビュー管理）
CLI = 自動化・構築（多段ステップ、ループ、ファイル I/O、再現性）
```
