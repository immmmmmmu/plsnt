// シフト割当: 新規作成時にシフト日を今日の日付でデフォルト設定
$p.events.on_editor_load = function() {
  var field = $('#Issues_DateA');
  if (field.length > 0 && !field.val()) {
    var today = new Date();
    var formatted = today.getFullYear() + '/' +
      ('0' + (today.getMonth() + 1)).slice(-2) + '/' +
      ('0' + today.getDate()).slice(-2);
    $p.set(field, formatted);
  }
};
