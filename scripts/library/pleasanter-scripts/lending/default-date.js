// #1 貸出 - 貸出日デフォルト（今日の日付）
// 対象: 貸出テーブル（Issues）
// 実行タイミング: New
$p.events.on_editor_load = function () {
  var dateField = $('#Issues_DateA');
  if (dateField.length > 0 && !dateField.val()) {
    var today = new Date();
    var formatted = today.getFullYear() + '/' +
      ('0' + (today.getMonth() + 1)).slice(-2) + '/' +
      ('0' + today.getDate()).slice(-2);
    $p.set(dateField, formatted);
  }
};
