package usecase

import (
	"context"
	"sort"

	"bitcrafttsbc/internal/api"
)

const HexiteItemID int64 = 828972621

type Progress struct {
	Done       int
	Total      int
	MemberName string
}

type TownMemberBalance struct {
	UserName string
	Wallet   int64
	Market   int64
	Total    int64
}

type TownResult struct {
	TownName     string
	TownEntityID string
	Treasury     int64
	Members      []TownMemberBalance
	TotalWallet  int64
	TotalMarket  int64
	Total        int64
}

type EmpireMemberHexite struct {
	UserName string
	Hexite   int64
}

type EmpireResult struct {
	EmpireName             string
	EmpireEntityID         string
	EmpireCurrencyTreasury int64
	Members                []EmpireMemberHexite
	TotalMemberHexite      int64
	Total                  int64
}

type Calculator struct {
	api *api.Client
}

func New(apiClient *api.Client) *Calculator {
	return &Calculator{api: apiClient}
}

func (c *Calculator) ComputeTownBalance(ctx context.Context, townEntityID string, onProgress func(Progress)) (TownResult, error) {
	detail, err := c.api.GetTownDetail(ctx, townEntityID)
	if err != nil {
		return TownResult{}, err
	}

	members, err := c.api.GetTownMembers(ctx, townEntityID)
	if err != nil {
		return TownResult{}, err
	}

	result := TownResult{
		TownName:     detail.Name,
		TownEntityID: townEntityID,
		Treasury:     detail.Treasury,
		Members:      make([]TownMemberBalance, 0, len(members)),
	}

	for idx, member := range members {
		wallet, err := c.api.GetPlayerWallet(ctx, member.PlayerEntityID)
		if err != nil {
			return TownResult{}, err
		}
		market, err := c.api.GetPlayerMarketCoins(ctx, member.PlayerEntityID)
		if err != nil {
			return TownResult{}, err
		}

		memberTotal := wallet + market
		result.Members = append(result.Members, TownMemberBalance{
			UserName: member.UserName,
			Wallet:   wallet,
			Market:   market,
			Total:    memberTotal,
		})
		result.TotalWallet += wallet
		result.TotalMarket += market

		if onProgress != nil {
			onProgress(Progress{Done: idx + 1, Total: len(members), MemberName: member.UserName})
		}
	}

	sort.Slice(result.Members, func(i, j int) bool {
		if result.Members[i].Total == result.Members[j].Total {
			return result.Members[i].UserName < result.Members[j].UserName
		}
		return result.Members[i].Total > result.Members[j].Total
	})

	result.Total = result.Treasury + result.TotalWallet + result.TotalMarket
	return result, nil
}

func (c *Calculator) ComputeEmpireHexite(ctx context.Context, empireEntityID string, includeRank9 bool, onProgress func(Progress)) (EmpireResult, error) {
	detail, err := c.api.GetEmpireDetail(ctx, empireEntityID)
	if err != nil {
		return EmpireResult{}, err
	}

	targetMembers := make([]api.Member, 0, len(detail.Members))
	for _, member := range detail.Members {
		if !includeRank9 && member.Rank == 9 {
			continue
		}
		targetMembers = append(targetMembers, member)
	}

	result := EmpireResult{
		EmpireName:             detail.Name,
		EmpireEntityID:         empireEntityID,
		EmpireCurrencyTreasury: detail.EmpireCurrencyTreasury,
		Members:                make([]EmpireMemberHexite, 0, len(targetMembers)),
	}

	for idx, member := range targetMembers {
		hexite, err := c.api.GetPlayerHexite(ctx, member.PlayerEntityID, HexiteItemID)
		if err != nil {
			return EmpireResult{}, err
		}
		result.Members = append(result.Members, EmpireMemberHexite{
			UserName: member.UserName,
			Hexite:   hexite,
		})
		result.TotalMemberHexite += hexite

		if onProgress != nil {
			onProgress(Progress{Done: idx + 1, Total: len(targetMembers), MemberName: member.UserName})
		}
	}

	sort.Slice(result.Members, func(i, j int) bool {
		if result.Members[i].Hexite == result.Members[j].Hexite {
			return result.Members[i].UserName < result.Members[j].UserName
		}
		return result.Members[i].Hexite > result.Members[j].Hexite
	})

	result.Total = result.TotalMemberHexite + result.EmpireCurrencyTreasury
	return result, nil
}
