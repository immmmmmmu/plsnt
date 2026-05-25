#!/bin/bash
# シフト管理システム UIスクリプト一括デプロイ
# 使い方: source scripts/shift-management/env.sh && bash scripts/shift-management/pleasanter-scripts/deploy-all.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 環境変数チェック
for var in SHIFT_ASSIGNMENTS_SITE_ID SITE_SLOTS_SITE_ID AVAILABILITY_SITE_ID; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: $var is not set. Run 'source scripts/shift-management/env.sh' first." >&2
    exit 1
  fi
done

deploy_scripts() {
  local site_id="$1"
  local table_name="$2"
  shift 2
  local scripts="$@"

  echo "=== $table_name (SiteID: $site_id) ==="

  # 現在のSiteSettings取得
  local current
  current=$(plsnt site get "$site_id" -o json)

  # Scripts配列を構築
  local script_array="["
  local id=1
  local first=true
  for script_file in $scripts; do
    local title
    title=$(basename "$script_file" .js)
    local body
    body=$(grep -v '^//' "$script_file" | jq -Rs .)

    # 実行タイミングを判定
    local is_index=false
    local is_new=true
    local is_edit=true
    if echo "$title" | grep -q "highlight\|index\|list"; then
      is_index=true
      is_new=false
      is_edit=false
    fi

    if [ "$first" = true ]; then
      first=false
    else
      script_array+=","
    fi
    script_array+="{\"Id\":$id,\"Title\":\"$title\",\"Body\":$body,\"New\":$is_new,\"Edit\":$is_edit,\"Index\":$is_index}"
    id=$((id + 1))
  done
  script_array+="]"

  # 既存SiteSettingsにScriptsをマージして更新
  local new_settings
  new_settings=$(echo "$current" | jq --argjson scripts "$script_array" '
    .SiteSettings | . + {"Scripts": $scripts}
  ')

  plsnt site update "$site_id" --json "{\"SiteSettings\":$new_settings}"
  echo "  Deployed $(echo "$scripts" | wc -w | tr -d ' ') scripts"
}

# シフト割当
deploy_scripts "$SHIFT_ASSIGNMENTS_SITE_ID" "シフト割当" \
  "$SCRIPT_DIR/shift-assignments/shift-date-default.js" \
  "$SCRIPT_DIR/shift-assignments/shift-status-highlight.js" \
  "$SCRIPT_DIR/shift-assignments/shift-hours-calc.js"

# 現場シフト枠
deploy_scripts "$SITE_SLOTS_SITE_ID" "現場シフト枠" \
  "$SCRIPT_DIR/site-slots/slot-name-auto.js" \
  "$SCRIPT_DIR/site-slots/slot-field-toggle.js"

# 稼働可能枠
deploy_scripts "$AVAILABILITY_SITE_ID" "稼働可能枠" \
  "$SCRIPT_DIR/availability/availability-name-auto.js" \
  "$SCRIPT_DIR/availability/availability-field-toggle.js"

echo ""
echo "=== All scripts deployed successfully ==="
