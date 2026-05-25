#!/usr/bin/env bash
# 返却ワークフロー: 返却レコード作成 → 貸出ステータス更新 → 蔵書数復元
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/return-books.sh --lending <LENDING_ID> [--status 正常|延滞|破損]
#
# 例:
#   bash scripts/library/return-books.sh --lending 32320 --status 正常
set -euo pipefail

: "${LIB_LENDING_SITE:?env.sh を source してください}"

# --- 引数パース ---
LENDING_ID=""
RETURN_STATUS="正常"
TODAY=$(date +%Y-%m-%d)

while [[ $# -gt 0 ]]; do
  case $1 in
    --lending) LENDING_ID="$2"; shift 2 ;;
    --status)  RETURN_STATUS="$2"; shift 2 ;;
    --date)    TODAY="$2"; shift 2 ;;
    *) echo "不明なオプション: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$LENDING_ID" ]]; then
  echo "使い方: $0 --lending <LENDING_ID> [--status 正常|延滞|破損]" >&2
  exit 1
fi

echo "=== 返却処理 ==="

# --- Step 1: 貸出情報を取得 ---
LENDING_JSON=$(plsnt record get "$LENDING_ID" --json '{}' -o json 2>/dev/null)
MEMBER_ID=$(echo "$LENDING_JSON" | jq -r '.[0].ClassHash.ClassA // empty')
DUE_DATE=$(echo "$LENDING_JSON" | jq -r '.[0].CompletionTime // empty' | cut -dT -f1)
LEND_DATE=$(echo "$LENDING_JSON" | jq -r '.[0].DateHash.DateA // empty' | cut -dT -f1)

MEMBER_NAME=$(plsnt record get "$MEMBER_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')

echo "  貸出ID: $LENDING_ID"
echo "  利用者: $MEMBER_NAME"
echo "  貸出日: $LEND_DATE, 返却期限: $DUE_DATE"
echo "  返却日: $TODAY, 状態: $RETURN_STATUS"

# 延滞日数を計算
OVERDUE_DAYS=0
if [[ -n "$DUE_DATE" && "$TODAY" > "$DUE_DATE" ]]; then
  OVERDUE_DAYS=$(( ($(date -d "$TODAY" +%s) - $(date -d "$DUE_DATE" +%s)) / 86400 ))
  echo "  延滞: ${OVERDUE_DAYS}日"
  RETURN_STATUS="延滞"
fi
echo ""

# --- Step 2: 貸出明細から書籍一覧を取得 ---
echo "[1/3] 貸出明細を取得..."
ITEMS_JSON=$(plsnt record list --site-id "$LIB_LENDING_ITEM_SITE" \
  --view "{\"ColumnFilterHash\":{\"ClassA\":\"[$LENDING_ID]\"}}" \
  -o json 2>/dev/null)

BOOK_IDS=$(echo "$ITEMS_JSON" | jq -r '.[].ClassHash.ClassB // empty')
BOOK_COUNT=$(echo "$ITEMS_JSON" | jq 'length')
echo "  ${BOOK_COUNT}冊の書籍を確認"

# --- Step 3: 返却レコード作成 ---
echo "[2/3] 返却レコード作成..."
RETURN_ID=$(plsnt record create --site-id "$LIB_RETURN_SITE" \
  --json "{\"ClassHash\":{\"ClassA\":\"$LENDING_ID\",\"ClassB\":\"$RETURN_STATUS\"},\"DateHash\":{\"DateA\":\"$TODAY\"},\"NumHash\":{\"NumA\":$OVERDUE_DAYS}}" \
  -o json | jq -r '.Id')
echo "  返却ID: $RETURN_ID ($RETURN_STATUS)"

# --- Step 4: 貸出ステータスを「完了」に更新 ---
echo "  貸出ステータスを 900（完了）に更新..."
plsnt record update "$LENDING_ID" --json '{"Status":900}' -o json > /dev/null

# --- Step 5: 蔵書数を復元 ---
echo "[3/3] 蔵書数復元..."
for BOOK_ID in $BOOK_IDS; do
  BOOK_NAME=$(plsnt record get "$BOOK_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')

  COL_JSON=$(plsnt record list --site-id "$LIB_COLLECTION_SITE" \
    --view "{\"ColumnFilterHash\":{\"ClassB\":\"[$BOOK_ID]\"}}" \
    -o json 2>/dev/null)

  COL_ID=$(echo "$COL_JSON" | jq -r '.[0].ResultId // empty')
  if [[ -n "$COL_ID" ]]; then
    CURRENT_QTY=$(echo "$COL_JSON" | jq -r '.[0].NumHash.NumA // 0')
    NEW_QTY=$(echo "$CURRENT_QTY + 1" | bc | cut -d. -f1)
    plsnt record update "$COL_ID" --json "{\"NumHash\":{\"NumA\":$NEW_QTY}}" -o json > /dev/null
    echo "  蔵書復元: $BOOK_NAME $CURRENT_QTY → $NEW_QTY"
  else
    echo "  警告: $BOOK_NAME の蔵書レコードが見つかりません" >&2
  fi
done

echo ""
echo "=== 返却完了 ==="
echo "  貸出ID: $LENDING_ID → 完了"
echo "  返却ID: $RETURN_ID"
echo "  状態: $RETURN_STATUS"
[[ $OVERDUE_DAYS -gt 0 ]] && echo "  延滞日数: ${OVERDUE_DAYS}日" || true
