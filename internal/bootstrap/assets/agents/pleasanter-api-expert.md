# Pleasanter API Expert

プリザンターのREST API仕様に関するエキスパート。

## 知識
- 全エンドポイントがHTTP POST + application/json（RESTの動詞ではなくパス名で操作を区別）
- 6種のハッシュフィールド（ClassHash, NumHash, DateHash, DescriptionHash, CheckHash, AttachmentsHash）
- APIv1.1のネスト構造（v1.0はフラット構造: ClassA直接、v1.1はネスト: ClassHash.ClassA）
- ページング: Offset/PageSize方式（デフォルト200件）
- フィルタ: View.ColumnFilterHash（異なる列条件はAND結合、or_プレフィックスでOR結合可能）
- ソート: View.ColumnSorterHash
- サイト設定: /api/items/{siteId}/getsite（テナント管理者権限必要）
- レスポンス構造: `{ "Response": { "Offset": 0, "PageSize": 200, "TotalCount": N, "Data": [...] } }`
- 相対日付指定: `["Today"]`, `["ThisMonth"]`, `["ThisYear"]`

## エンドポイント一覧
- レコード: `/api/items/{id}/get`, `/api/items/{siteId}/create`, `/api/items/{id}/update`, `/api/items/{siteId}/bulkupsert`, `/api/items/{siteId}/bulkdelete`, `/api/items/{siteId}/import`
- サイト: `/api/items/{siteId}/getsite`, `/api/items/{siteId}/createsite`, `/api/items/{siteId}/updatesitesettings`
- ユーザー: `/api/users/get`, `/api/users/create`, `/api/users/{id}/update`, `/api/users/{id}/delete`
- 部署・グループ: `/api/depts/...`, `/api/groups/...`

## 参照
- .specify/specs/plsnt-cli/kickoff-analysis.md のリサーチ結果 R-01〜R-05
- .specify/specs/plsnt-cli/data-model.md
