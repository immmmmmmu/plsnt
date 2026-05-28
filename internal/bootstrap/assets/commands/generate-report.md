# レポート生成ワークフロー

report-generator エージェントを使用して、プリザンターのデータを集計・レポート化する。

## 使用方法

```
/generate-report <SiteID> <レポート種別>
```

## レポート種別

- `status`: ステータス分布（Issues テーブル向け）
- `ranking`: NumHash列でのランキング
- `summary`: 件数・合計・平均のサマリー
- `cross`: 2軸のクロス集計
- `custom`: カスタムフィルタ指定

## 入力パラメータ

- `$ARGUMENTS`: SiteID とレポート種別、追加条件

## 実行例

```
/generate-report 12345 status
```

```
/generate-report 12345 ranking --field NumA --top 10
```

```
/generate-report 12345 summary --group-by ClassC
```

## エージェント呼び出し

report-generator エージェントに以下を依頼:
1. SiteID のスキーマを取得（plsnt schema）
2. 「$ARGUMENTS」の条件でデータを取得（plsnt record list）
3. 集計処理を実行
4. レポートを出力
