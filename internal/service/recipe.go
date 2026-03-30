package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"brewfather-mcp/internal/client"
)

type RecipeService struct {
	client *client.Client
}

func NewRecipeService(c *client.Client) *RecipeService {
	return &RecipeService{client: c}
}

func (s *RecipeService) ListRecipes(ctx context.Context, params *client.ListRecipesParams) (string, error) {
	raw, err := s.client.ListRecipes(ctx, params)
	if err != nil {
		return "", err
	}

	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(items) == 0 {
		return "No recipes found.", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d recipe(s):\n", len(items)))

	for i, item := range items {
		id := safeString(item, "_id")
		name := safeString(item, "name")
		author := safeString(item, "author")
		recipeType := safeString(item, "type")

		b.WriteString(fmt.Sprintf("\n%d. \"%s\" (ID: %s)\n", i+1, name, id))

		parts := make([]string, 0, 4)
		if recipeType != "" {
			parts = append(parts, "Type: "+recipeType)
		}
		if author != "" {
			parts = append(parts, "Author: "+author)
		}

		equip := safeMap(item, "equipment")
		if equip != nil {
			equipName := safeString(equip, "name")
			if equipName != "" {
				parts = append(parts, "Equipment: "+equipName)
			}
		}

		style := safeMap(item, "style")
		if style != nil {
			styleName := safeString(style, "name")
			if styleName != "" {
				parts = append(parts, "Style: "+styleName)
			}
		}

		if len(parts) > 0 {
			b.WriteString("   " + strings.Join(parts, " | ") + "\n")
		}
	}

	return b.String(), nil
}

func (s *RecipeService) GetRecipe(ctx context.Context, id string, params *client.GetItemParams) (string, error) {
	raw, err := s.client.GetRecipe(ctx, id, params)
	if err != nil {
		return "", err
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return formatRecipeDetail(m), nil
}

func formatRecipeDetail(m map[string]any) string {
	var b strings.Builder

	name := safeString(m, "name")
	id := safeString(m, "_id")
	author := safeString(m, "author")
	recipeType := safeString(m, "type")

	b.WriteString(fmt.Sprintf("Recipe: \"%s\" (ID: %s)\n", name, id))

	headerParts := make([]string, 0, 3)
	if recipeType != "" {
		headerParts = append(headerParts, "Type: "+recipeType)
	}
	if author != "" {
		headerParts = append(headerParts, "Author: "+author)
	}
	if len(headerParts) > 0 {
		b.WriteString(strings.Join(headerParts, " | ") + "\n")
	}

	teaser := safeString(m, "teaser")
	if teaser != "" {
		b.WriteString(teaser + "\n")
	}

	// Style
	style := safeMap(m, "style")
	if style != nil {
		styleName := safeString(style, "name")
		category := safeString(style, "category")
		guide := safeString(style, "styleGuide")
		if styleName != "" {
			writeSection(&b, "Style")
			line := "  " + styleName
			if category != "" {
				line += " (" + category + ")"
			}
			if guide != "" {
				line += " - " + guide
			}
			b.WriteString(line + "\n")

			styleParts := make([]string, 0)
			if v, ok := safeFloat(style, "ogMin"); ok {
				if v2, ok2 := safeFloat(style, "ogMax"); ok2 {
					styleParts = append(styleParts, fmt.Sprintf("OG: %.3f-%.3f", v, v2))
				}
			}
			if v, ok := safeFloat(style, "fgMin"); ok {
				if v2, ok2 := safeFloat(style, "fgMax"); ok2 {
					styleParts = append(styleParts, fmt.Sprintf("FG: %.3f-%.3f", v, v2))
				}
			}
			if v, ok := safeFloat(style, "ibuMin"); ok {
				if v2, ok2 := safeFloat(style, "ibuMax"); ok2 {
					styleParts = append(styleParts, fmt.Sprintf("IBU: %.0f-%.0f", v, v2))
				}
			}
			if v, ok := safeFloat(style, "abvMin"); ok {
				if v2, ok2 := safeFloat(style, "abvMax"); ok2 {
					styleParts = append(styleParts, fmt.Sprintf("ABV: %.1f-%.1f%%", v, v2))
				}
			}
			if v, ok := safeFloat(style, "colorMin"); ok {
				if v2, ok2 := safeFloat(style, "colorMax"); ok2 {
					styleParts = append(styleParts, fmt.Sprintf("Color: %.0f-%.0f SRM", v, v2))
				}
			}
			if len(styleParts) > 0 {
				b.WriteString("  Guidelines: " + strings.Join(styleParts, " | ") + "\n")
			}
		}
	}

	// Volumes and Efficiency
	vol := &strings.Builder{}
	if v, ok := safeFloat(m, "batchSize"); ok {
		writeLine(vol, "  ", "Batch Size", formatVolume(v))
	}
	if v, ok := safeFloat(m, "boilSize"); ok {
		writeLine(vol, "  ", "Pre-Boil Volume", formatVolume(v))
	}
	if v, ok := safeFloat(m, "boilTime"); ok {
		writeLine(vol, "  ", "Boil Time", fmt.Sprintf("%.0f min", v))
	}
	if v, ok := safeFloat(m, "efficiency"); ok {
		writeLine(vol, "  ", "Efficiency", formatPercent(v))
	}
	if vol.Len() > 0 {
		writeSection(&b, "Volumes & Efficiency")
		b.WriteString(vol.String())
	}

	// Gravity, Color, Bitterness
	stats := &strings.Builder{}
	statParts := make([]string, 0, 6)
	if v, ok := safeFloat(m, "og"); ok {
		statParts = append(statParts, "OG: "+formatGravity(v))
	}
	if v, ok := safeFloat(m, "fg"); ok {
		statParts = append(statParts, "FG: "+formatGravity(v))
	}
	if v, ok := safeFloat(m, "abv"); ok {
		statParts = append(statParts, "ABV: "+formatPercent(v))
	}
	if v, ok := safeFloat(m, "ibu"); ok {
		statParts = append(statParts, fmt.Sprintf("IBU: %.0f", v))
	}
	if v, ok := safeFloat(m, "color"); ok {
		statParts = append(statParts, fmt.Sprintf("Color: %.1f SRM", v))
	}
	if v, ok := safeFloat(m, "buGuRatio"); ok {
		statParts = append(statParts, fmt.Sprintf("BU:GU: %.2f", v))
	}
	if len(statParts) > 0 {
		stats.WriteString("  " + strings.Join(statParts, " | ") + "\n")
	}
	if v, ok := safeFloat(m, "carbonation"); ok {
		writeLine(stats, "  ", "Carbonation", fmt.Sprintf("%.1f volumes CO2", v))
	}
	if stats.Len() > 0 {
		writeSection(&b, "Stats")
		b.WriteString(stats.String())
	}

	// Ingredients
	hasIngredients := false
	ingredients := &strings.Builder{}
	if ferm := safeSlice(m, "fermentables"); len(ferm) > 0 {
		hasIngredients = true
		writeLine(ingredients, "  ", "Fermentables", formatIngredientSummary(ferm, "name", "amount", "kg"))
		if total, ok := safeFloat(m, "fermentablesTotalAmount"); ok {
			writeLine(ingredients, "    ", "Total", formatWeight(total, "kg"))
		}
	}
	if hops := safeSlice(m, "hops"); len(hops) > 0 {
		hasIngredients = true
		writeLine(ingredients, "  ", "Hops", formatHopSummary(hops))
		if total, ok := safeFloat(m, "hopsTotalAmount"); ok {
			writeLine(ingredients, "    ", "Total", fmt.Sprintf("%.0f g", total))
		}
	}
	if yeasts := safeSlice(m, "yeasts"); len(yeasts) > 0 {
		hasIngredients = true
		writeLine(ingredients, "  ", "Yeast", formatYeastSummary(yeasts))
	}
	if miscs := safeSlice(m, "miscs"); len(miscs) > 0 {
		hasIngredients = true
		writeLine(ingredients, "  ", "Misc", formatIngredientSummary(miscs, "name", "amount", ""))
	}
	if hasIngredients {
		writeSection(&b, "Ingredients")
		b.WriteString(ingredients.String())
	}

	// Mash Profile
	mash := safeMap(m, "mash")
	if mash != nil {
		mashName := safeString(mash, "name")
		if mashName != "" {
			writeSection(&b, "Mash Profile")
			b.WriteString(fmt.Sprintf("  \"%s\"\n", mashName))
			if v, ok := safeFloat(mash, "ph"); ok {
				writeLine(&b, "  ", "Target pH", formatPH(v))
			}
			steps := safeSlice(mash, "steps")
			for _, stepRaw := range steps {
				step, ok := stepRaw.(map[string]any)
				if !ok {
					continue
				}
				stepName := safeString(step, "name")
				stepType := safeString(step, "type")
				line := "  - " + stepName
				if stepType != "" {
					line += " (" + stepType + ")"
				}
				if v, ok := safeFloat(step, "stepTemp"); ok {
					line += " @ " + formatTemp(v)
				}
				if v, ok := safeFloat(step, "stepTime"); ok {
					line += fmt.Sprintf(" for %.0f min", v)
				}
				b.WriteString(line + "\n")
			}
		}
	}

	// Fermentation Profile
	ferm := safeMap(m, "fermentation")
	if ferm != nil {
		fermName := safeString(ferm, "name")
		if fermName != "" {
			writeSection(&b, "Fermentation Profile")
			b.WriteString(fmt.Sprintf("  \"%s\"\n", fermName))
			steps := safeSlice(ferm, "steps")
			for _, stepRaw := range steps {
				step, ok := stepRaw.(map[string]any)
				if !ok {
					continue
				}
				stepName := safeString(step, "name")
				stepType := safeString(step, "type")
				line := "  - " + stepName
				if stepType != "" {
					line += " (" + stepType + ")"
				}
				if v, ok := safeFloat(step, "stepTemp"); ok {
					line += " @ " + formatTemp(v)
				}
				if v, ok := safeFloat(step, "stepTime"); ok {
					line += fmt.Sprintf(" for %.0f days", v)
				}
				b.WriteString(line + "\n")
			}
		}
	}

	// Equipment
	equip := safeMap(m, "equipment")
	if equip != nil {
		equipName := safeString(equip, "name")
		if equipName != "" {
			writeSection(&b, "Equipment")
			b.WriteString(fmt.Sprintf("  \"%s\"\n", equipName))
			if v, ok := safeFloat(equip, "batchSize"); ok {
				writeLine(&b, "  ", "Batch Size", formatVolume(v))
			}
			if v, ok := safeFloat(equip, "efficiency"); ok {
				writeLine(&b, "  ", "Efficiency", formatPercent(v))
			}
			if v, ok := safeFloat(equip, "boilTime"); ok {
				writeLine(&b, "  ", "Boil Time", fmt.Sprintf("%.0f min", v))
			}
		}
	}

	// Notes
	notes := safeString(m, "notes")
	if notes != "" {
		writeSection(&b, "Notes")
		b.WriteString("  " + notes + "\n")
	}

	return b.String()
}
