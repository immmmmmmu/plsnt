// シフト割当一覧: ステータスに応じた行ハイライト
// 100=予定(黄), 200=確定(青), 300=欠勤(赤), 900=完了(緑)
$('.grid-row').each(function() {
  var statusCell = $(this).find('td[data-name="Status"]');
  var statusText = statusCell.text().trim();
  if (statusText === '予定' || statusText === '新規') {
    $(this).css('background-color', '#fff9c4');
  } else if (statusText === '確定' || statusText === '実施中') {
    $(this).css('background-color', '#e3f2fd');
  } else if (statusText === '欠勤') {
    $(this).css('background-color', '#ffebee');
  } else if (statusText === '完了') {
    $(this).css('background-color', '#e8f5e9');
  }
});
