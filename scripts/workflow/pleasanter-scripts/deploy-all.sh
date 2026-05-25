#!/usr/bin/env bash
# deploy-all.sh: ワークフローアプリのスクリプト/スタイル一括デプロイ
#
# Usage: ./scripts/workflow/pleasanter-scripts/deploy-all.sh <header_site_id> [detail_site_id]
#
#   header_site_id: 申請ヘッダの SiteID（Scripts/Styles をデプロイ）
#   detail_site_id: 申請明細の SiteID（total-calc.js の明細参照先）
#                   省略時は 0（total-calc.js 内の SiteID が未置換のまま）
#
# デプロイ内容:
#   Scripts:
#     Id:1 amount-helper  (New:true,  Edit:true,  Index:false) - 金額入力補助
#     Id:2 total-calc     (New:false, Edit:true,  Index:false) - 明細合計自動計算
#   Styles:
#     Id:1 approval-status (New:false, Edit:false, Index:true) - ステータス色分け
#
# 既存の SiteSettings (Links, Columns, Processes 等) を保持してマージします。

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
PLSNT="${PLSNT:-$PROJECT_ROOT/plsnt}"

HEADER_SITE_ID="${1:-}"
DETAIL_SITE_ID="${2:-0}"

if [ -z "$HEADER_SITE_ID" ]; then
  echo "Usage: $0 <header_site_id> [detail_site_id]" >&2
  echo "  header_site_id: 申請ヘッダの SiteID" >&2
  echo "  detail_site_id: 申請明細の SiteID (default: 0)" >&2
  exit 1
fi

echo "=== ワークフロー スクリプト/スタイル デプロイ ==="
echo "  申請ヘッダ SiteID: $HEADER_SITE_ID"
echo "  申請明細 SiteID:   $DETAIL_SITE_ID"
echo ""

# 1. 現在の SiteSettings を取得
echo "--- Phase 1: 現在の SiteSettings 取得 ---"
CURRENT=$("$PLSNT" site get "$HEADER_SITE_ID" -o json 2>/dev/null)
if [ -z "$CURRENT" ]; then
  echo "ERROR: site get $HEADER_SITE_ID が空を返しました" >&2
  exit 1
fi
echo "  OK"
echo ""

# 2. スクリプト/スタイルファイルを読み込み・JSON文字列化
echo "--- Phase 2: ファイル読み込み + SiteSettings マージ ---"

# amount-helper.js
AMOUNT_HELPER_BODY=$(jq -Rs . < "$SCRIPT_DIR/amount-helper.js")

# total-calc.js: DETAIL_SITE_ID を実際の SiteID に sed 置換してからJSON文字列化
if [ "$DETAIL_SITE_ID" != "0" ]; then
  TOTAL_CALC_BODY=$(sed -E "s/var DETAIL_SITE_ID = [0-9]+;/var DETAIL_SITE_ID = $DETAIL_SITE_ID;/" \
    "$SCRIPT_DIR/total-calc.js" | jq -Rs .)
else
  TOTAL_CALC_BODY=$(jq -Rs . < "$SCRIPT_DIR/total-calc.js")
fi

# approval-status.css
APPROVAL_CSS_BODY=$(jq -Rs . < "$SCRIPT_DIR/approval-status.css")

# 3. 既存 SiteSettings に Scripts/Styles をマージ（jq --argjson パターン）
#    shift-management-v3/pleasanter-scripts/deploy-all.sh と同じパターン
SCRIPTS_JSON="[
  {\"Id\":1,\"Title\":\"amount-helper\",\"Body\":$AMOUNT_HELPER_BODY,\"New\":true,\"Edit\":true,\"Index\":false},
  {\"Id\":2,\"Title\":\"total-calc\",\"Body\":$TOTAL_CALC_BODY,\"New\":false,\"Edit\":true,\"Index\":false}
]"

STYLES_JSON="[
  {\"Id\":1,\"Title\":\"approval-status\",\"Body\":$APPROVAL_CSS_BODY,\"New\":false,\"Edit\":false,\"Index\":true}
]"

NEW_SETTINGS=$(echo "$CURRENT" | jq \
  --argjson scripts "$SCRIPTS_JSON" \
  --argjson styles "$STYLES_JSON" '
  .SiteSettings | . + {"Scripts": $scripts, "Styles": $styles}
')

echo "  Scripts: 2件 (amount-helper, total-calc)"
echo "  Styles:  1件 (approval-status)"
echo ""

# 4. site update でデプロイ
echo "--- Phase 3: site update ---"
"$PLSNT" site update "$HEADER_SITE_ID" --json "{\"SiteSettings\":$NEW_SETTINGS}"
echo "  OK"
echo ""

echo "=== デプロイ完了 ==="
echo ""
echo "確認URL: http://localhost/items/$HEADER_SITE_ID"
echo ""
echo "デプロイ内容:"
echo "  Scripts:"
echo "    #1 amount-helper  (New, Edit)  - 金額入力補助（カンマ区切り）"
echo "    #2 total-calc     (Edit)       - 明細合計自動計算 (明細SiteID: $DETAIL_SITE_ID)"
echo "  Styles:"
echo "    #1 approval-status (Index)     - ステータス別行色分け"
