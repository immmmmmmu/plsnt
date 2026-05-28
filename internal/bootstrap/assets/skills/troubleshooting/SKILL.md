---
name: troubleshooting
description: plsnt CLIのエラーコード別対処法、よくあるミスと解決策。エラー発生時に自動参照される。
---

# plsnt トラブルシューティング

よくあるエラーと対処法。

## エラーコード別の対処

### HTTP 404 - Not Found
```
Error: request error (HTTP 404)
```
**原因**: 指定した SiteId または RecordId が存在しない
**対処**:
- `site search` や `record list` で正しいIDを確認する
- IDの桁数やタイプミスを確認

### HTTP 403 - Forbidden
```
Error: authentication failed (HTTP 403)
```
**原因**: APIキーに対象リソースへのアクセス権限がない
**対処**:
- `plsnt config list` でAPIキーを確認
- Pleasanter 管理画面でAPIキーの権限を確認
- 削除済みレコードにアクセスしようとしていないか確認

### HTTP 409 - Conflict
```
Error: request error (HTTP 409)
```
**原因**: `NoDuplication` 制約違反（重複不可フィールドに既存値と同じ値を設定）
**対処**:
- 重複している値を確認: `record list` で既存データを検索
- 別の値を使用するか、既存レコードを更新する

### JSON パースエラー
```
Error: JSON payload must start with '{' or '['
```
**原因**: `--json` に渡した値が有効な JSON ではない
**対処**:
- シングルクォートで JSON 全体を囲む: `--json '{"key":"value"}'`
- シェルの特殊文字（`!`, `$`, `` ` ``）に注意
- JSON のダブルクォート、カンマ、括弧の対応を確認

### 接続エラー
```
Error: failed to connect to <url>
```
**原因**: Pleasanter サーバーに接続できない
**対処**:
- URL が正しいか `plsnt config list` で確認
- サーバーが起動しているか確認
- 自己署名証明書の場合は `--insecure` フラグを追加

## よくあるミス

### コマンド引数の間違い

`--site-id` が必要なコマンドと不要なコマンドがある:

| 必要 | 不要 |
|------|------|
| `record list --site-id <id>` | `record get <record-id>` |
| `record create --site-id <id>` | `record update <record-id>` |
| | `record delete <record-id>` |

### EditorColumnHash の未設定

フィールドにデータを保存するには、サイトの `EditorColumnHash.General` にそのフィールド名を含める必要がある。

**症状**: レコード作成は成功するが、データが保存されない（空になる）

### 空 JSON でレコード作成

```bash
# 空JSONでもレコードが作成されてしまう（「タイトル無し」として）
plsnt record create --site-id 32147 --json '{}'
```
`ValidateRequired` を設定していても API 経由では検証されない場合がある。

### ChoicesText フィールドのフィルタ

`ChoicesText`（選択肢）で定義されたフィールドは `ColumnFilterHash` でうまくフィルタできない場合がある。
代替: 全件取得してクライアント側でフィルタ。

### リンクフィールドのフィルタ

リンクフィールドは ColumnFilterHash で **ブラケット付きレコードID** でフィルタする（例: `"ClassA":"[32274]"`）。リンク先の名前ではフィルタできない。
なお、レコード作成/更新時の値はブラケットなし（例: `"ClassA":"32274"`）。ブラケット形式は ColumnFilterHash 検索時のみ。

### リンクフィールドのUI表示

Links 設定だけではエディタがプレーンテキスト入力のまま（生IDが表示される）。
Columns の ChoicesText に `"[[SiteId]]"` を設定して初めてドロップダウン/リンクセレクタ表示になる。

```json
{"ColumnName": "ClassA", "LabelText": "顧客", "ChoicesText": "[[32262]]"}
```

### site create の parent-id に Results テーブルを指定

`--parent-id` に Results/Issues テーブルの SiteId を指定すると、サイトではなく **レコードが作成** される。
**確認方法**: `site get <parent-id>` で `ReferenceType` が `"Sites"` であることを確認。

### EditorColumnHash と EditorColumns の違い

`EditorColumnHash` はエラーになる。`EditorColumns`（配列形式）を使う。

### site create が 302 リダイレクトを返す

**原因**: フォルダ（Sites）の作成や、一部の環境でのテーブル作成で発生。
**対処**: site search で作成されたかどうかを確認する。

### フォルダに record list を実行

**原因**: フォルダ（ReferenceType: Sites）に対して `record list` を実行した
**対処**: フォルダの内容確認は `site get <folder-site-id>` を使用。

### Issues テーブルで Status が取得できない

**原因**: `--json` フラグなしでは Status/StartTime/CompletionTime が返されない
**対処**: `--json '{}'` を付ける。

### NumHash の科学記数法表示

**症状**: `-o table` で大きな数値が `1.2e+06` と表示される
**対処**: 正確な値が必要な場合は `-o json` を使用。

### Scripts/Styles がPleasanter UIの管理画面に表示されない

**症状**: API経由でScripts/Stylesを登録したが、テーブルの管理→スクリプト/スタイルに表示されない
**原因**: `Id` が `0` で登録されている。Pleasanterは `Id` が1以上でないとUI管理画面で認識しない
**対処**: 各Script/Styleに一意の `Id`（1以上の整数）を付与して再登録

```json
// 悪い例: Id 未指定 → 0 が設定される → UIに表示されない
{"Title": "スクリプト名", "Body": "...", "New": true}

// 良い例: Id を明示指定 → UIに表示される
{"Id": 1, "Title": "スクリプト名", "Body": "...", "New": true}
```

テーブルごとにIdは1から独立して採番可能。

### 日付範囲フィルタの不安定動作

`ColumnFilterHash` での日付範囲フィルタが正しく動作しない場合がある。
代替: クライアントサイドフィルタ。

### site update で SiteSettings の一部キーだけ送ると Columns が消える

**症状**: `plsnt site update <id> --json '{"SiteSettings":{"Summaries":[...]}}'` のように SiteSettings の一部キーだけ送ると 200 OK が返るが、明示送信しなかった `Columns` / `EditorColumnHash` / `GridColumns` / `Links` / `TitleColumns` が消失する。データ自体は無事だが、明細 API レスポンスから `ClassHash` / `NumHash` が空オブジェクトで返るようになり、UI も壊れる。

**原因**: Pleasanter の updatesite は SiteSettings 全置換動作。partial update を保証しない。

**対処**:
1. 必ず `plsnt site get <id>` で現状の SiteSettings 全体を取得
2. 変更したいキー (例: Summaries) を上書きしつつ、他キーを全部含めて送信
3. 200 OK でも `plsnt site get` で各キーの存在を再検証

**最小再現**:
```bash
# ⛔ 危険 — Columns が消える
plsnt site update 32822 --json '{"SiteSettings":{"Summaries":[{...}]}}'

# ✅ 安全 — 全キーを含めて送る
plsnt site get 32822 -o json > /tmp/cur.json
# /tmp/cur.json の SiteSettings に Summaries を追加してマージ
plsnt site update 32822 --json "$(jq '.SiteSettings | {SiteSettings: .}' /tmp/cur.json)"
```

**復旧手順** (既に消してしまった場合):
1. `plsnt site get <id>` で残っている設定を確認
2. 同種サイトまたはサイトパッケージ YAML から Columns 定義を取得
3. ChoicesText の `[[<SiteId>]]` 形式リンクを実数値に置換
4. EditorColumnHash / GridColumns / Links / Columns / Summaries をまとめて再送

**関連スキル**: `summary-aggregation` (Summaries 設定の本来用途), `sitepackage-management` (YAML 経由なら欠損しない)

#### ✅ 安全な実装パターン (merge 方式)

`--json` で SiteSettings の一部キーだけ送るのは避け、`getsite` → 既存設定とマージ → `updatesite` のパターンで実装する。Python 例:

```python
def get_site(site_id):
    return post(f"/api/items/{site_id}/getsite", {}).get("Response", {}).get("Data", {})

def update_site(site_id, ss, title):
    return post(f"/api/items/{site_id}/updatesite", {"SiteSettings": ss, "Title": title})

# 安全パターン: 全体を取得して、変更したいキーだけ書き換える
site = get_site(32822)
ss = site.get("SiteSettings") or {}
ss["Summaries"] = [{...新規 Summary 設定...}]  # 他のキーは触らない
update_site(32822, ss, site.get("Title"))
```

この get → 部分変更 → update のパターンにより、Columns / Scripts / Summaries 等を相互に壊さずに更新できる。

### Links の LabelText は黙って捨てられる

**症状**: `SiteSettings.Links` の各エントリに `LabelText` を含めて `updatesite` を投げると 200 OK が返るが、`getsite` で再取得すると `LabelText` が `null` になっている。

```json
// 送信
{"ColumnName": "ClassA", "SiteId": 32817, "LabelText": "商品"}
// getsite で返ってくる値
{"ColumnName": "ClassA", "SiteId": 32817, "LabelText": null}
```

**原因**: Pleasanter API は Links 配列要素の `LabelText` フィールドを受け付けない。`GridColumnHash` と同じ「黙って捨てる」パターン。

**対処**: UI 表示は Columns 側の LabelText が担うので、Links の LabelText 設定は**不要**。
- ❌ `"Links": [{"ColumnName": "ClassA", "SiteId": 32817, "LabelText": "商品"}]` ← LabelText 部分は無効
- ✅ `"Columns": [{"ColumnName": "ClassA", "LabelText": "商品", "ChoicesText": "[[32817]]"}]` ← こちらが効く

**確認方法**:
```bash
# Links に LabelText を載せる必要はない (どうせ捨てられる)
plsnt site get <site_id> -o json | jq '.SiteSettings.Links'
# UI 表示用のラベルは Columns 側を見る
plsnt site get <site_id> -o json | jq '.SiteSettings.Columns[] | select(.ColumnName | startswith("Class")) | {ColumnName, LabelText, ChoicesText}'
```

**関連**: `relational-modeling` スキルの「リンクフィールドの UI 表示」 — ChoicesText `"[[SiteId]]"` がドロップダウン表示の本体
