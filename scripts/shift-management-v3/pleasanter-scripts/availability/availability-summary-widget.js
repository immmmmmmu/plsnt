// 稼働可能枠: 一覧画面上部に枠種別サマリーを表示
(function() {
  if ($('#MainCommandContainer').length === 0) return;
  if ($('#avail-summary').length > 0) return;

  var counts = { total: 0 };
  var types = {};

  $('.grid-row').each(function() {
    counts.total++;
    var type = $(this).find('td[data-name="ClassB"]').text().trim();
    if (type) types[type] = (types[type] || 0) + 1;
  });

  var typeColors = {
    '定常': '#1565c0',
    '追加': '#2e7d32',
    '除外': '#c62828'
  };

  var html = '<div id="avail-summary" style="' +
    'display:flex;gap:10px;flex-wrap:wrap;padding:12px 16px;margin:0 0 12px 0;' +
    'background:#f8f9fa;border-radius:8px;border:1px solid #dee2e6;">' +
    '<div style="display:flex;align-items:center;gap:4px;font-size:13px;font-weight:bold;color:#333;">' +
    '稼働可能枠: ' + counts.total + '件</div>';

  Object.keys(types).forEach(function(type) {
    var color = typeColors[type] || '#666';
    html += '<div style="display:flex;align-items:center;gap:4px;' +
      'padding:4px 10px;background:#fff;border-radius:12px;' +
      'border:1px solid ' + color + ';font-size:12px;color:' + color + ';">' +
      '<span style="width:8px;height:8px;border-radius:50%;background:' + color + ';display:inline-block;"></span>' +
      type + ': ' + types[type] + '</div>';
  });

  html += '</div>';

  var $grid = $('#Grid');
  if ($grid.length > 0) {
    $grid.before(html);
  } else {
    $('#MainContainer').prepend(html);
  }
})();
