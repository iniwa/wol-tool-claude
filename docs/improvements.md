# プログラム改善チェックリスト

コードベースを調査して洗い出した改善候補の一覧（2026-07-08 調査）。

**運用方法**: 着手したい項目にチェック `[x]` を入れる → Codex が handoff
（`docs/handoffs/`）を作成し、Claude Code（Sonnet 実務・auto モード）が実装する。
handoff を挟むまでもない小粒な項目は Claude Code に直接依頼してもよい。
実装完了した項目は「完了アーカイブ」へ移動する。

- 機能追加・未検証項目はこのファイルの対象外（`docs/issues.md` 等で管理）。
- 優先度: **高** = 稼働中の安定性・データ保全に直結 / **中** = 保守性・性能 / **低** = 任意。

---

## 1. データ保全・安定性（Go バックエンド）

- [ ] **【高】devices.json 読み込み失敗時に既存データを空で上書きしないようにする**
  - 現状: `main.go:75-82` の `NewStore` が `json.Unmarshal` のエラーを無視している。破損した `devices.json` で起動すると空リストとして立ち上がり、ping ループの `Save`（`main.go:240`、デフォルト 30 秒間隔）が旧データを空で上書きして消失する。
  - 対応案: Unmarshal 失敗時は `log.Fatal` で起動を中止する（または破損ファイルを `.bak` に退避して警告ログ）。正常に読めた場合の挙動は不変。
  - 制約: `/data/devices.json` のフォーマット互換を維持。

- [ ] **【中】devices.json の保存をアトミック化し、エラーをログする**
  - 現状: `main.go:85-90` の `Save` が `MarshalIndent` / `os.WriteFile` のエラーを無ログで破棄。`O_TRUNC` 直書きのため書き込み中クラッシュで破損する。さらに `RLock` のため手動 `/api/ping/all` と自動ループの `Save` が並行実行され得る。
  - 対応案: 一時ファイルに書いて `os.Rename` で置換、エラーは `log.Printf`。`Save` を排他ロックにする。保存内容は不変。

- [ ] **【中】Device フィールドのデータ競合を解消する**
  - 現状: `pingAllDevices`（`main.go:222-230`）と `pingOneDevice`（`main.go:246`）が `d.IP` をロック外で読む一方、`Update`（`main.go:128-141`）が同フィールドをロック内で書く。`go run -race` で検出されるレベルの競合。
  - 対応案: ping 対象の ID・IP をロック内でコピーしたスナップショットに対して ping し、ロック外での `*Device` フィールド参照をなくす。挙動不変。

- [ ] **【中】shutdownWindows にタイムアウトを設ける**
  - 現状: `main.go:287-289` の `cmd.CombinedOutput()` に期限がなく、`net rpc shutdown` がハングすると HTTP ハンドラが無期限にブロックする（RPC/HTTP 呼び出しで唯一タイムアウトのない箇所）。
  - 対応案: `exec.CommandContext` + `context.WithTimeout`（15〜30 秒）に置き換え、タイムアウト時は 500 で理由を返す。正常系の挙動は不変。

- [ ] **【中】状態変化がないときの devices.json 書き込みを抑制する**
  - 現状: `main.go:240` で ping 巡回のたびに必ず `Save`。デフォルト 30 秒間隔だと約 2,880 回/日の全量書き込みが SSD に発生する（Online/LastSeen が変わらなくても書く）。
  - 対応案: 巡回中に `Online` / `LastSeen` が変化した場合のみ `Save` する。保存されるデータは不変。

- [ ] **【低】http.Server にタイムアウトを設定する**
  - 現状: `main.go:542` で `http.ListenAndServe` を直接使用しており `ReadHeaderTimeout` 等が未設定。
  - 対応案: `http.Server{ReadHeaderTimeout: 5 * time.Second, ...}` に置き換え。LAN 内前提のため優先度低。挙動不変。

## 2. セキュリティ（LAN 前提の範囲で）

- [ ] **【高】デバイス名のエスケープ不足による stored XSS / UI 破壊を修正する**
  - 現状: `static/index.html:584-586` の `esc()` が `'` をエスケープしないまま、`onclick="wake('${d.id}', '${esc(d.name)}')"`（`index.html:575,576,578`）でシングルクォート区切りの JS 文字列に埋め込んでいる。名前に `'` を含むデバイス（例: `John's PC`）で WAKE/SHUTDOWN/削除ボタンが壊れ、任意 JS の保存型注入も可能。
  - 対応案: inline `onclick` をやめ `data-id` 属性 + イベントデリゲーションにする（最小修正なら `esc()` に `'` → `&#39;` を追加）。表示・操作は同一。
  - 制約: プレーン JS 維持（フレームワーク不使用）。

- [ ] **【中】wake / shutdown / ping への CSRF 対策を追加する**
  - 現状: `main.go:402-462` の wake/shutdown/ping ハンドラはボディもヘッダも検証しないため、クロスサイトの単純 form POST がそのまま実行される（CORS ヘッダ不付与はレスポンスの読み取りを防ぐだけで、送信自体は防がない）。
  - 対応案: ミドルウェアで `Sec-Fetch-Site` が `cross-site` のリクエストを拒否（ヘッダ不在は許可）、または `Origin` と `Host` の一致チェック。同一オリジンの既存 UI は挙動不変。
  - 制約: 認証なし LAN 運用・curl 等の非ブラウザクライアントを壊さないこと。

## 3. 品質・体裁

- [ ] **【低】gofmt を適用し main.go の改行コードを LF に統一する**
  - 現状: `gofmt -l .` が `main.go` を報告（改行が CRLF、`main.go:34-43` の構造体タグ整列ずれ等）。
  - 対応案: `gofmt -w main.go` を実行し、`.gitattributes` に `*.go text eol=lf` を追加。挙動不変の機械的変更。

- [ ] **【低】純粋ロジックに最小ユニットテストを追加する**
  - 現状: `*_test.go` が 0 件（カバレッジ 0%）。`Store` CRUD（`main.go:92-163`）、`validIP`/`validMAC`（`main.go:167-177`）、マジックパケット構築（`main.go:181-202`）はテスト容易な純粋ロジック。
  - 対応案: `main_test.go` を追加し標準 `testing` のみで検証。依存追加・CI 変更なし。

---

## 完了アーカイブ

（まだなし）
