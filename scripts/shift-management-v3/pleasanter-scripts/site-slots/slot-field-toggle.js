// 現場シフト枠: 枠種別に応じてフィールドの表示/非表示を切替
// 定常: 曜日パターン必須, 特定日非表示
// スポット: 特定日必須, 曜日パターン非表示
// 期間: 曜日パターン+適用開始/終了日必須
// 除外: 特定日 or 適用期間
function toggleSlotFields() {
  var slotType = $('#Results_ClassB').val();
  var dayRow = $('#Results_ClassD').closest('.field-normal');
  var dateC = $('#Results_DateC').closest('.field-normal');
  var dateD = $('#Results_DateD').closest('.field-normal');
  var dateE = $('#Results_DateE').closest('.field-normal');
  dayRow.show(); dateC.show(); dateD.show(); dateE.show();
  if (slotType === '定常') {
    dateE.hide();
  } else if (slotType === 'スポット') {
    dayRow.hide(); dateC.hide(); dateD.hide();
  } else if (slotType === '期間') {
    dateE.hide();
  }
}
$(document).on('change', '#Results_ClassB', toggleSlotFields);
$p.events.on_editor_load = toggleSlotFields;
