// #4 蔵書 - 低蔵書アラート表示
// 対象: 蔵書テーブル（Results）
// 実行タイミング: Index, Edit
(function () {
  var LOW_THRESHOLD = 1;

  // 一覧画面
  if ($('.grid-row').length > 0) {
    var lowCount = 0;
    $('.grid-row').each(function () {
      var numACell = $(this).find('td[data-name="NumA"]');
      if (numACell.length > 0) {
        var qty = parseFloat(numACell.text().replace(/,/g, '')) || 0;
        if (qty <= LOW_THRESHOLD) {
          $(this).css({
            'background-color': '#ffebee',
            'border-left': '4px solid #f44336'
          });
          numACell.css({ 'color': '#d32f2f', 'font-weight': 'bold' });
          lowCount++;
        }
      }
    });

    if (lowCount > 0) {
      var alertHtml = '<div style="background:#fff3e0;border:1px solid #ff9800;border-radius:4px;padding:8px 16px;margin:8px 0;color:#e65100;">'
        + '<strong>蔵書アラート:</strong> ' + lowCount + '件の蔵書が' + LOW_THRESHOLD + '冊以下です'
        + '</div>';
      $('.grid').before(alertHtml);
    }
  }

  // 編集画面
  if ($('#Results_NumA').length > 0) {
    var qty = parseFloat($('#Results_NumA').val()) || 0;
    if (qty <= LOW_THRESHOLD) {
      $('#Results_NumA').closest('.field-normal')
        .append('<div class="low-stock-warning" style="color:#d32f2f;font-size:12px;margin-top:4px;">蔵書が少なくなっています（閾値: ' + LOW_THRESHOLD + '冊）</div>');
    }

    $(document).on('change', '#Results_NumA', function () {
      $('.low-stock-warning').remove();
      var newQty = parseFloat($(this).val()) || 0;
      if (newQty <= LOW_THRESHOLD) {
        $(this).closest('.field-normal')
          .append('<div class="low-stock-warning" style="color:#d32f2f;font-size:12px;margin-top:4px;">蔵書が少なくなっています（閾値: ' + LOW_THRESHOLD + '冊）</div>');
      }
    });
  }
})();
