---
name: column-design
description: Pleasanterテーブルのカラム割り当て戦略。テーブル設計時に参照するガイド。
---

# カラム設計ガイド

Pleasanter テーブルのカラム割り当て戦略。新しいアプリを設計する際に参照する。

## カラム型と上限

| 型 | プレフィックス | 範囲 | 最大数 | 用途 |
|----|-------------|------|--------|------|
| Class | ClassA〜ClassZ | A-Z | 26 | 文字列（短いテキスト、選択肢、リンク） |
| Num | NumA〜NumZ | A-Z | 26 | 数値（金額、数量、スコア） |
| Date | DateA〜DateZ | A-Z | 26 | 日付（カスタム日付フィールド） |
| Description | DescriptionA〜DescriptionZ | A-Z | 26 | 長文テキスト（備考、説明文） |
| Check | CheckA〜CheckZ | A-Z | 26 | チェックボックス（true/false） |

合計: 最大 **130 カラム**（5型 x 26）

## Issues テーブルの組み込みフィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| Status | 数値 | ステータス (100=新規, 200=実施中, 900=完了 等) |
| StartTime | 日時 | 開始日時 |
| CompletionTime | 日時 | 完了予定日時 |
| ProgressRate | 数値 | 進捗率 (0-100) |
| WorkValue | 数値 | 工数 |

これらは DateHash/NumHash ではなくトップレベルで指定する。

## カラム割り当ての原則

### 1. 役割ごとにブロック分け

```
ClassA-C : 識別情報（名前、コード、カテゴリ）
ClassD-F : リンクフィールド（他テーブルへの参照）
ClassG-J : 状態・分類（ステータス、フラグ）
ClassK-Z : 拡張用

NumA-C   : 主要数値（金額、数量）
NumD-F   : 副次数値（スコア、カウント）
NumG-Z   : 拡張用
```

### 2. リンクフィールドは Class を使う

リンク先テーブルのレコードIDをプレーンな数字文字列（例: `"32274"`、ブラケットなし）として格納するため、Class 型を使う。
リンク設定には Links に加えて、Columns の ChoicesText に `"[[SiteId]]"` を設定する必要がある（UIでドロップダウン表示にするため）。

### 3. 選択肢は ChoicesText で定義

```json
{"ColumnName": "ClassC", "LabelText": "カテゴリ", "ChoicesText": "選択肢1\n選択肢2\n選択肢3"}
```

## テーブル設計パターン

### パターン0: 判別子（多態）テーブル（10-15カラム）

1つのテーブルで複数の「種別」を持つレコードを管理するパターン。
種別に応じて使用するフィールドが変わる。詳細は `polymorphic-records` スキル参照。

```
ClassA: リンク先（マスタ参照）
ClassB: 種別（判別子）← ChoicesText で選択肢定義
ClassC: 時間帯区分
ClassD: 曜日パターン（複合選択肢: 月〜金/土日/毎日 等）
ClassE: 追加リンク（任意）
DateA-B: 開始/終了時刻
DateC-D: 適用開始/終了日（期間/除外で使用）
DateE: 特定日（スポット/除外で使用）
NumA: 数量/人数
NumB: 優先度（小さいほど高優先、競合解決用）
CheckA: 有効フラグ
DescriptionA: 表示名（TitleColumns、自動生成推奨）
DescriptionB: メモ（理由・背景の記録）
```

**ポイント**:
- ClassB（判別子）でフィールド表示/非表示を切替（UIスクリプト）
- NumBで定常/期間/スポットの優先度上書き
- DescriptionAに自動生成した枠名を格納（TitleColumns対策）

### パターン1: シンプルマスタ（5-8カラム）
```
ClassA: 名前（TitleColumns）、ClassB: コード（NoDuplication）、ClassC: カテゴリ
NumA: 主要数値、DescriptionA: 備考
```

### パターン2: リンク付きマスタ（8-12カラム）
```
ClassA-C: 識別情報、ClassD-E: リンク先
NumA-C: 数値、DateA: 日付、DescriptionA: 備考
```

### パターン3: トランザクション/Issues（10-15カラム）
```
ClassA: 伝票番号（NoDuplication）、ClassB: 区分、ClassC-E: リンク先
NumA: 数量/金額、Status/StartTime/CompletionTime: 組み込み
```

## 設計時のチェックリスト

- [ ] ReferenceType は Results か Issues か？
- [ ] TitleColumns は設定したか？
- [ ] EditorColumns に使用する全フィールドを含めたか？
- [ ] NoDuplication が必要なフィールドはあるか？
- [ ] ChoicesText が必要なフィールドはあるか？
- [ ] リンクが必要なフィールドはあるか？（Links + ChoicesText の両方が必要）

## Scripts / Styles の設定

テーブルにクライアントサイドスクリプト（JS/CSS）を登録する場合、SiteSettings に Scripts/Styles 配列を追加する。

```json
{
  "SiteSettings": {
    "Scripts": [
      {"Id": 1, "Title": "スクリプト名", "Body": "JSコード", "New": true, "Edit": true}
    ],
    "Styles": [
      {"Id": 1, "Title": "スタイル名", "Body": "CSSコード", "Index": true}
    ]
  }
}
```

**Id は1以上の整数を必ず指定する**（0だとUI管理画面に表示されない）。

詳細は `pleasanter-scripts` スキルを参照。

## Columns のオプション一覧

| オプション | 型 | 説明 |
|-----------|-----|------|
| `ColumnName` | string | フィールド名（ClassA等） |
| `LabelText` | string | 表示ラベル |
| `ChoicesText` | string | 選択肢（`\n` 区切り） |
| `ValidateRequired` | bool | 必須フィールド |
| `NoDuplication` | bool | 重複不可 |
