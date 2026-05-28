# 整合性チェックワークフロー

integrity-checker エージェントを使用して、プリザンターのデータ整合性を検証する。

## 使用方法

```
/check-integrity <SiteID> [追加SiteID...]
```

## チェック項目

1. **レコード件数**: 各テーブルの件数確認
2. **空Titleチェック**: Title が空のレコード検出
3. **EditorColumns漏れ**: データがあるがEditorColumnsにないカラム検出
4. **リンク切れ**: リンクフィールドが指すレコードの存在確認
5. **重複データ**: Title や ClassHash での重複検出

## 入力パラメータ

- `$ARGUMENTS`: 検査対象の SiteID（複数指定可）

## 実行例

```
/check-integrity 12345
```

```
/check-integrity 12345 12346 12347
```

## エージェント呼び出し

integrity-checker エージェントに以下を依頼:
1. 「$ARGUMENTS」の SiteID 群のスキーマとデータを取得
2. 各チェック項目を実行
3. 問題一覧をサマリー形式で出力
4. 修正が必要な場合は対処法を提示
