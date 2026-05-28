---
name: multi-environment
description: 複数のPleasanter環境（開発・ステージング・本番）をプロファイルで管理するパターン。
---

# マルチ環境管理

複数の Pleasanter 環境（開発・ステージング・本番）をプロファイルで管理する。

## プロファイル管理

```bash
# 名前付きプロファイル作成
plsnt config set --name production --url "https://prod.example.com" --api-key "prod-api-key"

# プロファイル一覧
plsnt config list

# プロファイル切替
plsnt config use production

# 接続テスト
plsnt config test
```

## プロファイルの一時切替

`-p` フラグでコマンド単位で切り替え:

```bash
plsnt record list --site-id 12345 -p production -o table
```

優先順位: `-p` フラグ > `PLSNT_PROFILE` 環境変数 > `current_profile`（config.yaml）

### HTTP 警告

HTTP（非HTTPS）のURLを設定すると警告が出る。本番環境では必ず HTTPS を使用する。

## 設定ファイルの構造

```yaml
current_profile: default
profiles:
    default:
        url: http://dev.example.com
        api_key: your-api-key-here
        api_version: "1.1"
    production:
        url: https://prod.example.com
        api_key: prod-api-key-here
        api_version: "1.1"
```

場所: `~/.config/plsnt/config.yaml`

## 環境構築の推奨パターン

| プロファイル名 | 用途 | URL |
|-------------|------|-----|
| `default` | 開発環境（日常作業用） | dev URL |
| `staging` | ステージング（テスト用） | staging URL |
| `production` | 本番環境（慎重に操作） | 本番 HTTPS URL |
