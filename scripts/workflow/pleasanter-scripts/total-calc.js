// #2 申請ヘッダ - 明細金額合計の自動計算
// 対象: 申請ヘッダテーブル（Issues）
// 実行タイミング: Edit
// 動作: 申請明細（Results, SiteID: 32552）の NumA（金額）を合計し、
//   申請ヘッダの NumB（合計金額）に自動反映する
//   - 画面ロード時に自動計算
//   - 「明細合計を再計算」ボタンで手動再計算も可能
//
// Pleasanter登録先: SiteSettings.Scripts[]

$(document).ready(function () {
  var DETAIL_SITE_ID = 32552;
  var recordId = $p.id();
  if (!recordId || recordId === 0) return;

  // 合計計算の実行関数
  function calcTotal(callback) {
    $.ajax({
      url: '/api/items/' + DETAIL_SITE_ID + '/get',
      type: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({
        ApiKey: $p.apiKey(),
        View: {
          ColumnFilterHash: {
            ClassA: '[' + recordId + ']'
          }
        }
      }),
      success: function (data) {
        var total = 0;
        if (data.Response && data.Response.Data) {
          data.Response.Data.forEach(function (record) {
            var amount =
              record.NumHash && record.NumHash.NumA;
            if (amount !== undefined && amount !== null) {
              total += parseFloat(String(amount));
            }
          });
        }
        // 小数点誤差を回避するため整数に丸める
        total = Math.round(total);
        $p.set($('#Issues_NumB'), total);
        // amount-helper.js のカンマフォーマットをトリガー
        $('#Issues_NumB').trigger('blur');
        if (typeof callback === 'function') callback(total);
      },
      error: function () {
        console.error('明細レコードの取得に失敗しました');
      }
    });
  }

  // 画面ロード時に自動計算
  calcTotal();

  // 再計算ボタンを追加
  var $calcBtn = $(
    '<button type="button" class="btn btn-default" ' +
      'style="margin-left:10px;font-size:12px;">' +
      '明細合計を再計算</button>'
  );
  $('#Issues_NumB')
    .closest('.field-normal')
    .find('.field-label')
    .append($calcBtn);

  $calcBtn.on('click', function () {
    calcTotal(function (total) {
      $calcBtn.text('再計算完了 (' + total.toLocaleString('ja-JP') + ')');
      setTimeout(function () {
        $calcBtn.text('明細合計を再計算');
      }, 2000);
    });
  });
});
