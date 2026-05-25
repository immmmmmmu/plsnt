package encoding

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetectAndConvert_UTF8(t *testing.T) {
	input := []byte("hello, world")
	got, enc, err := DetectAndConvert(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "UTF-8" {
		t.Errorf("encoding = %q, want %q", enc, "UTF-8")
	}
	if !bytes.Equal(got, input) {
		t.Errorf("data changed unexpectedly")
	}
}

func TestDetectAndConvert_UTF8WithJapanese(t *testing.T) {
	input := []byte("名前,年齢\n太郎,30\n")
	got, enc, err := DetectAndConvert(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "UTF-8" {
		t.Errorf("encoding = %q, want %q", enc, "UTF-8")
	}
	if !bytes.Equal(got, input) {
		t.Errorf("data changed unexpectedly")
	}
}

func TestDetectAndConvert_UTF8BOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := []byte("hello, world")
	input := append(bom, content...)

	got, enc, err := DetectAndConvert(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "UTF-8 (BOM)" {
		t.Errorf("encoding = %q, want %q", enc, "UTF-8 (BOM)")
	}
	if !bytes.Equal(got, content) {
		t.Errorf("BOM was not stripped: got %v, want %v", got, content)
	}
}

func TestDetectAndConvert_ShiftJIS(t *testing.T) {
	// "テスト" in Shift-JIS
	sjisData := []byte{0x83, 0x65, 0x83, 0x58, 0x83, 0x67}

	got, enc, err := DetectAndConvert(sjisData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "Shift-JIS" {
		t.Errorf("encoding = %q, want %q", enc, "Shift-JIS")
	}
	if string(got) != "テスト" {
		t.Errorf("decoded = %q, want %q", string(got), "テスト")
	}
}

func TestDetectAndConvert_ShiftJISCSV(t *testing.T) {
	// CSV with Shift-JIS encoding: "名前,年齢\nテスト,30\n"
	// "名前" in Shift-JIS: 0x96, 0xBC, 0x91, 0x4F
	// "年齢" in Shift-JIS: 0x94, 0x4E, 0x97, 0xEE
	// "テスト" in Shift-JIS: 0x83, 0x65, 0x83, 0x58, 0x83, 0x67
	sjisCSV := []byte{
		0x96, 0xBC, 0x91, 0x4F, // 名前
		0x2C,                   // ,
		0x94, 0x4E, 0x97, 0xEE, // 年齢
		0x0A,                               // \n
		0x83, 0x65, 0x83, 0x58, 0x83, 0x67, // テスト
		0x2C,       // ,
		0x33, 0x30, // 30
		0x0A, // \n
	}

	got, enc, err := DetectAndConvert(sjisCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "Shift-JIS" {
		t.Errorf("encoding = %q, want %q", enc, "Shift-JIS")
	}

	expected := "名前,年齢\nテスト,30\n"
	if string(got) != expected {
		t.Errorf("decoded = %q, want %q", string(got), expected)
	}
}

func TestNewReader_ShiftJIS(t *testing.T) {
	// "テスト" in Shift-JIS
	sjisData := []byte{0x83, 0x65, 0x83, 0x58, 0x83, 0x67}

	reader, enc, err := NewReader(bytes.NewReader(sjisData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "Shift-JIS" {
		t.Errorf("encoding = %q, want %q", enc, "Shift-JIS")
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if buf.String() != "テスト" {
		t.Errorf("decoded = %q, want %q", buf.String(), "テスト")
	}
}

func TestNewReader_UTF8Passthrough(t *testing.T) {
	input := "hello, world"
	reader, enc, err := NewReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "UTF-8" {
		t.Errorf("encoding = %q, want %q", enc, "UTF-8")
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if buf.String() != input {
		t.Errorf("data = %q, want %q", buf.String(), input)
	}
}

func TestIsValidUTF8(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{"ASCII", []byte("hello"), true},
		{"Japanese UTF-8", []byte("テスト"), true},
		{"Empty", []byte{}, true},
		{"Invalid bytes", []byte{0x80, 0x81}, false},
		{"Shift-JIS bytes", []byte{0x83, 0x65, 0x83, 0x58}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidUTF8(tt.input)
			if got != tt.want {
				t.Errorf("isValidUTF8(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
