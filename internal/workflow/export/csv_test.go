package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCSV_Normal(t *testing.T) {
	headers := []pleasanter.Record{
		{
			ResultId: 0,
			IssueId:  1001,
			Title:    "2026-0001 交通費",
			Status:   400,
			ClassHash: map[string]string{
				"ClassA": "5001",
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
			NumHash: map[string]json.Number{
				"NumB": json.Number("15000"),
			},
			Creator: 1,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			Title:    "電車代",
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("15000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-03",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "日付,申請番号,申請者,部署,申請種別,用途,金額,支払区分,ステータス")
	assert.Contains(t, output, "2026-04-03")
	assert.Contains(t, output, "2026-0001 交通費")
	assert.Contains(t, output, "1,営業部,5001,交通費") // 申請者=1, 部署=営業部, 申請種別=5001(未解決ID)
	assert.Contains(t, output, "15000")
	assert.Contains(t, output, "立替払い")
	assert.Contains(t, output, "承認完了")
}

func TestGenerateCSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateCSV(nil, nil, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	// ヘッダ行のみ
	assert.Contains(t, output, "日付,申請番号")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines))
}

func TestStatusName(t *testing.T) {
	tests := []struct {
		status int
		name   string
	}{
		{100, "下書き"},
		{200, "申請中"},
		{300, "承認中"},
		{350, "役員承認待ち"},
		{400, "承認完了"},
		{500, "差戻"},
		{600, "却下"},
		{900, "精算済"},
		{999, "不明(999)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, StatusName(tt.status))
		})
	}
}

func TestGenerateCSV_MultipleDetails(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001 交通費",
			Status:  200,
			ClassHash: map[string]string{
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			Title:    "電車代",
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("5000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
		{
			ResultId: 2002,
			Title:    "タクシー代",
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("3000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-02",
			},
		},
		{
			ResultId: 2003,
			Title:    "バス代",
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("500"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-03",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// ヘッダ行 + 明細3行 = 4行
	assert.Equal(t, 4, len(lines))
	assert.Contains(t, output, "5000")
	assert.Contains(t, output, "3000")
	assert.Contains(t, output, "500")
	// 全行同じステータス
	assert.Equal(t, 3, strings.Count(output, "申請中"))
}

func TestGenerateCSV_OrphanDetails(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
		},
	}
	// 明細のClassAが存在しないヘッダIDを参照
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "9999", // 孤立明細
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// ヘッダ行のみ（孤立明細はスキップ）
	assert.Equal(t, 1, len(lines))
}

func TestGenerateCSV_CommaInValue(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassB": "立替払い",
				"ClassC": "営業部,第1課", // カンマを含む値
			},
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	// encoding/csv はカンマを含む値をダブルクォートで囲む
	assert.Contains(t, output, "\"営業部,第1課\"")
}

func TestGenerateCSV_ResultIdHeader(t *testing.T) {
	// ResultIdベースのヘッダレコード（IssueId=0の場合）
	headers := []pleasanter.Record{
		{
			ResultId: 3001,
			IssueId:  0,
			Title:    "2026-0010",
			Status:   900,
			ClassHash: map[string]string{
				"ClassB": "会社カード利用",
				"ClassC": "総務部",
			},
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 4001,
			ClassHash: map[string]string{
				"ClassA": "3001",
				"ClassD": "消耗品",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("8500"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-05",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "精算済")
	assert.Contains(t, output, "8500")
	assert.Contains(t, output, "消耗品")
}

func TestGenerateCSV_CreatorOutput(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassB": "立替払い",
				"ClassC": "32630",
			},
			Creator: 42,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	// Creator ID が申請者列に出力される
	assert.Contains(t, output, ",42,")
}

func TestGenerateCSV_CreatorZero(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
			Creator: 0, // Creator未設定
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// データ行の申請者フィールドが空
	assert.Equal(t, 2, len(lines))
	// 申請番号の後に空の申請者カラム
	assert.Contains(t, lines[1], "2026-0001,,営業部")
}

func TestGenerateCSV_WithDepartmentResolve(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassA": "5001",
				"ClassB": "立替払い",
				"ClassC": "32630", // 部署マスタのレコードID
			},
			Creator: 10,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	opts := &ResolveOptions{
		Departments: map[string]string{
			"32630": "営業部",
			"32631": "総務部",
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, opts)
	require.NoError(t, err)

	output := buf.String()
	// 部署IDが名前に解決される
	assert.Contains(t, output, "営業部")
	assert.NotContains(t, output, "32630")
}

func TestGenerateCSV_WithAppTypeResolve(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassA": "32632", // 申請種別マスタのレコードID
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
			Creator: 10,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	opts := &ResolveOptions{
		AppTypes: map[string]string{
			"32632": "立替・支出依頼",
			"32633": "出張申請",
		},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, opts)
	require.NoError(t, err)

	output := buf.String()
	// 申請種別IDが名前に解決される
	assert.Contains(t, output, "立替・支出依頼")
	assert.NotContains(t, output, "32632")
}

func TestGenerateCSV_WithAllResolve(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassA": "32632",
				"ClassB": "立替払い",
				"ClassC": "32630",
			},
			Creator: 42,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("5000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	opts := &ResolveOptions{
		Departments: map[string]string{"32630": "営業部"},
		AppTypes:    map[string]string{"32632": "立替・支出依頼"},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, opts)
	require.NoError(t, err)

	output := buf.String()
	// 全てのリンクフィールドが解決される
	assert.Contains(t, output, "42")            // Creator ID
	assert.Contains(t, output, "営業部")          // 部署名
	assert.Contains(t, output, "立替・支出依頼")    // 申請種別名
	assert.Contains(t, output, "立替払い")         // 支払区分（テキスト値、変更なし）
	assert.NotContains(t, output, "32630")       // 部署IDが出ない
	assert.NotContains(t, output, "32632")       // 種別IDが出ない
}

func TestGenerateCSV_WithBOMPrefix(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001 交通費",
			Status:  400,
			ClassHash: map[string]string{
				"ClassA": "5001",
				"ClassB": "立替払い",
				"ClassC": "営業部",
			},
			NumHash: map[string]json.Number{
				"NumB": json.Number("15000"),
			},
			Creator: 1,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("15000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-03",
			},
		},
	}

	var buf bytes.Buffer
	// BOM を先に書き込んでから GenerateCSV を呼ぶ（cmd層と同じパターン）
	bom := []byte{0xEF, 0xBB, 0xBF}
	_, err := buf.Write(bom)
	require.NoError(t, err)

	err = GenerateCSV(headers, details, &buf, nil)
	require.NoError(t, err)

	raw := buf.Bytes()
	// 先頭3バイトが UTF-8 BOM
	require.True(t, len(raw) >= 3, "output should be at least 3 bytes")
	assert.Equal(t, byte(0xEF), raw[0], "BOM byte 1")
	assert.Equal(t, byte(0xBB), raw[1], "BOM byte 2")
	assert.Equal(t, byte(0xBF), raw[2], "BOM byte 3")

	// BOM の後に CSV ヘッダが続く
	output := string(raw[3:])
	assert.True(t, strings.HasPrefix(output, "日付,"), "CSV header should follow BOM")
	assert.Contains(t, output, "承認完了")
}

func TestGenerateCSV_WithBOMPrefix_Empty(t *testing.T) {
	var buf bytes.Buffer
	// BOM + 空データ（ヘッダ行のみ）
	bom := []byte{0xEF, 0xBB, 0xBF}
	_, err := buf.Write(bom)
	require.NoError(t, err)

	err = GenerateCSV(nil, nil, &buf, nil)
	require.NoError(t, err)

	raw := buf.Bytes()
	require.True(t, len(raw) >= 3)
	assert.Equal(t, byte(0xEF), raw[0])
	assert.Equal(t, byte(0xBB), raw[1])
	assert.Equal(t, byte(0xBF), raw[2])

	// BOM後はCSVヘッダ行のみ
	output := string(raw[3:])
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines))
	assert.Contains(t, output, "日付,申請番号")
}

func TestGenerateCSV_WithoutBOM(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateCSV(nil, nil, &buf, nil)
	require.NoError(t, err)

	raw := buf.Bytes()
	// BOM なしの場合、先頭は UTF-8 BOM ではない
	if len(raw) >= 3 {
		bomPresent := raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF
		assert.False(t, bomPresent, "BOM should not be present when not explicitly written")
	}
}

func TestGenerateCSV_UnknownIDFallback(t *testing.T) {
	headers := []pleasanter.Record{
		{
			IssueId: 1001,
			Title:   "2026-0001",
			Status:  400,
			ClassHash: map[string]string{
				"ClassA": "99999", // マスタに存在しないID
				"ClassB": "立替払い",
				"ClassC": "88888", // マスタに存在しないID
			},
			Creator: 10,
		},
	}
	details := []pleasanter.Record{
		{
			ResultId: 2001,
			ClassHash: map[string]string{
				"ClassA": "1001",
				"ClassD": "交通費",
			},
			NumHash: map[string]json.Number{
				"NumA": json.Number("1000"),
			},
			DateHash: map[string]string{
				"DateA": "2026-04-01",
			},
		},
	}

	opts := &ResolveOptions{
		Departments: map[string]string{"32630": "営業部"},
		AppTypes:    map[string]string{"32632": "立替・支出依頼"},
	}

	var buf bytes.Buffer
	err := GenerateCSV(headers, details, &buf, opts)
	require.NoError(t, err)

	output := buf.String()
	// マスタにないIDはそのまま出力（フォールバック）
	assert.Contains(t, output, "88888")
	assert.Contains(t, output, "99999")
}
