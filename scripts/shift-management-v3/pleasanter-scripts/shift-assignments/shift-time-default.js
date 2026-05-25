// シフト割当: 新規作成時に今日の日付+デフォルト時間をStartTime/CompletionTimeにセット
$p.events.on_editor_load = function() {
  var start = $('#Issues_StartTime');
  if (start.length > 0 && !start.val()) {
    var today = new Date();
    var y = today.getFullYear();
    var m = ('0' + (today.getMonth() + 1)).slice(-2);
    var d = ('0' + today.getDate()).slice(-2);
    $p.set(start, y + '/' + m + '/' + d + ' 09:00');
    $p.set($('#Issues_CompletionTime'), y + '/' + m + '/' + d + ' 18:00');
  }
};
