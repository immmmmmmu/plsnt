#!/usr/bin/env bash
# レコード詳細照会
#
# 使い方:
#   source scripts/library/env.sh
#   bash scripts/library/lookup.sh book <BOOK_ID>
#   bash scripts/library/lookup.sh member <MEMBER_ID>
#   bash scripts/library/lookup.sh lending <LENDING_ID>
set -euo pipefail

: "${LIB_BOOK_SITE:?env.sh を source してください}"

TYPE="${1:-}"
RECORD_ID="${2:-}"

if [[ -z "$TYPE" || -z "$RECORD_ID" ]]; then
  echo "使い方: $0 <book|member|lending> <RECORD_ID>" >&2
  exit 1
fi

case "$TYPE" in
  book)
    echo "=== 書籍詳細 ==="
    REC=$(plsnt record get "$RECORD_ID" --json '{}' -o json 2>/dev/null | jq '.[0]')
    BOOK_NAME=$(echo "$REC" | jq -r '.ClassHash.ClassA // "不明"')
    GENRE_ID=$(echo "$REC" | jq -r '.ClassHash.ClassB // empty')
    ISBN=$(echo "$REC" | jq -r '.ClassHash.ClassC // "不明"')
    PUB_ID=$(echo "$REC" | jq -r '.ClassHash.ClassD // empty')
    PRICE=$(echo "$REC" | jq -r '.NumHash.NumA // 0')

    GENRE_NAME="不明"
    [[ -n "$GENRE_ID" ]] && GENRE_NAME=$(plsnt record get "$GENRE_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')
    PUB_NAME="不明"
    [[ -n "$PUB_ID" ]] && PUB_NAME=$(plsnt record get "$PUB_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')

    echo "  書名: $BOOK_NAME"
    echo "  ジャンル: $GENRE_NAME"
    echo "  出版社: $PUB_NAME"
    echo "  ISBN: $ISBN"
    echo "  定価: ${PRICE}円"

    # 蔵書状況
    echo ""
    echo "  --- 蔵書状況 ---"
    COL_JSON=$(plsnt record list --site-id "$LIB_COLLECTION_SITE" \
      --view "{\"ColumnFilterHash\":{\"ClassB\":\"[$RECORD_ID]\"}}" \
      -o json 2>/dev/null)
    COL_COUNT=$(echo "$COL_JSON" | jq 'length')
    for i in $(seq 0 $((COL_COUNT - 1))); do
      DESC=$(echo "$COL_JSON" | jq -r ".[$i].DescriptionHash.DescriptionA // \"不明\"")
      QTY=$(echo "$COL_JSON" | jq -r ".[$i].NumHash.NumA // 0" | cut -d. -f1)
      echo "  $DESC: ${QTY}冊"
    done
    ;;

  member)
    echo "=== 利用者詳細 ==="
    REC=$(plsnt record get "$RECORD_ID" --json '{}' -o json 2>/dev/null | jq '.[0]')
    echo "  利用者名: $(echo "$REC" | jq -r '.ClassHash.ClassA // "不明"')"
    echo "  電話番号: $(echo "$REC" | jq -r '.ClassHash.ClassB // "不明"')"
    echo "  メール: $(echo "$REC" | jq -r '.ClassHash.ClassC // "不明"')"
    echo "  住所: $(echo "$REC" | jq -r '.DescriptionHash.DescriptionA // "不明"')"

    # 貸出履歴
    echo ""
    echo "  --- 貸出履歴 ---"
    LENDINGS_JSON=$(plsnt record list --site-id "$LIB_LENDING_SITE" \
      --view "{\"ColumnFilterHash\":{\"ClassA\":\"[$RECORD_ID]\"}}" \
      -o json 2>/dev/null)
    L_COUNT=$(echo "$LENDINGS_JSON" | jq 'length')
    if [[ "$L_COUNT" == "0" || "$L_COUNT" == "null" ]]; then
      echo "  貸出履歴なし"
    else
      for i in $(seq 0 $((L_COUNT - 1))); do
        L=$(echo "$LENDINGS_JSON" | jq ".[$i]")
        L_ID=$(echo "$L" | jq -r '.IssueId')
        L_DATE=$(echo "$L" | jq -r '.DateHash.DateA // empty' | cut -dT -f1)
        L_DUE=$(echo "$L" | jq -r '.CompletionTime // empty' | cut -dT -f1)
        L_STATUS=$(echo "$L" | jq -r '.Status')
        [[ "$L_STATUS" == "900" ]] && S="返却済" || S="貸出中"
        echo "  $L_ID: $L_DATE〜$L_DUE ($S)"
      done
    fi
    ;;

  lending)
    echo "=== 貸出詳細 ==="
    REC=$(plsnt record get "$RECORD_ID" --json '{}' -o json 2>/dev/null | jq '.[0]')
    MEMBER_ID=$(echo "$REC" | jq -r '.ClassHash.ClassA // empty')
    LEND_DATE=$(echo "$REC" | jq -r '.DateHash.DateA // empty' | cut -dT -f1)
    DUE_DATE=$(echo "$REC" | jq -r '.CompletionTime // empty' | cut -dT -f1)
    STATUS=$(echo "$REC" | jq -r '.Status')
    BOOK_COUNT=$(echo "$REC" | jq -r '.NumHash.NumA // 0')

    MEMBER_NAME="不明"
    [[ -n "$MEMBER_ID" ]] && MEMBER_NAME=$(plsnt record get "$MEMBER_ID" --json '{}' -o json 2>/dev/null | jq -r '.[0].ClassHash.ClassA // "不明"')

    [[ "$STATUS" == "900" ]] && S="返却済" || S="貸出中"

    echo "  貸出ID: $RECORD_ID"
    echo "  利用者: $MEMBER_NAME"
    echo "  貸出日: $LEND_DATE"
    echo "  返却期限: $DUE_DATE"
    echo "  冊数: ${BOOK_COUNT}冊"
    echo "  状態: $S"

    # 貸出明細
    echo ""
    echo "  --- 書籍一覧 ---"
    ITEMS_JSON=$(plsnt record list --site-id "$LIB_LENDING_ITEM_SITE" \
      --view "{\"ColumnFilterHash\":{\"ClassA\":\"[$RECORD_ID]\"}}" \
      -o json 2>/dev/null)
    I_COUNT=$(echo "$ITEMS_JSON" | jq 'length')
    for i in $(seq 0 $((I_COUNT - 1))); do
      DESC=$(echo "$ITEMS_JSON" | jq -r ".[$i].DescriptionHash.DescriptionA // \"不明\"")
      echo "  - $DESC"
    done

    # 返却情報
    echo ""
    echo "  --- 返却情報 ---"
    RET_JSON=$(plsnt record list --site-id "$LIB_RETURN_SITE" \
      --view "{\"ColumnFilterHash\":{\"ClassA\":\"[$RECORD_ID]\"}}" \
      -o json 2>/dev/null)
    R_COUNT=$(echo "$RET_JSON" | jq 'length')
    if [[ "$R_COUNT" == "0" || "$R_COUNT" == "null" ]]; then
      echo "  未返却"
    else
      R=$(echo "$RET_JSON" | jq '.[0]')
      R_DATE=$(echo "$R" | jq -r '.DateHash.DateA // empty' | cut -dT -f1)
      R_STATUS=$(echo "$R" | jq -r '.ClassHash.ClassB // "不明"')
      R_OVERDUE=$(echo "$R" | jq -r '.NumHash.NumA // 0' | cut -d. -f1)
      echo "  返却日: $R_DATE"
      echo "  状態: $R_STATUS"
      [[ "$R_OVERDUE" -gt 0 ]] && echo "  延滞日数: ${R_OVERDUE}日" || true
    fi
    ;;

  *)
    echo "不明なタイプ: $TYPE (book|member|lending)" >&2
    exit 1
    ;;
esac
