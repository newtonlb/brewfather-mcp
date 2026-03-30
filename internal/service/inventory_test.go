package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brewfather-mcp/internal/client"
)

func inventoryTestService(t *testing.T, handler http.HandlerFunc) *InventoryService {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := client.NewClientWithBaseURL("u", "k", ts.URL)
	return NewInventoryService(c)
}

// --- Fermentables ---

func TestListFermentables_Empty(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})

	text, err := svc.ListFermentables(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "No fermentable items found") {
		t.Errorf("text = %q", text)
	}
}

func TestListFermentables_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "f1",
			"name": "Pilsner Malt",
			"type": "Grain",
			"supplier": "Weyermann",
			"inventory": 10.5
		}
	]`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListFermentables(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 1 fermentable item(s)",
		`"Pilsner Malt" (ID: f1)`,
		"Type: Grain",
		"Supplier: Weyermann",
		"Stock: 10.50 kg",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetFermentable_Detail(t *testing.T) {
	payload := `{
		"_id": "f1",
		"name": "Maris Otter",
		"type": "Grain",
		"grainCategory": "Base (Pale Ale)",
		"origin": "United Kingdom",
		"supplier": "Crisp",
		"use": "Mash",
		"color": 3.0,
		"potential": 1.038,
		"attenuation": 81,
		"moisture": 4.0,
		"protein": 10.3,
		"diastaticPower": 120,
		"notFermentable": false,
		"inventory": 5.5,
		"costPerAmount": 2.50,
		"notes": "Classic English base malt"
	}`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetFermentable(context.Background(), "f1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Fermentable: "Maris Otter" (ID: f1)`,
		"Type: Grain",
		"Grain Category: Base (Pale Ale)",
		"Origin: United Kingdom",
		"Supplier: Crisp",
		"Use: Mash",
		"Color: 3.0 SRM",
		"Potential: 1.038 SG",
		"Attenuation: 81.0%",
		"Moisture: 4.0%",
		"Protein: 10.3%",
		"Diastatic Power: 120 Lintner",
		"Current Stock: 5.50",
		"Cost Per Unit: 2.50",
		"Classic English base malt",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestUpdateFermentableInventory(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	inv := 10.0
	text, err := svc.UpdateFermentableInventory(context.Background(), "f1", &client.UpdateInventoryBody{Inventory: &inv})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "Fermentable f1 inventory updated") {
		t.Errorf("text = %q", text)
	}
}

// --- Hops ---

func TestListHops_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "h1",
			"name": "Citra",
			"alpha": 12.5,
			"type": "Pellet",
			"use": "Dry Hop",
			"inventory": 200
		}
	]`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListHops(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 1 hop item(s)",
		`"Citra" (ID: h1)`,
		"Alpha: 12.5%",
		"Form: Pellet",
		"Use: Dry Hop",
		"Stock: 200 g",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetHop_Detail(t *testing.T) {
	payload := `{
		"_id": "h1",
		"name": "Citra",
		"alpha": 12.5,
		"beta": 4.0,
		"type": "Pellet",
		"use": "Dry Hop",
		"origin": "USA",
		"usage": "Aroma",
		"oil": 2.5,
		"myrcene": 60,
		"humulene": 12,
		"caryophyllene": 7,
		"inventory": 200,
		"notes": "Tropical and citrus"
	}`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetHop(context.Background(), "h1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Hop: "Citra" (ID: h1)`,
		"Alpha Acid: 12.5%",
		"Beta Acid: 4.0%",
		"Form: Pellet",
		"Origin: USA",
		"Usage Role: Aroma",
		"Total Oil: 2.5 ml/100g",
		"Myrcene: 60.0%",
		"Humulene: 12.0%",
		"Caryophyllene: 7.0%",
		"Current Stock: 200.00",
		"Tropical and citrus",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestUpdateHopInventory(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	adj := -50.0
	text, err := svc.UpdateHopInventory(context.Background(), "h1", &client.UpdateInventoryBody{InventoryAdjust: &adj})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "Hop h1 inventory updated") {
		t.Errorf("text = %q", text)
	}
}

// --- Miscs ---

func TestListMiscs_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "m1",
			"name": "Whirlfloc",
			"type": "Fining",
			"use": "Boil",
			"inventory": 15.0
		}
	]`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListMiscs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 1 misc ingredient item(s)",
		`"Whirlfloc" (ID: m1)`,
		"Type: Fining",
		"Use: Boil",
		"Stock: 15.0",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetMisc_Detail(t *testing.T) {
	payload := `{
		"_id": "m1",
		"name": "Lactic Acid",
		"type": "Water Agent",
		"use": "Mash",
		"unit": "ml",
		"concentration": 88.0,
		"waterAdjustment": true,
		"inventory": 250,
		"notes": "For pH adjustment"
	}`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetMisc(context.Background(), "m1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Misc Ingredient: "Lactic Acid" (ID: m1)`,
		"Type: Water Agent",
		"Use: Mash",
		"Unit: ml",
		"Concentration: 88.0%",
		"Water Adjustment: Yes",
		"Current Stock: 250.00",
		"For pH adjustment",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestUpdateMiscInventory(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	inv := 100.0
	text, err := svc.UpdateMiscInventory(context.Background(), "m1", &client.UpdateInventoryBody{Inventory: &inv})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "Misc ingredient m1 inventory updated") {
		t.Errorf("text = %q", text)
	}
}

// --- Yeasts ---

func TestListYeasts_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "y1",
			"name": "Safale US-05",
			"type": "Ale",
			"form": "Dry",
			"attenuation": 81,
			"laboratory": "Fermentis",
			"inventory": 3
		}
	]`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListYeasts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 1 yeast item(s)",
		`"Safale US-05" (ID: y1)`,
		"Type: Ale",
		"Form: Dry",
		"Atten: 81%",
		"Lab: Fermentis",
		"Stock: 3.0",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetYeast_Detail(t *testing.T) {
	payload := `{
		"_id": "y1",
		"name": "Safale US-05",
		"type": "Ale",
		"form": "Dry",
		"laboratory": "Fermentis",
		"productId": "US-05",
		"flocculation": "Medium",
		"attenuation": 81,
		"minAttenuation": 78,
		"maxAttenuation": 82,
		"minTemp": 15,
		"maxTemp": 24,
		"maxAbv": 11,
		"bestFor": "American Ales, IPAs",
		"inventory": 3,
		"costPerAmount": 4.50,
		"notes": "Clean fermenting ale yeast"
	}`
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetYeast(context.Background(), "y1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Yeast: "Safale US-05" (ID: y1)`,
		"Type: Ale",
		"Form: Dry",
		"Laboratory: Fermentis",
		"Product ID: US-05",
		"Flocculation: Medium",
		"Attenuation: 81.0%",
		"Attenuation Range: 78-82%",
		"Temperature Range: 15.0 °C - 24.0 °C",
		"Max ABV Tolerance: 11.0%",
		"American Ales, IPAs",
		"Current Stock: 3.00",
		"Cost Per Unit: 4.50",
		"Clean fermenting ale yeast",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestUpdateYeastInventory(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	adj := 2.0
	text, err := svc.UpdateYeastInventory(context.Background(), "y1", &client.UpdateInventoryBody{InventoryAdjust: &adj})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "Yeast y1 inventory updated") {
		t.Errorf("text = %q", text)
	}
}

// --- Error propagation ---

func TestListFermentables_APIError(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	})

	_, err := svc.ListFermentables(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFermentable_InvalidJSON(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	})

	_, err := svc.GetFermentable(context.Background(), "f1", &client.GetItemParams{})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// --- Inventory list path correctness ---

func TestListHops_CorrectPath(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/hops" {
			t.Errorf("path = %s, want /inventory/hops", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	})
	svc.ListHops(context.Background(), nil)
}

func TestListMiscs_CorrectPath(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/miscs" {
			t.Errorf("path = %s, want /inventory/miscs", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	})
	svc.ListMiscs(context.Background(), nil)
}

func TestListYeasts_CorrectPath(t *testing.T) {
	svc := inventoryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory/yeasts" {
			t.Errorf("path = %s, want /inventory/yeasts", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	})
	svc.ListYeasts(context.Background(), nil)
}
