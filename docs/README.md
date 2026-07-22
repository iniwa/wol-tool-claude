# docs

設計判断・トラブルシュート手順の蓄積先。作業開始時にまずここを確認する。

## improvements.md
コード調査で洗い出した改善候補のチェックリスト（優先度付き）。

- [プログラム改善チェックリスト](improvements.md) — 2026-07-08 調査。データ保全・安定性 / セキュリティ / 品質の 10 項目

## decisions/
設計判断と背景（durable な設計履歴）。

- [2026-05-08 Decision Log Scope](decisions/2026-05-08-decision-log-scope.md) — `AGENTS.md` を短く保ち、長い背景・却下案は `docs/decisions/` に置く方針

## troubleshooting/
再発しうる不具合の切り分け・対処手順。

- [Ping がオフラインのまま（Windows ファイアウォール）](troubleshooting/ping-offline-windows-firewall.md) — 実機起動中なのに ping 応答が返らない場合の ICMP / ネットワークプロファイル切り分け
