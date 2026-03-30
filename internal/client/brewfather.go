package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const defaultBaseURL = "https://api.brewfather.app/v2"

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("brewfather API error %d: %s", e.StatusCode, e.Message)
}

type Client struct {
	httpClient *http.Client
	authHeader string
	baseURL    string
}

func NewClient(userID, apiKey string) *Client {
	encoded := base64.StdEncoding.EncodeToString([]byte(userID + ":" + apiKey))
	return &Client{
		httpClient: &http.Client{},
		authHeader: "Basic " + encoded,
		baseURL:    defaultBaseURL,
	}
}

func NewClientWithBaseURL(userID, apiKey, baseURL string) *Client {
	c := NewClient(userID, apiKey)
	c.baseURL = baseURL
	return c
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(body))}
	}

	return json.RawMessage(body), nil
}

func (c *Client) doPatch(ctx context.Context, path string, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshalling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+path, strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(body))}
	}

	return strings.TrimSpace(string(body)), nil
}

// --- Parameter structs ---

type ListBatchesParams struct {
	Status          string
	Limit           int
	StartAfter      string
	Include         string
	Complete        bool
	OrderBy         string
	OrderByDirection string
}

func (p *ListBatchesParams) toQuery() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.StartAfter != "" {
		q.Set("start_after", p.StartAfter)
	}
	if p.Include != "" {
		q.Set("include", p.Include)
	}
	if p.Complete {
		q.Set("complete", "true")
	}
	if p.OrderBy != "" {
		q.Set("order_by", p.OrderBy)
	}
	if p.OrderByDirection != "" {
		q.Set("order_by_direction", p.OrderByDirection)
	}
	return q
}

type ListRecipesParams struct {
	Limit            int
	StartAfter       string
	Include          string
	Complete         bool
	OrderBy          string
	OrderByDirection string
}

func (p *ListRecipesParams) toQuery() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.StartAfter != "" {
		q.Set("start_after", p.StartAfter)
	}
	if p.Include != "" {
		q.Set("include", p.Include)
	}
	if p.Complete {
		q.Set("complete", "true")
	}
	if p.OrderBy != "" {
		q.Set("order_by", p.OrderBy)
	}
	if p.OrderByDirection != "" {
		q.Set("order_by_direction", p.OrderByDirection)
	}
	return q
}

type ListInventoryParams struct {
	InventoryExists  *bool
	InventoryNegative *bool
	Limit            int
	StartAfter       string
	Include          string
	Complete         bool
	OrderBy          string
	OrderByDirection string
}

func (p *ListInventoryParams) toQuery() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.InventoryExists != nil && *p.InventoryExists {
		q.Set("inventory_exists", "true")
	}
	if p.InventoryNegative != nil && *p.InventoryNegative {
		q.Set("inventory_negative", "true")
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.StartAfter != "" {
		q.Set("start_after", p.StartAfter)
	}
	if p.Include != "" {
		q.Set("include", p.Include)
	}
	if p.Complete {
		q.Set("complete", "true")
	}
	if p.OrderBy != "" {
		q.Set("order_by", p.OrderBy)
	}
	if p.OrderByDirection != "" {
		q.Set("order_by_direction", p.OrderByDirection)
	}
	return q
}

type GetItemParams struct {
	Include string
}

func (p *GetItemParams) toQuery() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	if p.Include != "" {
		q.Set("include", p.Include)
	}
	return q
}

type UpdateBatchBody struct {
	Status                   *string  `json:"status,omitempty"`
	MeasuredMashPh           *float64 `json:"measuredMashPh,omitempty"`
	MeasuredBoilSize         *float64 `json:"measuredBoilSize,omitempty"`
	MeasuredFirstWortGravity *float64 `json:"measuredFirstWortGravity,omitempty"`
	MeasuredPreBoilGravity   *float64 `json:"measuredPreBoilGravity,omitempty"`
	MeasuredPostBoilGravity  *float64 `json:"measuredPostBoilGravity,omitempty"`
	MeasuredKettleSize       *float64 `json:"measuredKettleSize,omitempty"`
	MeasuredOg               *float64 `json:"measuredOg,omitempty"`
	MeasuredFermenterTopUp   *float64 `json:"measuredFermenterTopUp,omitempty"`
	MeasuredBatchSize        *float64 `json:"measuredBatchSize,omitempty"`
	MeasuredFg               *float64 `json:"measuredFg,omitempty"`
	MeasuredBottlingSize     *float64 `json:"measuredBottlingSize,omitempty"`
	CarbonationTemp          *float64 `json:"carbonationTemp,omitempty"`
}

type UpdateInventoryBody struct {
	Inventory       *float64 `json:"inventory,omitempty"`
	InventoryAdjust *float64 `json:"inventory_adjust,omitempty"`
}

// --- Batch endpoints ---

func (c *Client) ListBatches(ctx context.Context, params *ListBatchesParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/batches", params.toQuery())
}

func (c *Client) GetBatch(ctx context.Context, id string, params *GetItemParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/batches/"+url.PathEscape(id), params.toQuery())
}

func (c *Client) UpdateBatch(ctx context.Context, id string, body *UpdateBatchBody) (string, error) {
	return c.doPatch(ctx, "/batches/"+url.PathEscape(id), body)
}

func (c *Client) GetBatchLastReading(ctx context.Context, id string) (json.RawMessage, error) {
	return c.doGet(ctx, "/batches/"+url.PathEscape(id)+"/readings/last", nil)
}

func (c *Client) GetBatchReadings(ctx context.Context, id string) (json.RawMessage, error) {
	return c.doGet(ctx, "/batches/"+url.PathEscape(id)+"/readings", nil)
}

func (c *Client) GetBatchBrewTracker(ctx context.Context, id string) (json.RawMessage, error) {
	return c.doGet(ctx, "/batches/"+url.PathEscape(id)+"/brewtracker", nil)
}

// --- Recipe endpoints ---

func (c *Client) ListRecipes(ctx context.Context, params *ListRecipesParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/recipes", params.toQuery())
}

func (c *Client) GetRecipe(ctx context.Context, id string, params *GetItemParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/recipes/"+url.PathEscape(id), params.toQuery())
}

// --- Inventory endpoints (generic by category) ---

func (c *Client) ListInventory(ctx context.Context, category string, params *ListInventoryParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/inventory/"+category, params.toQuery())
}

func (c *Client) GetInventoryItem(ctx context.Context, category, id string, params *GetItemParams) (json.RawMessage, error) {
	return c.doGet(ctx, "/inventory/"+category+"/"+url.PathEscape(id), params.toQuery())
}

func (c *Client) UpdateInventoryItem(ctx context.Context, category, id string, body *UpdateInventoryBody) (string, error) {
	return c.doPatch(ctx, "/inventory/"+category+"/"+url.PathEscape(id), body)
}
