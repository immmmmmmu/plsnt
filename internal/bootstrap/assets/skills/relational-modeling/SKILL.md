---
name: relational-modeling
description: ERモデル(1:N, M:N)をPleasanterのテーブル構成で実現するパターン。マスタ・トランザクション・明細テーブルの設計手法。
---

# リレーショナルモデリング in Pleasanter

ERモデルをPleasanterのClass/Num/Date/DescriptionHashで実現する設計パターン。

## 基本原則

### テーブル種別の選択

| ERエンティティ種別 | Pleasanter ReferenceType | 理由 |
|---|---|---|
| マスタ（顧客、商品、店舗） | Results | ステータス管理不要 |
| トランザクション（注文、発注） | Issues | Status/StartTime/CompletionTime が使える |
| 明細（注文明細、請求明細） | Results | 親トランザクションへのリンクで管理 |

### リレーション実現パターン

#### 1:N リレーション（マスタ → トランザクション）

トランザクション側の Class カラムにマスタのレコードIDを格納し、Links で紐付ける。

```json
{
  "SiteSettings": {
    "Columns": [
      {"ColumnName": "ClassA", "LabelText": "顧客", "ChoicesText": "[[CUSTOMER_SITE_ID]]"}
    ],
    "Links": [
      {"ColumnName": "ClassA", "SiteId": CUSTOMER_SITE_ID, "LabelText": "顧客"}
    ]
  }
}
```

#### 1:N リレーション（トランザクション → 明細）

明細テーブルの Class カラムに親トランザクションのレコードIDを格納。

```json
{
  "Columns": [
    {"ColumnName": "ClassA", "LabelText": "注文", "ChoicesText": "[[ORDER_SITE_ID]]"}
  ],
  "Links": [
    {"ColumnName": "ClassA", "SiteId": ORDER_SITE_ID, "LabelText": "注文"}
  ]
}
```

#### M:N リレーション（中間テーブル方式）

中間テーブル（明細テーブル）で2つのマスタを参照する。

```
商品マスタ ──┐
             ├── 注文明細（中間テーブル）
注文 ────────┘
```

## お買い物モデル（基本版 - 5テーブル）

### ER図 → Pleasanter マッピング

```
顧客マスタ (Results)          店舗マスタ (Results)
  ClassA: 顧客名                ClassA: 店舗名
  ClassB: 電話番号              ClassB: 住所
  ClassC: メールアドレス         ClassC: 電話番号
  DescriptionA: 住所
       ↓                            ↓
注文 (Issues) ←─ Status/StartTime/CompletionTime 活用
  DateA: 注文日
  ClassA: 顧客（Link→顧客マスタ）
  ClassB: 店舗（Link→店舗マスタ）
  NumA: 合計金額
       ↓
注文明細 (Results)
  ClassA: 注文（Link→注文）
  ClassB: 商品（Link→商品マスタ）
  NumA: 数量
  NumB: 単価
  NumC: 小計

商品マスタ (Results)
  ClassA: 商品名
  ClassB: カテゴリ
  NumA: 標準価格
  ClassC: 単位
```

```bash
# 上記構成を batch YAML（自作）に定義して一括構築
plsnt batch run shopping-model.yaml --var folder_id=<FOLDER_ID>
```

## お買い物モデル（拡張版 - 9テーブル）

基本版にカテゴリ・仕入先・支払い・在庫の4テーブルを追加した実務向けモデル。

### 追加エンティティの設計

```
カテゴリマスタ (Results)       仕入先マスタ (Results)
  ClassA: カテゴリ名             ClassA: 仕入先名
       ↓                            ↓
商品マスタ (Results) ←── 拡張: ClassB→カテゴリLink, ClassD→仕入先Link
  ClassA: 商品名
  ClassB: カテゴリ（Link→カテゴリマスタ）
  ClassC: 単位
  ClassD: 仕入先（Link→仕入先マスタ）
  NumA: 標準価格

支払い (Results) ←── 注文との1:1リレーション
  ClassA: 注文（Link→注文）
  ClassB: 支払方法（テキスト: クレジットカード/現金/電子マネー等）
  NumA: 支払金額
  DateA: 支払日

在庫 (Results) ←── 店舗×商品の中間テーブル（M:N）
  ClassA: 店舗（Link→店舗マスタ）
  ClassB: 商品（Link→商品マスタ）
  NumA: 在庫数量
```

### テンプレート構造（14ステップ）

| フェーズ | ステップ数 | 内容 |
|---------|-----------|------|
| 1. マスタ作成 | 4 | カテゴリ、仕入先、顧客、店舗（並列実行可） |
| 2. 商品作成 | 1 | カテゴリ・仕入先に依存 |
| 3. トランザクション/明細 | 4 | 注文、注文明細、支払い、在庫 |
| 4. リンク設定 | 5 | 全テーブルのLinks追加（SiteSettings全体上書き） |

```bash
# 9テーブル一括作成（batch YAML は自作）
plsnt batch run shopping-model-v2.yaml --var folder_id=<FOLDER_ID>

# dry-run で確認
plsnt batch run shopping-model-v2.yaml --var folder_id=<FOLDER_ID> --dry-run
```

### リンク設定の自動化

テンプレートはステップ出力参照 `{{step.Id}}` でリンクの SiteId を自動解決する。
テーブル作成 → リンク設定が1コマンドで完了。

### 1:1 リレーションの実現（注文 → 支払い）

Pleasanterに1:1制約はないため、設計で担保する：
- 支払いテーブルの ClassA に注文のRecordIDをLink
- アプリケーション側で1注文1支払いの整合性を管理
- 重複チェックが必要な場合は NoDuplication を ClassA に設定

## Pleasanter API の重要な制約

### 1. createsite API は SiteSettings 必須
SiteSettings 無しだと 302→/errors エラー。SiteSettings 付きなら 200 + Id が返る。

### 2. site update は SiteSettings 全体を上書き
Links だけ送ると EditorColumnHash/Columns が消える。update 時は全 SiteSettings を含める。

```json
// 悪い例: Links だけ送る → Columns が消える
{"SiteSettings":{"Links":[...]}}

// 良い例: Columns + Links を両方含める
{"SiteSettings":{"EditorColumnHash":{...},"Columns":[...],"Links":[...]}}
```

### 3. --view と --json の併用でフィルタ無効化
`plsnt record list` で `--view` と `--json` を同時指定すると `--json` が優先され View フィルタが無視される。
リンクフィールドの逆引き検索では `--view` のみを使うこと。

### 4. フォルダの子サイト一覧 API 無し
Sites 型に /api/items/{id}/get → フォルダ自身の情報のみ。子サイトの列挙は不可。

## テンプレート拡張のパターン

既存テンプレートを拡張する際の手順：

1. **新マスタテーブルを Phase 1 に追加**（依存なし、並列実行可）
2. **既存テーブルのカラムを拡張**（新しいClassカラムを追加）
3. **新テーブルの create ステップを追加**（depends_on で依存関係を設定）
4. **リンク設定ステップを追加**（EditorColumnHash + Columns + Links の全体を含める）

### 拡張時の注意

- 既存テーブルにカラムを追加する場合、EditorColumnHash の General 配列に新カラムを追加
- リンク設定の update では必ず既存の Columns 定義も含める（上書きで消えるため）
- 1:1 リレーションは Links + アプリケーション側制約で実現（DB制約なし）
- 中間テーブル（在庫 = 店舗×商品）は Results で2つのClassカラムにLinkを設定

## TitleColumns 設定（リンク表示に必須）

### なぜ TitleColumns が重要か

Pleasanter のリンクフィールドは、参照先レコードの **Title（ItemTitle）** を表示する。
TitleColumns が未設定（null）だとタイトルが空になり、リンク先が「タイトル無し」と表示される。

### TitleColumns のルール

1. **マスタテーブル**: `TitleColumns: ["ClassA"]`（名前フィールド）
2. **日付ベースのテーブル**: `TitleColumns: ["DateA"]`（注文日など）
3. **テキスト値のテーブル**: `TitleColumns: ["ClassB"]`（支払方法など、リンクでないClassカラム）
4. **リンクフィールドのみのテーブル**: `TitleColumns: ["DescriptionA"]` + DescriptionA フィールドを追加

### リンクフィールドは TitleColumns に使えない

リンクフィールド（Links で設定された ClassA/ClassB 等）を TitleColumns に指定すると、
生のレコードIDが Title に表示される（名前解決されない）。

**解決策**: DescriptionA フィールドを追加し、レコード作成時に解決済みの名前を格納する。

```
# 注文明細: DescriptionA に「商品名 x数量」を格納
DescriptionA = "おにぎり（鮭） x2"

# 在庫: DescriptionA に「店舗名 - 商品名」を格納
DescriptionA = "渋谷店 - おにぎり（鮭）"
```

### リンク設定後の既存レコード（重要）

Links 設定前に作成されたレコードは、リンクフィールドが有効化されない。
**空の update では不十分** — リンクフィールドの値を含めて再保存する必要がある。

```bash
# 悪い例: 空 update → リンク解決されない
plsnt record update $RECORD_ID --json '{}'

# 良い例: リンクフィールドの値を再設定 → リンク解決される
plsnt record update $RECORD_ID --json '{"ClassHash":{"ClassA":"32274","ClassB":"32275"}}'
```

TitleColumns 変更後も同様に、TitleColumns 対象フィールドの値を含めた update が必要。

## 対称テーブル設計パターン

需要と供給のように対になる2つのテーブルを、同じカラム構造で設計するパターン。
マッチングロジックやUIスクリプトの共通化が可能になる。

### 例: シフト管理の需要/供給

```
現場シフト枠（需要）           稼働可能枠（供給）
  ClassA: 現場（Link）          ClassA: 警備員（Link）
  ClassB: 枠種別                ClassB: 枠種別
  ClassC: 時間帯区分            ClassC: 時間帯区分
  ClassD: 曜日パターン          ClassD: 曜日パターン
  DateA-E: 時刻・期間・特定日   DateA-E: 時刻・期間・特定日
  CheckA: 有効                  CheckA: 有効
  DescriptionA: 枠名            DescriptionA: 枠名
```

### いつ使うか

- 需要と供給のマッチング（シフト、予約、リソース配分）
- 双方向の条件フィルタリングが必要な場合
- 同じ判定ロジック（枠種別、曜日パターン等）を両側に適用する場合

### メリット

- マッチング/照合ロジックの共通化
- フィールド切替・名前自動生成スクリプトの再利用
- 設計の対称性による理解しやすさ

## 優先度上書きパターン

同じ対象に対して複数のルール（定常/期間/スポット/除外）が存在する場合、
優先度フィールド（NumB等）で競合を解決する。

### 評価順序の例

```
除外(特定日) > 除外(期間) > スポット(特定日) > 期間 > 定常
同一レベルで複数ヒット → 優先度(NumB)が小さい方を採用
```

### カラム設計

```
ClassB: 種別（判別子）  → 定常/スポット/期間/除外
NumB: 優先度（小さいほど高優先）
  定常: 10, 期間: 5, スポット: 1
DateC/DateD: 適用期間（期間/除外で使用）
DateE: 特定日（スポット/除外で使用）
```

詳細は `polymorphic-records` スキルを参照。

## 設計チェックリスト

- [ ] エンティティごとにResults/Issuesを選択したか
- [ ] FK相当のカラムはClassを使用しているか
- [ ] 数値フィールドはNumを使用しているか
- [ ] 日付フィールドはDate（またはIssuesの組み込み）を使用しているか
- [ ] テーブル間リンク（Links + ChoicesText）を設定したか
- [ ] 依存関係の順序でテーブルを作成しているか（マスタ → トランザクション → 明細）
- [ ] バッチテンプレートの depends_on が正しいか
- [ ] 1:1 リレーションの一意性制約を検討したか（NoDuplication）
- [ ] 中間テーブルのリンクは双方のマスタを参照しているか
- [ ] 全テーブルに TitleColumns を設定したか
- [ ] リンクフィールドのみのテーブルに DescriptionA を追加したか

## カラム設計の注意

- 1テーブルに同じ型のカラムは最大26（A-Z）
- リンクフィールドに使うClassと通常のClassは混在可（例: ClassB=カテゴリLink, ClassC=単位テキスト）
- 計算フィールド（小計 = 数量 x 単価）はPleasanter側の数式設定か、CLI側で計算して格納
- NumHash の値は json.Number 型（decimal精度）で扱う
- 支払方法など選択肢の場合、ChoicesText 設定も可能だがLinkではなくテキスト値も実用的
- リンクフィールドには Links 設定に加えて ChoicesText: `"[[SiteId]]"` が必須（UIでドロップダウン表示にするため）
- テーブル設計後のUI改善（自動計算、バリデーション、アラート等）は `pleasanter-scripts` スキルを参照
