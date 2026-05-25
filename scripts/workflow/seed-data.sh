#!/bin/bash
# seed-data.sh: ワークフローアプリのサンプルデータ投入
# Usage: ./scripts/workflow/seed-data.sh <dept_site_id> <position_site_id> <app_type_site_id> <vendor_site_id> <rule_site_id>

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLSNT="${PLSNT:-plsnt}"

DEPT_ID=$1
POS_ID=$2
TYPE_ID=$3
VENDOR_ID=$4
RULE_ID=$5

if [ -z "$RULE_ID" ]; then
    echo "Usage: $0 <dept_site_id> <position_site_id> <app_type_site_id> <vendor_site_id> <rule_site_id>"
    exit 1
fi

echo "=== 部署マスタ投入 ==="
$PLSNT workflow master --site-id $DEPT_ID --file "$SCRIPT_DIR/seed-data/departments.csv" --key ClassA

echo "=== 役職マスタ投入 ==="
$PLSNT workflow master --site-id $POS_ID --file "$SCRIPT_DIR/seed-data/positions.csv" --key ClassA

echo "=== 申請種別マスタ投入 ==="
$PLSNT workflow master --site-id $TYPE_ID --file "$SCRIPT_DIR/seed-data/app-types.csv" --key ClassA

echo "=== 取引先マスタ投入 ==="
$PLSNT workflow master --site-id $VENDOR_ID --file "$SCRIPT_DIR/seed-data/vendors.csv" --key ClassA

echo "=== 承認ルールマスタ投入 ==="
$PLSNT workflow master --site-id $RULE_ID --file "$SCRIPT_DIR/seed-data/approval-rules.csv" --key Title

echo "=== 完了 ==="
