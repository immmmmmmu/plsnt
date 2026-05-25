// #12 返却 - 延滞日数ハイライト
// 対象: 返却テーブル（Results）
// 実行タイミング: Index
(function () {
  $('.grid-row').each(function () {
    var numACell = $(this).find('td[data-name="NumA"]');
    if (numACell.length > 0) {
      var days = parseFloat(numACell.text().replace(/,/g, '')) || 0;
      if (days > 0) {
        numACell.css({ 'color': '#d32f2f', 'font-weight': 'bold' });
        // 返却状態セルも赤くする
        var statusCell = $(this).find('td[data-name="ClassB"]');
        if (statusCell.length > 0) {
          statusCell.css({ 'color': '#d32f2f', 'font-weight': 'bold' });
        }
      }
    }
  });
})();
