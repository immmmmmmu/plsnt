# ADR-0001: Pleasanter MCP と plsnt MCP の棲み分け

- **ステータス**: Accepted
- **作成日**: 2026-03-14（初版） / 2026-04-16（判定改訂） / 2026-04-30（ADR としてリファクタ）
- **決定者**: plsnt 開発チーム
- **関連 Issue**: [#1](https://github.com/immmmmmmu/plsnt/issues/1) / [#2](https://github.com/immmmmmmu/plsnt/issues/2)
- **関連ドキュメント**:
  - `docs/proposal-mcp-coexistence-features.md`
  - `docs/mcp-coexistence-roadmap.md`
  - `docs/pleasanter-mcp-vs-plsnt-cli.md`
  - `docs/mcp-vs-cli-skill-integration.md`

---

## Context

Pleasanter 1.5.0 で本体に MCP (Model Context Protocol) Server が内蔵された。これにより以下 16 ツールが Claude Code / Claude Desktop から直接呼べるようになった:

- レコード参照系: `GetItem`, `GetItems`, `CreateViewJson`
- レコード書込: `CreateUpdateItemJson` + `UpdateItem`（2ステップ）
- サイト参照: `GetSite`, `GetSiteIdByTitle`
- ビュー CRUD: 6 ツール
- メール: `SendEmail`
- ユーザー: `GetUsers`, `GetUserIdByName`

一方 plsnt CLI は Pleasanter REST API の全操作を 100+ コマンドでカバーしてきた。両者の機能は一部重複するため、plsnt 自身が MCP Server を提供すべきか、提供する場合どこまでカバーするか、という設計判断が必要になった。

### 当初の仮説（Roadmap 2026-03-14）

> plsnt は **バッチ処理・一括操作・自動化・再現性** に強い。Pleasanter MCP は **対話的操作・日本語変換・ビュー管理** に強い。重複を避け、接点での摩擦を減らす拡張に集中する。

この仮説に基づくと「plsnt は CLI 専念、MCP は本家に委譲」が結論になり、当時の `requirements.md` も `plsnt mcp serve` をスコープ外と明示していた。

### スキル依存分析で見えた事実（2026-04-16 判定改訂）

`.claude/skills/` 群および `templates/`, `scripts/` の実コード解析（git grep）で以下が判明:

| コマンド | 使用箇所数 | 依存スキル数 |
|---|---|---|
| `record list` | **281+ 件** | 12+ スキル |
| `record create` | 181+ 件 | 10+ スキル |
| `site get` | 多数（デプロイ基盤） | 8+ スキル |
| `record get` | 30+ 件 | 6+ スキル |

スキルが内部で連続的に使うコマンドのほとんどは、**Pleasanter MCP に同等機能があっても plsnt 側にも実装されている**。スキル本体を 2 つの MCP の混在で書くと、以下の摩擦が現れる:

| 問題 | 影響 |
|---|---|
| パラメータ形式の不一致 | Pleasanter MCP `referenceId: integer` / plsnt MCP `record_id: string` |
| 2 ステップ vs 1 ステップ | 更新が `CreateUpdateItemJson` → `UpdateItem`、作成が `record_create` 1 ステップ |
| コンテキスト切り替え | 1 スキルで 2 MCP を使うとエージェントのツール選択判断コストが増える |
| SiteSettings 形式の不整合 | `GetSite`（生 JSON）→ `site update`（plsnt 形式）の変換が必要 |
| エラー分断 | plsnt の `CLIError`（code, message, suggestion）と Pleasanter MCP のエラー形式が異なる |
| 出力形式の不統一 | plsnt: 整形済み JSON / Pleasanter MCP: 生 API レスポンス |

つまり **「重複を避ける」原則を厳格に適用すると、スキル側に膨大なアダプタコードが必要になる**。これは保守不能。

---

## Decision

**plsnt MCP は「重複を避ける道具」ではなく「スキルとスクリプトのための統一インターフェース」として設計する。**

### 設計原則

1. **統一インターフェース優先** — Pleasanter MCP と機能が重複しても、スキル連携の一貫性・SiteSettings の形式統一・パラメータ命名規約の統一を優先する
2. **B 判定（保守モード）の維持** — スキル依存が 2 以下の操作（user list/get, group/dept CRUD, access, user import）は plsnt MCP に **含めない**。CLI のみで提供
3. **Pleasanter MCP 専用領域の尊重** — ビュー CRUD・メール送信・日本語列名自動変換は Pleasanter MCP に委譲。plsnt 側で再実装しない
4. **stdio 専念（Phase 1〜3）** — Claude Desktop / Claude Code 両対応のため stdio のみ。HTTP は本家 Pleasanter MCP が既に提供しており二重化しない

### plsnt MCP 公開ツール（20 個）

#### Tier 1: 差別化ツール（6 個）

| ツール | 対応 CLI | 役割 |
|---|---|---|
| `schema_get` | `schema` | カラム定義の構造化出力（型・選択肢・制約） |
| `batch_run` | `batch run` | YAML バッチ実行 |
| `site_create` | `site create` | サイト作成（SiteSettings 自動付与） |
| `record_upsert` | `record upsert` | 一括 upsert |
| `record_import` | `record import` | CSV 一括インポート |
| `config_test` | `config test` | 接続確認 |

#### Tier 2: CRUD + サイト管理（10 個）

| ツール | 対応 CLI | 役割 |
|---|---|---|
| `record_list` | `record list` | `--all-pages` `--view` `--fields` 対応 |
| `record_get` | `record get` | `--fields` `--output` 対応 |
| `record_create` | `record create` | Pleasanter MCP に非対応（plsnt 固有） |
| `record_update` | `record update` | 1 ステップ更新（Pleasanter MCP は 2 ステップ） |
| `record_delete` | `record delete` | Pleasanter MCP に非対応（plsnt 固有） |
| `record_bulk_delete` | `record bulk-delete` | 一括削除（plsnt 固有） |
| `site_get` | `site get` | デプロイパターンの起点 |
| `site_update` | `site update` | SiteSettings 全体上書き |
| `site_copy` | `site copy` | サイト複製 |
| `site_search` | `site search` | 子サイト検索 |

#### Tier 3: ワークフロー・移行（4 個）

| ツール | 対応 CLI | 役割 |
|---|---|---|
| `workflow_deploy` | `workflow deploy` | ワークフローアプリ展開 |
| `workflow_master` | `workflow master` | マスタデータ投入 |
| `workflow_export` | `workflow export` | 申請データ CSV エクスポート |
| `migrate_execute` | `migrate execute` | データ移行実行 |

### plsnt MCP に **含めない** 操作（B 判定）

| 操作 | 理由 |
|---|---|
| `user list` / `user get` | スキル依存が 2 以下。Pleasanter MCP に `GetUsers` あり |
| `user create/update/delete` / `user import` | スキル依存が低く、CLI でカバー |
| `group *` / `dept *` | スキル依存が 1 以下。低頻度 |
| `access *` | Access DB 読み取りは特殊用途。MCP より CLI が自然 |
| ビュー CRUD | Pleasanter MCP が完備（6 ツール） |
| メール送信 | Pleasanter MCP `SendEmail` に委譲 |

---

## ルーティングルール（運用時）

### plsnt MCP を使う

- **スキル・スクリプト連携の起点** — `/scaffold-app`, `/build-app-full`, `/deploy-scripts`, `/seed-data`, `/check-integrity` など
- **CRUD 全操作（特に create / delete / 一括系）** — Pleasanter MCP に非対応
- **`site get` → 加工 → `site update`** — SiteSettings 形式の一貫性
- **YAML バッチ実行・ワークフロー・データ移行**

### Pleasanter MCP を使う

- **対話的データ探索** — 自然言語で「このテーブル何件ある？」
- **ビュー作成・管理** — `AddView`, `CreateViewJson` の 6 ツール
- **メール送信** — `SendEmail`
- **不慣れなテーブルの理解** — 日本語列名 ↔ 内部名（ClassA 等）の自動変換

### plsnt CLI を直接使う（MCP を経由しない）

- **シェルスクリプト・CI/CD 自動化** — 再現性が必要な定型処理
- **大量データ処理** — `--all-pages` でストリーミング、`-o csv` で jq 連携
- **B 判定操作** — user / group / dept CRUD, access import
- **複数プロファイル切替** — `--profile production` などの環境制御

### Hybrid Pattern

「Pleasanter MCP で対象を特定 → plsnt MCP / CLI で実行」が基本パターン。

例: 「先月作成された未承認の経費申請を一括クローズ」
1. Pleasanter MCP `CreateViewJson` でフィルタ条件をビューとして可視化・確認
2. plsnt MCP `record_list` で対象 ID を取得
3. plsnt MCP `record_bulk_delete` または `record_update` で一括処理

---

## Consequences

### Positive

- **スキルの一貫性** — `.claude/skills/` は plsnt MCP / CLI 一本で書ける。アダプタコード不要
- **エージェントの判断負荷低減** — 「create / delete / 一括 / バッチ なら plsnt」「探索 / ビュー / メール なら Pleasanter」と明確
- **Claude Desktop 対応** — stdio なので `mcp-remote` 不要
- **既存スキル資産の保護** — 281+ 件の `record list` 呼び出しを書き換えずに済む

### Negative / Trade-off

- **コードの重複** — Pleasanter MCP と機能が重なる 6 ツール（record get/list/update, site get/search 等）を plsnt 側にも実装する必要がある（**実装完了済み** — Issue #1）
- **保守対象の増加** — Pleasanter 本体の API 変更時、plsnt MCP 側も追従が必要
- **ドキュメント上の説明コスト** — ユーザーに「2 つの MCP がある理由」を説明する必要がある（本 ADR + CLAUDE.md ルーティング章で対応）

### Neutral

- **20 ツール構成は固定ではない** — Pleasanter 本体側で create / delete が将来サポートされた場合、plsnt 側の優位性は薄れる。その時点で再評価する

---

## Alternatives Considered

### A. 完全委譲（Roadmap 当初案）

plsnt は CLI 専念し、MCP は本家のみ。スキルは Pleasanter MCP を使う。

**棄却理由**: スキル依存分析で 281+ 件の `record list`、181+ 件の `record create` が plsnt CLI を前提に書かれており、Pleasanter MCP には `create` / `delete` が無い。スキル側で大量のアダプタコードが必要になり、保守コストが破綻する。

### B. 削減版（差別化のみ）

plsnt MCP は Pleasanter MCP に無い操作だけ提供（15 ツール: Tier 1 + Tier 3 + Tier 2 のうち書込系のみ）。

**棄却理由**: スキルが `record_get`（plsnt MCP 非対応）→ `record_create`（plsnt MCP）の 2 MCP 横断を強いられる。パラメータ形式の不一致、SiteSettings の不整合、コンテキスト切り替えコストが残る。

### C. 統一版（採択）

plsnt MCP は CRUD 全操作 + 高レベル操作を統一インターフェースとして提供（20 ツール）。

**採択理由**: スキル依存に対する一貫性が最優先。コードの重複は発生するが、スキル層のシンプルさで相殺される。

---

## References

- [Pleasanter MCP Server 公式ドキュメント](https://pleasanter.org/manual/api-mcp-server)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP SDK（採用中）
- Issue [#1](https://github.com/immmmmmmu/plsnt/issues/1) — plsnt MCP 実装（20 ツール完了）
- Issue [#2](https://github.com/immmmmmmu/plsnt/issues/2) — 本 ADR の起点
