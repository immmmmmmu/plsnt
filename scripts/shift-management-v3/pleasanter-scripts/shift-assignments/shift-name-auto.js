// シフト割当: 警備員名+現場名+時間帯から割当名(DescriptionA)を自動生成
// カレンダーセルの表示を最適化（短い名前）
function generateShiftName() {
  var guardId = ($('#Issues_ClassB').val() || '').replace(/[\[\]]/g, '');
  var siteId = ($('#Issues_ClassA').val() || '').replace(/[\[\]]/g, '');
  var type = $('#Issues_ClassD').val() || '';
  var startTime = $('#Issues_StartTime').val() || '';

  var timeSlot = '';
  if (startTime) {
    var h = new Date(startTime).getHours();
    if (h >= 6 && h < 14) timeSlot = '日勤';
    else if (h >= 14 && h < 22) timeSlot = '遅番';
    else timeSlot = '夜勤';
  }

  var parts = [];

  function buildName(guardName, siteName) {
    if (guardName) parts.push(guardName.split(/\s/)[0]);
    if (siteName) {
      var short = siteName.replace(/オフィスビル|商業施設|建設現場|イベント会場/g, '').trim();
      parts.push(short || siteName);
    }
    if (timeSlot) parts.push(timeSlot);
    if (type === '代替') parts.push('(代替)');
    if (type === 'スポット') parts.push('(スポット)');
    $p.set($('#Issues_DescriptionA'), parts.join(' '));
  }

  if (guardId && /^\d+$/.test(guardId)) {
    $p.apiGet({
      id: parseInt(guardId),
      done: function(gData) {
        if (!gData.Response || !gData.Response.Data || gData.Response.Data.length === 0) return;
        var guardName = gData.Response.Data[0].ClassHash.ClassA || '';
        if (siteId && /^\d+$/.test(siteId)) {
          $p.apiGet({
            id: parseInt(siteId),
            done: function(sData) {
              if (!sData.Response || !sData.Response.Data || sData.Response.Data.length === 0) return;
              var siteName = sData.Response.Data[0].ClassHash.ClassA || '';
              buildName(guardName, siteName);
            }
          });
        } else {
          buildName(guardName, '');
        }
      }
    });
  } else {
    buildName('', '');
  }
}
$(document).on('change', '#Issues_ClassA, #Issues_ClassB, #Issues_ClassD, #Issues_StartTime', generateShiftName);
