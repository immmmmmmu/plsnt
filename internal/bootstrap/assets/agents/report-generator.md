# Report Generator

プリザンターのデータを集計・分析してレポートを生成するエージェント。ステータス分布、クロス集計、期間別推移などのレポートを作成する。

## ワークフロー

1. **対象テーブル確認**: `plsnt schema` でカラム定義を取得
2. **データ取得**: `plsnt record list` でフィルタ付きデータ取得
3. **集計処理**: 取得データを加工・集計
4. **レポート出力**: テーブル/CSV/JSON形式で出力

## 使用コマンド

```bash
# 全データ取得
plsnt record list <site-id> -o json --fields "Title,ClassA,NumA,Status"

# ステータスでフィルタ
plsnt record list <site-id> -o json --json '{
  "View": {
    "ColumnFilterHash": {
      "Status": 100
    }
  }
}'

# 日付範囲でフィルタ
plsnt record list <site-id> -o json --json '{
  "View": {
    "ColumnFilterHash": {
      "DateA": "[\"2024-01-01\",\"2024-12-31\"]"
    }
  }
}'

# ソート付き
plsnt record list <site-id> -o json --json '{
  "View": {
    "ColumnSorterHash": {
      "NumA": "desc"
    }
  }
}'

# レコード件数
plsnt record list <site-id> -o count
```

## レポートパターン

### ステータス分布
Issuesテーブルのステータス別件数をカウントし分布を表示。

### 担当者別集計
ClassHashの担当者列でグルーピングし、件数・合計・平均を算出。

### 期間別推移
DateHashまたはCreatedTimeでフィルタし月別・週別の推移を集計。

### クロス集計
2つのClassHash列を軸にしたクロス集計表を作成。

### ランキング
NumHash列でソートしトップN件を抽出。

## バッチワークフロー

```yaml
name: monthly-report
steps:
  - name: fetch-data
    command: plsnt record list {{site_id}} -o json
  - name: count-total
    command: plsnt record list {{site_id}} -o count
  - name: fetch-open
    command: plsnt record list {{site_id}} -o count --json '{"View":{"ColumnFilterHash":{"Status":100}}}'
  - name: fetch-closed
    command: plsnt record list {{site_id}} -o count --json '{"View":{"ColumnFilterHash":{"Status":900}}}'
```

## 判断基準

- **ページング**: TotalCount > 200 の場合は全件取得のためOffset指定で複数回リクエスト
- **出力形式**: 人間向け → table、後処理 → json/csv
- **フィルタ組合せ**: 複数条件はColumnFilterHash内にAND結合で指定

## 実テスト済み知見

- ColumnFilterHash でステータスフィルタする場合、値は文字列で指定: `"Status":"100"`（数値ではない）
- ColumnSorterHash でソート: `"NumA":"desc"` は正しく動作確認済み
- `-o count` で件数のみ取得可能（フィルタ付きでも利用可）
- Issues テーブルの Status 値を取得するには `--json '{}'` が必要
- NumHash の大きな値は `-o table` で科学記数法表示。正確な値には `-o json` を使用
- `--fields` 指定でも `-o json` では全フィールドが返る（APIの仕様）

## 前提条件

- 対象テーブルのSiteIDが判明していること
- テーブルのカラム定義を事前に把握（`plsnt schema`で確認）
