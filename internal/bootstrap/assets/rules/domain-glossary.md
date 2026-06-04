# Domain Glossary

| 用語 | 説明 |
|------|------|
| SiteID | プリザンターにおけるテーブル（サイト）の一意識別子 |
| RecordID / IssueId / ResultId | レコードの一意識別子 |
| ClassHash | 分類列（選択肢・ドロップダウン）のカスタムフィールドハッシュ（ClassA〜ClassZ）、値はstring型 |
| NumHash | 数値列のカスタムフィールドハッシュ（NumA〜NumZ）、値はnumber (decimal)型 |
| DateHash | 日付列のカスタムフィールドハッシュ（DateA〜DateZ）、値はstring (DateTime)型 |
| DescriptionHash | 説明列（長文テキスト）のカスタムフィールドハッシュ（DescriptionA〜DescriptionZ）、値はstring型 |
| CheckHash | チェック列（真偽値）のカスタムフィールドハッシュ（CheckA〜CheckZ）、値はboolean型 |
| AttachmentsHash | 添付ファイル列のカスタムフィールドハッシュ（AttachmentsA〜AttachmentsZ）、値はobject配列型 |
| Title | レコードのタイトルフィールド（必須） |
| Body | レコードの本文フィールド |
| Results / Issues | プリザンターのテーブルタイプ（期限管理なし / あり） |
| プロファイル | 接続先サーバ・APIキーのセット（production, staging等） |
| マッピング定義 | CSVカラムとプリザンターフィールドの対応表（YAML） |
| convenience flags | 人間向けの簡易フラグ（--title, --body, --status） |
| RAWペイロード | --json で渡すプリザンターAPIのリクエストJSON |
| NDJSON | Newline Delimited JSON（1行1レコードのストリーム形式） |
| View | APIリクエストに埋め込むフィルタ・ソート条件オブジェクト |
| ColumnFilterHash | View内のフィルタ条件（列名→値のハッシュ） |
| ColumnSorterHash | View内のソート条件（列名→asc/desc） |
| TotalCount | ページングレスポンスの総レコード数 |
| Offset / PageSize | ページング制御パラメータ（取得開始位置 / 1ページ件数） |
| getsite | サイト設定取得APIメソッド（テナント管理者権限が必要） |
| bulkupsert | 複数レコード一括作成・更新APIメソッド |
