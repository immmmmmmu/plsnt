// シフト割当: 一覧画面でステータスに応じて行を色分け
// 100=予定(黄), 200=確定(青), 900=完了(緑), 910=欠勤(赤)
$('.grid-row').each(function() {
  var status = $(this).find('td[data-name="Status"]').text().trim();
  if (status === '910') {
    $(this).css('background-color', '#ffcdd2');
  } else if (status === '200') {
    $(this).css('background-color', '#bbdefb');
  } else if (status === '100') {
    $(this).css('background-color', '#fff9c4');
  } else if (status === '900') {
    $(this).css('background-color', '#c8e6c9');
  }
});
