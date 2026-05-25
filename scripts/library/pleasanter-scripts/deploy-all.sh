#!/usr/bin/env bash
# 図書貸出管理モデル Pleasanterスクリプト一括デプロイ
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/pleasanter-scripts/deploy-all.sh
set -euo pipefail

: "${LIB_LENDING_SITE:?env.sh を source してください}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# --- ヘルパー関数 ---
deploy_scripts() {
  local SITE_ID="$1"
  local TABLE_NAME="$2"
  shift 2
  local JQ_FILTER="$1"

  echo ""
  echo "=== $TABLE_NAME (SiteID: $SITE_ID) ==="

  local CURRENT
  CURRENT=$(plsnt site get "$SITE_ID" -o json 2>/dev/null)
  if [[ -z "$CURRENT" ]]; then
    echo "  エラー: SiteID $SITE_ID の取得に失敗しました" >&2
    return 1
  fi

  local NEW_SETTINGS
  NEW_SETTINGS=$(echo "$CURRENT" | jq ".SiteSettings | $JQ_FILTER")

  plsnt site update "$SITE_ID" --json "{\"SiteSettings\":$NEW_SETTINGS}" -o json > /dev/null 2>&1
  echo "  デプロイ完了"
}

read_script() {
  local FILE="$1"
  if [[ ! -f "$FILE" ]]; then
    echo "  警告: $FILE が見つかりません" >&2
    echo '""'
    return
  fi
  grep -v '^//' "$FILE" | jq -Rs .
}

read_style() {
  local FILE="$1"
  if [[ ! -f "$FILE" ]]; then
    echo "  警告: $FILE が見つかりません" >&2
    echo '""'
    return
  fi
  jq -Rs . < "$FILE"
}

echo "=========================================="
echo " 図書貸出管理 スクリプト一括デプロイ"
echo "=========================================="

# --- 1. 貸出テーブル (#1, #2, #11) ---
DEFAULT_DATE=$(read_script "$SCRIPT_DIR/lending/default-date.js")
STATUS_COLORS=$(read_style "$SCRIPT_DIR/lending/status-colors.css")

deploy_scripts "$LIB_LENDING_SITE" "貸出" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"貸出日デフォルト（今日）\", \"Body\": $DEFAULT_DATE, \"New\": true, \"Edit\": false, \"Index\": false}
  ], \"Styles\": [
    {\"Id\": 1, \"Title\": \"ステータス別 行色分け\", \"Body\": $STATUS_COLORS, \"New\": false, \"Edit\": false, \"Index\": true}
  ],
  \"GridColumns\": [\"DateA\", \"ClassA\", \"NumA\", \"Status\", \"CompletionTime\"]}"

# --- 2. 貸出明細テーブル (#3) ---
AUTO_BOOK=$(read_script "$SCRIPT_DIR/lending-items/auto-book-name.js")

deploy_scripts "$LIB_LENDING_ITEM_SITE" "貸出明細" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"書籍選択時 書名自動セット\", \"Body\": $AUTO_BOOK, \"New\": true, \"Edit\": true, \"Index\": false}
  ]}"

# --- 3. 蔵書テーブル (#4, #7) ---
LOW_STOCK=$(read_script "$SCRIPT_DIR/collections/low-stock-alert.js")
PREVENT_NEG=$(read_script "$SCRIPT_DIR/collections/prevent-negative.js")

deploy_scripts "$LIB_COLLECTION_SITE" "蔵書" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"低蔵書アラート\", \"Body\": $LOW_STOCK, \"New\": false, \"Edit\": true, \"Index\": true},
    {\"Id\": 2, \"Title\": \"マイナス蔵書防止\", \"Body\": $PREVENT_NEG, \"New\": true, \"Edit\": true, \"Index\": false}
  ]}"

# --- 4. 返却テーブル (#5, #6, #12) ---
RET_DATE=$(read_script "$SCRIPT_DIR/returns/default-date.js")
OVERDUE_HL=$(read_script "$SCRIPT_DIR/returns/overdue-highlight.js")

deploy_scripts "$LIB_RETURN_SITE" "返却" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"返却日デフォルト（今日）\", \"Body\": $RET_DATE, \"New\": true, \"Edit\": false, \"Index\": false},
    {\"Id\": 2, \"Title\": \"延滞日数ハイライト\", \"Body\": $OVERDUE_HL, \"New\": false, \"Edit\": false, \"Index\": true}
  ]}"

# --- 返却テーブル: 返却状態ドロップダウン (#6) ---
echo ""
echo "=== 返却 - 返却状態ドロップダウン (SiteID: $LIB_RETURN_SITE) ==="
CURRENT_RET=$(plsnt site get "$LIB_RETURN_SITE" -o json 2>/dev/null)
NEW_RET_SETTINGS=$(echo "$CURRENT_RET" | jq '
  .SiteSettings |
  .Columns = [
    .Columns[] |
    if .ColumnName == "ClassB" then
      . + {"ChoicesText": "正常\n延滞\n破損\n紛失"}
    else
      .
    end
  ]
')
plsnt site update "$LIB_RETURN_SITE" --json "{\"SiteSettings\":$NEW_RET_SETTINGS}" -o json > /dev/null 2>&1
echo "  返却状態ドロップダウン設定完了"

# --- 5. 利用者マスタ (#8) ---
VALIDATE_EMAIL=$(read_script "$SCRIPT_DIR/members/validate-email.js")

deploy_scripts "$LIB_MEMBER_SITE" "利用者マスタ" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"メールアドレス形式チェック\", \"Body\": $VALIDATE_EMAIL, \"New\": true, \"Edit\": true, \"Index\": false}
  ]}"

# --- 6. 書籍マスタ (#9, #10) ---
VALIDATE_ISBN=$(read_script "$SCRIPT_DIR/books/validate-isbn.js")
VALIDATE_PRICE=$(read_script "$SCRIPT_DIR/books/validate-price.js")

deploy_scripts "$LIB_BOOK_SITE" "書籍マスタ" \
  ". + {\"Scripts\": [
    {\"Id\": 1, \"Title\": \"ISBN形式チェック\", \"Body\": $VALIDATE_ISBN, \"New\": true, \"Edit\": true, \"Index\": false},
    {\"Id\": 2, \"Title\": \"定価バリデーション\", \"Body\": $VALIDATE_PRICE, \"New\": true, \"Edit\": true, \"Index\": false}
  ]}"

echo ""
echo "=========================================="
echo " デプロイ完了"
echo "=========================================="
echo ""
echo "デプロイされたスクリプト:"
echo "  [貸出]     #1 貸出日デフォルト, #2 ステータス色分け, #11 GridColumns"
echo "  [貸出明細] #3 書籍選択時 書名自動セット"
echo "  [蔵書]     #4 低蔵書アラート, #7 マイナス蔵書防止"
echo "  [返却]     #5 返却日デフォルト, #6 返却状態ドロップダウン, #12 延滞日数ハイライト"
echo "  [利用者]   #8 メールアドレスチェック"
echo "  [書籍]     #9 ISBN形式チェック, #10 定価バリデーション"
echo ""
echo "Pleasanter UIで各テーブルを開いて動作を確認してください"
