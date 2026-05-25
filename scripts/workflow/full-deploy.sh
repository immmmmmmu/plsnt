#!/usr/bin/env bash
# full-deploy.sh: ワークフローアプリの完全デプロイ
#
# Usage: ./scripts/workflow/full-deploy.sh <folder_id>
#
# 実行内容:
#   1. テーブル作成 + リンク設定（batch run: full-deploy.yaml）
#   2. プロセス設定（site get → マージ → site update）
#   3. StatusControls設定（site get → マージ → site update）
#   4. サンプルデータ投入

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLSNT="$PROJECT_ROOT/plsnt"

# PATH に plsnt のディレクトリを追加（バッチエンジンのサブプロセス実行用）
export PATH="$PROJECT_ROOT:$PATH"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <folder_id>" >&2
  echo "  folder_id: デプロイ先フォルダの SiteID" >&2
  exit 1
fi

FOLDER_ID="$1"

echo "=== ワークフローアプリ フルデプロイ ==="
echo "デプロイ先フォルダ: $FOLDER_ID"
echo ""

# ==========================================================
# Phase 1: テーブル作成 + リンク設定（batch run）
# ==========================================================
echo "--- Phase 1: テーブル作成 + リンク設定 ---"
BATCH_OUTPUT=$("$PLSNT" batch run "$PROJECT_ROOT/templates/workflow/full-deploy.yaml" \
  --var "folder_id=$FOLDER_ID" \
  --log-file /tmp/workflow-full-deploy.log 2>&1)
echo "$BATCH_OUTPUT"

# ログファイルからステップ出力を解析して SiteID を取得
# batch run の出力から各テーブルの SiteID を抽出
extract_site_id() {
  local step_name="$1"
  local log_file="/tmp/workflow-full-deploy.log"
  if [ -f "$log_file" ]; then
    python3 -c "
import json, sys
with open('$log_file') as f:
    for line in f:
        try:
            entry = json.loads(line.strip())
            if entry.get('step') == '$step_name' and entry.get('event') == 'step_complete':
                out = entry.get('output', {})
                print(out.get('Id', ''))
                break
        except (json.JSONDecodeError, KeyError):
            continue
"
  fi
}

# SiteID の取得を試みる（ログファイルから）
SITE_DEPT=$(extract_site_id "create-dept")
SITE_POSITION=$(extract_site_id "create-position")
SITE_APP_TYPE=$(extract_site_id "create-app-type")
SITE_APPROVAL_RULE=$(extract_site_id "create-approval-rule")
SITE_VENDOR=$(extract_site_id "create-vendor")
SITE_HEADER=$(extract_site_id "create-header")
SITE_DETAIL=$(extract_site_id "create-detail")
SITE_HISTORY=$(extract_site_id "create-history")

# ログから取得できなかった場合、batch出力から推測する代替手段
if [ -z "$SITE_HEADER" ]; then
  echo ""
  echo "WARNING: ログファイルから SiteID を自動取得できませんでした。"
  echo "batch run 出力を確認してください。"
  echo ""
  echo "申請ヘッダの SiteID を入力してください:"
  read -r SITE_HEADER
  echo "部署マスタの SiteID を入力してください:"
  read -r SITE_DEPT
  echo "申請種別マスタの SiteID を入力してください:"
  read -r SITE_APP_TYPE
  echo "取引先マスタの SiteID を入力してください:"
  read -r SITE_VENDOR
fi

echo ""
echo "--- 作成テーブル一覧 ---"
echo "  部署マスタ:       $SITE_DEPT"
echo "  役職マスタ:       $SITE_POSITION"
echo "  申請種別マスタ:   $SITE_APP_TYPE"
echo "  承認ルールマスタ: $SITE_APPROVAL_RULE"
echo "  取引先マスタ:     $SITE_VENDOR"
echo "  申請ヘッダ:       $SITE_HEADER"
echo "  申請明細:         $SITE_DETAIL"
echo "  承認履歴:         $SITE_HISTORY"
echo ""

# ==========================================================
# Phase 2: プロセス設定 + StatusControls（site get → マージ → site update）
# ==========================================================
echo "--- Phase 2: プロセス設定 + StatusControls ---"

if [ -z "$SITE_HEADER" ]; then
  echo "ERROR: 申請ヘッダの SiteID が不明です。" >&2
  exit 1
fi

# 現在の SiteSettings を取得
CURRENT=$("$PLSNT" site get "$SITE_HEADER" -o json 2>/dev/null)
if [ -z "$CURRENT" ]; then
  echo "ERROR: site get $SITE_HEADER が空を返しました" >&2
  exit 1
fi

# Processes + StatusControls をマージ
MERGED=$(echo "$CURRENT" | python3 -c "
import json, sys
d = json.load(sys.stdin)
ss = d.get('SiteSettings', {})

ss['Processes'] = [
  {'Name':'申請','DisplayName':'申請する','CurrentStatus':100,'ChangedStatus':200,'OnClick':True},
  {'Name':'承認（単段）','DisplayName':'承認する','CurrentStatus':200,'ChangedStatus':900,'ConfirmationMessage':'承認してよろしいですか？','OnClick':True,'View':{'ColumnFilterHash':{'NumB':'[\"0\",\"200000\"]'}}},
  {'Name':'承認→部長へ','DisplayName':'承認（部長へ送付）','CurrentStatus':200,'ChangedStatus':300,'ConfirmationMessage':'課長承認として部長へ送付します','OnClick':True,'View':{'ColumnFilterHash':{'NumB':'[\"200000\",\"\"]'}}},
  {'Name':'部長承認','DisplayName':'承認する','CurrentStatus':300,'ChangedStatus':900,'ConfirmationMessage':'承認してよろしいですか？','OnClick':True,'View':{'ColumnFilterHash':{'NumB':'[\"200000\",\"1000000\"]'}}},
  {'Name':'部長承認→役員へ','DisplayName':'承認（役員へ送付）','CurrentStatus':300,'ChangedStatus':350,'ConfirmationMessage':'部長承認として役員へ送付します','OnClick':True,'View':{'ColumnFilterHash':{'NumB':'[\"1000000\",\"\"]'}}},
  {'Name':'役員承認','DisplayName':'承認する','CurrentStatus':350,'ChangedStatus':900,'ConfirmationMessage':'承認してよろしいですか？','OnClick':True},
  {'Name':'差戻200','DisplayName':'差し戻す','CurrentStatus':200,'ChangedStatus':500,'ConfirmationMessage':'差し戻してよろしいですか？','OnClick':True},
  {'Name':'差戻300','DisplayName':'差し戻す','CurrentStatus':300,'ChangedStatus':500,'ConfirmationMessage':'差し戻してよろしいですか？','OnClick':True},
  {'Name':'差戻350','DisplayName':'差し戻す','CurrentStatus':350,'ChangedStatus':500,'ConfirmationMessage':'差し戻してよろしいですか？','OnClick':True},
  {'Name':'却下200','DisplayName':'却下する','CurrentStatus':200,'ChangedStatus':600,'ConfirmationMessage':'却下してよろしいですか？','OnClick':True},
  {'Name':'却下300','DisplayName':'却下する','CurrentStatus':300,'ChangedStatus':600,'ConfirmationMessage':'却下してよろしいですか？','OnClick':True},
  {'Name':'却下350','DisplayName':'却下する','CurrentStatus':350,'ChangedStatus':600,'ConfirmationMessage':'却下してよろしいですか？','OnClick':True},
  {'Name':'再申請','DisplayName':'再申請する','CurrentStatus':500,'ChangedStatus':200,'OnClick':True}
]

ss['StatusControls'] = [
  {'Status':200,'Description':'申請中：主要フィールド読取専用','ColumnHash':{'ClassA':'ReadOnly','ClassB':'ReadOnly','ClassC':'ReadOnly','NumB':'ReadOnly'}},
  {'Status':300,'Description':'承認中：全カスタムフィールド読取専用','ColumnHash':{'ClassA':'ReadOnly','ClassB':'ReadOnly','ClassC':'ReadOnly','ClassD':'ReadOnly','ClassE':'ReadOnly','ClassF':'ReadOnly','NumA':'ReadOnly','NumB':'ReadOnly','NumC':'ReadOnly','DateA':'ReadOnly'}},
  {'Status':350,'Description':'役員承認待ち：全カスタムフィールド読取専用','ColumnHash':{'ClassA':'ReadOnly','ClassB':'ReadOnly','ClassC':'ReadOnly','ClassD':'ReadOnly','ClassE':'ReadOnly','ClassF':'ReadOnly','NumA':'ReadOnly','NumB':'ReadOnly','NumC':'ReadOnly','DateA':'ReadOnly'}},
  {'Status':400,'Description':'承認完了：全体読取専用','ReadOnly':True},
  {'Status':600,'Description':'却下：全体読取専用','ReadOnly':True},
  {'Status':900,'Description':'精算済：全体読取専用','ReadOnly':True}
]

print(json.dumps({'SiteSettings': ss}, ensure_ascii=False))
")

"$PLSNT" site update "$SITE_HEADER" --json "$MERGED" > /dev/null 2>&1
echo "  プロセス設定: 13件投入 OK"
echo "  StatusControls: 6件投入 OK"
echo ""

# ==========================================================
# Phase 3: サンプルデータ投入
# ==========================================================
echo "--- Phase 3: サンプルデータ投入 ---"

if [ -n "$SITE_DEPT" ]; then
  "$PLSNT" record create --site-id "$SITE_DEPT" --json '{"Title":"営業部","ClassHash":{"ClassA":"D001"}}' > /dev/null 2>&1
  "$PLSNT" record create --site-id "$SITE_DEPT" --json '{"Title":"総務部","ClassHash":{"ClassA":"D002"}}' > /dev/null 2>&1
  "$PLSNT" record create --site-id "$SITE_DEPT" --json '{"Title":"経理部","ClassHash":{"ClassA":"D003"}}' > /dev/null 2>&1
  echo "  部署マスタ: 3件 OK"
fi

if [ -n "$SITE_APP_TYPE" ]; then
  "$PLSNT" record create --site-id "$SITE_APP_TYPE" --json '{"Title":"立替・支出依頼","ClassHash":{"ClassA":"EXP"}}' > /dev/null 2>&1
  "$PLSNT" record create --site-id "$SITE_APP_TYPE" --json '{"Title":"出張申請","ClassHash":{"ClassA":"TRV"}}' > /dev/null 2>&1
  echo "  申請種別マスタ: 2件 OK"
fi

if [ -n "$SITE_VENDOR" ]; then
  "$PLSNT" record create --site-id "$SITE_VENDOR" --json '{"Title":"テスト商事","ClassHash":{"ClassA":"V001"}}' > /dev/null 2>&1
  echo "  取引先マスタ: 1件 OK"
fi

# ==========================================================
# Phase 4: スクリプト/スタイルデプロイ
# ==========================================================
echo "--- Phase 4: スクリプト/スタイルデプロイ ---"

if [ -n "$SITE_HEADER" ] && [ -n "$SITE_DETAIL" ]; then
  "$SCRIPT_DIR/pleasanter-scripts/deploy-all.sh" "$SITE_HEADER" "$SITE_DETAIL"
  echo "  スクリプト/スタイル: OK"
else
  echo "  WARNING: SITE_HEADER または SITE_DETAIL が不明のためスキップ" >&2
fi

echo ""
echo "=== フルデプロイ完了 ==="
echo ""
echo "申請ヘッダ SiteID: $SITE_HEADER"
echo "申請明細 SiteID:   $SITE_DETAIL"
echo "PleasanterのUI: http://localhost/items/$SITE_HEADER"
