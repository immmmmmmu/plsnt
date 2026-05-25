// シフト管理ダッシュボード: Pleasanter APIからデータを取得して表示
// DashboardsサイトのScript (Index=true) として実行
(function() {
  if ($('#shift-dashboard-main').length > 0) return;

  var SITES = {
    assignments: 32450,
    siteSlots: 32448,
    availability: 32449,
    sites: 32446,
    guards: 32447,
    qualifications: 32445
  };

  // API共通ヘッダー
  var apiHeaders = { 'Content-Type': 'application/json' };

  // MainContainerを置き換え
  var $container = $('<div id="shift-dashboard-main" style="padding:24px;max-width:1200px;margin:0 auto;"></div>');
  var $main = $('#MainContainer');
  if ($main.length === 0) $main = $('body');
  $main.prepend($container);

  // ヘッダー
  $container.append(
    '<div style="margin-bottom:24px;">' +
    '<h1 style="font-size:24px;font-weight:bold;color:#1a1a1a;margin:0 0 4px 0;">警備シフト管理</h1>' +
    '<p style="color:#888;font-size:13px;margin:0;">ダッシュボード</p>' +
    '</div>'
  );

  // KPIカードエリア
  var $kpi = $('<div id="kpi-cards" style="display:flex;gap:12px;flex-wrap:wrap;margin-bottom:24px;"></div>');
  $kpi.append('<div style="color:#999;font-size:13px;padding:12px;">読み込み中...</div>');
  $container.append($kpi);

  // リンクカードエリア
  var $links = $('<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(250px,1fr));gap:12px;margin-bottom:28px;"></div>');
  $container.append($links);

  var linkData = [
    { id: SITES.assignments, label: 'シフト割当', color: '#1976d2', desc: 'シフトの割当・確定・欠勤管理' },
    { id: SITES.assignments, label: 'カレンダー表示', color: '#0d47a1', desc: 'シフトをカレンダーで確認', suffix: '?View=Calendar' },
    { id: SITES.siteSlots, label: '現場シフト枠', color: '#6a1b9a', desc: '需要定義（定常/スポット/期間/除外）' },
    { id: SITES.availability, label: '稼働可能枠', color: '#2e7d32', desc: '供給定義（定常/追加/除外）' },
    { id: SITES.sites, label: '現場マスタ', color: '#e65100', desc: '警備対象の現場情報' },
    { id: SITES.guards, label: '警備員マスタ', color: '#c62828', desc: '警備員の基本情報・資格' },
    { id: SITES.qualifications, label: '資格マスタ', color: '#455a64', desc: '資格の管理' }
  ];

  linkData.forEach(function(l) {
    var href = '/items/' + l.id + (l.suffix || '');
    $links.append(
      '<a href="' + href + '" style="display:block;padding:14px 16px;background:#fff;border-radius:6px;' +
      'border-left:4px solid ' + l.color + ';box-shadow:0 1px 3px rgba(0,0,0,0.06);' +
      'text-decoration:none;color:#333;transition:box-shadow 0.2s,transform 0.2s;"' +
      ' onmouseover="this.style.boxShadow=\'0 3px 10px rgba(0,0,0,0.12)\';this.style.transform=\'translateY(-1px)\'"' +
      ' onmouseout="this.style.boxShadow=\'0 1px 3px rgba(0,0,0,0.06)\';this.style.transform=\'none\'">' +
      '<div style="font-size:14px;font-weight:bold;color:' + l.color + ';">' + l.label + '</div>' +
      '<div style="font-size:12px;color:#888;margin-top:4px;">' + l.desc + '</div>' +
      '</a>'
    );
  });

  // 最新シフトエリア
  $container.append('<h2 style="font-size:16px;font-weight:bold;color:#333;margin:0 0 12px 0;">直近のシフト割当</h2>');
  var $table = $('<div id="recent-shifts" style="background:#fff;border-radius:8px;border:1px solid #e0e0e0;overflow:hidden;"></div>');
  $table.html('<div style="padding:24px;text-align:center;color:#999;font-size:13px;">読み込み中...</div>');
  $container.append($table);

  // Pleasanter REST APIでレコード一覧を取得
  function apiGetRecords(siteId, requestBody, callback) {
    $.ajax({
      url: '/api/items/' + siteId + '/get',
      type: 'POST',
      contentType: 'application/json',
      data: JSON.stringify($.extend({
        ApiVersion: 1.1,
        ApiKey: $('#Token').val()
      }, requestBody)),
      success: function(resp) {
        if (resp && resp.Response && resp.Response.Data) {
          callback(resp.Response.Data, resp.Response.TotalCount);
        } else {
          callback([], 0);
        }
      },
      error: function() { callback([], 0); }
    });
  }

  // KPI取得
  apiGetRecords(SITES.assignments, { View: {} }, function(records, totalCount) {
    var planned = 0, confirmed = 0, completed = 0, absent = 0;

    records.forEach(function(r) {
      var s = r.Status;
      if (s === 100) planned++;
      else if (s === 200) confirmed++;
      else if (s === 900) completed++;
      else if (s === 910) absent++;
    });

    $kpi.empty();
    $kpi.append(buildKpiCard('シフト総数', totalCount + '件', '#333', '全レコード'));
    $kpi.append(buildKpiCard('予定', planned + '件', '#f9a825', '未確定'));
    $kpi.append(buildKpiCard('確定', confirmed + '件', '#1565c0', '確定済み'));
    $kpi.append(buildKpiCard('完了', completed + '件', '#2e7d32', '勤務終了'));
    if (absent > 0) {
      $kpi.append(buildKpiCard('欠勤', absent + '件', '#c62828', '要対応'));
    }
  });

  // 最新シフト取得
  apiGetRecords(SITES.assignments, {
    View: { ColumnSorterHash: { StartTime: 'desc' } },
    Offset: 0,
    PageSize: 10
  }, function(records) {
    if (records.length === 0) {
      $table.html('<div style="padding:24px;text-align:center;color:#999;">データがありません</div>');
      return;
    }

    var html = '<table style="width:100%;border-collapse:collapse;font-size:13px;">';
    html += '<thead><tr style="background:#fafafa;border-bottom:2px solid #e0e0e0;">';
    html += '<th style="padding:10px 14px;text-align:left;color:#666;font-weight:600;">割当名</th>';
    html += '<th style="padding:10px 14px;text-align:left;color:#666;font-weight:600;">開始</th>';
    html += '<th style="padding:10px 14px;text-align:left;color:#666;font-weight:600;">終了</th>';
    html += '<th style="padding:10px 14px;text-align:left;color:#666;font-weight:600;">種別</th>';
    html += '<th style="padding:10px 14px;text-align:center;color:#666;font-weight:600;">ステータス</th>';
    html += '</tr></thead><tbody>';

    records.forEach(function(r, i) {
      var name = (r.DescriptionHash && r.DescriptionHash.DescriptionA) || '(名称なし)';
      var start = formatDateTime(r.StartTime);
      var end = formatDateTime(r.CompletionTime);
      var type = (r.ClassHash && r.ClassHash.ClassD) || '';
      var statusInfo = getStatusInfo(r.Status);
      var recordUrl = '/items/' + (r.IssueId || r.ResultId);
      var bgColor = i % 2 === 0 ? '#fff' : '#fafafa';

      html += '<tr style="border-bottom:1px solid #f0f0f0;background:' + bgColor + ';cursor:pointer;" ' +
        'onclick="location.href=\'' + recordUrl + '\'">';
      html += '<td style="padding:10px 14px;font-weight:500;">' + esc(name) + '</td>';
      html += '<td style="padding:10px 14px;color:#666;">' + start + '</td>';
      html += '<td style="padding:10px 14px;color:#666;">' + end + '</td>';
      html += '<td style="padding:10px 14px;">' + esc(type) + '</td>';
      html += '<td style="padding:10px 14px;text-align:center;">' +
        '<span style="display:inline-block;padding:3px 10px;border-radius:12px;font-size:11px;font-weight:600;' +
        'background:' + statusInfo.bg + ';color:' + statusInfo.color + ';">' + statusInfo.label + '</span></td>';
      html += '</tr>';
    });

    html += '</tbody></table>';
    $table.html(html);
  });

  function buildKpiCard(label, value, color, sub) {
    return '<div style="flex:1;min-width:130px;padding:14px 16px;background:#fff;border-radius:8px;' +
      'border-left:4px solid ' + color + ';box-shadow:0 1px 3px rgba(0,0,0,0.06);">' +
      '<div style="font-size:11px;color:#999;margin-bottom:4px;">' + label + '</div>' +
      '<div style="font-size:26px;font-weight:bold;color:' + color + ';line-height:1.2;">' + value + '</div>' +
      '<div style="font-size:10px;color:#bbb;margin-top:2px;">' + sub + '</div>' +
      '</div>';
  }

  function getStatusInfo(status) {
    switch (status) {
      case 100: return { label: '予定', bg: '#fff9c4', color: '#f57f17' };
      case 200: return { label: '確定', bg: '#bbdefb', color: '#0d47a1' };
      case 900: return { label: '完了', bg: '#c8e6c9', color: '#1b5e20' };
      case 910: return { label: '欠勤', bg: '#ffcdd2', color: '#b71c1c' };
      default:  return { label: String(status), bg: '#f5f5f5', color: '#666' };
    }
  }

  function formatDateTime(dt) {
    if (!dt) return '-';
    var d = new Date(dt);
    if (isNaN(d.getTime())) return '-';
    return (d.getMonth() + 1) + '/' + d.getDate() + ' ' +
      ('0' + d.getHours()).slice(-2) + ':' + ('0' + d.getMinutes()).slice(-2);
  }

  function esc(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }
})();
