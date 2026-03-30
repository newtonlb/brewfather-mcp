package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"brewfather-mcp/internal/client"
)

type BatchService struct {
	client *client.Client
}

func NewBatchService(c *client.Client) *BatchService {
	return &BatchService{client: c}
}

func (s *BatchService) ListBatches(ctx context.Context, params *client.ListBatchesParams) (string, error) {
	raw, err := s.client.ListBatches(ctx, params)
	if err != nil {
		return "", err
	}

	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(items) == 0 {
		return "No batches found.", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d batch(es):\n", len(items)))

	for i, item := range items {
		id := safeString(item, "_id")
		name := safeString(item, "name")
		status := safeString(item, "status")
		brewer := safeString(item, "brewer")

		b.WriteString(fmt.Sprintf("\n%d. \"%s\" (ID: %s)\n", i+1, name, id))

		line := "   "
		if batchNo, ok := safeFloat(item, "batchNo"); ok {
			line += fmt.Sprintf("Batch #%.0f", batchNo)
		}
		if status != "" {
			if line != "   " {
				line += " | "
			}
			line += "Status: " + status
		}
		if brewDate, ok := safeFloat(item, "brewDate"); ok {
			line += " | Brewed: " + formatDate(brewDate)
		}
		if brewer != "" {
			line += " | Brewer: " + brewer
		}
		b.WriteString(line + "\n")

		recipe := safeMap(item, "recipe")
		if recipe != nil {
			recipeName := safeString(recipe, "name")
			if recipeName != "" {
				b.WriteString("   Recipe: " + recipeName + "\n")
			}
		}
	}

	return b.String(), nil
}

func (s *BatchService) GetBatch(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetBatch(ctx, id, params)
	if err != nil {
		return "", err
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return formatBatchDetail(m), nil
}

func formatBatchDetail(m map[string]any) string {
	var b strings.Builder

	name := safeString(m, "name")
	id := safeString(m, "_id")
	status := safeString(m, "status")
	brewer := safeString(m, "brewer")

	b.WriteString(fmt.Sprintf("Batch: \"%s\" (ID: %s)\n", name, id))

	headerParts := make([]string, 0, 4)
	if batchNo, ok := safeFloat(m, "batchNo"); ok {
		headerParts = append(headerParts, fmt.Sprintf("Batch #%.0f", batchNo))
	}
	if status != "" {
		headerParts = append(headerParts, "Status: "+status)
	}
	if brewer != "" {
		headerParts = append(headerParts, "Brewer: "+brewer)
	}
	if len(headerParts) > 0 {
		b.WriteString(strings.Join(headerParts, " | ") + "\n")
	}

	// Dates
	dates := &strings.Builder{}
	if v, ok := safeFloat(m, "brewDate"); ok {
		writeLine(dates, "  ", "Brew Date", formatDate(v))
	}
	if v, ok := safeFloat(m, "fermentationStartDate"); ok {
		writeLine(dates, "  ", "Fermentation Started", formatDate(v))
	}
	if v, ok := safeFloat(m, "bottlingDate"); ok {
		writeLine(dates, "  ", "Bottling Date", formatDate(v))
	}
	if dates.Len() > 0 {
		writeSection(&b, "Dates")
		b.WriteString(dates.String())
	}

	// Measured Values
	mv := &strings.Builder{}
	if v, ok := safeFloat(m, "measuredOg"); ok {
		writeLine(mv, "  ", "Original Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "measuredFg"); ok {
		writeLine(mv, "  ", "Final Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "measuredMashPh"); ok {
		writeLine(mv, "  ", "Mash pH", formatPH(v))
	}
	if v, ok := safeFloat(m, "measuredFirstWortGravity"); ok {
		writeLine(mv, "  ", "First Wort Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "measuredPreBoilGravity"); ok {
		writeLine(mv, "  ", "Pre-Boil Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "measuredPostBoilGravity"); ok {
		writeLine(mv, "  ", "Post-Boil Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "measuredBoilSize"); ok {
		writeLine(mv, "  ", "Pre-Boil Volume", formatVolume(v))
	}
	if v, ok := safeFloat(m, "measuredKettleSize"); ok {
		writeLine(mv, "  ", "Post-Boil Volume", formatVolume(v))
	}
	if v, ok := safeFloat(m, "measuredBatchSize"); ok {
		writeLine(mv, "  ", "Fermenter Volume", formatVolume(v))
	}
	if v, ok := safeFloat(m, "measuredFermenterTopUp"); ok {
		writeLine(mv, "  ", "Fermenter Top-Up", formatVolume(v))
	}
	if v, ok := safeFloat(m, "measuredBottlingSize"); ok {
		writeLine(mv, "  ", "Bottling Volume", formatVolume(v))
	}
	if mv.Len() > 0 {
		writeSection(&b, "Measured Values")
		b.WriteString(mv.String())
	}

	// Calculated from measurements
	calc := &strings.Builder{}
	if v, ok := safeFloat(m, "measuredAbv"); ok {
		writeLine(calc, "  ", "ABV", formatPercent(v))
	}
	if v, ok := safeFloat(m, "measuredAttenuation"); ok {
		writeLine(calc, "  ", "Attenuation", formatPercent(v))
	}
	if v, ok := safeFloat(m, "measuredEfficiency"); ok {
		writeLine(calc, "  ", "Brewhouse Efficiency", formatPercent(v))
	}
	if v, ok := safeFloat(m, "measuredMashEfficiency"); ok {
		writeLine(calc, "  ", "Mash Efficiency", formatPercent(v))
	}
	if calc.Len() > 0 {
		writeSection(&b, "Measured Calculations")
		b.WriteString(calc.String())
	}

	// Estimated Values
	est := &strings.Builder{}
	parts := make([]string, 0, 6)
	if v, ok := safeFloat(m, "estimatedOg"); ok {
		parts = append(parts, "OG: "+formatGravity(v))
	}
	if v, ok := safeFloat(m, "estimatedFg"); ok {
		parts = append(parts, "FG: "+formatGravity(v))
	}
	if v, ok := safeFloat(m, "estimatedIbu"); ok {
		parts = append(parts, fmt.Sprintf("IBU: %.0f", v))
	}
	if v, ok := safeFloat(m, "estimatedColor"); ok {
		parts = append(parts, fmt.Sprintf("Color: %.1f SRM", v))
	}
	if len(parts) > 0 {
		est.WriteString("  " + strings.Join(parts, " | ") + "\n")
	}
	if est.Len() > 0 {
		writeSection(&b, "Estimated Values")
		b.WriteString(est.String())
	}

	// Carbonation
	carb := &strings.Builder{}
	if v, ok := safeFloat(m, "carbonationTemp"); ok {
		writeLine(carb, "  ", "Carbonation Temp", formatTemp(v))
	}
	ct := safeString(m, "carbonationType")
	if ct != "" {
		writeLine(carb, "  ", "Carbonation Type", ct)
	}
	if v, ok := safeFloat(m, "carbonationForce"); ok {
		writeLine(carb, "  ", "Force Carbonation", fmt.Sprintf("%.1f PSI", v))
	}
	if carb.Len() > 0 {
		writeSection(&b, "Carbonation")
		b.WriteString(carb.String())
	}

	// Recipe summary
	recipe := safeMap(m, "recipe")
	if recipe != nil {
		recipeName := safeString(recipe, "name")
		recipeType := safeString(recipe, "type")
		if recipeName != "" {
			writeSection(&b, "Recipe")
			header := fmt.Sprintf("  \"%s\"", recipeName)
			if recipeType != "" {
				header += " (" + recipeType + ")"
			}
			b.WriteString(header + "\n")

			if ferm := safeSlice(recipe, "fermentables"); len(ferm) > 0 {
				writeLine(&b, "  ", "Fermentables", formatIngredientSummary(ferm, "name", "amount", "kg"))
			}
			if hops := safeSlice(recipe, "hops"); len(hops) > 0 {
				writeLine(&b, "  ", "Hops", formatHopSummary(hops))
			}
			if yeasts := safeSlice(recipe, "yeasts"); len(yeasts) > 0 {
				writeLine(&b, "  ", "Yeast", formatYeastSummary(yeasts))
			}

			style := safeMap(recipe, "style")
			if style != nil {
				styleName := safeString(style, "name")
				if styleName != "" {
					writeLine(&b, "  ", "Style", styleName)
				}
			}
		}
	}

	// Notes
	batchNotes := safeString(m, "batchNotes")
	if batchNotes != "" {
		writeSection(&b, "Notes")
		b.WriteString("  " + batchNotes + "\n")
	}

	tasteNotes := safeString(m, "tasteNotes")
	if tasteNotes != "" {
		if v, ok := safeFloat(m, "tasteRating"); ok {
			writeSection(&b, fmt.Sprintf("Taste Notes (Rating: %.0f/5)", v))
		} else {
			writeSection(&b, "Taste Notes")
		}
		b.WriteString("  " + tasteNotes + "\n")
	}

	return b.String()
}

func (s *BatchService) UpdateBatch(ctx context.Context, id string, body *client.UpdateBatchBody) (string, error) {
	msg, err := s.client.UpdateBatch(ctx, id, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Batch %s updated successfully. Server response: %s", id, msg), nil
}

func (s *BatchService) GetBatchLastReading(ctx context.Context, id string) (string, error) {
	raw, err := s.client.GetBatchLastReading(ctx, id)
	if err != nil {
		return "", err
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return formatReading(m, "Latest Reading"), nil
}

func (s *BatchService) GetBatchReadings(ctx context.Context, id string) (string, error) {
	raw, err := s.client.GetBatchReadings(ctx, id)
	if err != nil {
		return "", err
	}

	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(items) == 0 {
		return "No readings found for this batch.", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d reading(s) for batch:\n", len(items)))

	for i, item := range items {
		b.WriteString(fmt.Sprintf("\n--- Reading %d ---\n", i+1))
		b.WriteString(formatReading(item, ""))
	}

	return b.String(), nil
}

func formatReading(m map[string]any, title string) string {
	var b strings.Builder

	if title != "" {
		b.WriteString(title + ":\n")
	}

	if t, ok := safeFloat(m, "time"); ok {
		writeLine(&b, "  ", "Time", formatDateTime(t))
	}
	readingType := safeString(m, "type")
	if readingType != "" {
		writeLine(&b, "  ", "Source", readingType)
	}
	deviceID := safeString(m, "id")
	if deviceID != "" {
		writeLine(&b, "  ", "Device ID", deviceID)
	}
	if v, ok := safeFloat(m, "sg"); ok {
		writeLine(&b, "  ", "Specific Gravity", formatGravity(v))
	}
	if v, ok := safeFloat(m, "temp"); ok {
		writeLine(&b, "  ", "Temperature", formatTemp(v))
	}
	if v, ok := safeFloat(m, "ph"); ok {
		writeLine(&b, "  ", "pH", formatPH(v))
	}
	if v, ok := safeFloat(m, "pressure"); ok {
		writeLine(&b, "  ", "Pressure", fmt.Sprintf("%.2f", v))
	}
	if v, ok := safeFloat(m, "angle"); ok {
		writeLine(&b, "  ", "Angle", fmt.Sprintf("%.1f°", v))
	}
	if v, ok := safeFloat(m, "battery"); ok {
		writeLine(&b, "  ", "Battery", fmt.Sprintf("%.2f V", v))
	}
	if v, ok := safeFloat(m, "rssi"); ok {
		writeLine(&b, "  ", "Signal (RSSI)", fmt.Sprintf("%.0f dBm", v))
	}
	comment := safeString(m, "comment")
	if comment != "" {
		writeLine(&b, "  ", "Comment", comment)
	}

	return b.String()
}

func (s *BatchService) GetBatchBrewTracker(ctx context.Context, id string) (string, error) {
	raw, err := s.client.GetBatchBrewTracker(ctx, id)
	if err != nil {
		return "", err
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(m) == 0 {
		return "Brew tracker is not active for this batch.", nil
	}

	var b strings.Builder

	trackerID := safeString(m, "_id")
	b.WriteString(fmt.Sprintf("Brew Tracker (ID: %s)\n", trackerID))

	if enabled, ok := safeBool(m, "enabled"); ok {
		writeLine(&b, "  ", "Enabled", fmt.Sprintf("%v", enabled))
	}
	if completed, ok := safeBool(m, "completed"); ok {
		writeLine(&b, "  ", "Completed", fmt.Sprintf("%v", completed))
	}
	if stage, ok := safeFloat(m, "stage"); ok {
		writeLine(&b, "  ", "Current Stage", fmt.Sprintf("%.0f", stage))
	}

	stages := safeSlice(m, "stages")
	if len(stages) > 0 {
		writeSection(&b, "Stages")
		for i, raw := range stages {
			stage, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			name := safeString(stage, "name")
			stageType := safeString(stage, "type")
			b.WriteString(fmt.Sprintf("\n  %d. %s (%s)\n", i+1, name, stageType))

			if v, ok := safeFloat(stage, "duration"); ok {
				mins := v / 60
				writeLine(&b, "     ", "Duration", fmt.Sprintf("%.0f min", mins))
			}
			if paused, ok := safeBool(stage, "paused"); ok && paused {
				writeLine(&b, "     ", "Status", "Paused")
			}

			steps := safeSlice(stage, "steps")
			if len(steps) > 0 {
				for _, stepRaw := range steps {
					step, ok := stepRaw.(map[string]any)
					if !ok {
						continue
					}
					stepName := safeString(step, "name")
					stepType := safeString(step, "type")
					if stepName != "" {
						line := "     - " + stepName
						if stepType != "" {
							line += " [" + stepType + "]"
						}
						if v, ok := safeFloat(step, "value"); ok {
							line += fmt.Sprintf(" @ %s", formatTemp(v))
						}
						if v, ok := safeFloat(step, "duration"); ok {
							line += fmt.Sprintf(" for %.0f min", v/60)
						}
						b.WriteString(line + "\n")
					}
				}
			}
		}
	}

	return b.String(), nil
}
