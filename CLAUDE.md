# CLAUDE.md

> Detailed notes (Japanese): `CLAUDE_ja.md`（base 共通の詳細リファレンス）

WoL Tool — Wake-on-LAN + Ping 監視 + リモートシャットダウンの軽量 Web ツール（Go 製）。

## Project Overview
- **Language**: Go 1.22 (single binary, no external Go deps)
- **Runtime deps**: `iputils` (ping), `samba-client` (`net rpc shutdown`)
- **Storage**: `/data/devices.json` (JSON file, no DB)
- **Auth**: optional Basic Auth via `AUTH_USER` / `AUTH_PASS` env vars
- **Network**: `network_mode: host` 必須 — UDP ブロードキャスト (255.255.255.255:9) で WoL マジックパケット送信のため

## Coding Style
- 軽量・依存最小を維持。Go 標準ライブラリのみで完結させる方針。
- 静的アセットは `static/` 配下にバンドル。フロントは plain JS（フレームワーク不使用）。

## Environment
- Host: Raspberry Pi 4 (8GB RAM), `linux/arm64`
- Docker management: Portainer — Stack Web Editor only
- Build target: `linux/amd64,linux/arm64`（ローカル動作確認用に amd64 も含める）

### Work Location Detection
- `D:/Git/` → **Home**（メイン PC / サブ PC）
- `C:/Users/**/Documents/git/` → **Remote PC**（環境制限あり、コード修正のみ）
- ラズパイへは `ssh iniwapi` で接続可

## Build & Deploy
- Image: `ghcr.io/iniwa/wol-tool-claude:latest`
- Flow: push to `main` → GitHub Actions → GHCR → Portainer Stack paste
- Container requirements: `restart: unless-stopped`, `TZ=Asia/Tokyo`, `network_mode: host`

## Storage
| Data | Path | Backend |
|------|------|---------|
| `devices.json` | `/home/iniwa/docker/wol-tool/data/` | SSD (bind mount) |

## API Surface
- `GET  /api/devices` — 一覧取得（**パスワードは返却しない**、`has_shutdown_pass` のみ）
- `POST /api/devices` — 追加（name, mac 必須）
- `PUT  /api/devices/{id}` — 更新（`shutdown_pass` が空文字なら既存保持。`clear_shutdown_pass: true` で明示削除）
- `DELETE /api/devices/{id}` — 削除
- `POST /api/devices/{id}/wake` — マジックパケット送信
- `POST /api/devices/{id}/shutdown` — `net rpc shutdown` 実行（パスワードは stdin 経由）
- `POST /api/devices/{id}/ping` — 単体 ping
- `POST /api/ping/all` — 全台 ping
- `GET/PUT /api/config` — ping 間隔設定

## Security Notes（重要）
- このツールは **LAN 内信頼環境を前提**としている。Cloudflare Tunnel 等で外部公開する場合は **必ず `AUTH_USER`/`AUTH_PASS` を設定**すること。
- `shutdown_pass` は `devices.json` に平文保存される。`devices.json` はホスト側で適切なパーミッション（0600）にすること。
- API レスポンスからパスワードは除外済み（`has_shutdown_pass` フラグのみ返す）。
- CSRF: Same-origin 前提で CORS は撤去済み。クロスオリジンからの呼び出しはブロックされる。

## External Access
Cloudflared (Cloudflare Tunnel) はラズパイにインストール済み。ただし本ツールを **外部公開するなら `AUTH_USER`/`AUTH_PASS` 必須**。

## Knowledge Persistence
- 設計判断・トラブルシュート手順は `docs/*.md` に蓄積する
- 作業開始時に `docs/` に既存メモがないか確認

## Tooling
- Use **Serena MCP** for code navigation/edit (symbol search, replace, insert)
- Use **Tavily MCP** for web search/research:
  - `tavily_search` — general docs / error messages / library usage
  - `tavily_crawl` — site-specific deep crawl
  - `tavily_extract` — URL → structured content
  - `tavily_research` — multi-faceted in-depth research

## Checklist (新規ツール作成時の参考)
- [ ] arm64-compatible base image (`alpine` preferred)
- [ ] `TZ=Asia/Tokyo` in environment
- [ ] `restart: unless-stopped`
- [ ] Image: `ghcr.io/iniwa/{tool-name}:latest`
- [ ] GitHub Actions workflow at `.github/workflows/`
- [ ] `.gitignore` + `.claudeignore` + `.dockerignore` をルートに配置
- [ ] 認証情報を扱う場合は `AUTH_*` 等の env var で外出し、平文を API レスポンスに含めない
- [ ] Portainer Stack 貼り付けで動作確認
- [ ] 外部公開する場合は Cloudflare Tunnel + 認証必須
