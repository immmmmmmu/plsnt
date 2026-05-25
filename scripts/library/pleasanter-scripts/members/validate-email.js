// #8 利用者 - メールアドレス形式チェック
// 対象: 利用者マスタテーブル（Results）
// 実行タイミング: New, Edit
$p.events.before_validate = function () {
  var email = $('#Results_ClassC').val();
  if (!email) return true;

  var emailPattern = /^[a-zA-Z0-9.!#$%&'*+\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

  if (!emailPattern.test(email)) {
    alert('メールアドレスの形式が正しくありません: ' + email);
    $('#Results_ClassC').closest('.field-normal').addClass('error');
    return false;
  }

  $('#Results_ClassC').closest('.field-normal').removeClass('error');
  return true;
};
