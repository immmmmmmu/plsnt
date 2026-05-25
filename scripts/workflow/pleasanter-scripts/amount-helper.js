// #1 申請ヘッダ - 金額入力補助
// 対象: 申請ヘッダテーブル（Issues）
// 実行タイミング: New, Edit
// 動作: NumB（合計金額）, NumC（概算金額）にカンマ区切りフォーマットを適用
//   - フォーカスアウト時: 数値をカンマ区切り表示（例: 100000 → 100,000）
//   - フォーカス時: カンマを除去して編集しやすく
//
// Pleasanter登録先: SiteSettings.Scripts[]

$(document).ready(function () {
  var amountFields = ['NumB', 'NumC'];

  amountFields.forEach(function (field) {
    var selector = '#Issues_' + field;

    // フォーカスアウト時: カンマ区切りフォーマット
    $(selector).on('blur', function () {
      var raw = $(this).val().replace(/,/g, '');
      if (raw === '' || isNaN(raw)) {
        return;
      }
      var num = Number(raw);
      $(this).val(num.toLocaleString('ja-JP'));
    });

    // フォーカス時: カンマ除去して編集モード
    $(selector).on('focus', function () {
      var raw = $(this).val().replace(/,/g, '');
      $(this).val(raw);
    });

    // 初期表示時にもフォーマット適用
    var el = $(selector);
    if (el.length > 0 && el.val()) {
      var raw = el.val().replace(/,/g, '');
      if (raw !== '' && !isNaN(raw)) {
        el.val(Number(raw).toLocaleString('ja-JP'));
      }
    }
  });
});

// 保存前にカンマを除去して数値として保存
$p.events.before_validate = function () {
  var amountFields = ['NumB', 'NumC'];
  amountFields.forEach(function (field) {
    var selector = '#Issues_' + field;
    var el = $(selector);
    if (el.length > 0) {
      var raw = el.val().replace(/,/g, '');
      el.val(raw);
    }
  });
  return true;
};
