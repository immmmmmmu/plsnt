// #3 貸出明細 - 書籍選択時に明細名(DescriptionA)へ書名を自動セット
// 対象: 貸出明細テーブル（Results）
// 実行タイミング: New, Edit
$(document).on('change', '#Results_ClassB', function () {
  var bookId = $(this).val();
  if (!bookId) return;

  bookId = bookId.replace(/[\[\]]/g, '');
  if (!bookId || !/^\d+$/.test(bookId)) return;

  $p.apiGet({
    id: parseInt(bookId),
    done: function (data) {
      if (data && data.Response && data.Response.Data && data.Response.Data.length > 0) {
        var book = data.Response.Data[0];
        var bookName = book.ClassHash && book.ClassHash.ClassA ? book.ClassHash.ClassA : '';
        if (bookName) {
          $p.set($('#Results_DescriptionA'), bookName);
        }
      }
    }
  });
});
