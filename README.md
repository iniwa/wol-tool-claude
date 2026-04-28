# WoL Tool

LAN 内の PC を Wake-on-LAN で起動・ping で生存監視・SMB 経由でリモートシャットダウンできる軽量 Web ツール。Raspberry Pi 4 上の Docker コンテナでの稼働を想定。

![arch: linux/arm64 + linux/amd64](https://img.shields.io/badge/arch-arm64%20%7C%20amd64-blue) ![go](https://img.shields.io/badge/go-1.22-00ADD8) ![license](https://img.shields.io/badge/license-MIT-green)

## ⚠️ セキュリティに関する重要な警告

**このツールは LAN 内の信頼された環境での使用を前提に設計されています。**

- 認証は `AUTH_USER` / `AUTH_PASS` 環境変数を設定したときのみ有効になる **オプション** です。
- リモートシャットダウン用の Windows ローカル管理者パスワードは `devices.json` に **平文** で保存されます。
- API レスポンスにはパスワードを含めませんが、`devices.json` ファイル自体を読める者には平文で見えます。

### 外部公開時のチェック

Cloudflare Tunnel 等でインターネット公開する場合、以下を **必ず** 実施してください:

- [ ] `AUTH_USER` と `AUTH_PASS` を強力な値に設定する（未設定だと無認証で誰でも起動・シャットダウンできます）
- [ ] Cloudflare Access 等で追加の認証層を重ねる
- [ ] `devices.json` を保存するホスト側ディレクトリを `chmod 0700` にする
- [ ] シャットダウン用アカウントには最小限の権限のみ付与する

LAN 内専用で使う場合でも、認証なしで動かすときは「LAN に接続している全員が WoL/シャットダウン操作可能」であることを認識してください。

## 機能

- **WoL マジックパケット送信**（UDP ブロードキャスト 255.255.255.255:9）
- **ping による生存監視**（自動巡回間隔は UI から設定可、0 で無効化）
- **リモートシャットダウン**（`net rpc shutdown` 経由、Windows のリモート RPC 設定が必要）
- **デバイス追加・編集・削除** と最終確認時刻の記録

## クイックスタート（Docker Compose）

```yaml
services:
  wol-tool:
    image: ghcr.io/iniwa/wol-tool-claude:latest
    container_name: wol-tool
    restart: unless-stopped
    network_mode: host  # WoL のブロードキャストに必要
    volumes:
      - /home/iniwa/docker/wol-tool/data:/data
    environment:
      - TZ=Asia/Tokyo
      - PORT=8080
      - AUTH_USER=admin       # ← 外部公開時は必須
      - AUTH_PASS=change-me   # ← 外部公開時は必須
```

```bash
mkdir -p /home/iniwa/docker/wol-tool/data
chmod 700 /home/iniwa/docker/wol-tool/data
docker compose up -d
```

ブラウザで `http://<host>:8080` にアクセス。

## 環境変数

| 変数 | 既定値 | 説明 |
|------|--------|------|
| `PORT` | `8080` | Listen ポート |
| `DATA_PATH` | `/data/devices.json` | デバイス情報の保存先 |
| `PING_INTERVAL` | `30` | 起動時の自動 ping 間隔（秒）。0 で無効 |
| `AUTH_USER` | （未設定） | Basic 認証ユーザー名。両方設定で認証有効 |
| `AUTH_PASS` | （未設定） | Basic 認証パスワード |
| `TZ` | UTC | タイムゾーン（コンテナ向け）|

## API

すべて JSON 形式。`AUTH_USER`/`AUTH_PASS` 設定時は Basic 認証必須。

| メソッド | パス | 説明 |
|---------|------|------|
| `GET` | `/api/devices` | 一覧取得（パスワードは含まれません）|
| `POST` | `/api/devices` | 追加（`name`, `mac` 必須）|
| `PUT` | `/api/devices/{id}` | 更新（`shutdown_pass` 空文字なら既存保持、`clear_shutdown_pass:true` で削除）|
| `DELETE` | `/api/devices/{id}` | 削除 |
| `POST` | `/api/devices/{id}/wake` | WoL マジックパケット送信 |
| `POST` | `/api/devices/{id}/shutdown` | リモートシャットダウン |
| `POST` | `/api/devices/{id}/ping` | 単体 ping |
| `POST` | `/api/ping/all` | 全台 ping |
| `GET`/`PUT` | `/api/config` | 自動 ping 間隔の取得・設定 |

## リモートシャットダウンの前提条件

シャットダウン対象の Windows PC で以下が必要です:

1. ファイアウォールで「Remote Service Management」を許可
2. `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System\LocalAccountTokenFilterPolicy = 1` を設定（リモート UAC 制限の回避）
3. ローカル管理者アカウントの認証情報

## ローカルビルド

```bash
go build -o wol-server .
./wol-server
```

または Docker:

```bash
docker build -t wol-tool .
docker run --rm --network host -v $(pwd)/data:/data wol-tool
```

## ライセンス

MIT — [LICENSE](LICENSE) を参照
