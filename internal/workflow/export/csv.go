// Package export は申請ヘッダ+明細レコードからCSVを生成する。
// Pleasanter API に依存しない純粋Go実装。
package export

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/immmmmmmu/plsnt/internal/pleasanter"
)

// csvHeaders はCSV出力の列ヘッダ。
// cli-spec.md の CSV列定義に準拠。
var csvHeaders = []string{
	"日付", "申請番号", "申請者", "部署", "申請種別", "用途", "金額", "支払区分", "ステータス",
}

// StatusName はステータスコードを日本語名に変換する。
// interfaces.go の Status定数に対応。
func StatusName(status int) string {
	switch status {
	case 100:
		return "下書き"
	case 200:
		return "申請中"
	case 300:
		return "承認中"
	case 350:
		return "役員承認待ち"
	case 400:
		return "承認完了"
	case 500:
		return "差戻"
	case 600:
		return "却下"
	case 900:
		return "精算済"
	default:
		return fmt.Sprintf("不明(%d)", status)
	}
}

// recordID はレコードのIDを返す。IssueId優先、0ならResultIdを使用。
func recordID(r pleasanter.Record) string {
	if r.IssueId != 0 {
		return fmt.Sprintf("%d", r.IssueId)
	}
	return fmt.Sprintf("%d", r.ResultId)
}

// ResolveOptions はCSV出力時のリンクフィールド名前解決オプション。
// 各マップは レコードID(文字列) → 表示名 の対応。
// nil の場合は名前解決せずID値をそのまま出力する。
type ResolveOptions struct {
	Departments map[string]string // 部署マスタ: ID → 部署名
	AppTypes    map[string]string // 申請種別マスタ: ID → 種別名
}

// GenerateCSV は申請ヘッダ+明細レコードから申請明細一覧CSVを生成する。
// ヘッダと明細はClassA（申請ヘッダへのリンク）で紐付ける。
// 明細のClassAに対応するヘッダが見つからない場合、その明細行はスキップする。
// opts が nil の場合、リンクフィールドは解決せずにID値をそのまま出力する。
func GenerateCSV(headers []pleasanter.Record, details []pleasanter.Record, w io.Writer, opts *ResolveOptions) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// ヘッダ行出力
	if err := writer.Write(csvHeaders); err != nil {
		return err
	}

	// ヘッダレコードをIDでインデックス化
	headerMap := make(map[string]pleasanter.Record, len(headers))
	for _, h := range headers {
		headerMap[recordID(h)] = h
	}

	// 明細ごとにCSV行出力
	for _, d := range details {
		headerID := d.ClassHash["ClassA"]
		h, ok := headerMap[headerID]
		if !ok {
			continue // リンク先ヘッダなし: スキップ
		}

		row := []string{
			d.DateHash["DateA"],           // 日付
			h.Title,                       // 申請番号
			resolveCreator(h),             // 申請者
			resolveDepartment(h, opts),    // 部署
			resolveAppType(h, opts),       // 申請種別
			d.ClassHash["ClassD"],         // 用途
			d.NumHash["NumA"].String(),    // 金額
			h.ClassHash["ClassB"],         // 支払区分
			StatusName(h.Status),          // ステータス
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// resolveCreator はヘッダレコードの Creator ID を文字列で返す。
// Creator が 0 の場合は空文字を返す。
func resolveCreator(h pleasanter.Record) string {
	if h.Creator == 0 {
		return ""
	}
	return fmt.Sprintf("%d", h.Creator)
}

// resolveDepartment はヘッダのClassC（部署リンクID）を部署名に解決する。
// opts が nil または Departments が nil の場合はClassCの値をそのまま返す。
func resolveDepartment(h pleasanter.Record, opts *ResolveOptions) string {
	raw := h.ClassHash["ClassC"]
	if opts == nil || opts.Departments == nil {
		return raw
	}
	if name, ok := opts.Departments[raw]; ok {
		return name
	}
	return raw
}

// resolveAppType はヘッダのClassA（申請種別リンクID）を種別名に解決する。
// opts が nil または AppTypes が nil の場合はClassAの値をそのまま返す。
func resolveAppType(h pleasanter.Record, opts *ResolveOptions) string {
	raw := h.ClassHash["ClassA"]
	if opts == nil || opts.AppTypes == nil {
		return raw
	}
	if name, ok := opts.AppTypes[raw]; ok {
		return name
	}
	return raw
}
