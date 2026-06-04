# サイト構築ワークフロー

site-builder エージェントを使用して、プリザンターのサイト（テーブル）を構築する。

## 使用方法

```
/site-build <用途の説明>
```

## ワークフロー

1. site-builder エージェントに用途を伝える
2. エージェントが以下を自動実行:
   - ReferenceType（Results/Issues）の選択
   - カラム設計（ClassHash/NumHash/DateHash等の割当）
   - EditorColumnsの構成
   - テーブル作成コマンドの生成と実行
   - スキーマ確認

## 入力パラメータ

- `$ARGUMENTS`: 構築したいテーブルの用途・要件の説明

## 実行例

```
/site-build 顧客マスタを作りたい。会社名、担当者、電話番号、メール、年間売上の管理
```

```
/site-build タスク管理テーブル。プロジェクト名とステータス、期限の管理が必要
```

## エージェント呼び出し

site-builder エージェントに以下を依頼:
1. 「$ARGUMENTS」の要件を分析
2. 適切なReferenceTypeとカラム設計を提案
3. ユーザー確認後、plsnt site create コマンドで構築
4. plsnt schema で結果を確認
