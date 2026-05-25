// #9 書籍 - ISBN形式チェック
// 対象: 書籍マスタテーブル（Results）
// 実行タイミング: New, Edit
$p.events.before_validate = function () {
  var isbn = $('#Results_ClassC').val();
  if (!isbn) return true;

  // ハイフン除去して数字のみにする
  var digits = isbn.replace(/[-\s]/g, '');

  // ISBN-13 (13桁) または ISBN-10 (10桁)
  if (digits.length !== 13 && digits.length !== 10) {
    alert('ISBNは10桁または13桁で入力してください（現在: ' + digits.length + '桁）\n入力値: ' + isbn);
    $('#Results_ClassC').closest('.field-normal').addClass('error');
    return false;
  }

  // 数字のみかチェック（ISBN-10 は末尾Xあり）
  if (digits.length === 13 && !/^\d{13}$/.test(digits)) {
    alert('ISBN-13は数字13桁で入力してください\n入力値: ' + isbn);
    $('#Results_ClassC').closest('.field-normal').addClass('error');
    return false;
  }

  if (digits.length === 10 && !/^\d{9}[\dX]$/.test(digits)) {
    alert('ISBN-10は数字9桁+チェック桁(数字またはX)で入力してください\n入力値: ' + isbn);
    $('#Results_ClassC').closest('.field-normal').addClass('error');
    return false;
  }

  $('#Results_ClassC').closest('.field-normal').removeClass('error');
  return true;
};
