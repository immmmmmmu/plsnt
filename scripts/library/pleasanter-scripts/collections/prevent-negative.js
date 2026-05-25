// #7 蔵書 - マイナス蔵書防止バリデーション
// 対象: 蔵書テーブル（Results）
// 実行タイミング: New, Edit
$p.events.before_validate = function () {
  var qty = parseFloat($('#Results_NumA').val());

  if (isNaN(qty)) {
    alert('所蔵数を入力してください');
    $('#Results_NumA').closest('.field-normal').addClass('error');
    return false;
  }

  if (qty < 0) {
    alert('所蔵数は0以上の値を入力してください（現在の値: ' + qty + '）');
    $('#Results_NumA').closest('.field-normal').addClass('error');
    return false;
  }

  $('#Results_NumA').closest('.field-normal').removeClass('error');
  return true;
};
