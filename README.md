# BitCraft_Total_Settlement_Balance_Calculator

BitCraft Online の町総資産計算ツール（TSBC）です。
町の所持金（treasury）・メンバーの Wallet・マーケット資産を合算して "総資産" を算出します。

## 概要

- 町名で検索して候補を表示
- 候補から町を選択
- 選択した町の詳細・メンバー情報・各メンバーの Wallet / Market を取得
- 合算して総資産を表示

## ダウンロード・実行方法（exe版）

1. `main.exe` をダウンロード
1. ダブルクリックで実行
1. コマンドプロンプトが開くので画面の指示に従って操作してください

> 注意: EXE は Windows 向けです。Python が不要でそのまま実行できます。

## 使い方（CLI版）

1. 必要なパッケージをインストール

```bash
pip install requests
```

1. CLIで実行

```bash
python main.py
```

1. 町名を入力して候補から選択します

```text
町名を入力してください: musashi
```

```text
0: Musashi T7 (ID: 576460752386460204)
番号で町を選択してください: 0
```

1. 実行完了後は下部に「Enterキーで終了します」と表示されるので、Enterキーを押してウィンドウを閉じてください。

## 処理フロー

1. 町名で検索: `GET /api/claims?q={町名}`
2. 町詳細: `GET /api/claims/{entityId}`
3. メンバー一覧: `GET /api/claims/{entityId}/members`
4. 各メンバーの Wallet: `GET /api/players/{playerId}/inventories` (`inventoryName == "Wallet"` を探し、`pockets[0]["contents"]["quantity"]` を使用)
5. 各プレイヤーの Market: `GET /api/market/player/{playerId}` (`sellOrders` と `buyOrders` の `storedCoins` を合算)
6. 最終的に `treasury + Σ(wallet) + Σ(market)` を表示

## API仕様（抜粋）

- 町検索: `/api/claims?q={町名}`
- 町詳細: `/api/claims/{entityId}`
- メンバー一覧: `/api/claims/{entityId}/members`
- Wallet: `/api/players/{playerId}/inventories`
  - `"inventoryName": "Wallet"` の `"pockets"[0]["contents"]["quantity"]`
- Market: `/api/market/player/{playerId}`
  - `sellOrders`/`buyOrders` の `"storedCoins"` 合計

## クライアント識別ヘッダー

このクライアントでは API へアクセスする際にアプリを識別するための HTTP ヘッダーを送信します。

- `User-Agent`: 例 `BitJita (BitCraft_TSBC)`
- `x-app-identifier`: デフォルト `BitCraft_TSBC`

アプリ識別子を変更したい場合は、`api_client.BitCraftAPIClient` のコンストラクタに文字列を渡してください。

```py
from api_client import BitCraftAPIClient
api = BitCraftAPIClient(app_identifier="YourName_YourApp")
```

## ライセンス

MIT License

---

ご要望・バグ報告は Issues へお願いします。

