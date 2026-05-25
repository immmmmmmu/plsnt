// 稼働可能枠: 枠種別に応じてフィールドの表示/非表示を切替
// 定常: 曜日パターン必須, 特定日非表示
// 追加: 特定日必須, 曜日パターン非表示
// 除外: 特定日 or 適用期間
function toggleAvailFields() {
  var slotType = $('#Results_ClassB').val();
  var dayRow = $('#Results_ClassD').closest('.field-normal');
  var dateC = $('#Results_DateC').closest('.field-normal');
  var dateD = $('#Results_DateD').closest('.field-normal');
  var dateE = $('#Results_DateE').closest('.field-normal');
  dayRow.show(); dateC.show(); dateD.show(); dateE.show();
  if (slotType === '定常') {
    dateE.hide();
  } else if (slotType === '追加') {
    dayRow.hide(); dateC.hide(); dateD.hide();
  }
}
$(document).on('change', '#Results_ClassB', toggleAvailFields);
$p.events.on_editor_load = toggleAvailFields;
