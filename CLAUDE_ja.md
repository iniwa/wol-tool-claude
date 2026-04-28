# CLAUDE_ja.md - Docker Tool Development on Raspberry Pi 4（詳細版）

> AI が実際に読む指示書は `CLAUDE.md`（英語・簡潔版）です。このファイルは人間向けの詳細リファレンスです。

このファイルは、Raspberry Pi 4 上で動作する Docker ツール開発時の環境・規約をまとめた日本語ドキュメントです。本プロジェクト（WoL Tool）固有の補足は末尾の「本プロジェクト固有の事項」を参照。

---

## 実行環境

| 項目 | 詳細 |
|------|------|
| ホストデバイス | Raspberry Pi 4 Model B (RAM 8GB) |
| アーキテクチャ | `linux/arm64` |
| OS | Raspberry Pi OS (64-bit) |
| ストレージ | SSD 接続済み（コンテナデータ・DB 置き場: `/home/iniwa/docker/`） |
| Docker 管理 | Portainer |
| 外部アクセス | Cloudflared (Cloudflare Tunnel) |

> **注意**: イメージをビルドする際は `linux/arm64` をターゲットアーキテクチャとすること。
> マルチアーキテクチャ対応が必要な場合は `linux/amd64,linux/arm64` を指定する。

---

## 作業環境の判定

- 作業ディレクトリが `D:/Git/` → **自宅**（メイン PC / サブ PC を使用可能）
- 作業ディレクトリが `C:/Users/**/Documents/git/` → **リモート PC**
  - リモート PC には必要な環境（例: ollama）がない。コード修正のみに集中すること。
- ラズパイには `ssh iniwapi` で接続できるため、ラズパイからコードやログを読み取っても良い

---

## Docker / Portainer 運用ルール

- コンテナの管理は **Portainer の Stack (Web Editor)** で行う
- `docker-compose.yml` は直接ファイルとして置かず、**Stack の Web Editor に貼り付ける形**を基本とする
- Stack 名はツール名と合わせる（例: `tool-name`）
- ローカルの compose ファイルを書く場合も、**Portainer Stack にそのまま貼れる形式**にすること

### compose ファイルの基本構造

```yaml
services:
  tool-name:
    image: ghcr.io/iniwa/TOOL_NAME:latest
    container_name: tool-name
    restart: unless-stopped
    ports:
      - "XXXX:XXXX"
    volumes:
      - /home/iniwa/docker/TOOL_NAME/data:/data
    environment:
      - TZ=Asia/Tokyo
```

---

## GitHub Actions / GHCR によるデプロイフロー

### 基本方針

1. ソースコードを GitHub リポジトリで管理する
2. `main` ブランチへの push をトリガーに GitHub Actions が起動する
3. Actions がコンテナイメージをビルドし、GHCR (GitHub Container Registry) へ push する
4. Portainer の Stack で `image: ghcr.io/...` を指定してデプロイする

### GHCR イメージ命名規則

```
ghcr.io/iniwa/{tool-name}:latest
```

- GitHub ユーザー名: `iniwa`
- `{tool-name}` はリポジトリ名と合わせる（小文字ケバブケース）
- タグは基本 `latest` のみ。バージョン管理が必要な場合は `v1.0.0` 形式を追加

### GitHub Actions ワークフロー雛形

`.github/workflows/docker-publish.yml` に以下を配置する:

```yaml
name: Build and Push Docker Image

on:
  push:
    branches:
      - main
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Set up QEMU (for arm64 cross-build)
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=raw,value=latest,enable={{is_default_branch}}
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          platforms: linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

---

## NAS (Synology DS420j) マウント構成

Synology DS420j を Raspberry Pi にマウントして利用している。

### マウント方式

| 用途 | プロトコル | 対象 |
|------|-----------|------|
| Windows PC と共有するフォルダ | SMB | Windows 2台 + Raspberry Pi で共用 |
| Raspberry Pi 専用フォルダ | NFS | Raspberry Pi のみ使用 |

### マウントポイント一覧

NAS の IP アドレス: `192.168.1.190`

#### SMB マウント（Windows PC 2台 + Raspberry Pi で共用）

| NAS 共有名 | マウントポイント | 用途 |
|-----------|----------------|------|
| `photo` | `/mnt/nas/photo` | 写真データ |
| `pi_backup` | `/mnt/nas/pi_backup` | Raspberry Pi のバックアップ |
| `video` | `/mnt/nas/video` | 動画データ |
| `docker` | `/mnt/nas/docker` | ※旧環境の名残。現在はほぼ未使用 |

#### NFS マウント（Raspberry Pi 専用）

| NAS ボリュームパス | マウントポイント | 用途 |
|------------------|----------------|------|
| `/volume1/git-data` | `/mnt/nas/git-data` | Git リポジトリ本体・LFS などの大容量データ |
| `/volume1/NetBackup` | `/mnt/nas/NetBackup` | ネットワークバックアップ |

### ストレージ使い分け方針

| データ種別 | 保存先 | 理由 |
|-----------|--------|------|
| コンテナデータ全般・DB | `/home/iniwa/docker/{tool-name}/` | SSD 直結で I/O 速度が高い |
| Git リポジトリ・LFS | `/mnt/nas/git-data/` | 大容量のため NAS（NFS）に退避 |
| 写真・動画（参照のみ） | `/mnt/nas/photo/`, `/mnt/nas/video/` | Windows とも共有する既存データ |

> **基本原則**: Docker ツールのデータ置き場は `/home/iniwa/docker/{tool-name}/` を使う。
> NAS を使うのは「大容量データの参照・保存」か「Windows との共有」が必要な場合のみ。

### compose でのボリューム指定例

```yaml
volumes:
  # 通常のコンテナデータ・DB（SSD）← 基本はこちら
  - /home/iniwa/docker/{tool-name}/data:/data
  - /home/iniwa/docker/{tool-name}/db:/var/lib/postgresql/data  # DB 例

  # 大容量データ（NFS）← リポジトリ・LFS など
  - /mnt/nas/git-data/{tool-name}:/repo

  # メディア参照（SMB・読み取り専用推奨）
  - /mnt/nas/photo:/media/photo:ro
  - /mnt/nas/video:/media/video:ro
```

> **注意**: NAS マウントが前提のコンテナは、マウントが外れた場合に備えて `restart: unless-stopped` を使うこと。

---

## ネットワーク / 外部アクセス

- **ローカルアクセス**: 各コンテナのポートへ直接アクセス（例: `http://raspberrypi.local:8080`）
- **外部アクセス**: Cloudflared (Cloudflare Tunnel) 経由でインターネットから安全にアクセス可能
  - Cloudflared は Raspberry Pi にインストール済み
  - 新しいツールを外部公開する場合は Cloudflare の設定も必要

---

## 開発時の注意事項

### アーキテクチャ

- ベースイメージは `arm64` 対応のものを選ぶこと
- `alpine` ベースは軽量で arm64 対応済みのため推奨
- `debian`/`ubuntu` ベースも arm64 対応済み
- **注意**: `amd64` のみのイメージは Raspberry Pi で動作しない

### 推奨ベースイメージ例

できる限りその時点の最新安定版を使うこと。

```dockerfile
# Python ツールの場合
FROM python:<latest>-slim

# Node.js ツールの場合
FROM node:<latest-lts>-alpine

# Go ツールの場合 (マルチステージビルド)
FROM golang:<latest>-alpine AS builder
FROM alpine:<latest>
```

### リソース制限

RAM 8GB の Raspberry Pi でも、複数コンテナが同時稼働するため、必要に応じてリソース制限を設ける:

```yaml
services:
  tool-name:
    image: ghcr.io/iniwa/TOOL_NAME:latest
    deploy:
      resources:
        limits:
          memory: 512m
```

### タイムゾーン

全コンテナに `TZ=Asia/Tokyo` を設定すること:

```yaml
environment:
  - TZ=Asia/Tokyo
```

### 非 root ユーザー実行

セキュリティ強化のため、可能なら Dockerfile で非 root ユーザーを作成して実行する:

```dockerfile
RUN adduser -D -u 1000 app
USER app
```

bind mount を使う場合は、ホスト側ディレクトリの所有者も UID 1000 に揃えること:

```bash
sudo chown -R 1000:1000 /home/iniwa/docker/{tool-name}/
```

---

## .gitignore / .claudeignore / .dockerignore

各プロジェクトのルートに以下を配置すること。

- **`.gitignore`** — Git にコミットしないファイル（ビルド成果物・実行時データ・機密情報・エディタ設定など）
- **`.claudeignore`** — Claude Code がコードを読む際に除外するファイル
- **`.dockerignore`** — Docker ビルドコンテキストから除外するファイル

### 共通除外対象

```gitignore
# Git 内部ファイル
.git/

# ログファイル
*.log
*.log.*
logs/

# 一時ファイル
*.tmp
*.temp
*.bak
*.backup
*.orig
*~
tmp/
temp/
.cache/

# ビルド成果物
dist/
build/
out/

# 言語別キャッシュ・依存関係
__pycache__/
*.pyc
*.pyo
*.pyd
.venv/
venv/
node_modules/
.next/
target/

# エディタ・OS 生成ファイル
.DS_Store
Thumbs.db
*.swp
*.swo
.idea/
.vscode/

# 機密情報（誤って読まれない/コミットされないように）
.env
.env.*
*.pem
*.key
secrets/

# Claude Code / MCP ローカル状態
.claude/settings.local.json
.serena/
```

> 機密情報を含むファイルは必ず除外しておくこと。

---

## プロジェクト構成の雛形

```
tool-name/
├── .github/
│   └── workflows/
│       └── docker-publish.yml   # GitHub Actions
├── Dockerfile
├── docker-compose.yml           # ローカルテスト用 or Portainer 貼り付け用
├── .dockerignore
├── .gitignore
├── .claudeignore                # Claude Code の除外設定
├── README.md
├── CLAUDE.md                    # AI 向け簡潔指示書
├── CLAUDE_ja.md                 # 人間向け詳細リファレンス（このファイル）
├── docs/                        # 知見・設計判断の蓄積
└── src/                         # アプリケーションコード
```

---

## 知見の永続化

- 設計判断・アーキテクチャの選定理由・利用フレームワークの知見など、流用できる情報は `docs/*.md` に積極的に残す
- 作業開始時には `docs/` に既存の文脈がないか確認する
- `/clear` で会話をリセットしても、`CLAUDE.md` と `docs/` に残した情報が次の会話で引き継がれる

---

## ツール活用

- コードの読み取り・編集には **Serena MCP** ツールを積極的に使う（シンボル検索・概要取得・置換・挿入など）
- Web 上の情報収集には **Tavily MCP** ツールを使う:
  - `tavily_search` — ドキュメント、エラーメッセージ、ライブラリの使い方などの一般的な Web 検索
  - `tavily_crawl` — 特定の Web サイトを巡回して詳細な情報を取得
  - `tavily_extract` — URL から構造化されたコンテンツを抽出
  - `tavily_research` — トピックについての詳細なリサーチ（複雑・多面的な調査に使用）

---

## チェックリスト（新規ツール作成時）

- [ ] ベースイメージが `arm64` 対応か確認
- [ ] `TZ=Asia/Tokyo` を environment に追加
- [ ] `restart: unless-stopped` を設定
- [ ] GHCR イメージ名を正しい形式で記述
- [ ] GitHub Actions ワークフローを `.github/workflows/` に配置
- [ ] `.gitignore` + `.claudeignore` + `.dockerignore` をプロジェクトルートに配置
- [ ] 認証情報を扱う場合は `AUTH_*` 等の env var で外出しし、平文を API レスポンスに含めない
- [ ] 非 root ユーザーで実行（bind mount 先の UID も合わせる）
- [ ] Portainer Stack への貼り付けで動作確認
- [ ] 外部公開が必要なら Cloudflare Tunnel + 認証層を追加
- [ ] NAS マウントを使う場合、マウントポイントのパスが正しいか確認

---

## 本プロジェクト固有の事項（WoL Tool）

| 項目 | 値 |
|------|-----|
| 言語 | Go 1.22（標準ライブラリのみ、外部依存なし） |
| ベースイメージ | `golang:1.22-alpine`（builder）→ `alpine:3.19`（runtime） |
| ランタイム依存 | `iputils`（ping）, `samba-client`（`net rpc shutdown`）, `tzdata` |
| データ永続化 | `/data/devices.json`（DB なし） |
| ネットワーク | `network_mode: host`（WoL マジックパケットの UDP ブロードキャスト 255.255.255.255:9 のため必須） |
| 認証 | `AUTH_USER` / `AUTH_PASS` 設定時のみ Basic 認証有効 |
| イメージ | `ghcr.io/iniwa/wol-tool-claude:latest` |
| bind mount | `/home/iniwa/docker/wol-claude/:/data`（UID 1000 で chown 必須） |

### セキュリティ前提

- LAN 内信頼環境を前提に設計されている
- `shutdown_pass` は `devices.json` に平文保存される（API レスポンスからは除外済み）
- 外部公開する場合は **必ず `AUTH_USER`/`AUTH_PASS` を設定**すること
- CSRF: Same-origin 前提で CORS は撤去済み

### 主要 API

| メソッド | パス | 用途 |
|---------|------|------|
| `GET` | `/api/devices` | 一覧取得（パスワード非返却、`has_shutdown_pass` のみ） |
| `POST` | `/api/devices` | 追加 |
| `PUT` | `/api/devices/{id}` | 更新（`shutdown_pass` 空文字なら既存保持、`clear_shutdown_pass: true` で削除） |
| `DELETE` | `/api/devices/{id}` | 削除 |
| `POST` | `/api/devices/{id}/wake` | WoL マジックパケット送信 |
| `POST` | `/api/devices/{id}/shutdown` | リモートシャットダウン（パスワードは stdin 経由で `net rpc` に渡す） |
| `POST` | `/api/devices/{id}/ping` | 単体 ping |
| `POST` | `/api/ping/all` | 全台 ping |
| `GET`/`PUT` | `/api/config` | ping 間隔設定 |
