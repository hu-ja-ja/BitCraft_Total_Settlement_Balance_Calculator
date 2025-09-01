from api_client import BitCraftAPIClient
from calculator import SettlementBalanceCalculator


def main():
    api = BitCraftAPIClient()
    calc = SettlementBalanceCalculator(api)
    town_name = input("町名を入力してください: ")
    towns = api.search_town(town_name)
    if not towns:
        print("該当する町が見つかりませんでした。")
        return
    print("候補:")
    for idx, town in enumerate(towns):
        print(f"{idx}: {town['name']} (ID: {town['entityId']})")
    sel = input("番号で町を選択してください: ")
    try:
        sel_idx = int(sel)
        town_entity_id = towns[sel_idx]["entityId"]
    except (ValueError, IndexError):
        print("選択が不正です。")
        return

    print(f"選択された町: {towns[sel_idx]['name']} (ID: {town_entity_id})")
    town_detail = api.get_town_detail(town_entity_id)
    treasury = int(town_detail.get("treasury", "0"))
    print(f"町の所持金 (treasury): {treasury}")

    members = api.get_town_members(town_entity_id)
    print(f"メンバー数: {len(members)}")
    for m in members:
        print(f"  - {m.get('userName', '')} (PlayerID: {m.get('playerEntityId', '')})")

    total_wallet = 0
    total_market = 0
    for m in members:
        player_id = m.get("playerEntityId")
        wallet = api.get_player_wallet(player_id)
        market = api.get_player_market_coins(player_id)
        print(f"{m.get('userName', '')} のWallet: {wallet}, Market: {market}")
        total_wallet += wallet
        total_market += market

    print(f"全メンバーWallet合計: {total_wallet}")
    print(f"全メンバーMarket合計: {total_market}")
    total = treasury + total_wallet + total_market
    print(f"総資産: {total}")

if __name__ == "__main__":
    main()
