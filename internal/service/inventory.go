package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"brewfather-mcp/internal/client"
)

type InventoryService struct {
	client *client.Client
}

func NewInventoryService(c *client.Client) *InventoryService {
	return &InventoryService{client: c}
}

// --- Fermentables ---

func (s *InventoryService) ListFermentables(ctx context.Context, params *client.ListInventoryParams) (string, error) {
	raw, err := s.client.ListInventory(ctx, "fermentables", params)
	if err != nil {
		return "", err
	}
	return formatInventoryList(raw, "fermentable", formatFermentableSummaryLine)
}

func (s *InventoryService) GetFermentable(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetInventoryItem(ctx, "fermentables", id, params)
	if err != nil {
		return "", err
	}
	return formatInventoryDetail(raw, "Fermentable", formatFermentableDetail)
}

func (s *InventoryService) UpdateFermentableInventory(ctx context.Context, id string, body *client.UpdateInventoryBody) (string, error) {
	msg, err := s.client.UpdateInventoryItem(ctx, "fermentables", id, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Fermentable %s inventory updated. Server response: %s", id, msg), nil
}

// --- Hops ---

func (s *InventoryService) ListHops(ctx context.Context, params *client.ListInventoryParams) (string, error) {
	raw, err := s.client.ListInventory(ctx, "hops", params)
	if err != nil {
		return "", err
	}
	return formatInventoryList(raw, "hop", formatHopSummaryLine)
}

func (s *InventoryService) GetHop(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetInventoryItem(ctx, "hops", id, params)
	if err != nil {
		return "", err
	}
	return formatInventoryDetail(raw, "Hop", formatHopDetail)
}

func (s *InventoryService) UpdateHopInventory(ctx context.Context, id string, body *client.UpdateInventoryBody) (string, error) {
	msg, err := s.client.UpdateInventoryItem(ctx, "hops", id, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Hop %s inventory updated. Server response: %s", id, msg), nil
}

// --- Miscs ---

func (s *InventoryService) ListMiscs(ctx context.Context, params *client.ListInventoryParams) (string, error) {
	raw, err := s.client.ListInventory(ctx, "miscs", params)
	if err != nil {
		return "", err
	}
	return formatInventoryList(raw, "misc ingredient", formatMiscSummaryLine)
}

func (s *InventoryService) GetMisc(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetInventoryItem(ctx, "miscs", id, params)
	if err != nil {
		return "", err
	}
	return formatInventoryDetail(raw, "Misc Ingredient", formatMiscDetail)
}

func (s *InventoryService) UpdateMiscInventory(ctx context.Context, id string, body *client.UpdateInventoryBody) (string, error) {
	msg, err := s.client.UpdateInventoryItem(ctx, "miscs", id, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Misc ingredient %s inventory updated. Server response: %s", id, msg), nil
}

// --- Yeasts ---

func (s *InventoryService) ListYeasts(ctx context.Context, params *client.ListInventoryParams) (string, error) {
	raw, err := s.client.ListInventory(ctx, "yeasts", params)
	if err != nil {
		return "", err
	}
	return formatInventoryList(raw, "yeast", formatYeastSummaryLine)
}

func (s *InventoryService) GetYeast(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetInventoryItem(ctx, "yeasts", id, params)
	if err != nil {
		return "", err
	}
	return formatInventoryDetail(raw, "Yeast", formatYeastDetail)
}

func (s *InventoryService) UpdateYeastInventory(ctx context.Context, id string, body *client.UpdateInventoryBody) (string, error) {
	msg, err := s.client.UpdateInventoryItem(ctx, "yeasts", id, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Yeast %s inventory updated. Server response: %s", id, msg), nil
}

// --- Generic list/detail helpers ---

type summaryLineFunc func(m map[string]any) string
type detailFunc func(b *strings.Builder, m map[string]any)

func formatInventoryList(raw json.RawMessage, singular string, summaryLine summaryLineFunc) (string, error) {
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(items) == 0 {
		return fmt.Sprintf("No %s items found in inventory.", singular), nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d %s item(s) in inventory:\n", len(items), singular))

	for i, item := range items {
		id := safeString(item, "_id")
		name := safeString(item, "name")
		b.WriteString(fmt.Sprintf("\n%d. \"%s\" (ID: %s)\n", i+1, name, id))
		b.WriteString(summaryLine(item))
	}

	return b.String(), nil
}

func formatInventoryDetail(raw json.RawMessage, label string, detailFn detailFunc) (string, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	var b strings.Builder
	name := safeString(m, "name")
	id := safeString(m, "_id")
	b.WriteString(fmt.Sprintf("%s: \"%s\" (ID: %s)\n", label, name, id))

	detailFn(&b, m)

	return b.String(), nil
}

// --- Fermentable formatting ---

func formatFermentableSummaryLine(m map[string]any) string {
	parts := make([]string, 0, 4)
	if t := safeString(m, "type"); t != "" {
		parts = append(parts, "Type: "+t)
	}
	if s := safeString(m, "supplier"); s != "" {
		parts = append(parts, "Supplier: "+s)
	}
	if inv, ok := safeFloat(m, "inventory"); ok {
		parts = append(parts, "Stock: "+formatWeight(inv, "kg"))
	}
	if len(parts) == 0 {
		return ""
	}
	return "   " + strings.Join(parts, " | ") + "\n"
}

func formatFermentableDetail(b *strings.Builder, m map[string]any) {
	// Core info
	core := &strings.Builder{}
	writeLine(core, "  ", "Type", safeString(m, "type"))
	writeLine(core, "  ", "Grain Category", safeString(m, "grainCategory"))
	writeLine(core, "  ", "Origin", safeString(m, "origin"))
	writeLine(core, "  ", "Supplier", safeString(m, "supplier"))
	writeLine(core, "  ", "Use", safeString(m, "use"))
	if core.Len() > 0 {
		writeSection(b, "Details")
		b.WriteString(core.String())
	}

	// Properties
	props := &strings.Builder{}
	if v, ok := safeFloat(m, "color"); ok {
		writeLine(props, "  ", "Color", fmt.Sprintf("%.1f SRM", v))
	}
	if v, ok := safeFloat(m, "potential"); ok {
		writeLine(props, "  ", "Potential", formatGravity(v))
	}
	if v, ok := safeFloat(m, "attenuation"); ok {
		writeLine(props, "  ", "Attenuation", formatPercent(v))
	}
	if v, ok := safeFloat(m, "moisture"); ok {
		writeLine(props, "  ", "Moisture", formatPercent(v))
	}
	if v, ok := safeFloat(m, "protein"); ok {
		writeLine(props, "  ", "Protein", formatPercent(v))
	}
	if v, ok := safeFloat(m, "diastaticPower"); ok {
		writeLine(props, "  ", "Diastatic Power", fmt.Sprintf("%.0f Lintner", v))
	}
	if nf, ok := safeBool(m, "notFermentable"); ok && nf {
		writeLine(props, "  ", "Non-Fermentable", "Yes")
	}
	if props.Len() > 0 {
		writeSection(b, "Properties")
		b.WriteString(props.String())
	}

	writeInventorySection(b, m)

	notes := safeString(m, "notes")
	if notes != "" {
		writeSection(b, "Notes")
		b.WriteString("  " + notes + "\n")
	}
}

// --- Hop formatting ---

func formatHopSummaryLine(m map[string]any) string {
	parts := make([]string, 0, 4)
	if v, ok := safeFloat(m, "alpha"); ok {
		parts = append(parts, fmt.Sprintf("Alpha: %.1f%%", v))
	}
	if t := safeString(m, "type"); t != "" {
		parts = append(parts, "Form: "+t)
	}
	if u := safeString(m, "use"); u != "" {
		parts = append(parts, "Use: "+u)
	}
	if inv, ok := safeFloat(m, "inventory"); ok {
		parts = append(parts, fmt.Sprintf("Stock: %.0f g", inv))
	}
	if len(parts) == 0 {
		return ""
	}
	return "   " + strings.Join(parts, " | ") + "\n"
}

func formatHopDetail(b *strings.Builder, m map[string]any) {
	core := &strings.Builder{}
	if v, ok := safeFloat(m, "alpha"); ok {
		writeLine(core, "  ", "Alpha Acid", formatPercent(v))
	}
	if v, ok := safeFloat(m, "beta"); ok {
		writeLine(core, "  ", "Beta Acid", formatPercent(v))
	}
	writeLine(core, "  ", "Form", safeString(m, "type"))
	writeLine(core, "  ", "Use", safeString(m, "use"))
	writeLine(core, "  ", "Origin", safeString(m, "origin"))
	writeLine(core, "  ", "Usage Role", safeString(m, "usage"))
	if core.Len() > 0 {
		writeSection(b, "Details")
		b.WriteString(core.String())
	}

	// Oil profile
	oils := &strings.Builder{}
	if v, ok := safeFloat(m, "oil"); ok {
		writeLine(oils, "  ", "Total Oil", fmt.Sprintf("%.1f ml/100g", v))
	}
	oilFields := []struct{ key, label string }{
		{"myrcene", "Myrcene"}, {"humulene", "Humulene"}, {"caryophyllene", "Caryophyllene"},
		{"farnesene", "Farnesene"}, {"geraniol", "Geraniol"}, {"linalool", "Linalool"},
	}
	for _, f := range oilFields {
		if v, ok := safeFloat(m, f.key); ok {
			writeLine(oils, "  ", f.label, formatPercent(v))
		}
	}
	if oils.Len() > 0 {
		writeSection(b, "Oil Profile")
		b.WriteString(oils.String())
	}

	writeInventorySection(b, m)

	notes := safeString(m, "notes")
	if notes != "" {
		writeSection(b, "Notes")
		b.WriteString("  " + notes + "\n")
	}
}

// --- Misc formatting ---

func formatMiscSummaryLine(m map[string]any) string {
	parts := make([]string, 0, 4)
	if t := safeString(m, "type"); t != "" {
		parts = append(parts, "Type: "+t)
	}
	if u := safeString(m, "use"); u != "" {
		parts = append(parts, "Use: "+u)
	}
	if inv, ok := safeFloat(m, "inventory"); ok {
		parts = append(parts, fmt.Sprintf("Stock: %.1f", inv))
	}
	if len(parts) == 0 {
		return ""
	}
	return "   " + strings.Join(parts, " | ") + "\n"
}

func formatMiscDetail(b *strings.Builder, m map[string]any) {
	core := &strings.Builder{}
	writeLine(core, "  ", "Type", safeString(m, "type"))
	writeLine(core, "  ", "Use", safeString(m, "use"))
	writeLine(core, "  ", "Unit", safeString(m, "unit"))
	if v, ok := safeFloat(m, "concentration"); ok {
		writeLine(core, "  ", "Concentration", formatPercent(v))
	}
	if wa, ok := safeBool(m, "waterAdjustment"); ok && wa {
		writeLine(core, "  ", "Water Adjustment", "Yes")
	}
	if core.Len() > 0 {
		writeSection(b, "Details")
		b.WriteString(core.String())
	}

	writeInventorySection(b, m)

	notes := safeString(m, "notes")
	if notes != "" {
		writeSection(b, "Notes")
		b.WriteString("  " + notes + "\n")
	}
}

// --- Yeast formatting ---

func formatYeastSummaryLine(m map[string]any) string {
	parts := make([]string, 0, 5)
	if t := safeString(m, "type"); t != "" {
		parts = append(parts, "Type: "+t)
	}
	if f := safeString(m, "form"); f != "" {
		parts = append(parts, "Form: "+f)
	}
	if v, ok := safeFloat(m, "attenuation"); ok {
		parts = append(parts, fmt.Sprintf("Atten: %.0f%%", v))
	}
	if lab := safeString(m, "laboratory"); lab != "" {
		parts = append(parts, "Lab: "+lab)
	}
	if inv, ok := safeFloat(m, "inventory"); ok {
		parts = append(parts, fmt.Sprintf("Stock: %.1f", inv))
	}
	if len(parts) == 0 {
		return ""
	}
	return "   " + strings.Join(parts, " | ") + "\n"
}

func formatYeastDetail(b *strings.Builder, m map[string]any) {
	core := &strings.Builder{}
	writeLine(core, "  ", "Type", safeString(m, "type"))
	writeLine(core, "  ", "Form", safeString(m, "form"))
	writeLine(core, "  ", "Laboratory", safeString(m, "laboratory"))
	writeLine(core, "  ", "Product ID", safeString(m, "productId"))
	writeLine(core, "  ", "Flocculation", safeString(m, "flocculation"))
	if core.Len() > 0 {
		writeSection(b, "Details")
		b.WriteString(core.String())
	}

	// Fermentation characteristics
	ferm := &strings.Builder{}
	if v, ok := safeFloat(m, "attenuation"); ok {
		writeLine(ferm, "  ", "Attenuation", formatPercent(v))
	}
	if minA, ok := safeFloat(m, "minAttenuation"); ok {
		if maxA, ok2 := safeFloat(m, "maxAttenuation"); ok2 {
			writeLine(ferm, "  ", "Attenuation Range", fmt.Sprintf("%.0f-%.0f%%", minA, maxA))
		}
	}
	if minT, ok := safeFloat(m, "minTemp"); ok {
		if maxT, ok2 := safeFloat(m, "maxTemp"); ok2 {
			writeLine(ferm, "  ", "Temperature Range", fmt.Sprintf("%s - %s", formatTemp(minT), formatTemp(maxT)))
		}
	}
	if v, ok := safeFloat(m, "maxAbv"); ok {
		writeLine(ferm, "  ", "Max ABV Tolerance", formatPercent(v))
	}
	if ferm.Len() > 0 {
		writeSection(b, "Fermentation")
		b.WriteString(ferm.String())
	}

	bestFor := safeString(m, "bestFor")
	if bestFor != "" {
		writeSection(b, "Best For")
		b.WriteString("  " + bestFor + "\n")
	}

	writeInventorySection(b, m)

	notes := safeString(m, "notes")
	if notes != "" {
		writeSection(b, "Notes")
		b.WriteString("  " + notes + "\n")
	}
}

// --- Shared inventory section ---

func writeInventorySection(b *strings.Builder, m map[string]any) {
	inv := &strings.Builder{}
	if v, ok := safeFloat(m, "inventory"); ok {
		writeLine(inv, "  ", "Current Stock", fmt.Sprintf("%.2f", v))
	}
	if v, ok := safeFloat(m, "costPerAmount"); ok {
		writeLine(inv, "  ", "Cost Per Unit", fmt.Sprintf("%.2f", v))
	}
	if inv.Len() > 0 {
		writeSection(b, "Inventory")
		b.WriteString(inv.String())
	}
}
