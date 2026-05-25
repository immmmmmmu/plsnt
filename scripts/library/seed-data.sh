#!/usr/bin/env bash
# 図書貸出管理モデル サンプルデータ投入
# 使い方: source scripts/library/env.sh && bash scripts/library/seed-data.sh
set -euo pipefail

: "${LIB_GENRE_SITE:?env.sh を source してください}"

echo "=== ジャンルマスタ ==="
GEN_LIT=$(plsnt record create --site-id "$LIB_GENRE_SITE" --json '{"ClassHash":{"ClassA":"文学"}}' -o json | jq -r '.Id')
GEN_SCI=$(plsnt record create --site-id "$LIB_GENRE_SITE" --json '{"ClassHash":{"ClassA":"科学"}}' -o json | jq -r '.Id')
GEN_HIST=$(plsnt record create --site-id "$LIB_GENRE_SITE" --json '{"ClassHash":{"ClassA":"歴史"}}' -o json | jq -r '.Id')
GEN_KIDS=$(plsnt record create --site-id "$LIB_GENRE_SITE" --json '{"ClassHash":{"ClassA":"児童書"}}' -o json | jq -r '.Id')
echo "  文学=$GEN_LIT, 科学=$GEN_SCI, 歴史=$GEN_HIST, 児童書=$GEN_KIDS"

echo "=== 出版社マスタ ==="
PUB_KODANSHA=$(plsnt record create --site-id "$LIB_PUBLISHER_SITE" --json '{"ClassHash":{"ClassA":"講談社"}}' -o json | jq -r '.Id')
PUB_IWANAMI=$(plsnt record create --site-id "$LIB_PUBLISHER_SITE" --json '{"ClassHash":{"ClassA":"岩波書店"}}' -o json | jq -r '.Id')
PUB_SHINCHOSHA=$(plsnt record create --site-id "$LIB_PUBLISHER_SITE" --json '{"ClassHash":{"ClassA":"新潮社"}}' -o json | jq -r '.Id')
echo "  講談社=$PUB_KODANSHA, 岩波=$PUB_IWANAMI, 新潮=$PUB_SHINCHOSHA"

echo "=== 利用者マスタ ==="
MEM_YAMADA=$(plsnt record create --site-id "$LIB_MEMBER_SITE" --json '{"ClassHash":{"ClassA":"山田太郎","ClassB":"03-1111-2222","ClassC":"yamada@example.com"},"DescriptionHash":{"DescriptionA":"東京都世田谷区1-2-3"}}' -o json | jq -r '.Id')
MEM_SUZUKI=$(plsnt record create --site-id "$LIB_MEMBER_SITE" --json '{"ClassHash":{"ClassA":"鈴木花子","ClassB":"03-3333-4444","ClassC":"suzuki@example.com"},"DescriptionHash":{"DescriptionA":"東京都杉並区4-5-6"}}' -o json | jq -r '.Id')
MEM_SATO=$(plsnt record create --site-id "$LIB_MEMBER_SITE" --json '{"ClassHash":{"ClassA":"佐藤次郎","ClassB":"03-5555-6666","ClassC":"sato@example.com"},"DescriptionHash":{"DescriptionA":"東京都練馬区7-8-9"}}' -o json | jq -r '.Id')
echo "  山田=$MEM_YAMADA, 鈴木=$MEM_SUZUKI, 佐藤=$MEM_SATO"

echo "=== 書架マスタ ==="
SH_1F_A=$(plsnt record create --site-id "$LIB_SHELF_SITE" --json '{"ClassHash":{"ClassA":"1F-A 文学棚","ClassB":"1階","ClassC":"小説・エッセイ"}}' -o json | jq -r '.Id')
SH_1F_B=$(plsnt record create --site-id "$LIB_SHELF_SITE" --json '{"ClassHash":{"ClassA":"1F-B 児童書棚","ClassB":"1階","ClassC":"絵本・児童文学"}}' -o json | jq -r '.Id')
SH_2F_A=$(plsnt record create --site-id "$LIB_SHELF_SITE" --json '{"ClassHash":{"ClassA":"2F-A 科学棚","ClassB":"2階","ClassC":"自然科学・技術"}}' -o json | jq -r '.Id')
SH_2F_B=$(plsnt record create --site-id "$LIB_SHELF_SITE" --json '{"ClassHash":{"ClassA":"2F-B 歴史棚","ClassB":"2階","ClassC":"日本史・世界史"}}' -o json | jq -r '.Id')
echo "  1F-A=$SH_1F_A, 1F-B=$SH_1F_B, 2F-A=$SH_2F_A, 2F-B=$SH_2F_B"

echo "=== 書籍マスタ ==="
BK_KOKORO=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"こころ\",\"ClassB\":\"$GEN_LIT\",\"ClassC\":\"978-4-10-101001-1\",\"ClassD\":\"$PUB_SHINCHOSHA\"},\"NumHash\":{\"NumA\":407}}" -o json | jq -r '.Id')
BK_SNOW=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"雪国\",\"ClassB\":\"$GEN_LIT\",\"ClassC\":\"978-4-10-101002-8\",\"ClassD\":\"$PUB_SHINCHOSHA\"},\"NumHash\":{\"NumA\":473}}" -o json | jq -r '.Id')
BK_COSMOS=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"コスモス\",\"ClassB\":\"$GEN_SCI\",\"ClassC\":\"978-4-02-260000-1\",\"ClassD\":\"$PUB_IWANAMI\"},\"NumHash\":{\"NumA\":1100}}" -o json | jq -r '.Id')
BK_SAPIENS=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"サピエンス全史\",\"ClassB\":\"$GEN_HIST\",\"ClassC\":\"978-4-309-22671-2\",\"ClassD\":\"$PUB_KODANSHA\"},\"NumHash\":{\"NumA\":2090}}" -o json | jq -r '.Id')
BK_MOMO=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"モモ\",\"ClassB\":\"$GEN_KIDS\",\"ClassC\":\"978-4-00-114006-9\",\"ClassD\":\"$PUB_IWANAMI\"},\"NumHash\":{\"NumA\":880}}" -o json | jq -r '.Id')
BK_FIRE=$(plsnt record create --site-id "$LIB_BOOK_SITE" --json "{\"ClassHash\":{\"ClassA\":\"火花\",\"ClassB\":\"$GEN_LIT\",\"ClassC\":\"978-4-16-390582-1\",\"ClassD\":\"$PUB_KODANSHA\"},\"NumHash\":{\"NumA\":1320}}" -o json | jq -r '.Id')
echo "  こころ=$BK_KOKORO, 雪国=$BK_SNOW, コスモス=$BK_COSMOS, サピエンス=$BK_SAPIENS, モモ=$BK_MOMO, 火花=$BK_FIRE"

echo "=== 蔵書（書架×書籍）==="
declare -A SHELF_NAMES=(["$SH_1F_A"]="1F-A 文学棚" ["$SH_1F_B"]="1F-B 児童書棚" ["$SH_2F_A"]="2F-A 科学棚" ["$SH_2F_B"]="2F-B 歴史棚")
declare -A BOOK_NAMES=(["$BK_KOKORO"]="こころ" ["$BK_SNOW"]="雪国" ["$BK_COSMOS"]="コスモス" ["$BK_SAPIENS"]="サピエンス全史" ["$BK_MOMO"]="モモ" ["$BK_FIRE"]="火花")

# 文学棚: こころ, 雪国, 火花
for BK_ID in "$BK_KOKORO" "$BK_SNOW" "$BK_FIRE"; do
  QTY=$((RANDOM % 3 + 2))
  COL_DESC="${SHELF_NAMES[$SH_1F_A]} - ${BOOK_NAMES[$BK_ID]}"
  plsnt record create --site-id "$LIB_COLLECTION_SITE" \
    --json "{\"ClassHash\":{\"ClassA\":\"$SH_1F_A\",\"ClassB\":\"$BK_ID\"},\"DescriptionHash\":{\"DescriptionA\":\"$COL_DESC\"},\"NumHash\":{\"NumA\":$QTY}}" \
    -o json > /dev/null
done
# 児童書棚: モモ
COL_DESC="${SHELF_NAMES[$SH_1F_B]} - ${BOOK_NAMES[$BK_MOMO]}"
plsnt record create --site-id "$LIB_COLLECTION_SITE" \
  --json "{\"ClassHash\":{\"ClassA\":\"$SH_1F_B\",\"ClassB\":\"$BK_MOMO\"},\"DescriptionHash\":{\"DescriptionA\":\"$COL_DESC\"},\"NumHash\":{\"NumA\":3}}" \
  -o json > /dev/null
# 科学棚: コスモス
COL_DESC="${SHELF_NAMES[$SH_2F_A]} - ${BOOK_NAMES[$BK_COSMOS]}"
plsnt record create --site-id "$LIB_COLLECTION_SITE" \
  --json "{\"ClassHash\":{\"ClassA\":\"$SH_2F_A\",\"ClassB\":\"$BK_COSMOS\"},\"DescriptionHash\":{\"DescriptionA\":\"$COL_DESC\"},\"NumHash\":{\"NumA\":2}}" \
  -o json > /dev/null
# 歴史棚: サピエンス全史
COL_DESC="${SHELF_NAMES[$SH_2F_B]} - ${BOOK_NAMES[$BK_SAPIENS]}"
plsnt record create --site-id "$LIB_COLLECTION_SITE" \
  --json "{\"ClassHash\":{\"ClassA\":\"$SH_2F_B\",\"ClassB\":\"$BK_SAPIENS\"},\"DescriptionHash\":{\"DescriptionA\":\"$COL_DESC\"},\"NumHash\":{\"NumA\":2}}" \
  -o json > /dev/null
echo "  4書架 x 各書籍 = 7件の蔵書レコード作成"

echo ""
echo "=== シードデータ投入完了 ==="
echo ""
echo "環境変数として使える RecordID:"
echo "  export MEM_YAMADA=$MEM_YAMADA MEM_SUZUKI=$MEM_SUZUKI MEM_SATO=$MEM_SATO"
echo "  export SH_1F_A=$SH_1F_A SH_1F_B=$SH_1F_B SH_2F_A=$SH_2F_A SH_2F_B=$SH_2F_B"
echo "  export BK_KOKORO=$BK_KOKORO BK_SNOW=$BK_SNOW BK_COSMOS=$BK_COSMOS"
echo "  export BK_SAPIENS=$BK_SAPIENS BK_MOMO=$BK_MOMO BK_FIRE=$BK_FIRE"
