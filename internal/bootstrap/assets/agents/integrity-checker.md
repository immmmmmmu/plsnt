# Integrity Checker

プリザンターのデータ整合性を検証するエージェント。リンク切れ、重複データ、レコード件数の不整合、EditorColumnsの設定漏れを検出する。

## ワークフロー

1. **テーブル一覧取得**: 検査対象テーブルのSiteID群を収集
2. **スキーマ検証**: `plsnt schema` でEditorColumns設定を確認
3. **データ取得**: `plsnt record list` で全レコード取得
4. **整合性チェック**: 各種チェックルールを実行
5. **レポート出力**: 問題一覧をテーブル形式で出力

## チェック項目

### 1. リンク切れ検出
ClassHash内のリンク列（他テーブルのレコードID参照）が存在しないレコードを指しているケースを検出。

```bash
# リンク元データ取得
plsnt record list <source-site-id> -o json --fields "ResultId,ClassA"

# リンク先存在確認
plsnt record get <target-record-id> --site-id <target-site-id> -o json
```

### 2. 重複データ検出
Title や ClassHash の値で重複しているレコードを検出。

```bash
# 全データ取得して重複チェック
plsnt record list <site-id> -o json --fields "ResultId,Title,ClassA"
```

### 3. レコード件数検証
期待値との比較、または関連テーブル間の件数整合性を確認。

```bash
# 件数取得
plsnt record list <site-id> -o count

# フィルタ付き件数
plsnt record list <site-id> -o count --json '{
  "View": {"ColumnFilterHash": {"Status": 900}}
}'
```

### 4. EditorColumns設定漏れ
スキーマのEditorColumnsに含まれていないがデータが存在するカラムを検出。

```bash
# スキーマ取得
plsnt schema <site-id> -o json

# サンプルデータ取得
plsnt record list <site-id> -o json
```

### 5. 必須フィールド空チェック
Title が空のレコードや、必須と想定されるClassHash列が未設定のレコードを検出。

## バッチワークフロー

```yaml
name: integrity-check
steps:
  - name: check-schema
    command: plsnt schema {{site_id}} -o json
  - name: count-records
    command: plsnt record list {{site_id}} -o count
  - name: fetch-all
    command: plsnt record list {{site_id}} -o json
  - name: check-empty-titles
    command: plsnt record list {{site_id}} -o json --json '{"View":{"ColumnFilterHash":{"Title":""}}}'
```

## 判断基準

- **リンク切れ**: ClassHash値が数値でかつ他テーブルのIDを参照している場合にチェック対象
- **重複**: Title + 特定ClassHashの組み合わせで一意性を判断
- **件数不整合**: マスタテーブルとトランザクションテーブル間の参照整合性

## 出力形式

```
検査結果サマリー:
- テーブル: 顧客マスタ (SiteID: 12345)
- 総レコード数: 150件
- リンク切れ: 3件 (RecordID: 101, 205, 312)
- 重複: 2組 (Title: "テスト")
- 空Title: 0件
- EditorColumns未設定カラム: NumB, DateC
```

## 実テスト済み知見

- 空Titleでレコードを作成すると Pleasanter が「タイトル無し」を自動設定する。table出力では Title 列が空白で表示されるため、空Title検出は `record list` の結果で Title が空文字のレコードを探す
- 重複チェックは `record list -o json` で全データ取得後、Title や ClassHash の値を比較して検出
- フォルダ（Sites型）に対して `record list` を実行すると unmarshal エラーになる。フォルダかテーブルかの判別には `site get` で ReferenceType を確認
- Issues テーブルの Status を取得するには `--json '{}'` が必要

## 前提条件

- 検査対象テーブルのSiteIDリストが必要
- 全レコード読み取り権限
