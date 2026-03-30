package service

import (
	"strings"
	"testing"
)

func TestSafeString(t *testing.T) {
	m := map[string]any{
		"name":   "Pilsner",
		"count":  42.0,
		"empty":  nil,
		"number": 123,
	}

	tests := []struct {
		key  string
		want string
	}{
		{"name", "Pilsner"},
		{"count", "42"},
		{"missing", ""},
		{"empty", ""},
		{"number", "123"},
	}

	for _, tt := range tests {
		got := safeString(m, tt.key)
		if got != tt.want {
			t.Errorf("safeString(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestSafeFloat(t *testing.T) {
	m := map[string]any{
		"gravity": 1.050,
		"text":    "hello",
		"null":    nil,
	}

	if v, ok := safeFloat(m, "gravity"); !ok || v != 1.050 {
		t.Errorf("safeFloat(gravity) = (%v, %v), want (1.050, true)", v, ok)
	}
	if _, ok := safeFloat(m, "text"); ok {
		t.Error("safeFloat(text) should return false")
	}
	if _, ok := safeFloat(m, "null"); ok {
		t.Error("safeFloat(null) should return false")
	}
	if _, ok := safeFloat(m, "missing"); ok {
		t.Error("safeFloat(missing) should return false")
	}
}

func TestSafeBool(t *testing.T) {
	m := map[string]any{
		"enabled":  true,
		"disabled": false,
		"text":     "yes",
		"null":     nil,
	}

	if v, ok := safeBool(m, "enabled"); !ok || !v {
		t.Errorf("safeBool(enabled) = (%v, %v), want (true, true)", v, ok)
	}
	if v, ok := safeBool(m, "disabled"); !ok || v {
		t.Errorf("safeBool(disabled) = (%v, %v), want (false, true)", v, ok)
	}
	if _, ok := safeBool(m, "text"); ok {
		t.Error("safeBool(text) should return false")
	}
	if _, ok := safeBool(m, "null"); ok {
		t.Error("safeBool(null) should return false")
	}
}

func TestSafeMap(t *testing.T) {
	sub := map[string]any{"name": "IPA"}
	m := map[string]any{
		"recipe": sub,
		"text":   "hello",
		"null":   nil,
	}

	got := safeMap(m, "recipe")
	if got == nil || safeString(got, "name") != "IPA" {
		t.Error("safeMap(recipe) should return sub map")
	}
	if safeMap(m, "text") != nil {
		t.Error("safeMap(text) should return nil")
	}
	if safeMap(m, "null") != nil {
		t.Error("safeMap(null) should return nil")
	}
	if safeMap(m, "missing") != nil {
		t.Error("safeMap(missing) should return nil")
	}
}

func TestSafeSlice(t *testing.T) {
	items := []any{"a", "b"}
	m := map[string]any{
		"items":  items,
		"text":   "hello",
		"null":   nil,
	}

	got := safeSlice(m, "items")
	if len(got) != 2 {
		t.Errorf("safeSlice(items) len = %d, want 2", len(got))
	}
	if safeSlice(m, "text") != nil {
		t.Error("safeSlice(text) should return nil")
	}
	if safeSlice(m, "null") != nil {
		t.Error("safeSlice(null) should return nil")
	}
}

func TestFormatDate(t *testing.T) {
	// 2026-01-15 00:00:00 UTC in milliseconds
	ms := float64(1768435200000)
	got := formatDate(ms)
	if got != "2026-01-15" {
		t.Errorf("formatDate(%f) = %q, want %q", ms, got, "2026-01-15")
	}
}

func TestFormatDateTime(t *testing.T) {
	ms := float64(1768435200000)
	got := formatDateTime(ms)
	if !strings.HasPrefix(got, "2026-01-15") {
		t.Errorf("formatDateTime(%f) = %q, want prefix 2026-01-15", ms, got)
	}
	if !strings.HasSuffix(got, "UTC") {
		t.Errorf("formatDateTime(%f) = %q, want suffix UTC", ms, got)
	}
}

func TestFormatGravity(t *testing.T) {
	got := formatGravity(1.050)
	if got != "1.050 SG" {
		t.Errorf("formatGravity(1.050) = %q, want %q", got, "1.050 SG")
	}
}

func TestFormatVolume(t *testing.T) {
	got := formatVolume(23.5)
	if got != "23.5 L" {
		t.Errorf("formatVolume(23.5) = %q, want %q", got, "23.5 L")
	}
}

func TestFormatTemp(t *testing.T) {
	got := formatTemp(68.0)
	if got != "68.0 °C" {
		t.Errorf("formatTemp(68.0) = %q, want %q", got, "68.0 °C")
	}
}

func TestFormatPercent(t *testing.T) {
	got := formatPercent(72.5)
	if got != "72.5%" {
		t.Errorf("formatPercent(72.5) = %q, want %q", got, "72.5%")
	}
}

func TestFormatWeight(t *testing.T) {
	if got := formatWeight(4.50, "kg"); got != "4.50 kg" {
		t.Errorf("formatWeight(4.50, kg) = %q, want %q", got, "4.50 kg")
	}
	if got := formatWeight(100.00, ""); got != "100.00 kg" {
		t.Errorf("formatWeight(100.00, '') = %q, want %q", got, "100.00 kg")
	}
	if got := formatWeight(50.00, "g"); got != "50.00 g" {
		t.Errorf("formatWeight(50.00, g) = %q, want %q", got, "50.00 g")
	}
}

func TestFormatPH(t *testing.T) {
	got := formatPH(5.35)
	if got != "5.35 pH" {
		t.Errorf("formatPH(5.35) = %q, want %q", got, "5.35 pH")
	}
}

func TestWriteLine(t *testing.T) {
	var b strings.Builder
	writeLine(&b, "  ", "Name", "Pilsner")
	if b.String() != "  Name: Pilsner\n" {
		t.Errorf("writeLine = %q, want %q", b.String(), "  Name: Pilsner\n")
	}
}

func TestWriteLine_EmptyValueSkipped(t *testing.T) {
	var b strings.Builder
	writeLine(&b, "  ", "Name", "")
	if b.String() != "" {
		t.Errorf("writeLine with empty value should produce empty string, got %q", b.String())
	}
}

func TestWriteSection(t *testing.T) {
	var b strings.Builder
	writeSection(&b, "Dates")
	if b.String() != "\nDates:\n" {
		t.Errorf("writeSection = %q, want %q", b.String(), "\nDates:\n")
	}
}

func TestMapSliceToMaps(t *testing.T) {
	items := []any{
		map[string]any{"name": "a"},
		"not a map",
		map[string]any{"name": "b"},
	}
	result := mapSliceToMaps(items)
	if len(result) != 2 {
		t.Errorf("mapSliceToMaps len = %d, want 2", len(result))
	}
}

func TestFormatIngredientSummary(t *testing.T) {
	items := []any{
		map[string]any{"name": "Pilsner Malt", "amount": 4.5},
		map[string]any{"name": "Wheat Malt", "amount": 1.0},
	}
	got := formatIngredientSummary(items, "name", "amount", "kg")
	if !strings.Contains(got, "Pilsner Malt") {
		t.Error("should contain Pilsner Malt")
	}
	if !strings.Contains(got, "Wheat Malt") {
		t.Error("should contain Wheat Malt")
	}
	if !strings.Contains(got, "4.50 kg") {
		t.Error("should contain 4.50 kg")
	}
}

func TestFormatIngredientSummary_Empty(t *testing.T) {
	got := formatIngredientSummary(nil, "name", "amount", "kg")
	if got != "" {
		t.Errorf("empty input should return empty string, got %q", got)
	}
}

func TestFormatIngredientSummary_NoAmount(t *testing.T) {
	items := []any{
		map[string]any{"name": "Irish Moss"},
	}
	got := formatIngredientSummary(items, "name", "amount", "g")
	if got != "Irish Moss" {
		t.Errorf("got %q, want %q", got, "Irish Moss")
	}
}

func TestFormatHopSummary(t *testing.T) {
	items := []any{
		map[string]any{"name": "Citra", "amount": 50.0, "use": "Dry Hop"},
		map[string]any{"name": "Mosaic", "amount": 30.0, "use": "Boil", "time": 10.0},
	}
	got := formatHopSummary(items)
	if !strings.Contains(got, "Citra") {
		t.Error("should contain Citra")
	}
	if !strings.Contains(got, "50g") {
		t.Error("should contain 50g")
	}
	if !strings.Contains(got, "Dry Hop") {
		t.Error("should contain Dry Hop")
	}
	if !strings.Contains(got, "10min") {
		t.Error("should contain 10min")
	}
}

func TestFormatHopSummary_Empty(t *testing.T) {
	got := formatHopSummary(nil)
	if got != "" {
		t.Errorf("empty input should return empty string, got %q", got)
	}
}

func TestFormatYeastSummary(t *testing.T) {
	items := []any{
		map[string]any{"name": "US-05", "form": "Dry"},
		map[string]any{"name": "WLP001"},
	}
	got := formatYeastSummary(items)
	if !strings.Contains(got, "US-05 (Dry)") {
		t.Errorf("should contain 'US-05 (Dry)', got %q", got)
	}
	if !strings.Contains(got, "WLP001") {
		t.Error("should contain WLP001")
	}
}

func TestFormatYeastSummary_Empty(t *testing.T) {
	got := formatYeastSummary(nil)
	if got != "" {
		t.Errorf("empty input should return empty string, got %q", got)
	}
}
