# BitCraft_Total_Settlement_Balance_Calculator

BitCraft Onlineの町総資産計算ツール（TSBC）
町の所持金・メンバーのWallet・マーケット資産を合算し、総資産を算出します。

## 概要

- 町名で検索し、候補から町を選択
- APIから町の詳細・メンバー情報・各メンバーのWallet/Market資産を取得
- すべての金額を合算し、総資産を表示

## 使い方

1. 必要なPythonパッケージをインストール
   ```
   pip install requests
   ```
2. CLIで実行
   ```
   python main.py
   ```
3. 町名を入力
   ```
   町名を入力してください: musashi
   ```
4. 候補から町を選択
   ```
   0: Musashi T7 (ID: 576460752386460204)
   番号で町を選択してください: 0
   ```
5. 総資産が表示されます

## 処理フロー

1. 町名でAPI検索
   `GET /api/claims?q={町名}`
2. 町候補リスト表示・選択
3. 選択した町の詳細取得
   `GET /api/claims/{entityId}`
4. メンバー一覧取得
   `GET /api/claims/{entityId}/members`
5. 各メンバーのWallet取得
   `GET /api/players/{playerId}/inventories`
6. 各メンバーのMarket資産取得
   `GET /api/market/player/{playerId}`
7. treasury + 全メンバーWallet + 全メンバーMarket を合算

## API仕様（抜粋）

- 町検索: `/api/claims?q={町名}`
- 町詳細: `/api/claims/{entityId}`
- メンバー一覧: `/api/claims/{entityId}/members`
- Wallet: `/api/players/{playerId}/inventories`
  - `"inventoryName": "Wallet"` の `"pockets"[0]["contents"]["quantity"]`
- Market: `/api/market/player/{playerId}`
  - `sellOrders`/`buyOrders` の `"storedCoins"` 合計

## ライセンス

MIT License
Copyright (c) 2025 HU_JA_JA

---

ご要望・バグ報告は [Issues](https://github.com/your-repo/issues) へどうぞ。
