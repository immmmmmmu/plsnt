# サンプルデータ投入

アプリモデルにサンプルデータを投入する。

## 使用方法

```
/seed-data <アプリモデル名>
```

## ワークフロー

1. `scripts/<モデル名>/env.sh` を確認し、環境変数を特定
2. `scripts/<モデル名>/seed-data.sh` の内容を確認
3. ユーザーに投入データの概要を報告
4. 確認後、実行:
   ```bash
   source scripts/<モデル名>/env.sh && bash scripts/<モデル名>/seed-data.sh
   ```
5. 投入結果を `plsnt record list` で確認・報告

## seed-data.sh の共通パターン

```bash
#!/bin/bash
set -euo pipefail

# 1. 環境変数チェック
# 2. マスタデータ投入（資格、現場、商品等）
# 3. マスタIDを取得（plsnt record list -o ids）
# 4. リンク参照を使ってトランザクションデータ投入
# 5. 完了報告
```

### ID取得パターン

```bash
IDS=$(plsnt record list --site-id "$MASTER_SITE_ID" -o ids)
ID1=$(echo "$IDS" | sed -n '1p')
ID2=$(echo "$IDS" | sed -n '2p')
```

### 日付計算（日跨ぎ対応）

```bash
# 月末日跨ぎを安全に処理する
NEXT_DAY=$(date -d "2026-03-${DAY} +1 day" +%Y-%m-%dT09:00:00)
```

## 入力パラメータ

- `$ARGUMENTS`: アプリモデル名（shopping, library, shift-management-v3 等）

## 実行例

```
/seed-data shift-management-v3
/seed-data shopping
```

## 注意事項

- 既存データがあるテーブルに重複投入される可能性あり
- リンクフィールドの値はブラケットなし `"32274"` で格納
- Issues テーブルの Status/StartTime/CompletionTime はトップレベルに指定
