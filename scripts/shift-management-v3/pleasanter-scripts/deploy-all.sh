#!/bin/bash
# シフト管理システム v3 UIスクリプト一括デプロイ（ダッシュボード対応版）
# 使い方: source scripts/shift-management-v3/env.sh && bash scripts/shift-management-v3/pleasanter-scripts/deploy-all.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 環境変数チェック
for var in SHIFT_ASSIGNMENTS_SITE_ID SITE_SLOTS_SITE_ID AVAILABILITY_SITE_ID; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: $var is not set. Run 'source scripts/shift-management-v3/env.sh' first." >&2
    exit 1
  fi
done

# Scripts/Styles をデプロイする汎用関数
# 引数: site_id table_name scripts_json [style_files...]
# scripts_json: '[{"file":"path.js","title":"名前","new":true,"edit":true,"index":false},...]'
deploy_to_site() {
  local site_id="$1"
  local table_name="$2"
  local scripts_def="$3"
  shift 3
  local css_files=("$@")

  echo "=== $table_name (SiteID: $site_id) ==="

  # 現在のSiteSettings取得
  local current
  current=$(plsnt site get "$site_id" -o json 2>/dev/null)
  if [[ -z "$current" ]]; then
    echo "  ERROR: SiteID $site_id の取得に失敗しました" >&2
    return 1
  fi

  # Scripts配列を構築（定義JSONから）
  local script_array
  script_array=$(echo "$scripts_def" | jq -c '[
    to_entries[] | {
      Id: (.key + 1),
      Title: .value.title,
      Body: (.value.file | ltrimstr("'$SCRIPT_DIR'/") | . as $f | input_line_number),
      New: .value.new,
      Edit: .value.edit,
      Index: .value.index
    }
  ]' 2>/dev/null || echo "[]")

  # jqでファイル読み込みは複雑なので、シンプルにループで構築
  script_array="["
  local id=1
  local first=true
  local count=0
  for row in $(echo "$scripts_def" | jq -r '.[] | @base64'); do
    local decoded
    decoded=$(echo "$row" | base64 -d)
    local file title is_new is_edit is_index
    file=$(echo "$decoded" | jq -r '.file')
    title=$(echo "$decoded" | jq -r '.title')
    is_new=$(echo "$decoded" | jq -r '.new')
    is_edit=$(echo "$decoded" | jq -r '.edit')
    is_index=$(echo "$decoded" | jq -r '.index')

    local body
    body=$(jq -Rs . < "$file")

    if [ "$first" = true ]; then first=false; else script_array+=","; fi
    script_array+="{\"Id\":$id,\"Title\":\"$title\",\"Body\":$body,\"New\":$is_new,\"Edit\":$is_edit,\"Index\":$is_index}"
    id=$((id + 1))
    count=$((count + 1))
  done
  script_array+="]"

  # Styles配列を構築
  local style_array="[]"
  local style_count=0
  if [ ${#css_files[@]} -gt 0 ]; then
    style_array="["
    local sid=1
    local sfirst=true
    for css_file in "${css_files[@]}"; do
      local stitle sbody
      stitle=$(basename "$css_file" .css)
      sbody=$(jq -Rs . < "$css_file")
      if [ "$sfirst" = true ]; then sfirst=false; else style_array+=","; fi
      style_array+="{\"Id\":$sid,\"Title\":\"$stitle\",\"Body\":$sbody,\"New\":true,\"Edit\":true,\"Index\":true}"
      sid=$((sid + 1))
      style_count=$((style_count + 1))
    done
    style_array+="]"
  fi

  # 既存SiteSettingsにScripts/Stylesをマージして更新
  local new_settings
  new_settings=$(echo "$current" | jq \
    --argjson scripts "$script_array" \
    --argjson styles "$style_array" '
    .SiteSettings | . + {"Scripts": $scripts, "Styles": $styles}
  ')

  plsnt site update "$site_id" --json "{\"SiteSettings\":$new_settings}"
  echo "  Deployed $count scripts, $style_count styles"
}

SA="$SCRIPT_DIR/shift-assignments"
SS="$SCRIPT_DIR/site-slots"
AV="$SCRIPT_DIR/availability"

# シフト割当: 各スクリプトの実行タイミングを明示指定
deploy_to_site "$SHIFT_ASSIGNMENTS_SITE_ID" "シフト割当" '[
  {"file":"'"$SA"'/shift-time-default.js",     "title":"デフォルト時間",    "new":true,  "edit":false, "index":false},
  {"file":"'"$SA"'/shift-status-highlight.js",  "title":"ステータス色分け",  "new":false, "edit":false, "index":true},
  {"file":"'"$SA"'/shift-hours-calc.js",        "title":"勤務時間計算",      "new":true,  "edit":true,  "index":false},
  {"file":"'"$SA"'/shift-name-auto.js",         "title":"割当名自動生成",    "new":true,  "edit":true,  "index":false},
  {"file":"'"$SA"'/dashboard-widget.js",        "title":"ダッシュボード",    "new":false, "edit":false, "index":true},
  {"file":"'"$SA"'/dashboard-calendar.js",      "title":"カレンダー色分け",  "new":false, "edit":false, "index":true}
]' "$SA/dashboard-style.css"

# 現場シフト枠
deploy_to_site "$SITE_SLOTS_SITE_ID" "現場シフト枠" '[
  {"file":"'"$SS"'/slot-name-auto.js",       "title":"枠名自動生成",     "new":true,  "edit":true,  "index":false},
  {"file":"'"$SS"'/slot-field-toggle.js",    "title":"フィールド切替",   "new":true,  "edit":true,  "index":false},
  {"file":"'"$SS"'/slot-summary-widget.js",  "title":"枠種別サマリー",   "new":false, "edit":false, "index":true}
]'

# 稼働可能枠
deploy_to_site "$AVAILABILITY_SITE_ID" "稼働可能枠" '[
  {"file":"'"$AV"'/availability-name-auto.js",       "title":"枠名自動生成",     "new":true,  "edit":true,  "index":false},
  {"file":"'"$AV"'/availability-field-toggle.js",    "title":"フィールド切替",   "new":true,  "edit":true,  "index":false},
  {"file":"'"$AV"'/availability-summary-widget.js",  "title":"枠種別サマリー",   "new":false, "edit":false, "index":true}
]'

echo ""
echo "=== All scripts deployed successfully ==="
echo ""
echo "確認URL:"
echo "  シフト割当(一覧):     http://localhost/items/$SHIFT_ASSIGNMENTS_SITE_ID"
echo "  シフト割当(カレンダー): http://localhost/items/$SHIFT_ASSIGNMENTS_SITE_ID?View=Calendar"
echo "  現場シフト枠:          http://localhost/items/$SITE_SLOTS_SITE_ID"
echo "  稼働可能枠:            http://localhost/items/$AVAILABILITY_SITE_ID"
