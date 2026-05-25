#!/bin/bash
# シフト管理システム v3 5月サンプルデータ投入
# 使い方: source scripts/shift-management-v3/env.sh && bash scripts/shift-management-v3/seed-data-may.sh

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

echo "=== 5月 シフト割当データ投入 ==="
echo "現場: 渋谷=$SITE1, 新宿=$SITE2, 品川=$SITE3"
echo "警備員: 田中=$G1, 鈴木=$G2, 佐藤=$G3, 山田=$G4, 高橋=$G5"

# --- 5/1(金) GW前最終営業日 ---
echo ""
echo "--- 5/1(金) GW前最終営業日 ---"
# 渋谷ビル 日勤 田中
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-01T09:00:00\",\"CompletionTime\":\"2026-05-01T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"
# 渋谷ビル 日勤 鈴木
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-01T09:00:00\",\"CompletionTime\":\"2026-05-01T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"
# 渋谷ビル 夜勤 高橋
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-01T18:00:00\",\"CompletionTime\":\"2026-05-02T09:00:00\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"
# 品川建設 日勤 佐藤
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-01T08:00:00\",\"CompletionTime\":\"2026-05-01T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
# 新宿商業施設 日勤 山田（金曜）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-01T10:00:00\",\"CompletionTime\":\"2026-05-01T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"

# --- GW期間: 5/3(日)〜5/6(水) 新宿商業施設 特別シフト ---
echo ""
echo "--- GW特別シフト: 5/3〜5/6 新宿商業施設 ---"
for DAY in 03 04 05 06; do
  # 新宿 GW特別 田中（日勤）
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G1\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 新宿 日勤(GW)\"}}"
  # 新宿 GW特別 高橋（日勤）
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G5\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 新宿 日勤(GW)\"}}"
  # 新宿 GW特別 山田（遅番）
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-${DAY}T14:00:00\",\"CompletionTime\":\"2026-05-${DAY}T22:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 遅番(GW)\"}}"
done

# --- GW期間: 5/3〜5/6 渋谷ビル 最低限警備（夜勤のみ） ---
echo ""
echo "--- GW期間: 5/3〜5/6 渋谷ビル 夜勤のみ ---"
for DAY in 03 04 05 06; do
  NEXT_DAY=$(date -d "2026-05-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G3\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 渋谷 夜勤(GW)\"}}"
done

# --- GW明け: 5/7(木)〜5/8(金) ---
echo ""
echo "--- GW明け: 5/7(木)〜5/8(金) ---"
for DAY in 07 08; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"
  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"
  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-05-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"
  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T08:00:00\",\"CompletionTime\":\"2026-05-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
done
# 山田 5/8(金)のみ新宿
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-08T10:00:00\",\"CompletionTime\":\"2026-05-08T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"

# --- 第2週: 5/11(月)〜5/15(金) ---
echo ""
echo "--- 第2週: 5/11(月)〜5/15(金) ---"
for DAY in 11 12 13 14 15; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"
  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"
  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-05-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"
  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T08:00:00\",\"CompletionTime\":\"2026-05-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
  # 新宿商業施設 日勤 山田（月・水・金のみ）
  DOW=$(date -d "2026-05-${DAY}" +%u)
  if [ "$DOW" = "1" ] || [ "$DOW" = "3" ] || [ "$DOW" = "5" ]; then
    plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"
  fi
done

# --- 第3週: 5/18(月)〜5/22(金) ---
echo ""
echo "--- 第3週: 5/18(月)〜5/22(金) ---"
for DAY in 18 19 20 21 22; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"
  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"
  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-05-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"
  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T08:00:00\",\"CompletionTime\":\"2026-05-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
  # 新宿商業施設 日勤 山田（月・水・金のみ）
  DOW=$(date -d "2026-05-${DAY}" +%u)
  if [ "$DOW" = "1" ] || [ "$DOW" = "3" ] || [ "$DOW" = "5" ]; then
    plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"
  fi
done

# --- 第4週: 5/25(月)〜5/29(金) ---
echo ""
echo "--- 第4週: 5/25(月)〜5/29(金) ---"
for DAY in 25 26 27 28 29; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"
  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T09:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"
  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-05-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"
  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T08:00:00\",\"CompletionTime\":\"2026-05-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
  # 新宿商業施設 日勤 山田（月・水・金のみ）
  DOW=$(date -d "2026-05-${DAY}" +%u)
  if [ "$DOW" = "1" ] || [ "$DOW" = "3" ] || [ "$DOW" = "5" ]; then
    plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G4\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T16:00:00\",\"NumHash\":{\"NumA\":6},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 新宿 日勤\"}}"
  fi
done

# --- 5/9(土)・5/10(日) 新宿商業施設 土日シフト ---
echo ""
echo "--- 5/9-10 新宿 土日シフト ---"
for DAY in 09 10; do
  # 田中 土日は新宿へ
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 新宿 日勤\"}}"
  # 高橋 土日は新宿日勤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 新宿 日勤\"}}"
done

# --- 5/16(土)・5/17(日) 新宿商業施設 土日シフト ---
echo ""
echo "--- 5/16-17 新宿 土日シフト ---"
for DAY in 16 17; do
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 新宿 日勤\"}}"
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-05-${DAY}T10:00:00\",\"CompletionTime\":\"2026-05-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":8},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 新宿 日勤\"}}"
done

# --- 5/20(水) 佐藤有給 → 鈴木が品川代替 ---
echo ""
echo "--- 5/20 佐藤有給 → 鈴木代替 ---"
# 欠勤レコード
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\",\"ClassE\":\"有給休暇\"},\"StartTime\":\"2026-05-20T08:00:00\",\"CompletionTime\":\"2026-05-20T17:00:00\",\"NumHash\":{\"NumA\":0},\"Status\":910,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤(有給)\"}}"
# 代替レコード（鈴木が渋谷の代わりに品川へ）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G2\",\"ClassD\":\"代替\",\"ClassE\":\"有給休暇\"},\"StartTime\":\"2026-05-20T08:00:00\",\"CompletionTime\":\"2026-05-20T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 品川 日勤(代替)\"}}"

# --- 5/22(金) 品川建設 増員（田中・山田応援）---
echo ""
echo "--- 5/22 品川建設 増員 ---"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G1\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-22T08:00:00\",\"CompletionTime\":\"2026-05-22T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 品川 日勤(応援)\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G4\",\"ClassD\":\"期間\"},\"StartTime\":\"2026-05-22T08:00:00\",\"CompletionTime\":\"2026-05-22T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"山田 品川 日勤(応援)\"}}"

echo ""
echo "=== 5月サンプルデータ投入完了 ==="
echo "投入件数:"
echo "  5/1 GW前: 田中1 + 鈴木1 + 高橋夜勤1 + 佐藤1 + 山田1 = 5件"
echo "  GW特別(5/3-6): 新宿(田中4+高橋4+山田4) + 渋谷夜勤(佐藤4) = 16件"
echo "  GW明け(5/7-8): 田中2 + 鈴木2 + 高橋夜勤2 + 佐藤2 + 山田1(金) = 9件"
echo "  第2週(5/11-15): 田中5 + 鈴木5 + 高橋夜勤5 + 佐藤5 + 山田3(月水金) = 23件"
echo "  第3週(5/18-22): 田中5 + 鈴木5 + 高橋夜勤5 + 佐藤5 + 山田3(月水金) = 23件"
echo "  第4週(5/25-29): 田中5 + 鈴木5 + 高橋夜勤5 + 佐藤5 + 山田3(月水金) = 23件"
echo "  土日(5/9-10, 5/16-17): 田中4 + 高橋4 = 8件"
echo "  欠勤/代替: 佐藤有給1 + 鈴木代替1 = 2件"
echo "  増員: 田中品川1 + 山田品川1 = 2件"
echo "  合計: 111件"
echo ""
echo "カレンダービューで確認: http://localhost/items/$SHIFT_ASSIGNMENTS_SITE_ID"
