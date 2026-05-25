#!/bin/bash
# シフト管理システム v2 サンプルデータ投入
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/env.sh"

echo "=== 資格マスタ ==="
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"施設警備業務検定1級"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"施設警備業務検定2級"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"交通誘導警備業務検定1級"}}'
plsnt record create --site-id "$QUALIFICATIONS_SITE_ID" --json '{"ClassHash":{"ClassA":"交通誘導警備業務検定2級"}}'

echo ""
echo "=== 現場マスタ ==="
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"渋谷オフィスビル","ClassB":"渋谷区渋谷1-1-1","ClassC":"鈴木太郎","ClassD":"03-1111-2222","ClassE":"定常契約"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"地下1F〜10F、夜間は正面玄関施錠"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"新宿商業施設","ClassB":"新宿区新宿3-3-3","ClassC":"田中花子","ClassD":"03-3333-4444","ClassE":"定常契約"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"地下2F〜8F、土日は来客多数"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"品川工事現場","ClassB":"品川区南品川5-5-5","ClassC":"山本一郎","ClassD":"03-5555-6666","ClassE":"期間契約"},"DateHash":{"DateA":"2026-04-01","DateB":"2026-06-30"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"交通誘導、工事車両出入管理"}}'
plsnt record create --site-id "$SITES_SITE_ID" --json '{"ClassHash":{"ClassA":"代々木公園イベント会場","ClassB":"渋谷区代々木神園町","ClassC":"佐々木健","ClassD":"03-7777-8888","ClassE":"スポット"},"DateHash":{"DateA":"2026-03-15"},"CheckHash":{"CheckA":true},"DescriptionHash":{"DescriptionA":"春祭りイベント警備、1日限り"}}'

echo ""
echo "=== 警備員マスタ ==="
QUAL_LIST=$(plsnt record list --site-id "$QUALIFICATIONS_SITE_ID" -o json)
QUAL1=$(echo "$QUAL_LIST" | jq -r '.[0].ResultId')
QUAL2=$(echo "$QUAL_LIST" | jq -r '.[1].ResultId')
QUAL3=$(echo "$QUAL_LIST" | jq -r '.[2].ResultId')
QUAL4=$(echo "$QUAL_LIST" | jq -r '.[3].ResultId')

plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"佐藤一郎\",\"ClassB\":\"G001\",\"ClassC\":\"090-1111-1111\",\"ClassD\":\"sato@example.com\",\"ClassE\":\"$QUAL1\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"高橋次郎\",\"ClassB\":\"G002\",\"ClassC\":\"090-2222-2222\",\"ClassD\":\"takahashi@example.com\",\"ClassE\":\"$QUAL2\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"伊藤三郎\",\"ClassB\":\"G003\",\"ClassC\":\"090-3333-3333\",\"ClassD\":\"ito@example.com\",\"ClassE\":\"$QUAL3\",\"ClassF\":\"契約社員\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"渡辺四郎\",\"ClassB\":\"G004\",\"ClassC\":\"090-4444-4444\",\"ClassD\":\"watanabe@example.com\",\"ClassE\":\"$QUAL4\",\"ClassF\":\"パート\"},\"CheckHash\":{\"CheckA\":true}}"
plsnt record create --site-id "$GUARDS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"中村五郎\",\"ClassB\":\"G005\",\"ClassC\":\"090-5555-5555\",\"ClassD\":\"nakamura@example.com\",\"ClassE\":\"$QUAL1\",\"ClassF\":\"正社員\"},\"CheckHash\":{\"CheckA\":true}}"

echo ""
echo "=== 現場シフト枠 ==="
SITE_LIST=$(plsnt record list --site-id "$SITES_SITE_ID" -o json)
SITE1=$(echo "$SITE_LIST" | jq -r '.[0].ResultId')
SITE2=$(echo "$SITE_LIST" | jq -r '.[1].ResultId')
SITE3=$(echo "$SITE_LIST" | jq -r '.[2].ResultId')
SITE4=$(echo "$SITE_LIST" | jq -r '.[3].ResultId')

# 渋谷オフィスビル: 定常 月〜金 日勤2名
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL1\"},\"DateHash\":{\"DateA\":\"09:00\",\"DateB\":\"18:00\"},\"NumHash\":{\"NumA\":2,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渋谷オフィスビル 定常 月〜金 日勤\"}}"

# 渋谷オフィスビル: 定常 月〜金 夜勤1名
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"定常\",\"ClassC\":\"夜勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL1\"},\"DateHash\":{\"DateA\":\"18:00\",\"DateB\":\"09:00\"},\"NumHash\":{\"NumA\":1,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渋谷オフィスビル 定常 月〜金 夜勤\"}}"

# 渋谷オフィスビル: 除外 年末年始
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateC\":\"2026-12-29\",\"DateD\":\"2027-01-03\"},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渋谷オフィスビル 除外 年末年始\",\"DescriptionB\":\"ビル全館休館のため\"}}"

# 渋谷オフィスビル: 期間 12月は8-21に延長（優先度高）
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"期間\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL1\"},\"DateHash\":{\"DateA\":\"08:00\",\"DateB\":\"21:00\",\"DateC\":\"2026-12-01\",\"DateD\":\"2026-12-28\"},\"NumHash\":{\"NumA\":3,\"NumB\":5},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渋谷オフィスビル 期間 12月繁忙期\",\"DescriptionB\":\"12月繁忙期対応: 時間延長+増員\"}}"

# 新宿商業施設: 定常 土日 日勤3名
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"土日\",\"ClassE\":\"$QUAL2\"},\"DateHash\":{\"DateA\":\"08:00\",\"DateB\":\"20:00\"},\"NumHash\":{\"NumA\":3,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"新宿商業施設 定常 土日 日勤\"}}"

# 新宿商業施設: 定常 月〜金 日勤1名
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL2\"},\"DateHash\":{\"DateA\":\"09:00\",\"DateB\":\"18:00\"},\"NumHash\":{\"NumA\":1,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"新宿商業施設 定常 月〜金 日勤\"}}"

# 品川工事現場: 期間 4/1〜6/30 月〜金 日勤1名（交通誘導）
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE3\",\"ClassB\":\"期間\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\",\"ClassE\":\"$QUAL3\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"17:00\",\"DateC\":\"2026-04-01\",\"DateD\":\"2026-06-30\"},\"NumHash\":{\"NumA\":1,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"品川工事現場 期間 月〜金 日勤\",\"DescriptionB\":\"4〜6月期間契約\"}}"

# 代々木公園: スポット 3/15 日勤5名
plsnt record create --site-id "$SITE_SLOTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"スポット\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"20:00\",\"DateE\":\"2026-03-15\"},\"NumHash\":{\"NumA\":5,\"NumB\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"代々木公園 スポット 2026-03-15 日勤\",\"DescriptionB\":\"春祭りイベント警備\"}}"

echo ""
echo "=== 稼働可能枠 ==="
GUARD_LIST=$(plsnt record list --site-id "$GUARDS_SITE_ID" -o json)
G1=$(echo "$GUARD_LIST" | jq -r '.[0].ResultId')
G2=$(echo "$GUARD_LIST" | jq -r '.[1].ResultId')
G3=$(echo "$GUARD_LIST" | jq -r '.[2].ResultId')
G4=$(echo "$GUARD_LIST" | jq -r '.[3].ResultId')
G5=$(echo "$GUARD_LIST" | jq -r '.[4].ResultId')

# 佐藤: 定常 月〜金 日勤
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"09:00\",\"DateB\":\"18:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"佐藤一郎 定常 月〜金 日勤\"}}"

# 佐藤: 追加 3/15（土曜だが出勤可能）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"20:00\",\"DateE\":\"2026-03-15\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"佐藤一郎 追加 2026-03-15 日勤\",\"DescriptionB\":\"イベント対応で休日出勤\"}}"

# 佐藤: 除外 3/20（有給）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G1\",\"ClassB\":\"除外\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateE\":\"2026-03-20\"},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"佐藤一郎 除外 2026-03-20\",\"DescriptionB\":\"有給取得\"}}"

# 高橋: 定常 月水金 → 個別レコード
for day in 月 水 金; do
  plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"$day\"},\"DateHash\":{\"DateA\":\"08:00\",\"DateB\":\"20:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"高橋次郎 定常 $day 日勤\"}}"
done

# 高橋: 定常 土日
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G2\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"土日\"},\"DateHash\":{\"DateA\":\"08:00\",\"DateB\":\"20:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"高橋次郎 定常 土日 日勤\"}}"

# 伊藤: 定常 月〜金 日勤（交通誘導）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"17:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"伊藤三郎 定常 月〜金 日勤\"}}"

# 伊藤: 追加 3/15
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"20:00\",\"DateE\":\"2026-03-15\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"伊藤三郎 追加 2026-03-15 日勤\",\"DescriptionB\":\"イベント対応\"}}"

# 渡辺: 定常 火木土 → 個別レコード
for day in 火 木; do
  plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"$day\"},\"DateHash\":{\"DateA\":\"09:00\",\"DateB\":\"18:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渡辺四郎 定常 $day 日勤\"}}"
done
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"定常\",\"ClassC\":\"日勤\",\"ClassD\":\"土\"},\"DateHash\":{\"DateA\":\"09:00\",\"DateB\":\"18:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渡辺四郎 定常 土 日勤\"}}"

# 渡辺: 追加 3/15
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"追加\",\"ClassC\":\"日勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateA\":\"07:00\",\"DateB\":\"20:00\",\"DateE\":\"2026-03-15\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"渡辺四郎 追加 2026-03-15 日勤\",\"DescriptionB\":\"イベント対応\"}}"

# 中村: 定常 月〜金 夜勤
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G5\",\"ClassB\":\"定常\",\"ClassC\":\"夜勤\",\"ClassD\":\"月〜金\"},\"DateHash\":{\"DateA\":\"18:00\",\"DateB\":\"09:00\"},\"NumHash\":{\"NumA\":10},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"中村五郎 定常 月〜金 夜勤\"}}"

# 中村: 除外 4月は夜勤不可（研修）
plsnt record create --site-id "$AVAILABILITY_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G5\",\"ClassB\":\"除外\",\"ClassC\":\"夜勤\",\"ClassD\":\"指定なし\"},\"DateHash\":{\"DateC\":\"2026-04-01\",\"DateD\":\"2026-04-30\"},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"中村五郎 除外 4月 夜勤\",\"DescriptionB\":\"研修期間のため夜勤不可\"}}"

echo ""
echo "=== シフト割当（サンプル）==="
# 月曜日のシフト例
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G1\",\"ClassD\":\"定常\"},\"DateHash\":{\"DateA\":\"2026-03-09\",\"DateB\":\"2026-03-09T09:00:00\",\"DateC\":\"2026-03-09T18:00:00\"},\"NumHash\":{\"NumA\":9},\"DescriptionHash\":{\"DescriptionA\":\"佐藤一郎 渋谷オフィスビル 3/9\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G2\",\"ClassD\":\"定常\"},\"DateHash\":{\"DateA\":\"2026-03-09\",\"DateB\":\"2026-03-09T09:00:00\",\"DateC\":\"2026-03-09T18:00:00\"},\"NumHash\":{\"NumA\":9},\"DescriptionHash\":{\"DescriptionA\":\"高橋次郎 渋谷オフィスビル 3/9\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE1\",\"ClassB\":\"$G5\",\"ClassD\":\"定常\"},\"DateHash\":{\"DateA\":\"2026-03-09\",\"DateB\":\"2026-03-09T18:00:00\",\"DateC\":\"2026-03-10T09:00:00\"},\"NumHash\":{\"NumA\":15},\"DescriptionHash\":{\"DescriptionA\":\"中村五郎 渋谷オフィスビル 3/9 夜勤\"}}"

# 3/15 スポット（イベント）
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"$G1\",\"ClassD\":\"スポット\"},\"DateHash\":{\"DateA\":\"2026-03-15\",\"DateB\":\"2026-03-15T07:00:00\",\"DateC\":\"2026-03-15T20:00:00\"},\"NumHash\":{\"NumA\":13},\"DescriptionHash\":{\"DescriptionA\":\"佐藤一郎 代々木公園イベント 3/15\"}}"
plsnt record create --site-id "$SHIFT_ASSIGNMENTS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$SITE4\",\"ClassB\":\"$G3\",\"ClassD\":\"スポット\"},\"DateHash\":{\"DateA\":\"2026-03-15\",\"DateB\":\"2026-03-15T07:00:00\",\"DateC\":\"2026-03-15T20:00:00\"},\"NumHash\":{\"NumA\":13},\"DescriptionHash\":{\"DescriptionA\":\"伊藤三郎 代々木公園イベント 3/15\"}}"

echo ""
echo "=== サンプルデータ投入完了 ==="
echo "資格: 4件, 現場: 4件, 警備員: 5名"
echo "現場シフト枠: 8件（定常3/期間2/スポット1/除外2）"
echo "稼働可能枠: 16件（定常9/追加3/除外2）"
echo "シフト割当: 5件（定常3/スポット2）"
