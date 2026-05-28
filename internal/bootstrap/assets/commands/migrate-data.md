# データ移行ワークフロー

data-migrator エージェントを使用して、CSV/Access からプリザンターへデータを移行する。

## 使用方法

```
/migrate-data <ソースファイルパス> --site-id <移行先SiteID>
```

## ワークフロー

1. data-migrator エージェントにソースとターゲットを伝える
2. エージェントが以下を自動実行:
   - ソースファイルの構造分析
   - 移行先スキーマの取得（plsnt schema）
   - マッピングYAMLの生成（plsnt migrate generate-mapping）
   - マッピング内容の確認提示
   - ドライラン実行（plsnt migrate execute --dry-run）
   - ユーザー確認後、本番移行

## 入力パラメータ

- `$ARGUMENTS`: ソースファイルパスと移行先SiteID

## 実行例

```
/migrate-data data/customers.csv --site-id 12345
```

```
/migrate-data legacy.mdb テーブル「顧客」を SiteID 12345 に移行
```

## エージェント呼び出し

data-migrator エージェントに以下を依頼:
1. ソースファイル「$ARGUMENTS」を分析
2. マッピングを生成し確認を求める
3. ドライランで問題がなければ本番実行
4. 移行後のレコード件数を確認
