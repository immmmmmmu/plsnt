#!/usr/bin/env bash
# setup-links.sh - ワークフローアプリのテーブル間リンク設定
#
# 使い方:
#   ./scripts/workflow/setup-links.sh
#
# SiteSettings 全体上書き問題を回避するため、
# site get → python3 マージ → site update の安全パターンで実装。
# (relational-modeling / plsnt-guide / app-build-workflow スキル準拠)
#
# リンクフィールド3点セット:
#   1. Links: テーブル間リレーション定義（参照元テーブルの SiteSettings に設定）
#   2. ChoicesText: "[[SiteId]]" を参照元テーブルの Columns に設定（UIドロップダウン表示）
#   3. TitleColumns: 参照先テーブルに設定（リンク表示に参照先の ItemTitle を使用）
#
# 設定するリンク:
#   1. 申請ヘッダ(32524).ClassA → 申請種別マスタ(32521)
#   2. 申請ヘッダ(32524).ClassC → 部署マスタ(32519)
#   3. 申請ヘッダ(32524).ClassD → 取引先マスタ(32523)
#   4. 申請明細(32525).ClassA → 申請ヘッダ(32524)
#   5. 承認履歴(32526).ClassA → 申請ヘッダ(32524)
#   6. 承認ルールマスタ(32522).ClassB → 申請種別マスタ(32521)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLSNT="$PROJECT_ROOT/plsnt"

# --- SiteID 定義 ---
SITE_DEPT=32519          # 部署マスタ
SITE_POSITION=32520      # 役職マスタ
SITE_APP_TYPE=32521      # 申請種別マスタ
SITE_APPROVAL_RULE=32522 # 承認ルールマスタ
SITE_VENDOR=32523        # 取引先マスタ
SITE_APP_HEADER=32524    # 申請ヘッダ
SITE_APP_DETAIL=32525    # 申請明細
SITE_APPROVAL_HIST=32526 # 承認履歴

echo "=== ワークフローアプリ: テーブル間リンク設定 ==="
echo ""

# --- ヘルパー関数 ---

# site get → python3 で SiteSettings をマージ → site update
# relational-modeling スキル: site update は SiteSettings 全体を上書きするため、
# 必ず site get で現在の設定を取得してからマージする
merge_site_settings() {
  local site_id="$1"
  local site_name="$2"
  local python_script="$3"

  echo "--- $site_name ($site_id) ---"

  # 1. 現在のSiteSettings取得
  local current
  current=$("$PLSNT" site get "$site_id" -o json 2>/dev/null)
  if [ -z "$current" ]; then
    echo "  ERROR: site get $site_id が空を返しました" >&2
    return 1
  fi

  # 2. python3 で SiteSettings をマージ
  local new_ss
  new_ss=$(echo "$current" | python3 -c "$python_script")
  if [ -z "$new_ss" ]; then
    echo "  ERROR: python3 マージが空を返しました" >&2
    return 1
  fi

  # 3. site update
  "$PLSNT" site update "$site_id" --json "$new_ss" > /dev/null 2>&1
  echo "  OK"
}

# ==========================================================
# Phase 0: EditorColumnHash 修正 + TitleColumns 設定
# ==========================================================
# Issues型テーブルでも EditorColumnHash のキーは "General" が正しい
# "Issues" キーだと ClassHash/NumHash 等のカスタムフィールドが保存されない
echo "--- Phase 0: EditorColumnHash 修正 + TitleColumns 設定 ---"
echo ""

# 全テーブルに TitleColumns を設定し、EditorColumnHash の Issues→General 修正も行う
# relational-modeling スキル: 参照先テーブルに TitleColumns が必要（リンク表示に必須）
for site_info in \
  "$SITE_DEPT:部署マスタ" \
  "$SITE_POSITION:役職マスタ" \
  "$SITE_APP_TYPE:申請種別マスタ" \
  "$SITE_APPROVAL_RULE:承認ルールマスタ" \
  "$SITE_VENDOR:取引先マスタ" \
  "$SITE_APP_HEADER:申請ヘッダ" \
  "$SITE_APP_DETAIL:申請明細" \
  "$SITE_APPROVAL_HIST:承認履歴"; do

  site_id="${site_info%%:*}"
  site_name="${site_info#*:}"

  merge_site_settings "$site_id" "$site_name (TitleColumns + EditorColumnHash修正)" "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

# TitleColumns 設定（table-schema.md: 全テーブル TitleColumns: ['Title']）
ss['TitleColumns'] = ['Title']

# EditorColumnHash の Issues→General 修正
ech = ss.get('EditorColumnHash', {})
if 'Issues' in ech:
    ech['General'] = ech.pop('Issues')
    ss['EditorColumnHash'] = ech

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
"
done

echo ""

# ==========================================================
# Phase 1: リンク設定（Links + ChoicesText）
# ==========================================================
# relational-modeling スキル: リンクフィールド3点セット
#   - Links: テーブル間リレーション定義
#   - ChoicesText: "[[SiteId]]" でUIドロップダウン表示
#   - TitleColumns: Phase 0 で設定済み
echo "--- Phase 1: リンク設定 (Links + ChoicesText) ---"
echo ""

# --- 1. 申請ヘッダ (32524) のリンク設定 ---
# ClassA → 申請種別マスタ(32521), ClassC → 部署マスタ(32519), ClassD → 取引先マスタ(32523)
merge_site_settings "$SITE_APP_HEADER" "申請ヘッダ リンク設定" "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

# Links 追加（scaffold-shift-management-v3.yaml Phase 4 パターン準拠）
ss['Links'] = [
    {'ColumnName': 'ClassA', 'SiteId': $SITE_APP_TYPE, 'LabelText': '申請種別'},
    {'ColumnName': 'ClassC', 'SiteId': $SITE_DEPT, 'LabelText': '申請部署'},
    {'ColumnName': 'ClassD', 'SiteId': $SITE_VENDOR, 'LabelText': '取引先'}
]

# ChoicesText 設定（Columns 内の該当カラムに追加）
choices_map = {
    'ClassA': '[[${SITE_APP_TYPE}]]',
    'ClassC': '[[${SITE_DEPT}]]',
    'ClassD': '[[${SITE_VENDOR}]]'
}
for col in ss.get('Columns', []):
    cn = col.get('ColumnName', '')
    if cn in choices_map:
        col['ChoicesText'] = choices_map[cn]

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
"

# --- 2. 申請明細 (32525) のリンク設定 ---
# ClassA → 申請ヘッダ(32524)
merge_site_settings "$SITE_APP_DETAIL" "申請明細 リンク設定" "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

ss['Links'] = [
    {'ColumnName': 'ClassA', 'SiteId': $SITE_APP_HEADER, 'LabelText': '申請ヘッダ'}
]

for col in ss.get('Columns', []):
    if col.get('ColumnName') == 'ClassA':
        col['ChoicesText'] = '[[${SITE_APP_HEADER}]]'
        break

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
"

# --- 3. 承認履歴 (32526) のリンク設定 ---
# ClassA → 申請ヘッダ(32524)
merge_site_settings "$SITE_APPROVAL_HIST" "承認履歴 リンク設定" "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

ss['Links'] = [
    {'ColumnName': 'ClassA', 'SiteId': $SITE_APP_HEADER, 'LabelText': '申請ヘッダ'}
]

for col in ss.get('Columns', []):
    if col.get('ColumnName') == 'ClassA':
        col['ChoicesText'] = '[[${SITE_APP_HEADER}]]'
        break

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
"

# --- 4. 承認ルールマスタ (32522) のリンク設定 ---
# ClassB → 申請種別マスタ(32521)
merge_site_settings "$SITE_APPROVAL_RULE" "承認ルールマスタ リンク設定" "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

ss['Links'] = [
    {'ColumnName': 'ClassB', 'SiteId': $SITE_APP_TYPE, 'LabelText': '対象申請種別'}
]

for col in ss.get('Columns', []):
    if col.get('ColumnName') == 'ClassB':
        col['ChoicesText'] = '[[${SITE_APP_TYPE}]]'
        break

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
"

echo ""
echo "=== 全リンク設定完了 ==="
echo ""
echo "設定済みリンク一覧:"
echo "  [Phase 0] TitleColumns: 全8テーブルに [\"Title\"] を設定"
echo "  [Phase 0] EditorColumnHash: Issues→General 修正（該当テーブルのみ）"
echo "  [Phase 1] 申請ヘッダ($SITE_APP_HEADER).ClassA → 申請種別マスタ($SITE_APP_TYPE)"
echo "  [Phase 1] 申請ヘッダ($SITE_APP_HEADER).ClassC → 部署マスタ($SITE_DEPT)"
echo "  [Phase 1] 申請ヘッダ($SITE_APP_HEADER).ClassD → 取引先マスタ($SITE_VENDOR)"
echo "  [Phase 1] 申請明細($SITE_APP_DETAIL).ClassA → 申請ヘッダ($SITE_APP_HEADER)"
echo "  [Phase 1] 承認履歴($SITE_APPROVAL_HIST).ClassA → 申請ヘッダ($SITE_APP_HEADER)"
echo "  [Phase 1] 承認ルールマスタ($SITE_APPROVAL_RULE).ClassB → 申請種別マスタ($SITE_APP_TYPE)"
