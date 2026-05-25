#!/bin/bash
# シフト管理システム v3 サンプルデータ投入
# 使い方: source scripts/shift-management-v3/env.sh && bash scripts/shift-management-v3/seed-data.sh

set -euo pipefail

# 環境変数チェック
for var in QUALIFICATIONS_SITE_ID SITES_SITE_ID GUARDS_SITE_ID SITE_SLOTS_SITE_ID AVAILABILITY_SITE_ID SHIFT_ASSIGNMENTS_SITE_ID; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: $var is not set. Run 'source scripts/shift-management-v3/env.sh' first." >&2
    exit 1
  fi
done

echo "=== 資格マスタ ==="
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"施設警備業務検定1級"},"DescriptionHash":{"DescriptionA":"施設警備の国家資格"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"施設警備業務検定2級"},"DescriptionHash":{"DescriptionA":"施設警備の基本資格"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"交通誘導警備業務検定"},"DescriptionHash":{"DescriptionA":"交通誘導の国家資格"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"普通救命講習修了"},"DescriptionHash":{"DescriptionA":"AED・心肺蘇生法"}}'

echo ""
echo "=== 現場マスタ ==="
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"渋谷オフィスビル","ClassB":"東京都渋谷区渋谷1-1-1","ClassC":"山本部長","ClassD":"03-1111-2222","ClassE":"定常契約"},"DateHash":{"DateA":"2025-04-01T00:00:00","DateB":"2027-03-31T00:00:00"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"地下1F〜地上20F、常時2名体制"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"新宿商業施設","ClassB":"東京都新宿区新宿3-3-3","ClassC":"田村マネージャー","ClassD":"03-3333-4444","ClassE":"定常契約"},"DateHash":{"DateA":"2025-06-01T00:00:00","DateB":"2027-05-31T00:00:00"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"B1〜5F商業フロア、日勤2名+夜勤1名"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"品川建設現場","ClassB":"東京都品川区東品川4-4-4","ClassC":"工藤現場監督","ClassD":"03-5555-6666","ClassE":"期間契約"},"DateHash":{"DateA":"2026-01-15T00:00:00","DateB":"2026-06-30T00:00:00"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"大規模建設現場、交通誘導資格必須"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"六本木イベント会場","ClassB":"東京都港区六本木5-5-5","ClassC":"伊藤プロデューサー","ClassD":"03-7777-8888","ClassE":"スポット"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"単発イベント警備"}}'

echo ""
echo "=== 警備員マスタ ==="
QUAL_IDS=$(plsnt record list --site-id "$QUALIFICATIONS_SITE_ID" -o ids)
QUAL1=$(echo "$QUAL_IDS" | sed -n '1p')
QUAL2=$(echo "$QUAL_IDS" | sed -n '2p')
QUAL3=$(echo "$QUAL_IDS" | sed -n '3p')
QUAL4=$(echo "$QUAL_IDS" | sed -n '4p')

plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"田中太郎\",\"ClassB\":\"G001\",\"ClassC\":\"090-1111-1111\",\"ClassD\":\"tanaka@example.com\",\"ClassE\":\"$QUAL1\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"鈴木花子\",\"ClassB\":\"G002\",\"ClassC\":\"090-2222-2222\",\"ClassD\":\"suzuki@example.com\",\"ClassE\":\"$QUAL2\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"佐藤次郎\",\"ClassB\":\"G003\",\"ClassC\":\"090-3333-3333\",\"ClassD\":\"sato@example.com\",\"ClassE\":\"$QUAL3\",\"ClassF\":\"契約社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"山田美咲\",\"ClassB\":\"G004\",\"ClassC\":\"090-4444-4444\",\"ClassD\":\"yamada@example.com\",\"ClassE\":\"$QUAL4\",\"ClassF\":\"パート\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"高橋健一\",\"ClassB\":\"G005\",\"ClassC\":\"090-5555-5555\",\"ClassD\":\"takahashi@example.com\",\"ClassE\":\"$QUAL1\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"

echo ""
echo "=== 現場シフト枠 ==="
SITE_IDS=$(plsnt record list --site-id "$SITES_SITE_ID" -o ids)
SITE1=$(echo "$SITE_IDS" | sed -n '1p')
SITE2=$(echo "$SITE_IDS" | sed -n '2p')
SITE3=$(echo "$SITE_IDS" | sed -n '3p')
SITE4=$(echo "$SITE_IDS" | sed -n '4p')

# 定常枠
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL2\"},\"DateHash\":{\"DateA\":\"2026-01-01T09:00:00\",\"DateB\":\"2026-01-01T18:00:00\"},\"NumHash\":{\"NumA\":2,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 渋谷オフィスビル 月〜金 日勤\"}}"
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"定常\",\"ClassC\":\"夜勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL2\"},\"DateHash\":{\"DateA\":\"2026-01-01T18:00:00\",\"DateB\":\"2026-01-02T09:00:00\"},\"NumHash\":{\"NumA\":1,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 渋谷オフィスビル 月〜金 夜勤\"}}"
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"毎日\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T20:00:00\"},\"NumHash\":{\"NumA\":2,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 新宿商業施設 毎日 日勤\"}}"

# 期間枠
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"期間\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL3\"},\"DateHash\":{\"DateA\":\"2026-01-01T08:00:00\",\"DateB\":\"2026-01-01T17:00:00\",\"DateC\":\"2026-01-15T00:00:00\",\"DateD\":\"2026-06-30T00:00:00\"},\"NumHash\":{\"NumA\":1,\"NumB\":5},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"期間 品川建設現場 月〜金 日勤\"}}"
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"期間\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL2\"},\"DateHash\":{\"DateA\":\"2026-01-01T08:00:00\",\"DateB\":\"2026-01-01T21:00:00\",\"DateC\":\"2026-03-01T00:00:00\",\"DateD\":\"2026-03-31T00:00:00\"},\"NumHash\":{\"NumA\":3,\"NumB\":5},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"期間 渋谷オフィスビル 月〜金 日勤 3月繁忙期\",\"DescriptionB\":\"3月は年度末で来客が多いため増員・時間延長\"}}"

# スポット枠
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"スポット\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T22:00:00\",\"DateE\":\"2026-03-15T00:00:00\"},\"NumHash\":{\"NumA\":5,\"NumB\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"スポット 六本木イベント会場 2026-03-15\",\"DescriptionB\":\"音楽フェス警備\"}}"

# 除外枠
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateC\":\"2026-04-29T00:00:00\",\"DateD\":\"2026-05-06T00:00:00\"},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"除外 渋谷オフィスビル GW休止\",\"DescriptionB\":\"GW期間はビル全館休業\"}}"
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateE\":\"2026-01-01T00:00:00\"},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"除外 新宿商業施設 2026-01-01\",\"DescriptionB\":\"元旦は全館休業\"}}"

echo ""
echo "=== 稼働可能枠 ==="
GUARD_IDS=$(plsnt record list --site-id "$GUARDS_SITE_ID" -o ids)
G1=$(echo "$GUARD_IDS" | sed -n '1p')
G2=$(echo "$GUARD_IDS" | sed -n '2p')
G3=$(echo "$GUARD_IDS" | sed -n '3p')
G4=$(echo "$GUARD_IDS" | sed -n '4p')
G5=$(echo "$GUARD_IDS" | sed -n '5p')

# 定常枠（基本シフト）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"2026-01-01T09:00:00\",\"DateB\":\"2026-01-01T18:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 田中太郎 月〜金 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"2026-01-01T09:00:00\",\"DateB\":\"2026-01-01T18:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 鈴木花子 月〜金 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"2026-01-01T08:00:00\",\"DateB\":\"2026-01-01T17:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 佐藤次郎 月〜金 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T16:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 山田美咲 月 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"水\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T16:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 山田美咲 水 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"金\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T16:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 山田美咲 金 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G5\",\"ClassB\":\"定常\",\"ClassC\":\"夜勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"2026-01-01T18:00:00\",\"DateB\":\"2026-01-02T09:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 高橋健一 月〜金 夜勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"土日\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T18:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 田中太郎 土日 日勤\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G5\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"土日\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T18:00:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"定常 高橋健一 土日 日勤\"}}"

# 追加枠（特定日の出勤可能）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G2\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T22:00:00\",\"DateE\":\"2026-03-15T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"追加 鈴木花子 2026-03-15\",\"DescriptionB\":\"イベント警備に参加可能\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T22:00:00\",\"DateE\":\"2026-03-15T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"追加 山田美咲 2026-03-15\",\"DescriptionB\":\"イベント警備に参加可能\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G5\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"2026-01-01T10:00:00\",\"DateB\":\"2026-01-01T22:00:00\",\"DateE\":\"2026-03-15T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"追加 高橋健一 2026-03-15\",\"DescriptionB\":\"イベント警備に参加可能\"}}"

# 除外枠
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateE\":\"2026-03-10T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"除外 田中太郎 2026-03-10\",\"DescriptionB\":\"有給休暇\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"除外\",\"ClassC\":\"夜勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateC\":\"2026-04-01T00:00:00\",\"DateD\":\"2026-04-30T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"除外 佐藤次郎 4月夜勤不可\",\"DescriptionB\":\"研修参加のため夜勤不可\"}}"
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G2\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateC\":\"2026-03-20T00:00:00\",\"DateD\":\"2026-03-25T00:00:00\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"除外 鈴木花子 3/20-25\",\"DescriptionB\":\"旅行のため不可\"}}"

echo ""
echo "=== シフト割当（StartTime/CompletionTime使用 - カレンダー最適化）==="
# 3/9(月)〜3/13(金) の1週間分サンプルデータ
for DAY in 09 10 11 12 13; do
  # 渋谷ビル 日勤 田中
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-03-${DAY}T09:00:00\",\"CompletionTime\":\"2026-03-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤\"}}"

  # 渋谷ビル 日勤 鈴木
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-03-${DAY}T09:00:00\",\"CompletionTime\":\"2026-03-${DAY}T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 渋谷 日勤\"}}"

  # 渋谷ビル 夜勤 高橋
  NEXT_DAY=$(date -d "2026-03-${DAY} +1 day" +%Y-%m-%dT09:00:00)
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-03-${DAY}T18:00:00\",\"CompletionTime\":\"${NEXT_DAY}\",\"NumHash\":{\"NumA\":15},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"高橋 渋谷 夜勤\"}}"

  # 品川建設 日勤 佐藤
  plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"$G3\",\"ClassD\":\"定常\"},\"StartTime\":\"2026-03-${DAY}T08:00:00\",\"CompletionTime\":\"2026-03-${DAY}T17:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"佐藤 品川 日勤\"}}"
done

# 3/10 田中欠勤 → 山田が代替
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G4\",\"ClassD\":\"代替\",\"ClassE\":\"私用\"},\"StartTime\":\"2026-03-10T09:00:00\",\"CompletionTime\":\"2026-03-10T18:00:00\",\"NumHash\":{\"NumA\":9},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"山田 渋谷 日勤(代替)\"}}"

# 3/15 イベント警備（スポット）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"$G1\",\"ClassD\":\"スポット\"},\"StartTime\":\"2026-03-15T10:00:00\",\"CompletionTime\":\"2026-03-15T22:00:00\",\"NumHash\":{\"NumA\":12},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"田中 六本木 イベント\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"$G2\",\"ClassD\":\"スポット\"},\"StartTime\":\"2026-03-15T10:00:00\",\"CompletionTime\":\"2026-03-15T22:00:00\",\"NumHash\":{\"NumA\":12},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"鈴木 六本木 イベント\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"$G5\",\"ClassD\":\"スポット\"},\"StartTime\":\"2026-03-15T10:00:00\",\"CompletionTime\":\"2026-03-15T22:00:00\",\"NumHash\":{\"NumA\":12},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"高橋 六本木 イベント\"}}"

# 欠勤レコード（3/10 田中）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\",\"ClassE\":\"私用\"},\"StartTime\":\"2026-03-10T09:00:00\",\"CompletionTime\":\"2026-03-10T18:00:00\",\"NumHash\":{\"NumA\":0},\"Status\":910,\"DescriptionHash\":{\"DescriptionA\":\"田中 渋谷 日勤(欠勤)\"}}"

echo ""
echo "=== サンプルデータ投入完了 ==="
echo "カレンダービューで確認: http://localhost/items/$SHIFT_ASSIGNMENTS_SITE_ID"
