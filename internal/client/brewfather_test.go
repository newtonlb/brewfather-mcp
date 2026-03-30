package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := NewClientWithBaseURL("testuser", "testapikey", ts.URL)
	return c, ts
}

func TestNewClient_AuthHeader(t *testing.T) {
	c := NewClient("myuser", "mykey")
	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("myuser:mykey"))
	if c.authHeader != expected {
		t.Errorf("authHeader = %q, want %q", c.authHeader, expected)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	c := NewClientWithBaseURL("u", "k", "http://localhost:9999")
	if c.baseURL != "http://localhost:9999" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:9999")
	}
}

func TestDoGet_Success(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth == "" {
			t.Error("missing Authorization header")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"_id":"abc"}]`))
	})

	raw, err := c.ListBatches(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(raw) != `[{"_id":"abc"}]` {
		t.Errorf("body = %s, want %s", string(raw), `[{"_id":"abc"}]`)
	}
}

func TestDoGet_APIError(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	})

	_, err := c.ListBatches(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Message != "Unauthorized" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Unauthorized")
	}
}

func TestDoPatch_Success(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		json.Unmarshal(body, &m)
		if m["status"] != "Fermenting" {
			t.Errorf("body status = %v, want Fermenting", m["status"])
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Updated"))
	})

	status := "Fermenting"
	msg, err := c.UpdateBatch(context.Background(), "abc123", &UpdateBatchBody{Status: &status})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Updated" {
		t.Errorf("msg = %q, want %q", msg, "Updated")
	}
}

func TestDoPatch_APIError(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid status"))
	})

	status := "InvalidStatus"
	_, err := c.UpdateBatch(context.Background(), "abc", &UpdateBatchBody{Status: &status})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestListBatchesParams_ToQuery(t *testing.T) {
	p := &ListBatchesParams{
		Status:           "Fermenting",
		Limit:            25,
		StartAfter:       "abc123",
		Include:          "recipe.mash",
		Complete:         true,
		OrderBy:          "name",
		OrderByDirection: "desc",
	}
	q := p.toQuery()

	checks := map[string]string{
		"status":             "Fermenting",
		"limit":              "25",
		"start_after":        "abc123",
		"include":            "recipe.mash",
		"complete":           "true",
		"order_by":           "name",
		"order_by_direction": "desc",
	}
	for key, want := range checks {
		if got := q.Get(key); got != want {
			t.Errorf("query[%s] = %q, want %q", key, got, want)
		}
	}
}

func TestListBatchesParams_NilReturnsEmpty(t *testing.T) {
	var p *ListBatchesParams
	q := p.toQuery()
	if len(q) != 0 {
		t.Errorf("expected empty query for nil params, got %v", q)
	}
}

func TestListBatchesParams_EmptyFieldsOmitted(t *testing.T) {
	p := &ListBatchesParams{Limit: 10}
	q := p.toQuery()
	if q.Get("status") != "" {
		t.Error("empty status should not be in query")
	}
	if q.Get("limit") != "10" {
		t.Errorf("limit = %q, want 10", q.Get("limit"))
	}
}

func TestListInventoryParams_ToQuery(t *testing.T) {
	trueVal := true
	p := &ListInventoryParams{
		InventoryExists:  &trueVal,
		InventoryNegative: nil,
		Limit:            50,
		Complete:         true,
	}
	q := p.toQuery()
	if q.Get("inventory_exists") != "true" {
		t.Errorf("inventory_exists = %q, want true", q.Get("inventory_exists"))
	}
	if q.Get("inventory_negative") != "" {
		t.Error("nil inventory_negative should not be in query")
	}
}

func TestGetItemParams_NilReturnsEmpty(t *testing.T) {
	var p *GetItemParams
	q := p.toQuery()
	if len(q) != 0 {
		t.Errorf("expected empty query for nil params, got %v", q)
	}
}

func TestListBatches_QueryParams(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batches" {
			t.Errorf("path = %s, want /batches", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "Brewing" {
			t.Errorf("status = %q, want Brewing", r.URL.Query().Get("status"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	c.ListBatches(context.Background(), &ListBatchesParams{Status: "Brewing"})
}

func TestGetBatch_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batches/abc123" {
			t.Errorf("path = %s, want /batches/abc123", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c.GetBatch(context.Background(), "abc123", &GetItemParams{})
}

func TestGetBatchLastReading_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batches/b1/readings/last" {
			t.Errorf("path = %s, want /batches/b1/readings/last", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c.GetBatchLastReading(context.Background(), "b1")
}

func TestGetBatchReadings_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batches/b1/readings" {
			t.Errorf("path = %s, want /batches/b1/readings", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	c.GetBatchReadings(context.Background(), "b1")
}

func TestGetBatchBrewTracker_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batches/b1/brewtracker" {
			t.Errorf("path = %s, want /batches/b1/brewtracker", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c.GetBatchBrewTracker(context.Background(), "b1")
}

func TestListRecipes_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recipes" {
			t.Errorf("path = %s, want /recipes", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	c.ListRecipes(context.Background(), &ListRecipesParams{Limit: 5})
}

func TestGetRecipe_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/recipes/r1" {
			t.Errorf("path = %s, want /recipes/r1", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c.GetRecipe(context.Background(), "r1", &GetItemParams{})
}

func TestListInventory_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/hops" {
			t.Errorf("path = %s, want /inventory/hops", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	c.ListInventory(context.Background(), "hops", &ListInventoryParams{})
}

func TestGetInventoryItem_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/fermentables/f1" {
			t.Errorf("path = %s, want /inventory/fermentables/f1", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	c.GetInventoryItem(context.Background(), "fermentables", "f1", &GetItemParams{})
}

func TestUpdateInventoryItem_Path(t *testing.T) {
	c, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/yeasts/y1" {
			t.Errorf("path = %s, want /inventory/yeasts/y1", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Updated"))
	})

	inv := 5.0
	c.UpdateInventoryItem(context.Background(), "yeasts", "y1", &UpdateInventoryBody{Inventory: &inv})
}

func TestAPIError_ErrorString(t *testing.T) {
	e := &APIError{StatusCode: 403, Message: "Forbidden"}
	want := "brewfather API error 403: Forbidden"
	if e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}
