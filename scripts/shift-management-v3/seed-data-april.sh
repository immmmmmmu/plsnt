#!/bin/bash
# シフト管理システム v3 4月サンプルデータ投入
# 使い方: source scripts/shift-management-v3/env.sh && bash scripts/shift-management-v3/seed-data-april.sh

set -euo pipefail

# 環境変数チェック
for var in SITES_SITE_ID GUARDS_SITE_ID SHIFT_ASSIGNMENTS_SITE_ID; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: $var is not set. Run 'source scripts/shift-management-v3/env.sh' first." >&2
    exit 1
  fi
done

# マスタデータのID取得
SITE_IDS=$(plsnt record list --site-id "$SITES_SITE_ID" -o ids)
SITE1=$(echo "$SITE_IDS" | sed -n '1p')  # 渋谷オフィスビル
SITE2=$(echo "$SITE_IDS" | sed -n '2p')  # 新宿商業施設
SITE3=$(echo "$SITE_IDS" | sed -n '3p')  # 品川建設現場

GUARD_IDS=$(plsnt record list --site-id "$GUARDS_SITE_ID" -o ids)
G1=$(echo "$GUARD_IDS" | sed -n '1p')  # 田中太郎
G2=$(echo "$GUARD_IDS" | sed -n '2p')  # 鈴木花子
G3=$(echo "$GUARD_IDS" | sed -n '3p')  # 佐藤次郎
G4=$(echo "$GUARD_IDS" | sed -n '4p')  # 山田美咲
G5=$(echo "$GUARD_IDS" | sed -n '5p')  # 高橋健一

echo "=== 4月 シフト割当データ投入 ==="
echo "現場: 渋谷=$SITE1, 新宿=$SITE2, 品川=$SITE3"
echo "警備員: 田中=$G1, 鈴木=$G2, 佐藤=$G3, 山田=$G4, 高橋=$G5"

# --- 第1週: 4/6(月)〜4/10(金) ---
echo ""
echo "--- 第1週: 4/6(月)〜4/10(金) ---"
for DAY in 06 07 08 09 10; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T09:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"

  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T09:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"

  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-04-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"

  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T08:00:00\",\"CompletionTime\":\"2026-04-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"

  # 新宿商業施設 日勤 山田（月・水・金のみ）
  DOW=$(date -d "2026-04-${DAY}" +%u)  # 1=月, 3=水, 5=金
  if [ "$DOW" = "1" ] || [ "$DOW" = "3" ] || [ "$DOW" = "5" ]; then
    plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T10:00:00\",\"CompletionTime\":\"2026-04-${DAY}T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"
  fi
done

# --- 第2週: 4/13(月)〜4/17(金) ---
echo ""
echo "--- 第2週: 4/13(月)〜4/17(金) ---"
for DAY in 13 14 15 16 17; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T09:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"

  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T09:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"

  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-04-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"

  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T08:00:00\",\"CompletionTime\":\"2026-04-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"

  # 新宿商業施設 日勤 山田（月・水・金のみ）
  DOW=$(date -d "2026-04-${DAY}" +%u)
  if [ "$DOW" = "1" ] || [ "$DOW" = "3" ] || [ "$DOW" = "5" ]; then
    plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T10:00:00\",\"CompletionTime\":\"2026-04-${DAY}T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"
  fi
done

# --- 4/8(水) 鈴木体調不良で欠勤 → 山田が代替 ---
echo ""
echo "--- 4/8 鈴木欠勤 → 山田代替 ---"
# 欠勤レコード
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\",\"ClassE\":\"体調不良\"},\"StartTime\":\"2026-04-08T09:00:00\",\"CompletionTime\":\"2026-04-08T18:00:00\",\"NumHash\":{\"NumA\":0},\"Status\":910,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤(欠勤)\"}}"
# 代替レコード（山田は水曜新宿の代わりに渋谷へ）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G4\",\"ClassD\":\"代替\",\"ClassE\":\"体調不良\"},\"StartTime\":\"2026-04-08T09:00:00\",\"CompletionTime\":\"2026-04-08T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"山田 渋谷 日勤(代替)\"}}"

# --- 4/11(土)・4/12(日) 新宿商業施設 土日シフト ---
echo ""
echo "--- 4/11-12 新宿 土日シフト ---"
for DAY in 11 12; do
  # 田中 土日は新宿へ
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T10:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"田中 新宿 日勤\"}}"
  # 高橋 土日は新宿日勤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-04-${DAY}T10:00:00\",\"CompletionTime\":\"2026-04-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"高橋 新宿 日勤\"}}"
done

# --- 4/15(水) 品川建設 増員（田中応援） ---
echo ""
echo "--- 4/15 品川建設 増員 ---"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G1\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-04-15T08:00:00\",\"CompletionTime\":\"2026-04-15T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 品川 日勤(応援)\"}}"

echo ""
echo "=== 4月サンプルデータ投入完了 ==="
echo "投入件数:"
echo "  第1週 定常: 田中5 + 鈴木5 + 高橋夜勤5 + 佐藤5 + 山田3(月水金) = 23件"
echo "  第2週 定常: 田中5 + 鈴木5 + 高橋夜勤5 + 佐藤5 + 山田3(月水金) = 23件"
echo "  欠勤/代替: 鈴木欠勤1 + 山田代替1 = 2件"
echo "  土日: 田中2 + 高橋2 = 4件"
echo "  増員: 田中品川1 = 1件"
echo "  合計: 53件"
echo ""
echo "カレンダービューで確認: http://localhost/items/$SHIFT_ASSIGNMENTS_SITE_ID"
