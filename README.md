# BitCraft_Total_Settlement_Balance_Calculator

TSBC outputs the total amount of money in the town, combining funds held by members and those stored in the market.

## なにをしてるか

<https://bitjita.com/api/claims?q=musashi>
musashiを含むクレームを検索、出力されたclaimsを一旦出力し、ユーザーが町を選択する。
選択した町のentityIdを次のAPIで使用する。

一部抜粋
```json
{
  "claims": [
    {
      "entityId": "576460752386460204",
      "ownerPlayerEntityId": "576460752315732635",
      "ownerBuildingEntityId": "576460752386460183",
      "name": "Musashi T7",
      "neutral": false,
      "regionId": 8,
      "regionName": "Draxionne",
      "createdAt": "2025-08-29 21:22:39.942428+00",
      "updatedAt": "2025-08-31 13:11:55.449196+00",
      "supplies": 32170,
      "buildingMaintenance": 0,
      "numTiles": 6398,
      "locationX": 12061,
      "locationZ": 17397,
      "locationDimension": 1,
      "treasury": "92060",
      "learned": [
        1,
        200,
        300,
        400,
        1826500486,
        500,
        1161565644,
        477458683,
        2099304788,
        956489539,
        600,
        1157053499,
        473237479,
        688169271,
        2136582402,
        10000,
        700
      ],
      "researching": 0,
      "startTimestamp": null,
      "tier": 7
    }
  ],
  "count": "1"
}
```

<https://bitjita.com/api/claims/576460752386460204>
treasuryが町の所持金、保存しておく。

一部抜粋
```json
{
  "claim": {
    "entityId": "576460752386460204",
    "ownerPlayerEntityId": "576460752315732635",
    "ownerBuildingEntityId": "576460752386460183",
    "name": "Musashi T7",
    "neutral": false,
    "regionId": 8,
    "regionName": "Draxionne",
    "supplies": 32170,
    "buildingMaintenance": 0,
    "numTiles": 6398,
    "locationX": 12061,
    "locationZ": 17397,
    "locationDimension": 1,
    "treasury": "92060",
    "ownerPlayerUsername": "Akagi",
    "techResearching": 0,
    "techStartTimestamp": "0",
    "tileCost": 0.035,
    "upkeepCost": 223.93000000000004,
    "suppliesRunOut": 1757247484353,
    "tier": 7,
```

<https://bitjita.com/api/claims/576460752386460204/members>
各プレイヤーのentityIdを取得、次のAPIで使用する。

一部抜粋
```json
{
  "members": [
    {
      "entityId": "576460752386461022",
      "claimEntityId": "576460752386460204",
      "playerEntityId": "576460752315732635",
      "userName": "Akagi",
      "inventoryPermission": 1,
      "buildPermission": 1,
      "officerPermission": 1,
      "coOwnerPermission": 1,
      "createdAt": "2025-08-29 21:22:44.59334+00",
      "updatedAt": "2025-08-29 21:22:44.59334+00",
      "lastLoginTimestamp": "2025-09-01 06:44:50+00"
    },
    {
      "entityId": "576460752387946641",
      "claimEntityId": "576460752386460204",
      "playerEntityId": "288230376164052258",
      "userName": "Neet",
      "inventoryPermission": 1,
      "buildPermission": 1,
      "officerPermission": 1,
      "coOwnerPermission": 1,
      "createdAt": "2025-08-29 21:22:44.59334+00",
      "updatedAt": "2025-08-29 21:22:44.59334+00",
      "lastLoginTimestamp": "2025-09-01 02:48:58+00"
    },
```

<https://bitjita.com/api/players/216172782125155383/inventories>
プレイヤーのWalletを取得する。
contents:{quantity}が所持金だが、jsonがかなり長いので"inventoryName": "Wallet"の部分を探す必要あり

一部抜粋
```json
    {
      "entityId": "216172782125155385",
      "playerOwnerEntityId": "0",
      "ownerEntityId": "216172782125155383",
      "pockets": [
        {
          "locked": false,
          "volume": 6000,
          "contents": {
            "itemId": 1,
            "itemType": 0,
            "quantity": 21325
          }
        }
      ],
      "inventoryIndex": 2,
      "cargoIndex": 1,
      "buildingName": null,
      "claimEntityId": null,
      "claimName": null,
      "claimLocationX": null,
      "claimLocationZ": null,
      "claimLocationDimension": null,
      "regionId": 8,
      "inventoryName": "Wallet"
    },
```

<https://bitjita.com/api/market/player/216172782125155383>
各プレイヤーのマーケットにあるお金を取得する。
storedCoinsがマーケットにあるお金で、sellOrdersとbuyOrdersすべてのstoredCoinsを合算する。

一部抜粋
```json
{
  "playerId": "216172782125155383",
  "playerUsername": "HUJAJA",
  "sellOrders": [
    {
      "entityId": "576460752371162763",
      "ownerEntityId": "216172782125155383",
      "claimEntityId": "576460752316505261",
      "itemId": 1682637898,
      "itemType": 0,
      "priceThreshold": "100",
      "quantity": "1",
      "timestamp": "1753690529393639",
      "storedCoins": "0",
      "createdAt": "2025-08-29 21:22:41.664381+00",
      "updatedAt": "2025-08-29 21:22:41.664381+00",
      "itemName": "Pyrelite Plated Belt",
      "itemDescription": "",
      "itemTier": 2,
      "itemTag": "Metal Armor",
      "itemRarity": 1,
      "itemRarityStr": "Common",
      "iconAssetName": "GeneratedIcons/Other/GeneratedIcons/Items/MetalBeltT1",
      "claimName": "Tokyo",
      "ownerUsername": "HUJAJA",
      "regionId": 8,
      "regionName": "Draxionne"
    },
      ],
  "buyOrders": []
}
```