#!/usr/bin/env bash
# 延滞チェック: 返却期限超過の貸出を一覧表示
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/overdue-check.sh [--date <CHECK_DATE>]
set -euo pipefail

: "${LIB_LENDING_SITE:?env.sh を source してください}"

TODAY=$(date +%Y-%m-%d)
while [[ $# -gt 0 ]]; do
  case $1 in
    --date) TODAY="$2"; shift 2 ;;
    *) echo "不明なオプション: $1" >&2; exit 1 ;;
  esac
done

echo "=== 延滞チェック ($TODAY 時点) ==="
echo ""

# 未完了の貸出（Status != 900）を取得
LENDINGS_JSON=$(plsnt record list --site-id "$LIB_LENDING_SITE" \
  --view '{"ColumnFilterHash":{"Status":"100,200,300,400,500,600,700,800"}}' \
  -o json 2>/dev/null)

COUNT=$(echo "$LENDINGS_JSON" | jq 'length')
OVERDUE_COUNT=0

if [[ "$COUNT" == "0" || "$COUNT" == "null" ]]; then
  echo "  現在の貸出はありません"
  exit 0
fi

printf "%-10s %-12s %-10s %-12s %s\n" "貸出ID" "利用者" "冊数" "返却期限" "状態"
printf "%-10s %-12s %-10s %-12s %s\n" "------" "------" "----" "--------" "----"

for i in $(seq 0 $((COUNT - 1))); do
  LENDING=$(echo "$LENDINGS_JSON" | jq ".[$i]")
  L_ID=$(echo "$LENDING" | jq -r '.IssueId')
  MEMBER_ID=$(echo "$LENDING" | jq -r '.ClassHash.ClassA // empty')
  BOOK_COUNT=$(echo "$LENDING" | jq -r '.NumHash.NumA // 0')
  DUE=$(echo "$LENDING" | jq -r '.CompletionTime // empty' | cut -dT -f1)

  MEMBER_NAME="不明"
  if [[ -n "$MEMBER_ID" ]]; then
    MEMBER_NAME=$(plsnt record get "$MEMBER_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')
  fi

  if [[ -n "$DUE" && "$TODAY" > "$DUE" ]]; then
    DAYS_OVER=$(( ($(date -d "$TODAY" +%s) - $(date -d "$DUE" +%s)) / 86400 ))
    STATUS="延滞${DAYS_OVER}日"
    OVERDUE_COUNT=$((OVERDUE_COUNT + 1))
  else
    STATUS="貸出中"
  fi

  printf "%-10s %-12s %-10s %-12s %s\n" "$L_ID" "$MEMBER_NAME" "${BOOK_COUNT}冊" "$DUE" "$STATUS"
done

echo ""
echo "  合計: ${COUNT}件の貸出、うち${OVERDUE_COUNT}件が延滞"
