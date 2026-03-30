package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brewfather-mcp/internal/client"
)

func batchTestService(t *testing.T, handler http.HandlerFunc) *BatchService {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := client.NewClientWithBaseURL("u", "k", ts.URL)
	return NewBatchService(c)
}

func TestListBatches_Empty(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})

	text, err := svc.ListBatches(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "No batches found." {
		t.Errorf("text = %q, want %q", text, "No batches found.")
	}
}

func TestListBatches_WithResults(t *testing.T) {
	payload := `[
		{
			"_id": "batch1",
			"name": "Hazy IPA",
			"status": "Fermenting",
			"batchNo": 42,
			"brewDate": 1768435200000,
			"brewer": "Newton",
			"recipe": {"name": "NE IPA"}
		},
		{
			"_id": "batch2",
			"name": "Stout",
			"status": "Completed"
		}
	]`
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.ListBatches(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Found 2 batch(es)",
		`"Hazy IPA" (ID: batch1)`,
		"Batch #42",
		"Status: Fermenting",
		"Brewer: Newton",
		"Recipe: NE IPA",
		`"Stout" (ID: batch2)`,
		"Status: Completed",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestListBatches_APIError(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	})

	_, err := svc.ListBatches(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBatch_Detail(t *testing.T) {
	payload := `{
		"_id": "b1",
		"name": "Hefeweizen",
		"status": "Completed",
		"batchNo": 10,
		"brewer": "Newton",
		"brewDate": 1768435200000,
		"measuredOg": 1.052,
		"measuredFg": 1.012,
		"measuredMashPh": 5.35,
		"measuredAbv": 5.2,
		"estimatedOg": 1.050,
		"estimatedIbu": 15,
		"carbonationTemp": 4.0,
		"carbonationType": "Forced",
		"recipe": {
			"name": "Bavarian Wheat",
			"type": "All Grain",
			"fermentables": [{"name": "Wheat Malt", "amount": 3.0}],
			"hops": [{"name": "Hallertau", "amount": 25, "use": "Boil", "time": 60}],
			"yeasts": [{"name": "WB-06", "form": "Dry"}],
			"style": {"name": "Hefeweizen"}
		},
		"batchNotes": "Great brew day",
		"tasteNotes": "Banana and clove",
		"tasteRating": 4
	}`
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetBatch(context.Background(), "b1", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`Batch: "Hefeweizen" (ID: b1)`,
		"Status: Completed",
		"Brewer: Newton",
		"Brew Date:",
		"Original Gravity: 1.052 SG",
		"Final Gravity: 1.012 SG",
		"Mash pH: 5.35 pH",
		"ABV: 5.2%",
		"OG: 1.050 SG",
		"IBU: 15",
		"Carbonation Temp: 4.0 °C",
		"Carbonation Type: Forced",
		`"Bavarian Wheat" (All Grain)`,
		"Wheat Malt (3.00 kg)",
		"Hallertau",
		"WB-06 (Dry)",
		"Style: Hefeweizen",
		"Great brew day",
		"Banana and clove",
		"Rating: 4/5",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetBatch_MinimalFields(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"_id": "b2", "name": "Simple Batch"}`))
	})

	text, err := svc.GetBatch(context.Background(), "b2", &client.GetItemParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, `Batch: "Simple Batch" (ID: b2)`) {
		t.Errorf("output missing header, got:\n%s", text)
	}
}

func TestUpdateBatch(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	status := "Fermenting"
	text, err := svc.UpdateBatch(context.Background(), "b1", &client.UpdateBatchBody{Status: &status})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(text, "b1 updated successfully") {
		t.Errorf("text = %q, missing update confirmation", text)
	}
}

func TestGetBatchLastReading(t *testing.T) {
	payload := `{
		"time": 1768435200000,
		"type": "iSpindel",
		"sg": 1.022,
		"temp": 20.5,
		"battery": 3.85,
		"angle": 45.2,
		"comment": "Looking good"
	}`
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetBatchLastReading(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Latest Reading:",
		"Source: iSpindel",
		"Specific Gravity: 1.022 SG",
		"Temperature: 20.5 °C",
		"Battery: 3.85 V",
		"Angle: 45.2°",
		"Comment: Looking good",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}

func TestGetBatchReadings_Empty(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	})

	text, err := svc.GetBatchReadings(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "No readings found for this batch." {
		t.Errorf("text = %q", text)
	}
}

func TestGetBatchReadings_Multiple(t *testing.T) {
	payload := `[
		{"time": 1768435200000, "sg": 1.050, "temp": 20.0},
		{"time": 1768521600000, "sg": 1.030, "temp": 19.5}
	]`
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetBatchReadings(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(text, "Found 2 reading(s)") {
		t.Errorf("missing count header, got:\n%s", text)
	}
	if !strings.Contains(text, "--- Reading 1 ---") {
		t.Errorf("missing reading 1 header, got:\n%s", text)
	}
	if !strings.Contains(text, "1.050 SG") {
		t.Errorf("missing gravity value, got:\n%s", text)
	}
}

func TestGetBatchBrewTracker_Empty(t *testing.T) {
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})

	text, err := svc.GetBatchBrewTracker(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Brew tracker is not active for this batch." {
		t.Errorf("text = %q", text)
	}
}

func TestGetBatchBrewTracker_Active(t *testing.T) {
	payload := `{
		"_id": "bt1",
		"enabled": true,
		"completed": false,
		"stage": 2,
		"stages": [
			{
				"name": "Mash",
				"type": "mash",
				"duration": 3600,
				"steps": [
					{"name": "Sacch Rest", "type": "temperature", "value": 67.0, "duration": 3600}
				]
			},
			{
				"name": "Boil",
				"type": "boil",
				"paused": true
			}
		]
	}`
	svc := batchTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	})

	text, err := svc.GetBatchBrewTracker(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Brew Tracker (ID: bt1)",
		"Enabled: true",
		"Completed: false",
		"Current Stage: 2",
		"1. Mash (mash)",
		"60 min",
		"Sacch Rest [temperature]",
		"67.0 °C",
		"2. Boil (boil)",
		"Status: Paused",
	}
	for _, s := range checks {
		if !strings.Contains(text, s) {
			t.Errorf("output missing %q\nfull output:\n%s", s, text)
		}
	}
}
