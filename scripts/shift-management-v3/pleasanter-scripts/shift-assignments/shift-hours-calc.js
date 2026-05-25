// シフト割当: StartTime/CompletionTimeから勤務時間(NumA)を自動計算
// 日跨ぎ（夜勤）にも対応
$(document).on('change', '#Issues_StartTime, #Issues_CompletionTime', function() {
  var start = new Date($('#Issues_StartTime').val());
  var end = new Date($('#Issues_CompletionTime').val());
  if (isNaN(start.getTime()) || isNaN(end.getTime())) return;
  var diff = (end - start) / (1000 * 60 * 60);
  if (diff <= 0) {
    $p.set($('#Issues_NumA'), 0);
    return;
  }
  $p.set($('#Issues_NumA'), Math.round(diff * 10) / 10);
});
