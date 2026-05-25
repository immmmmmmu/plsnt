// #10 書籍 - 定価バリデーション
// 対象: 書籍マスタテーブル（Results）
// 実行タイミング: New, Edit
$p.events.before_validate = function () {
  var price = parseFloat($('#Results_NumA').val());

  if (isNaN(price) || price <= 0) {
    var proceed = confirm('定価が0以下です（現在の値: ' + (price || 0) + '円）。\nこのまま保存しますか？');
    if (!proceed) {
      $('#Results_NumA').closest('.field-normal').addClass('error');
      return false;
    }
  }

  $('#Results_NumA').closest('.field-normal').removeClass('error');
  return true;
};
