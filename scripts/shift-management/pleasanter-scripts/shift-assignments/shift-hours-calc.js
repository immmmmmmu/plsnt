// シフト割当: 開始時刻・終了時刻から勤務時間(NumA)を自動計算
$(document).on('change', '#Issues_DateB, #Issues_DateC', function() {
  var startVal = $('#Issues_DateB').val();
  var endVal = $('#Issues_DateC').val();
  if (!startVal || !endVal) return;
  var start = new Date(startVal);
  var end = new Date(endVal);
  if (isNaN(start.getTime()) || isNaN(end.getTime())) return;
  var diffMs = end.getTime() - start.getTime();
  if (diffMs < 0) diffMs += 24 * 60 * 60 * 1000;
  var hours = Math.round(diffMs / (60 * 60 * 1000) * 10) / 10;
  $p.set($('#Issues_NumA'), hours);
});
