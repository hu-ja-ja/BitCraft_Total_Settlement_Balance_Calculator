import requests

BASE_URL = "https://bitjita.com/api"

class BitCraftAPIClient:
    def search_town(self, town_name: str) -> list:
        resp = requests.get(f"{BASE_URL}/claims", params={"q": town_name})
        resp.raise_for_status()
        return resp.json().get("claims", [])

    def get_town_detail(self, entity_id: str) -> dict:
        resp = requests.get(f"{BASE_URL}/claims/{entity_id}")
        resp.raise_for_status()
        return resp.json().get("claim", {})

    def get_town_members(self, entity_id: str) -> list:
        resp = requests.get(f"{BASE_URL}/claims/{entity_id}/members")
        resp.raise_for_status()
        return resp.json().get("members", [])

    def get_player_wallet(self, player_id: str) -> int:
        resp = requests.get(f"{BASE_URL}/players/{player_id}/inventories")
        resp.raise_for_status()
        data = resp.json()
        # inventoriesがリストか、辞書の"inventories"キーか両方対応
        if isinstance(data, dict):
            inventories = data.get("inventories", [])
        else:
            inventories = data
        for inv in inventories:
            if isinstance(inv, dict) and inv.get("inventoryName") == "Wallet":
                pockets = inv.get("pockets", [])
                if pockets:
                    return pockets[0].get("contents", {}).get("quantity", 0)
        return 0

    def get_player_market_coins(self, player_id: str) -> int:
        resp = requests.get(f"{BASE_URL}/market/player/{player_id}")
        resp.raise_for_status()
        data = resp.json()
        total = 0
        for order in data.get("sellOrders", []):
            total += int(order.get("storedCoins", "0"))
        for order in data.get("buyOrders", []):
            total += int(order.get("storedCoins", "0"))
        return total
