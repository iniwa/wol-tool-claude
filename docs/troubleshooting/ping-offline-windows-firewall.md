# Ping がオフラインのまま（実機は起動中）— Windows ファイアウォール

## 症状
WoL ツール（Pi 上の Go）から Windows PC への生存確認 (ping) が、実機起動中でも「オフライン」のまま。Pi から直接 `ping` しても応答が返らない。

## 原因
PC 側 Windows Defender ファイアウォールの ICMP Echo 受信ルールが、**対象インターフェースのネットワークプロファイルでは無効**になっていた。

- 対象 IP のインターフェースのプロファイル: **Private**
- `File and Printer Sharing (Echo Request - ICMPv4-In)` の有効状態:
  - Public: 有効 ✅
  - **Private: 無効 ❌**（← 対象インターフェースはこちら）
  - Domain: 無効

ping を受けるプロファイル（Private）でルールが無効のため、ICMP Echo がすべて破棄され、ツールはオフライン判定になる。WoL ツール／Pi 側に問題はない。

## 切り分け手順（PowerShell, PC 側）
```powershell
# 対象 IP がどのインターフェース/プロファイルか
Get-NetIPAddress -AddressFamily IPv4 | Select IPAddress, InterfaceAlias
Get-NetConnectionProfile | Select InterfaceAlias, NetworkCategory

# ICMP Echo 受信ルールのプロファイル別有効状態
Get-NetFirewallRule -DisplayName "File and Printer Sharing (Echo Request - ICMPv4-In)" |
  Select Name, Profile, Enabled
```

## 対処
該当プロファイルのルールを有効化（要管理者権限）:
```powershell
# 例: Private プロファイル用ルール
Enable-NetFirewallRule -Name "FPS-ICMP4-ERQ-In_1"
```
※ ルール名（`FPS-ICMP4-ERQ-In*`）はプロファイルごとに分かれている。上の切り分けで対象プロファイルの `Name` を確認してから有効化する。

GUI の場合: コントロールパネル → Windows Defender ファイアウォール → 詳細設定 → 受信の規則 → 「ファイルとプリンターの共有 (エコー要求 - ICMPv4 受信)」の該当プロファイルを有効化。

## 確認
```bash
ssh iniwapi
ping -c 4 <対象IP>
```
応答が返ればツールの生存確認もオンラインになる。

## 再発時の注意
ネットワークプロファイルが Public/Private 間で変わったり、ファイアウォール設定がリセットされると再発する。まず「対象インターフェースのプロファイル」と「そのプロファイルで ICMP Echo 受信が有効か」を確認する。
