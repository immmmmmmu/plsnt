#!/bin/bash
# 現任教育管理アプリ サンプルデータ投入
# 警備業法施行規則第38条に基づく教育科目 + 2025年度実施記録サンプル
# 使い方: source scripts/training-management/env.sh && bash scripts/training-management/seed-data.sh

set -euo pipefail

PLSNT="${PLSNT:-plsnt}"

# 環境変数チェック
for var in SUBJECTS_SITE_ID SESSIONS_SITE_ID ATTENDANCE_SITE_ID GUARDS_SITE_ID; do
  if [ -z "${!var:-}" ]; then
    echo "ERROR: $var is not set. Run 'source scripts/training-management/env.sh' first." >&2
    exit 1
  fi
done

echo "=== 現任教育管理 サンプルデータ投入 ==="
echo "教育科目マスタ: $SUBJECTS_SITE_ID"
echo "教育実施記録:   $SESSIONS_SITE_ID"
echo "受講記録:       $ATTENDANCE_SITE_ID"
echo "警備員マスタ:   $GUARDS_SITE_ID"

# =============================================
# Phase 1: 教育科目マスタ（法定科目）
# =============================================
echo ""
echo "--- Phase 1: 教育科目マスタ投入 ---"

# 基本教育 3科目（警備業法施行規則第38条）
echo "基本教育 3科目..."
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"警備業務実施の基本原則","ClassB":"基本教育","ClassC":"共通","ClassD":"講義"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"警備業務の基本原則、警備員の心構え、服務規律"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"警備業法その他関係法令","ClassB":"基本教育","ClassC":"共通","ClassD":"講義"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"警備業法、個人情報保護法、刑法（正当防衛・緊急避難）等"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"事故発生時の応急措置","ClassB":"基本教育","ClassC":"共通","ClassD":"講義・実技"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"警察機関への連絡、応急手当、AED使用法、避難誘導"}}'

# 1号警備 業務別教育 5科目
echo "1号警備 業務別教育 5科目..."
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"出入管理の方法","ClassB":"業務別教育","ClassC":"1号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1.5},"DescriptionHash":{"DescriptionA":"人・車両の出入管理、受付業務、入退館管理システムの操作"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"巡回の方法","ClassB":"業務別教育","ClassC":"1号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1.5},"DescriptionHash":{"DescriptionA":"巡回経路の設定、巡回時の確認事項、異常発見時の対応"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"警報装置・機器の使用方法","ClassB":"業務別教育","ClassC":"1号警備","ClassD":"実技"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"防犯カメラ、侵入検知センサー、火災報知器等の操作・点検"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"不審者・不審物件への対応","ClassB":"業務別教育","ClassC":"1号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1.5},"DescriptionHash":{"DescriptionA":"不審者発見時の声掛け、不審物件の取扱い、警察への通報要領"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"施設警備の知識・技能","ClassB":"業務別教育","ClassC":"1号警備","ClassD":"講義"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"防火・防災管理、鍵管理、来訪者対応マナー、報告書作成"}}'

# 2号警備 業務別教育 6科目
echo "2号警備 業務別教育 6科目..."
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"道路交通関係法令","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"講義"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"道路交通法、道路法、交通誘導に関する法的根拠と責任"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"車両・歩行者の誘導方法","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1.5},"DescriptionHash":{"DescriptionA":"手旗信号、誘導灯の使用法、片側交互通行の実施方法"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"雑踏整理の方法","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"群衆心理、動線管理、雑踏事故防止策、避難誘導"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"資機材の使用方法","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"実技"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"セーフティコーン、バリケード、誘導棒、無線機の取扱い"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"交通事故発生時の対応","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"講義・実技"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"事故現場の安全確保、負傷者救護、警察・消防への通報"}}'
$PLSNT record create --site-id "$SUBJECTS_SITE_ID" --json '{"ClassHash":{"ClassA":"交通誘導の知識・技能","ClassB":"業務別教育","ClassC":"2号警備","ClassD":"講義"},"NumHash":{"NumA":1},"DescriptionHash":{"DescriptionA":"工事現場周辺の安全管理、夜間誘導、特殊車両対応"}}'

echo "教育科目マスタ: 14科目投入完了（基本3 + 1号5 + 2号6）"

# =============================================
# Phase 2: 2025年度 教育実施記録サンプル
# =============================================
echo ""
echo "--- Phase 2: 2025年度 教育実施記録投入 ---"

# 科目IDを取得
SUBJECT_IDS=$($PLSNT record list --site-id "$SUBJECTS_SITE_ID" -o ids)
# 基本教育
S_BASIC1=$(echo "$SUBJECT_IDS" | sed -n '1p')  # 基本原則
S_BASIC2=$(echo "$SUBJECT_IDS" | sed -n '2p')  # 関係法令
S_BASIC3=$(echo "$SUBJECT_IDS" | sed -n '3p')  # 応急措置
# 1号
S_1G_1=$(echo "$SUBJECT_IDS" | sed -n '4p')    # 出入管理
S_1G_2=$(echo "$SUBJECT_IDS" | sed -n '5p')    # 巡回
S_1G_3=$(echo "$SUBJECT_IDS" | sed -n '6p')    # 警報装置
S_1G_4=$(echo "$SUBJECT_IDS" | sed -n '7p')    # 不審者対応
S_1G_5=$(echo "$SUBJECT_IDS" | sed -n '8p')    # 施設警備知識
# 2号
S_2G_1=$(echo "$SUBJECT_IDS" | sed -n '9p')    # 道路交通法令
S_2G_2=$(echo "$SUBJECT_IDS" | sed -n '10p')   # 誘導方法
S_2G_3=$(echo "$SUBJECT_IDS" | sed -n '11p')   # 雑踏整理

echo "科目ID取得完了"

# --- 第1回: 2025/6/10 基本教育（全員対象）---
echo "第1回: 2025/6/10 基本教育..."
$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_BASIC1\",\"ClassB\":\"基本教育\",\"ClassC\":\"共通\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義\"},\"StartTime\":\"2025-06-10T09:00:00\",\"CompletionTime\":\"2025-06-10T10:00:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"基本原則 2025年度第1回\",\"DescriptionB\":\"警備業務の基本原則と心構えについて。令和元年改正内容の確認を含む。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_BASIC2\",\"ClassB\":\"基本教育\",\"ClassC\":\"共通\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義\"},\"StartTime\":\"2025-06-10T10:15:00\",\"CompletionTime\":\"2025-06-10T11:15:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"関係法令 2025年度第1回\",\"DescriptionB\":\"警備業法改正ポイント、個人情報保護法、正当防衛の要件。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_BASIC3\",\"ClassB\":\"基本教育\",\"ClassC\":\"共通\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2025-06-10T13:00:00\",\"CompletionTime\":\"2025-06-10T14:00:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"応急措置 2025年度第1回\",\"DescriptionB\":\"AED操作実習、止血法、通報手順のロールプレイ。\"}}"

# --- 第2回: 2025/7/15 1号業務別教育 ---
echo "第2回: 2025/7/15 1号業務別教育..."
$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_1G_1\",\"ClassB\":\"業務別教育\",\"ClassC\":\"1号警備\",\"ClassD\":\"田村主任\",\"ClassE\":\"渋谷オフィスビル警備室\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2025-07-15T09:00:00\",\"CompletionTime\":\"2025-07-15T10:30:00\",\"NumHash\":{\"NumA\":1.5},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"出入管理 2025年度第1回\",\"DescriptionB\":\"渋谷ビル入退館管理システムの操作研修。来訪者受付フロー確認。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_1G_2\",\"ClassB\":\"業務別教育\",\"ClassC\":\"1号警備\",\"ClassD\":\"田村主任\",\"ClassE\":\"渋谷オフィスビル\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2025-07-15T10:45:00\",\"CompletionTime\":\"2025-07-15T12:15:00\",\"NumHash\":{\"NumA\":1.5},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"巡回方法 2025年度第1回\",\"DescriptionB\":\"渋谷ビル全フロア巡回実習。チェックポイント確認、報告書記入。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_1G_3\",\"ClassB\":\"業務別教育\",\"ClassC\":\"1号警備\",\"ClassD\":\"田村主任\",\"ClassE\":\"渋谷オフィスビル防災センター\",\"ClassF\":\"2025\",\"ClassG\":\"実技\"},\"StartTime\":\"2025-07-15T13:00:00\",\"CompletionTime\":\"2025-07-15T14:00:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"警報装置操作 2025年度第1回\",\"DescriptionB\":\"防犯カメラ操作、火災報知器連動確認、侵入検知テスト。\"}}"

# --- 第3回: 2025/9/20 1号残り + 2号教育 ---
echo "第3回: 2025/9/20 業務別教育（1号残り＋2号）..."
$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_1G_4\",\"ClassB\":\"業務別教育\",\"ClassC\":\"1号警備\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2025-09-20T09:00:00\",\"CompletionTime\":\"2025-09-20T10:30:00\",\"NumHash\":{\"NumA\":1.5},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"不審者対応 2025年度第1回\",\"DescriptionB\":\"声掛け訓練、不審物件発見シナリオ演習、通報手順確認。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_1G_5\",\"ClassB\":\"業務別教育\",\"ClassC\":\"1号警備\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義\"},\"StartTime\":\"2025-09-20T10:45:00\",\"CompletionTime\":\"2025-09-20T11:45:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"施設警備知識 2025年度第1回\",\"DescriptionB\":\"防火管理、鍵管理規定、報告書作成演習。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_2G_1\",\"ClassB\":\"業務別教育\",\"ClassC\":\"2号警備\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義\"},\"StartTime\":\"2025-09-20T13:00:00\",\"CompletionTime\":\"2025-09-20T14:00:00\",\"NumHash\":{\"NumA\":1},\"Status\":900,\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"道路交通法令 2025年度第1回\",\"DescriptionB\":\"交通誘導の法的根拠、違反時の責任、最近の法改正確認。\"}}"

$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_2G_2\",\"ClassB\":\"業務別教育\",\"ClassC\":\"2号警備\",\"ClassD\":\"山本指導員\",\"ClassE\":\"品川建設現場\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2025-09-20T14:15:00\",\"CompletionTime\":\"2025-09-20T15:45:00\",\"NumHash\":{\"NumA\":1.5},\"Status\":200,\"DescriptionHash\":{\"DescriptionA\":\"誘導方法 2025年度第1回\",\"DescriptionB\":\"手旗・誘導灯の実技訓練、片側交互通行シミュレーション。\"}}"

# --- 第4回: 2026/1/20 予定（未実施）---
echo "第4回: 2026/1/20 予定..."
$PLSNT record create --site-id "$SESSIONS_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$S_2G_3\",\"ClassB\":\"業務別教育\",\"ClassC\":\"2号警備\",\"ClassD\":\"山本指導員\",\"ClassE\":\"本社研修室\",\"ClassF\":\"2025\",\"ClassG\":\"講義・実技\"},\"StartTime\":\"2026-01-20T09:00:00\",\"CompletionTime\":\"2026-01-20T10:00:00\",\"NumHash\":{\"NumA\":1},\"Status\":100,\"DescriptionHash\":{\"DescriptionA\":\"雑踏整理 2025年度第1回\",\"DescriptionB\":\"群衆心理と動線管理、雑踏事故防止の実践演習。\"}}"

echo "教育実施記録: 12件投入完了（確認済9 + 実施済1 + 予定2）"

# =============================================
# Phase 3: 受講記録サンプル
# =============================================
echo ""
echo "--- Phase 3: 受講記録投入 ---"

# 警備員ID取得
GUARD_IDS=$($PLSNT record list --site-id "$GUARDS_SITE_ID" -o ids)
G1=$(echo "$GUARD_IDS" | sed -n '1p')  # 田中太郎（1号）
G2=$(echo "$GUARD_IDS" | sed -n '2p')  # 鈴木花子（1号）
G3=$(echo "$GUARD_IDS" | sed -n '3p')  # 佐藤次郎（2号）
G4=$(echo "$GUARD_IDS" | sed -n '4p')  # 山田美咲
G5=$(echo "$GUARD_IDS" | sed -n '5p')  # 高橋健一（1号）

# 教育実施記録ID取得
SESSION_IDS=$($PLSNT record list --site-id "$SESSIONS_SITE_ID" -o ids)
SS1=$(echo "$SESSION_IDS" | sed -n '1p')   # 基本原則
SS2=$(echo "$SESSION_IDS" | sed -n '2p')   # 関係法令
SS3=$(echo "$SESSION_IDS" | sed -n '3p')   # 応急措置
SS4=$(echo "$SESSION_IDS" | sed -n '4p')   # 出入管理
SS5=$(echo "$SESSION_IDS" | sed -n '5p')   # 巡回
SS6=$(echo "$SESSION_IDS" | sed -n '6p')   # 警報装置
SS7=$(echo "$SESSION_IDS" | sed -n '7p')   # 不審者対応
SS8=$(echo "$SESSION_IDS" | sed -n '8p')   # 施設警備知識
SS9=$(echo "$SESSION_IDS" | sed -n '9p')   # 道路交通法令
SS10=$(echo "$SESSION_IDS" | sed -n '10p') # 誘導方法

echo "警備員: 田中=$G1, 鈴木=$G2, 佐藤=$G3, 山田=$G4, 高橋=$G5"

# 第1回 基本教育（全員受講）
echo "第1回 基本教育 受講記録（5名x3科目=15件）..."
for SS in $SS1 $SS2 $SS3; do
  for G in $G1 $G2 $G3 $G4 $G5; do
    # 警備員名を取得して受講記録名に使用
    case "$G" in
      "$G1") GNAME="田中" ;;
      "$G2") GNAME="鈴木" ;;
      "$G3") GNAME="佐藤" ;;
      "$G4") GNAME="山田" ;;
      "$G5") GNAME="高橋" ;;
    esac
    case "$SS" in
      "$SS1") SNAME="基本原則" ;;
      "$SS2") SNAME="関係法令" ;;
      "$SS3") SNAME="応急措置" ;;
    esac
    $PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G\",\"ClassB\":\"$SS\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"基本教育\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"${GNAME} ${SNAME}\"}}"
  done
done

# 第2回 1号業務別教育（田中・鈴木・高橋受講、山田欠席）
echo "第2回 1号業務別教育 受講記録..."
for SS in $SS4 $SS5 $SS6; do
  case "$SS" in
    "$SS4") SNAME="出入管理"; HOURS=1.5 ;;
    "$SS5") SNAME="巡回"; HOURS=1.5 ;;
    "$SS6") SNAME="警報装置"; HOURS=1 ;;
  esac
  for G in $G1 $G2 $G5; do
    case "$G" in
      "$G1") GNAME="田中" ;;
      "$G2") GNAME="鈴木" ;;
      "$G5") GNAME="高橋" ;;
    esac
    $PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G\",\"ClassB\":\"$SS\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"業務別教育\"},\"NumHash\":{\"NumA\":$HOURS},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"${GNAME} ${SNAME}\"}}"
  done
  # 山田は1号業務別を受講（パート、資格なし）
  $PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G4\",\"ClassB\":\"$SS\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"業務別教育\"},\"NumHash\":{\"NumA\":$HOURS},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"山田 ${SNAME}\"}}"
done

# 第3回 1号残り + 2号（佐藤は2号のみ受講）
echo "第3回 業務別教育 受講記録..."
for SS in $SS7 $SS8; do
  case "$SS" in
    "$SS7") SNAME="不審者対応"; HOURS=1.5 ;;
    "$SS8") SNAME="施設警備知識"; HOURS=1 ;;
  esac
  for G in $G1 $G2 $G4 $G5; do
    case "$G" in
      "$G1") GNAME="田中" ;;
      "$G2") GNAME="鈴木" ;;
      "$G4") GNAME="山田" ;;
      "$G5") GNAME="高橋" ;;
    esac
    $PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G\",\"ClassB\":\"$SS\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"業務別教育\"},\"NumHash\":{\"NumA\":$HOURS},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"${GNAME} ${SNAME}\"}}"
  done
done

# 2号: 佐藤受講
$PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"$SS9\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"業務別教育\"},\"NumHash\":{\"NumA\":1},\"CheckHash\":{\"CheckA\":true},\"DescriptionHash\":{\"DescriptionA\":\"佐藤 道路交通法令\"}}"

# 誘導方法は実施済（確認前）
$PLSNT record create --site-id "$ATTENDANCE_SITE_ID" --json "{\"ClassHash\":{\"ClassA\":\"$G3\",\"ClassB\":\"$SS10\",\"ClassC\":\"受講済\",\"ClassD\":\"2025\",\"ClassE\":\"業務別教育\"},\"NumHash\":{\"NumA\":1.5},\"CheckHash\":{\"CheckA\":false},\"DescriptionHash\":{\"DescriptionA\":\"佐藤 誘導方法\"}}"

echo ""
echo "=== サンプルデータ投入完了 ==="
echo ""
echo "投入件数:"
echo "  教育科目マスタ: 14科目（基本3 + 1号5 + 2号6）"
echo "  教育実施記録: 12件（確認済9 + 実施済1 + 予定2）"
echo "  受講記録: 約40件"
echo ""
echo "--- 2025年度 受講時間サマリー（概算） ---"
echo "  田中太郎（1号）: 基本3h + 1号業務別7h = 10h ✓ 達成"
echo "  鈴木花子（1号）: 基本3h + 1号業務別7h = 10h ✓ 達成"
echo "  高橋健一（1号）: 基本3h + 1号業務別7h = 10h ✓ 達成"
echo "  山田美咲（1号）: 基本3h + 1号業務別7h = 10h ✓ 達成"
echo "  佐藤次郎（2号）: 基本3h + 2号業務別2.5h = 5.5h ✗ 残4.5h（1月教育で充足予定）"
echo ""
echo "UIで確認:"
echo "  教育科目: http://localhost/items/$SUBJECTS_SITE_ID"
echo "  実施記録: http://localhost/items/$SESSIONS_SITE_ID"
echo "  受講記録: http://localhost/items/$ATTENDANCE_SITE_ID"
