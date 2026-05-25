// シフト割当: カレンダービューのセルをステータスに応じて色分け
// カレンダービュー（CalendarBody）が表示されている場合のみ実行
(function() {
  var $calendar = $('#CalendarBody');
  if ($calendar.length === 0) return;

  // カレンダー内の各セル（a/div/span等、Pleasanterバージョンにより異なる）
  var cellStyle = {
    'padding-left': '4px',
    'margin-bottom': '2px',
    'border-radius': '3px',
    'font-size': '11px'
  };

  // Pleasanterのカレンダーセルを検索（複数セレクタでバージョン差異を吸収）
  var $items = $calendar.find('.calendar-item, .calendar-title, td[class*="calendar"] a');
  if ($items.length === 0) $items = $calendar.find('td a[href*="/items/"]');

  $items.each(function() {
    var $item = $(this);
    var text = $item.text();

    // テキスト内容からキーワードマッチで色分け（DescriptionAの短縮名に依存）
    var style = $.extend({}, cellStyle);
    if (text.indexOf('欠勤') >= 0) {
      style['background-color'] = '#ffcdd2';
      style['border-left'] = '3px solid #c62828';
    } else if (text.indexOf('代替') >= 0 || text.indexOf('スポット') >= 0) {
      style['background-color'] = '#fff3e0';
      style['border-left'] = '3px solid #e65100';
    } else {
      style['background-color'] = '#e3f2fd';
      style['border-left'] = '3px solid #1976d2';
    }
    $item.css(style);
  });
})();
