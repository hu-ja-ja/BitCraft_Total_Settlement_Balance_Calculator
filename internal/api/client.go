package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://bitjita.com/api"
)

type Candidate struct {
	EntityID string
	Name     string
}

type Member struct {
	PlayerEntityID string
	UserName       string
	Rank           int
}

type TownDetail struct {
	EntityID string
	Name     string
	Treasury int64
}

type EmpireDetail struct {
	EntityID              string
	Name                  string
	EmpireCurrencyTreasury int64
	Members               []Member
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
}

func New(appIdentifier string) *Client {
	if strings.TrimSpace(appIdentifier) == "" {
		appIdentifier = "BitCraft_TSBC_Go (discord: hu_ja_ja_)"
	}

	headers := make(http.Header)
	headers.Set("User-Agent", fmt.Sprintf("(%s)", appIdentifier))
	headers.Set("x-app-identifier", appIdentifier)

	return &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		headers: headers,
	}
}

func (c *Client) SearchTowns(ctx context.Context, query string) ([]Candidate, error) {
	var payload struct {
		Claims []struct {
			EntityID string `json:"entityId"`
			Name     string `json:"name"`
		} `json:"claims"`
	}
	if err := c.getJSON(ctx, "/claims", map[string]string{"q": query}, &payload); err != nil {
		return nil, err
	}

	candidates := make([]Candidate, 0, len(payload.Claims))
	for _, claim := range payload.Claims {
		candidates = append(candidates, Candidate{EntityID: claim.EntityID, Name: claim.Name})
	}
	return candidates, nil
}

func (c *Client) SearchEmpires(ctx context.Context, query string) ([]Candidate, error) {
	var payload struct {
		Empires []struct {
			EntityID string `json:"entityId"`
			Name     string `json:"name"`
		} `json:"empires"`
	}
	if err := c.getJSON(ctx, "/empires", map[string]string{"q": query}, &payload); err != nil {
		return nil, err
	}

	candidates := make([]Candidate, 0, len(payload.Empires))
	for _, empire := range payload.Empires {
		candidates = append(candidates, Candidate{EntityID: empire.EntityID, Name: empire.Name})
	}
	return candidates, nil
}

func (c *Client) GetTownDetail(ctx context.Context, entityID string) (TownDetail, error) {
	var payload struct {
		Claim map[string]any `json:"claim"`
	}
	if err := c.getJSON(ctx, "/claims/"+entityID, nil, &payload); err != nil {
		return TownDetail{}, err
	}

	return TownDetail{
		EntityID: entityID,
		Name:     toString(payload.Claim["name"]),
		Treasury: toInt64(payload.Claim["treasury"]),
	}, nil
}

func (c *Client) GetTownMembers(ctx context.Context, entityID string) ([]Member, error) {
	var payload struct {
		Members []struct {
			PlayerEntityID string `json:"playerEntityId"`
			UserName       string `json:"userName"`
		} `json:"members"`
	}
	if err := c.getJSON(ctx, "/claims/"+entityID+"/members", nil, &payload); err != nil {
		return nil, err
	}

	members := make([]Member, 0, len(payload.Members))
	for _, m := range payload.Members {
		members = append(members, Member{PlayerEntityID: m.PlayerEntityID, UserName: m.UserName, Rank: 0})
	}
	return members, nil
}

func (c *Client) GetEmpireDetail(ctx context.Context, entityID string) (EmpireDetail, error) {
	var payload map[string]any
	if err := c.getJSON(ctx, "/empires/"+entityID, nil, &payload); err != nil {
		return EmpireDetail{}, err
	}

	empireObj, _ := payload["empire"].(map[string]any)
	name := toString(payload["name"])
	if name == "" {
		name = toString(empireObj["name"])
	}
	treasury := toInt64(payload["empireCurrencyTreasury"])
	if treasury == 0 {
		treasury = toInt64(empireObj["empireCurrencyTreasury"])
	}

	membersAny := toSlice(payload["members"])
	members := make([]Member, 0, len(membersAny))
	for _, raw := range membersAny {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		playerID := toString(m["playerEntityId"])
		if playerID == "" {
			playerID = toString(m["entityId"])
		}
		userName := toString(m["userName"])
		if userName == "" {
			userName = toString(m["playerName"])
		}
		members = append(members, Member{
			PlayerEntityID: playerID,
			UserName:       userName,
			Rank:           int(toInt64(m["rank"])),
		})
	}

	return EmpireDetail{
		EntityID:               entityID,
		Name:                   name,
		EmpireCurrencyTreasury: treasury,
		Members:                members,
	}, nil
}

func (c *Client) GetPlayerWallet(ctx context.Context, playerID string) (int64, error) {
	inventories, err := c.getInventories(ctx, playerID)
	if err != nil {
		return 0, err
	}

	for _, inv := range inventories {
		if toString(inv["inventoryName"]) != "Wallet" {
			continue
		}
		pockets := toSlice(inv["pockets"])
		if len(pockets) == 0 {
			return 0, nil
		}
		pocket, ok := pockets[0].(map[string]any)
		if !ok {
			return 0, nil
		}
		contents, _ := pocket["contents"].(map[string]any)
		return toInt64(contents["quantity"]), nil
	}

	return 0, nil
}

func (c *Client) GetPlayerMarketCoins(ctx context.Context, playerID string) (int64, error) {
	var payload struct {
		SellOrders []map[string]any `json:"sellOrders"`
		BuyOrders  []map[string]any `json:"buyOrders"`
	}
	if err := c.getJSON(ctx, "/market/player/"+playerID, nil, &payload); err != nil {
		return 0, err
	}

	var total int64
	for _, order := range payload.SellOrders {
		total += toInt64(order["storedCoins"])
	}
	for _, order := range payload.BuyOrders {
		total += toInt64(order["storedCoins"])
	}
	return total, nil
}

func (c *Client) GetPlayerHexite(ctx context.Context, playerID string, itemID int64) (int64, error) {
	inventories, err := c.getInventories(ctx, playerID)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, inv := range inventories {
		pockets := toSlice(inv["pockets"])
		for _, p := range pockets {
			pocket, ok := p.(map[string]any)
			if !ok {
				continue
			}
			contents, _ := pocket["contents"].(map[string]any)
			if toInt64(contents["itemId"]) == itemID {
				total += toInt64(contents["quantity"])
			}
		}
	}

	return total, nil
}

func (c *Client) getInventories(ctx context.Context, playerID string) ([]map[string]any, error) {
	var raw any
	if err := c.getJSON(ctx, "/players/"+playerID+"/inventories", nil, &raw); err != nil {
		return nil, err
	}

	switch typed := raw.(type) {
	case []any:
		return anySliceToMapSlice(typed), nil
	case map[string]any:
		return anySliceToMapSlice(toSlice(typed["inventories"])), nil
	default:
		return nil, nil
	}
}

func (c *Client) getJSON(ctx context.Context, path string, params map[string]string, target any) error {
	parsedBase, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	// ResolveReference with a leading slash would drop "/api" from baseURL.
	// Build a stable absolute URL so endpoints always stay under /api.
	endpointPath := strings.TrimRight(parsedBase.Path, "/") + "/" + strings.TrimLeft(path, "/")
	parsedPath, err := url.Parse(endpointPath)
	if err != nil {
		return err
	}

	endpoint := &url.URL{
		Scheme: parsedBase.Scheme,
		Host:   parsedBase.Host,
		Path:   parsedPath.Path,
	}
	if len(params) > 0 {
		query := endpoint.Query()
		for k, v := range params {
			query.Set(k, v)
		}
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}

	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("api request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 160 {
			snippet = snippet[:160]
		}
		return fmt.Errorf("invalid json response: %w (body: %q)", err, snippet)
	}

	return nil
}

func anySliceToMapSlice(raw []any) []map[string]any {
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		mapped, ok := item.(map[string]any)
		if ok {
			out = append(out, mapped)
		}
	}
	return out
}

func toSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func toString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

func toInt64(v any) int64 {
	switch typed := v.(type) {
	case nil:
		return 0
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		n, err := typed.Int64()
		if err == nil {
			return n
		}
		f, err := typed.Float64()
		if err == nil {
			return int64(f)
		}
		return 0
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err == nil {
			return n
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return int64(f)
		}
		return 0
	default:
		return 0
	}
}
