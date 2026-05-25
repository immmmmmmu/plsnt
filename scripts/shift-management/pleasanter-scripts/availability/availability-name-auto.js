// 稼働可能枠: 枠種別・警備員・曜日・時間帯から枠名(DescriptionA)を自動生成
$(document).on('change', '#Results_ClassA, #Results_ClassB, #Results_ClassC, #Results_ClassD, #Results_DateE', function() {
  var guardId = $('#Results_ClassA').val();
  var slotType = $('#Results_ClassB').val();
  var period = $('#Results_ClassC').val();
  var dayPattern = $('#Results_ClassD').val();
  var specificDate = $('#Results_DateE').val();
  if (!guardId) return;
  var id = guardId.replace(/[\[\]]/g, '');
  if (!id || !/^\d+$/.test(id)) return;
  $p.apiGet({
    id: parseInt(id),
    done: function(data) {
      var guardName = data.Response.Data[0].ClassHash.ClassA || '';
      var parts = [guardName];
      if (slotType) parts.push(slotType);
      if (dayPattern && dayPattern !== '指定なし') parts.push(dayPattern);
      if (period) parts.push(period);
      if (specificDate) parts.push(specificDate.split('T')[0]);
      $p.set($('#Results_DescriptionA'), parts.join(' '));
    }
  });
});
