# BitCraft Total Settlement Balance Calculator

BitCraft Online 向けの集計ツールです。

Go + Bubbletea 版では次の2モードを実装しています。

1. 町の総資産
2. 国のヘキサイト総量

## プロジェクト構成

```text
.
├─ .github/
│  └─ workflows/
│     └─ release-go.yml
├─ cmd/
│  └─ bitcraft-tsbc/
│     └─ main.go
├─ internal/
│  ├─ api/
│  │  └─ client.go
│  ├─ tui/
│  │  └─ model.go
│  └─ usecase/
│     └─ calculator.go
├─ scripts/
│  └─ build-go.ps1
├─ go.mod
├─ go.sum
└─ README.md
```

役割:

1. `cmd/bitcraft-tsbc`: エントリーポイント
2. `internal/api`: BitCraft API クライアント
3. `internal/usecase`: 集計ロジック
4. `internal/tui`: Bubbletea 画面遷移と表示
5. `scripts/build-go.ps1`: ローカル/CI 共通のビルドスクリプト

## 機能

### 町の総資産モード

1. 町検索: `GET /api/claims?q={town_name}`
2. 町詳細: `GET /api/claims/{entityId}` から `treasury`
3. 町メンバー: `GET /api/claims/{entityId}/members`
4. 各メンバーの Wallet: `GET /api/players/{playerId}/inventories` から `inventoryName == "Wallet"` の `pockets[0].contents.quantity`
5. 各メンバーの Market: `GET /api/market/player/{playerId}` の `sellOrders` と `buyOrders` の `storedCoins` 合計
6. 総資産: `treasury + Σ(wallet) + Σ(market)`

### 国のヘキサイト総量モード

1. 帝国検索: `GET /api/empires?q={name}`
2. 帝国詳細: `GET /api/empires/{id}` から `members` と `empireCurrencyTreasury`
3. 各メンバーの inventories: `GET /api/players/{playerId}/inventories`
4. `inventories` 全体の `pockets` を走査し、`itemId == 828972621` の `quantity` を抽出して合算
5. 帝国候補選択後に `rank == 9`（一般市民）を含めるか選択
6. 総量: `Σ(member_hexite) + empireCurrencyTreasury`

## 実行方法（Go版 / 推奨）

要件:

1. Go 1.24 以上

ローカル実行:

```bash
go mod tidy
go run ./cmd/bitcraft-tsbc
```

ビルド（Windows x64）:

```powershell
./scripts/build-go.ps1
```

テストを省略してビルド:

```powershell
./scripts/build-go.ps1 -SkipTest
```

### 操作キー（Bubbletea）

1. `↑` / `↓`: カーソル移動
2. `Enter`: 決定
3. `Esc`: 戻る
4. `r`: 結果画面から再検索
5. `m`: モード選択へ戻る
6. `q` または `Ctrl+C`: 終了

## 配布設計（Windows x64）

タグ起点での配布を想定しています。

1. `v1.0.0` 形式でタグ作成
2. GitHub Actions が `scripts/build-go.ps1` で `dist/bitcraft-tsbc-windows-amd64.exe` をビルド
3. SHA256 を生成
4. Release に exe と checksum を添付

ワークフロー定義は `.github/workflows/release-go.yml` を参照してください。

## ライセンス

MIT License
