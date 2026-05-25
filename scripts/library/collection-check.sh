#!/usr/bin/env bash
# 蔵書一覧・低在庫アラート
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/collection-check.sh [--threshold <NUM>]
set -euo pipefail

: "${LIB_COLLECTION_SITE:?env.sh を source してください}"

THRESHOLD=1
while [[ $# -gt 0 ]]; do
  case $1 in
    --threshold) THRESHOLD="$2"; shift 2 ;;
    *) echo "不明なオプション: $1" >&2; exit 1 ;;
  esac
done

echo "=== 蔵書一覧 (低在庫閾値: ${THRESHOLD}冊以下) ==="
echo ""

COL_JSON=$(plsnt record list --site-id "$LIB_COLLECTION_SITE" -o json 2>/dev/null)
COUNT=$(echo "$COL_JSON" | jq 'length')

if [[ "$COUNT" == "0" || "$COUNT" == "null" ]]; then
  echo "  蔵書データがありません"
  exit 0
fi

LOW_COUNT=0
printf "%-20s %-18s %8s %s\n" "書架" "書籍" "所蔵数" "状態"
printf "%-20s %-18s %8s %s\n" "----------" "----------" "------" "----"

for i in $(seq 0 $((COUNT - 1))); do
  REC=$(echo "$COL_JSON" | jq ".[$i]")
  DESC=$(echo "$REC" | jq -r '.DescriptionHash.DescriptionA // "不明"')
  QTY=$(echo "$REC" | jq -r '.NumHash.NumA // 0')
  QTY_INT=$(echo "$QTY" | cut -d. -f1)

  # DescriptionA は "書架名 - 書籍名" 形式（" - " で分割）
  SHELF_NAME=$(echo "$DESC" | sed 's/ - .*//')
  BOOK_NAME=$(echo "$DESC" | sed 's/.* - //')

  if [[ "$QTY_INT" -le "$THRESHOLD" ]]; then
    STATUS="*** 低在庫 ***"
    LOW_COUNT=$((LOW_COUNT + 1))
  else
    STATUS=""
  fi

  printf "%-20s %-18s %8s %s\n" "$SHELF_NAME" "$BOOK_NAME" "${QTY_INT}冊" "$STATUS"
done

echo ""
echo "  合計: ${COUNT}件、うち${LOW_COUNT}件が低在庫（${THRESHOLD}冊以下）"
