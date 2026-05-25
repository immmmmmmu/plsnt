#!/usr/bin/env bash
# 図書貸出管理モデルの SiteID 環境変数
# batch run 実行後に出力される SiteID をここに設定する
#
# 使い方:
#   1. plsnt batch run templates/scaffold-library.yaml --var folder_id=<FOLDER_ID>
#   2. 出力から各テーブルの SiteID を確認
#   3. 以下の値を更新
#   4. source scripts/library/env.sh

export LIB_GENRE_SITE=32293
export LIB_PUBLISHER_SITE=32294
export LIB_MEMBER_SITE=32295
export LIB_SHELF_SITE=32296
export LIB_BOOK_SITE=32297
export LIB_LENDING_SITE=32298
export LIB_COLLECTION_SITE=32299
export LIB_LENDING_ITEM_SITE=32300
export LIB_RETURN_SITE=32301
