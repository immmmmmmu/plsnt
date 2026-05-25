#!/usr/bin/env bash
# 貸出ワークフロー: 貸出 → 貸出明細 → 蔵書更新 を一括実行
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/lend-books.sh \
#     --member <MEMBER_ID> \
#     --books '<BOOK_ID> <BOOK_ID> ...' \
#     --due <RETURN_DUE_DATE>
#
# 例:
#   bash scripts/library/lend-books.sh \
#     --member 32302 \
#     --books '32308 32309' \
#     --due 2026-03-21
set -euo pipefail

: "${LIB_LENDING_SITE:?env.sh を source してください}"

# --- 引数パース ---
MEMBER_ID=""
BOOKS=""
TODAY=$(date +%Y-%m-%d)
DUE_DATE=$(date -d "+14 days" +%Y-%m-%d 2>/dev/null || date -v+14d +%Y-%m-%d)

while [[ $# -gt 0 ]]; do
  case $1 in
    --member) MEMBER_ID="$2"; shift 2 ;;
    --books)  BOOKS="$2"; shift 2 ;;
    --date)   TODAY="$2"; shift 2 ;;
    --due)    DUE_DATE="$2"; shift 2 ;;
    *) echo "不明なオプション: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$MEMBER_ID" || -z "$BOOKS" ]]; then
  echo "使い方: $0 --member <ID> --books '<BOOK_ID> ...' [--due <DATE>]" >&2
  exit 1
fi

# 利用者名を取得
MEMBER_NAME=$(plsnt record get "$MEMBER_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')

echo "=== 貸出処理 ==="
echo "  利用者: $MEMBER_NAME ($MEMBER_ID)"
echo "  貸出日: $TODAY, 返却期限: $DUE_DATE"

# --- Step 1: 書籍情報を取得 ---
BOOK_COUNT=0
declare -a BOOK_RECORDS=()

for BOOK_ID in $BOOKS; do
  BOOK_NAME=$(plsnt record get "$BOOK_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')
  echo "  書籍: $BOOK_NAME ($BOOK_ID)"
  BOOK_RECORDS+=("$BOOK_ID:$BOOK_NAME")
  BOOK_COUNT=$((BOOK_COUNT + 1))
done

echo "  合計: ${BOOK_COUNT}冊"
echo ""

# --- Step 2: 貸出レコード作成 (Issues) ---
echo "[1/3] 貸出レコード作成..."
LENDING_ID=$(plsnt record create --site-id "$LIB_LENDING_SITE" \
  --json "{\"ClassHash\":{\"ClassA\":\"$MEMBER_ID\"},\"DateHash\":{\"DateA\":\"$TODAY\"},\"NumHash\":{\"NumA\":$BOOK_COUNT},\"CompletionTime\":\"${DUE_DATE}T00:00:00\"}" \
  -o json | jq -r '.Id')
echo "  貸出ID: $LENDING_ID (期限: $DUE_DATE)"

# --- Step 3: 貸出明細レコード作成 ---
echo "[2/3] 貸出明細作成..."
for RECORD in "${BOOK_RECORDS[@]}"; do
  IFS=: read -r BOOK_ID BOOK_NAME <<< "$RECORD"
  ITEM_DESC="${BOOK_NAME}"
  ITEM_ID=$(plsnt record create --site-id "$LIB_LENDING_ITEM_SITE" \
    --json "{\"ClassHash\":{\"ClassA\":\"$LENDING_ID\",\"ClassB\":\"$BOOK_ID\"},\"DescriptionHash\":{\"DescriptionA\":\"$ITEM_DESC\"},\"NumHash\":{\"NumA\":1}}" \
    -o json | jq -r '.Id')
  echo "  明細ID: $ITEM_ID ($ITEM_DESC)"
done

# --- Step 4: 蔵書数更新 ---
echo "[3/3] 蔵書更新..."
for RECORD in "${BOOK_RECORDS[@]}"; do
  IFS=: read -r BOOK_ID BOOK_NAME <<< "$RECORD"

  # 当該書籍の蔵書レコードを検索
  COL_JSON=$(plsnt record list --site-id "$LIB_COLLECTION_SITE" \
    --view "{\"ColumnFilterHash\":{\"ClassB\":\"[$BOOK_ID]\"}}" \
    -o json 2>/dev/null)

  COL_ID=$(echo "$COL_JSON" | jq -r '.[0].ResultId // empty')
  if [[ -n "$COL_ID" ]]; then
    CURRENT_QTY=$(echo "$COL_JSON" | jq -r '.[0].NumHash.NumA // 0')
    NEW_QTY=$(echo "$CURRENT_QTY - 1" | bc | cut -d. -f1)
    plsnt record update "$COL_ID" --json "{\"NumHash\":{\"NumA\":$NEW_QTY}}" -o json > /dev/null
    echo "  蔵書更新: $BOOK_NAME $CURRENT_QTY → $NEW_QTY"
  else
    echo "  警告: $BOOK_NAME の蔵書レコードが見つかりません" >&2
  fi
done

echo ""
echo "=== 貸出完了 ==="
echo "  貸出ID: $LENDING_ID"
echo "  利用者: $MEMBER_NAME"
echo "  冊数: ${BOOK_COUNT}冊"
echo "  返却期限: $DUE_DATE"
