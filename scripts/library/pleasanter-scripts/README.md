# 図書貸出管理 Pleasanter UI改善スクリプト

PleasanterのWeb UI上でのユーザー体験を向上させるスクリプト集。

## 一括デプロイ

```bash
source scripts/library/env.sh
bash scripts/library/pleasanter-scripts/deploy-all.sh
```

## スクリプト一覧

### 貸出テーブル (`lending/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 1 | `default-date.js` | Script (New) | 新規作成時に今日の日付を自動セット |
| 2 | `status-colors.css` | Style (Index) | ステータスに応じた一覧行の色分け |

### 貸出明細テーブル (`lending-items/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 3 | `auto-book-name.js` | Script (New/Edit) | 書籍選択時に書名を明細名へ自動セット |

### 蔵書テーブル (`collections/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 4 | `low-stock-alert.js` | Script (Index/Edit) | 蔵書1冊以下で赤ハイライト+警告 |
| 7 | `prevent-negative.js` | Script (New/Edit) | 蔵書数マイナス防止バリデーション |

### 返却テーブル (`returns/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 5 | `default-date.js` | Script (New) | 新規作成時に今日の日付を自動セット |
| 12 | `overdue-highlight.js` | Script (Index) | 延滞日数>0で赤ハイライト |

### 利用者マスタ (`members/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 8 | `validate-email.js` | Script (New/Edit) | メールアドレス形式チェック |

### 書籍マスタ (`books/`)

| # | ファイル | 種類 | 効果 |
|---|---------|------|------|
| 9 | `validate-isbn.js` | Script (New/Edit) | ISBN-10/13形式チェック |
| 10 | `validate-price.js` | Script (New/Edit) | 定価が0以下の場合に確認ダイアログ |

### 設定変更（deploy-all.sh 内で処理）

| # | 対象 | 内容 |
|---|------|------|
| 6 | 返却テーブル | 返却状態をドロップダウン化（正常/延滞/破損/紛失） |
| 11 | 貸出テーブル | GridColumns に主要カラム+返却期限を設定 |

## カスタマイズ

- 低蔵書閾値: `collections/low-stock-alert.js` の `LOW_THRESHOLD` を変更
- ステータス色: `lending/status-colors.css` の色コードを変更
- 返却状態の選択肢: `deploy-all.sh` の ChoicesText を変更
