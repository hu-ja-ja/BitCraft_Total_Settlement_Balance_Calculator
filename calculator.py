from api_client import BitCraftAPIClient

class SettlementBalanceCalculator:
    def __init__(self, api_client: BitCraftAPIClient):
        self.api = api_client

    def calculate_total_balance(self, town_entity_id: str) -> int:
        town = self.api.get_town_detail(town_entity_id)
        treasury = int(town.get("treasury", "0"))
        members = self.api.get_town_members(town_entity_id)
        total_wallet = 0
        total_market = 0
        for member in members:
            player_id = member.get("playerEntityId")
            total_wallet += self.api.get_player_wallet(player_id)
            total_market += self.api.get_player_market_coins(player_id)
        return treasury + total_wallet + total_market
