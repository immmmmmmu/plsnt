// シフト割当: 一覧画面上部にダッシュボードウィジェットを表示
// ステータス別集計、本日のシフト数、欠勤アラートをリアルタイム表示
(function() {
  if ($('#MainCommandContainer').length === 0) return;
  if ($('#shift-dashboard').length > 0) return;

  var counts = { total: 0, planned: 0, confirmed: 0, completed: 0, absent: 0 };
  var todayShifts = 0;
  var today = new Date();
  var todayStr = today.getFullYear() + '/' +
    ('0' + (today.getMonth() + 1)).slice(-2) + '/' +
    ('0' + today.getDate()).slice(-2);

  $('.grid-row').each(function() {
    counts.total++;
    var statusText = $(this).find('td[data-name="Status"]').text().trim();
    if (statusText === '100' || statusText === '予定') counts.planned++;
    else if (statusText === '200' || statusText === '確定') counts.confirmed++;
    else if (statusText === '900' || statusText === '完了') counts.completed++;
    else if (statusText === '910' || statusText === '欠勤') counts.absent++;

    var startTime = $(this).find('td[data-name="StartTime"]').text().trim();
    if (startTime && startTime.indexOf(todayStr) === 0) todayShifts++;
  });

  var html = '<div id="shift-dashboard" style="' +
    'display:flex;gap:12px;flex-wrap:wrap;padding:16px;margin:0 0 16px 0;' +
    'background:#f8f9fa;border-radius:8px;border:1px solid #dee2e6;">';

  html += buildCard('本日のシフト', todayShifts + '件', '#1976d2', todayShifts === 0 ? '(データなし)' : todayStr);
  html += buildCard('予定', counts.planned + '件', '#f9a825', '未確定シフト');
  html += buildCard('確定', counts.confirmed + '件', '#1565c0', '確定済みシフト');
  html += buildCard('完了', counts.completed + '件', '#2e7d32', '勤務終了');

  if (counts.absent > 0) {
    html += buildCard('欠勤', counts.absent + '件', '#c62828', '要対応');
  }

  html += '</div>';

  var $grid = $('#Grid');
  if ($grid.length > 0) {
    $grid.before(html);
  } else {
    $('#MainContainer').prepend(html);
  }

  function buildCard(label, value, color, sub) {
    return '<div style="' +
      'flex:1;min-width:140px;padding:12px 16px;' +
      'background:#fff;border-radius:6px;border-left:4px solid ' + color + ';' +
      'box-shadow:0 1px 3px rgba(0,0,0,0.1);">' +
      '<div style="font-size:11px;color:#666;margin-bottom:4px;">' + label + '</div>' +
      '<div style="font-size:24px;font-weight:bold;color:' + color + ';">' + value + '</div>' +
      '<div style="font-size:10px;color:#999;margin-top:2px;">' + sub + '</div>' +
      '</div>';
  }
})();
