# GoFyneInstaller

ウィザード形式のインストーラを、Go言語とFyneライブラリを使用して作成するサンプルです。

## 📋 目次

- [概要](#概要)
- [機能](#機能)
- [インストール](#インストール)
- [インストーラ作成手順](#インストーラ作成手順)
- [YAML設定ファイル仕様](#yaml設定ファイル仕様)

---

## 概要

**GoFyneInstaller**は、Go言語で開発ビルド可能なインストーラウィザードを作成する試みです。  
Go言語で作成されたプログラムは実行にランタイムが不要なのでインストール動作にも適するでしょう。  
また、FyneはGo向けのGUIライブラリで、Goで実行できるので他にランタイムが不要です。

**特徴：**
- ✅ **インストールウィザードのUI** — ウェルカム画面から完了画面まで5ステップのウィザード
- ✅ **自己完結型** — setup.exeに全てのファイルとスクリプトが埋め込まれている
- ✅ **YAML設定ベース** — JSON/YAML形式で簡単にカスタマイズ可能
- ✅ **Windowsサービス対応** — インストール時にサービスを登録可能
- ✅ **完全なアンインストール** — `setup.exe --uninstall`で削除可能
- ✅ **進捗表示** — ファイルコピーの進度をリアルタイム表示
- ✅ **管理者権限対応** — Program Filesへのインストール対応，管理者昇格ダイアログ表示

---

## 機能

### インストーラ画面フロー

```
1️⃣  Welcome（ウェルカム）
    ├─ アプリケーション名・説明表示
    ├─ 「続行」を確認
    └─ Next / Cancel ボタン

2️⃣  Folder Selection（インストール先選択）
    ├─ デフォルトパス表示（C:\Program Files\...）
    ├─ テキストボックスで手動編集可能
    ├─ "Browse" ボタンでフォルダ選択ダイアログ
    └─ Previous / Next / Cancel ボタン

3️⃣  Confirm（インストール確認）
    ├─ アプリ名・バージョン・パブリッシャー表示
    ├─ インストール先フォルダを読み取り専用で表示
    └─ Previous / Next / Cancel ボタン

4️⃣  Installing（インストール中）
    ├─ プログレスバーでファイルコピー進捗表示
    ├─ ステータスメッセージ表示
    ├─ インストールログ表示
    ├─ ボタンなし（完全自動）
    └─ 完了後、次ステップへ自動遷移

5️⃣  Complete（完了）
    ├─ 「インストール完了」メッセージ表示
    └─ Finish ボタンでウィザード終了
```

### アンインストーラ画面フロー

```
1️⃣  Confirm（アンインストール確認）
    ├─ 「削除してもよいか？」確認メッセージ
    └─ Previous / Next / Cancel ボタン

2️⃣  Uninstalling（アンインストール中）
    ├─ プログレスバーで削除進捗表示
    └─ ログ表示

3️⃣  Complete（完了）
    ├─ 「アンインストール完了」メッセージ表示
    └─ Finish でウィザード終了
```

### 管理者権限対応

**GoFyneInstaller は Program Files へのインストールに対応するため、自動的に管理者権限の昇格を行います。**

#### 動作フロー

1. **権限チェック** — アプリケーション起動時に管理者権限を確認
2. **権限がない場合** — ユーザーに通知ダイアログを表示
3. **自動昇格** — ユーザーの許可を得て、管理者権限で再起動
4. **インストール実行** — 管理者権限で Program Files へのアクセス可能

#### マニフェスト埋め込み

`manifest.xml` をリソースとして埋め込むことで、Windows がこのアプリケーションに管理者権限が必要であることを認識します。

```xml
<requestedExecutionLevel level="requireAdministrator" uiAccess="false"/>
```

#### コード実装（admin パッケージ）

- `IsAdminAlt()` — 現在のプロセスが管理者権限で実行されているか確認
- `RelaunchAsAdmin()` — 管理者権限で再起動（ShellExecute "runas" を使用）
- `MarkFileForDeletionOnReboot()` — ファイルを PC 再起動後に削除するようマーク（Windows API MoveFileEx 使用）

---

## アンインストール実行ファイルの再起動後削除

**GoFyneInstaller は、アンインストール後に実行ファイル自体が PC 再起動後に自動削除されるようマークします。**

#### 動作フロー

1. **アンインストール実行** — ユーザーが `setup.exe --uninstall` を実行
2. **クリーンアップ処理** — ファイル・Registry・サービスを削除
3. **実行ファイルマーク** — アンインストール実行ファイル自体を削除予定とマーク
4. **PC 再起動** — ユーザーが PC を再起動
5. **自動削除** — 再起動時に Windows が削除対象ファイルを自動削除

#### 実装仕様

- Windows API の `MoveFileEx()` を使用
- `MOVEFILE_DELAY_UNTIL_REBOOT` フラグを指定
- キャリー言語化（異なるインストーラバージョンでも対応）

---

## インストール

### 前提条件

- **Go 1.16以上** — Go embed 機能が必要
- **Windows 7以上** — Fyne v2が対応するOSバージョン
- **git** — リポジトリクローン用

### リポジトリのクローンと準備

```bash
# リポジトリクローン
git clone https://github.com/maishiro/GoFyneInstaller.git
cd GoFyneInstaller

# 依存パッケージ自動ダウンロード
go mod tidy
```

---

## インストーラ作成手順

### ステップ1️⃣: YAML設定ファイル作成

`assets/installer-config.yaml` を編集して、以下の項目を設定：

```yaml
metadata:
  name: "MyApplication"                    # アプリケーション名
  version: "1.0.0"                        # バージョン
  description: "My awesome app"           # 説明文
  publisher: "My Company"                 # パブリッシャー

files:                                     # インストールするファイル
  - source: "myapp.exe"
    target: "myapp.exe"
    type: "executable"
  - source: "config.ini"
    target: "config.ini"
    type: "config"

install:                                   # インストール時の処理
  commands:
    - type: "copy_file"
      params:
        source: "myapp.exe"
        destination: "{INSTALL_DIR}\\myapp.exe"
    - type: "register_service"            # Windowsサービス登録
      params:
        name: "MyAppService"
        binary_path: "{INSTALL_DIR}\\myapp.exe"
        display_name: "My Application Service"

uninstall:                                 # アンインストール時の処理
  commands:
    - type: "unregister_service"
      params:
        name: "MyAppService"
    - type: "delete_folder"
      params:
        path: "{INSTALL_DIR}"
```

### ステップ2️⃣: アセットファイル配置

`assets/` フォルダに以下を配置：

```
assets/
├── installer-config.yaml   ← 上記で作成した設定ファイル
├── myapp.exe               ← 実行ファイル
├── config.ini              ← 設定ファイル
├── readme.txt              ← ドキュメント
└── ... その他のファイル
```

**重要:** `installer-config.yaml` に記述した `files.[].source` は、実際のファイル名と一致する必要があります。

### ステップ3️⃣: ビルド

```bash
# 前提: assets/ フォルダにファイルが配置されていること
build.bat

# ビルド完了 → setup.exe が生成される
```

### ステップ4️⃣: テスト

```bash
# インストーラー実行
./setup.exe

# 各ステップで以下を確認:
# 1. アプリ名・説明が正しく表示されるか
# 2. デフォルトインストール先が設定した値か
# 3. ファイルが指定フォルダにコピーされるか
# 4. Windowsサービスが登録されたか（レジストリ確認）

# アンインストーラーテスト
setup.exe --uninstall
```

### ステップ5️⃣: 配布

`setup.exe` をユーザーに配布します。単一ファイルで全てを含んでいます。

---

## YAML設定ファイル仕様

### メタデータセクション

```yaml
metadata:
  name: string            # アプリケーション名（表示用）
  version: string         # バージョン（例: "1.0.0"）
  description: string     # 説明文（ウェルカム画面に表示）
  publisher: string       # パブリッシャー名
```

### ファイルセクション

インストールするファイルを指定：

```yaml
files:
  - source: string        # assets/ フォルダ内のファイル名
    target: string        # インストール先での名前（通常 source と同じ）
    type: string          # ファイルタイプ（"executable", "config", "documentation" など）
```

### インストール/アンインストールコマンド

#### 利用可能なコマンドタイプ

| コマンド | 説明 | パラメータ |
|---------|------|-----------|
| `copy_file` | ファイルコピー | `source`, `destination` |
| `delete_folder` | フォルダ削除 | `path` |
| `register_service` | Windowsサービス登録 | `name`, `binary_path`, `display_name` |
| `unregister_service` | Windowsサービス削除 | `name` |
| `start_service` | サービス開始 | `name` |
| `stop_service` | サービス停止 | `name` |
| `run_command` | 任意のコマンド実行 | `command`, `args` |

#### 変数置換

以下の変数はコマンド実行時に自動置換されます：

| 変数 | 説明 |
|-----|------|
| `{INSTALL_DIR}` | ユーザーが選択したインストール先ディレクトリ |
| `{PROGRAM_FILES}` | `C:\Program Files` |
| `{WINDOWS}` | Windowsシステムフォルダ（%WINDIR%） |

#### コマンド例

**ファイルコピー:**
```yaml
- type: "copy_file"
  params:
    source: "app.exe"
    destination: "{INSTALL_DIR}\\app.exe"
```

**Windowsサービス登録:**
```yaml
- type: "register_service"
  params:
    name: "MyService"
    binary_path: "{INSTALL_DIR}\\app.exe"
    display_name: "My Application"
```

**フォルダ削除（アンインストール時）:**
```yaml
- type: "delete_folder"
  params:
    path: "{INSTALL_DIR}"
```

**任意コマンド実行:**
```yaml
- type: "run_command"
  params:
    command: "cmd"
    args: ["/c", "echo Installation complete"]
```

---

## サンプル設定ファイル

### 最小構成

```yaml
metadata:
  name: "Simple App"
  version: "0.1.0"
  description: "A simple application"
  publisher: "Example Corp"

files:
  - source: "app.exe"
    target: "app.exe"
    type: "executable"

install:
  commands:
    - type: "copy_file"
      params:
        source: "app.exe"
        destination: "{INSTALL_DIR}\\app.exe"

uninstall:
  commands:
    - type: "delete_folder"
      params:
        path: "{INSTALL_DIR}"
```

### サービス登録を含む構成

```yaml
metadata:
  name: "Service Application"
  version: "1.0.0"
  description: "A Windows service application"
  publisher: "Example Corp"

files:
  - source: "service.exe"
    target: "service.exe"
    type: "executable"
  - source: "config.yaml"
    target: "config.yaml"
    type: "config"

install:
  commands:
    - type: "copy_file"
      params:
        source: "service.exe"
        destination: "{INSTALL_DIR}\\service.exe"
    - type: "copy_file"
      params:
        source: "config.yaml"
        destination: "{INSTALL_DIR}\\config.yaml"
    - type: "register_service"
      params:
        name: "MyService"
        binary_path: "{INSTALL_DIR}\\service.exe"
        display_name: "My Custom Service"

uninstall:
  commands:
    - type: "stop_service"
      params:
        name: "MyService"
    - type: "unregister_service"
      params:
        name: "MyService"
    - type: "delete_folder"
      params:
        path: "{INSTALL_DIR}"
```

---

## 技術スタック

| 技術 | バージョン | 用途 |
|------|-----------|------|
| Go | 1.16+ | メイン言語 |
| Fyne | v2.7.3 | GUI フレームワーク |
| YAML | via gopkg.in/yaml.v3 | 設定ファイル形式 |
| Go embed | 標準 | アセット埋め込み |
|  |  |  |

---

## ライセンス

このプロジェクトは **MITライセンス** の下で公開されています。詳細は [LICENSE](LICENSE) ファイルをご覧ください。

### 第三者ライセンス

本ソフトウェアには、以下の第三者コンポーネントが含まれています。

* **[Fyne](https://fyne.io/)** - [BSD 3-Clause License](https://opensource.org/licenses/BSD-3-Clause) の下で提供されています。
* **[yaml.v3](https://github.com/go-yaml/yaml)** - Apache License 2.0
* **[rsrc](https://github.com/akavel/rsrc)** - [MIT License](https://opensource.org/licenses/MIT)の下で提供されています。

各ライブラリの著作権表示およびライセンス全文については、[THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md) をご確認ください。

---

## 参考リンク

- [Fyne Documentation](https://fyne.io/doc/)
- [Go embed Package](https://pkg.go.dev/embed)
- [YAML 仕様](https://yaml.org/)
- [Windows Service Management (sc.exe)](https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/sc-create)
