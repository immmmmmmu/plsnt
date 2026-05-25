// 現場シフト枠: 枠種別・現場・曜日・時間帯から枠名(DescriptionA)を自動生成
$(document).on('change', '#Results_ClassA, #Results_ClassB, #Results_ClassC, #Results_ClassD, #Results_DateE', function() {
  var siteId = $('#Results_ClassA').val();
  var slotType = $('#Results_ClassB').val();
  var period = $('#Results_ClassC').val();
  var dayPattern = $('#Results_ClassD').val();
  var specificDate = $('#Results_DateE').val();
  if (!siteId) return;
  var id = siteId.replace(/[\[\]]/g, '');
  if (!id || !/^\d+$/.test(id)) return;
  $p.apiGet({
    id: parseInt(id),
    done: function(data) {
      var siteName = data.Response.Data[0].ClassHash.ClassA || '';
      var parts = [siteName];
      if (slotType) parts.push(slotType);
      if (dayPattern && dayPattern !== '指定なし') parts.push(dayPattern);
      if (period) parts.push(period);
      if (specificDate) parts.push(specificDate.split('T')[0]);
      $p.set($('#Results_DescriptionA'), parts.join(' '));
    }
  });
});
