package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brewfather-mcp/internal/client"
)

func recipeTestService(t *testing.T, handler http.HandlerFunc) *RecipeService {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := client.NewClientWithBaseURL("u", "k", ts.URL)
	return NewRecipeService(c)
}

func TestListRecipes_Empty(t *testing.T) {
	svc := recipeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})

	text, err := svc.ListRecipes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "No recipes found." {
		t.Errorf("text = %q, want %q", text, "No recipes found.")
	}
}

func TestListRecipes_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "r1",
			"name": "American IPA",
			"type": "All Grain",
			"author": "Newton",
			"equipment": {"name": "My Brewhouse"},
			"style": {"name": "American IPA"}
		},
		{
			"_id": "r2",
			"name": "Porter"
		}
	]`
	svc := recipeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListRecipes(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 2 recipe(s)",
		`"American IPA" (ID: r1)`,
		"Type: All Grain",
		"Author: Newton",
		"Equipment: My Brewhouse",
		"Style: American IPA",
		`"Porter" (ID: r2)`,
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetRecipe_FullDetail(t *testing.T) {
	payload := `{
		"_id": "r1",
		"name": "Hefeweizen",
		"type": "All Grain",
		"author": "Newton",
		"teaser": "A refreshing wheat beer",
		"style": {
			"name": "Weissbier",
			"category": "German Wheat Beer",
			"styleGuide": "BJCP 2021",
			"ogMin": 1.044,
			"ogMax": 1.052,
			"fgMin": 1.010,
			"fgMax": 1.014,
			"ibuMin": 8,
			"ibuMax": 15,
			"abvMin": 4.3,
			"abvMax": 5.6,
			"colorMin": 2,
			"colorMax": 6
		},
		"batchSize": 23,
		"boilSize": 28,
		"boilTime": 60,
		"efficiency": 72,
		"og": 1.048,
		"fg": 1.012,
		"abv": 4.7,
		"ibu": 12,
		"color": 4,
		"buGuRatio": 0.25,
		"carbonation": 3.5,
		"fermentables": [
			{"name": "Pilsner Malt", "amount": 2.5},
			{"name": "Wheat Malt", "amount": 2.5}
		],
		"fermentablesTotalAmount": 5.0,
		"hops": [
			{"name": "Hallertau Mittelfrueh", "amount": 20, "use": "Boil", "time": 60}
		],
		"hopsTotalAmount": 20,
		"yeasts": [{"name": "Safbrew WB-06", "form": "Dry"}],
		"miscs": [{"name": "Whirlfloc", "amount": 1}],
		"mash": {
			"name": "Single Infusion",
			"ph": 5.4,
			"steps": [
				{"name": "Mash In", "type": "Infusion", "stepTemp": 66, "stepTime": 60}
			]
		},
		"fermentation": {
			"name": "Ale Primary",
			"steps": [
				{"name": "Primary", "type": "Primary", "stepTemp": 19, "stepTime": 14}
			]
		},
		"equipment": {
			"name": "My System",
			"batchSize": 23,
			"efficiency": 72,
			"boilTime": 60
		},
		"notes": "Ferment warm for more banana"
	}`
	svc := recipeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetRecipe(context.Background(), "r1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Recipe: "Hefeweizen" (ID: r1)`,
		"Type: All Grain",
		"Author: Newton",
		"A refreshing wheat beer",
		"Weissbier (German Wheat Beer) - BJCP 2021",
		"OG: 1.044-1.052",
		"FG: 1.010-1.014",
		"IBU: 8-15",
		"ABV: 4.3-5.6%",
		"Color: 2-6 SRM",
		"Batch Size: 23.0 L",
		"Pre-Boil Volume: 28.0 L",
		"Boil Time: 60 min",
		"Efficiency: 72.0%",
		"OG: 1.048 SG",
		"FG: 1.012 SG",
		"ABV: 4.7%",
		"IBU: 12",
		"Color: 4.0 SRM",
		"BU:GU: 0.25",
		"Carbonation: 3.5 volumes CO2",
		"Pilsner Malt (2.50 kg)",
		"Wheat Malt (2.50 kg)",
		"Total: 5.00 kg",
		"Hallertau Mittelfrueh",
		"Total: 20 g",
		"Safbrew WB-06 (Dry)",
		"Whirlfloc",
		`"Single Infusion"`,
		"Target pH: 5.40 pH",
		"Mash In (Infusion) @ 66.0 °C for 60 min",
		`"Ale Primary"`,
		"Primary (Primary) @ 19.0 °C for 14 days",
		`"My System"`,
		"Batch Size: 23.0 L",
		"Ferment warm for more banana",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetRecipe_MinimalFields(t *testing.T) {
	svc := recipeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"_id": "r2", "name": "Simple Recipe"}`))
	})

	text, err := svc.GetRecipe(context.Background(), "r2", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, `Recipe: "Simple Recipe" (ID: r2)`) {
		t.Errorf("missing header, got:\n%s", text)
	}
}

func TestListRecipes_APIError(t *testing.T) {
	svc := recipeTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	})

	_, err := svc.ListRecipes(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
