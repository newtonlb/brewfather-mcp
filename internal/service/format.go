package service

import (
	"fmt"
	"strings"
	"time"
)

func safeString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func safeFloat(m map[string]any, key string) (float64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

func safeBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func safeMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	sub, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return sub
}

func safeSlice(m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.([]any)
	if !ok {
		return nil
	}
	return s
}

func formatDate(ms float64) string {
	t := time.UnixMilli(int64(ms)).UTC()
	return t.Format("2006-01-02")
}

func formatDateTime(ms float64) string {
	t := time.UnixMilli(int64(ms)).UTC()
	return t.Format("2006-01-02 15:04 UTC")
}

func formatGravity(v float64) string {
	return fmt.Sprintf("%.3f SG", v)
}

func formatVolume(v float64) string {
	return fmt.Sprintf("%.1f L", v)
}

func formatTemp(v float64) string {
	return fmt.Sprintf("%.1f °C", v)
}

func formatPercent(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

func formatWeight(v float64, unit string) string {
	if unit == "" {
		unit = "kg"
	}
	return fmt.Sprintf("%.2f %s", v, unit)
}

func formatPH(v float64) string {
	return fmt.Sprintf("%.2f pH", v)
}

// writeLine writes a labeled value to the builder with indentation.
// It does nothing if value is empty.
func writeLine(b *strings.Builder, indent, label, value string) {
	if value == "" {
		return
	}
	b.WriteString(indent)
	b.WriteString(label)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteString("\n")
}

// writeSection writes a section header if the section has content.
func writeSection(b *strings.Builder, title string) {
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString(":\n")
}

func mapSliceToMaps(items []any) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

func formatIngredientSummary(items []any, nameKey, amountKey, amountUnit string) string {
	maps := mapSliceToMaps(items)
	if len(maps) == 0 {
		return ""
	}
	parts := make([]string, 0, len(maps))
	for _, m := range maps {
		name := safeString(m, nameKey)
		if name == "" {
			continue
		}
		if amt, ok := safeFloat(m, amountKey); ok {
			parts = append(parts, fmt.Sprintf("%s (%s)", name, formatWeight(amt, amountUnit)))
		} else {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, ", ")
}

func formatHopSummary(items []any) string {
	maps := mapSliceToMaps(items)
	if len(maps) == 0 {
		return ""
	}
	parts := make([]string, 0, len(maps))
	for _, m := range maps {
		name := safeString(m, "name")
		if name == "" {
			continue
		}
		detail := name
		if amt, ok := safeFloat(m, "amount"); ok {
			detail += fmt.Sprintf(" (%.0fg", amt)
			use := safeString(m, "use")
			if use != "" {
				detail += ", " + use
				if t, ok := safeFloat(m, "time"); ok && t > 0 {
					detail += fmt.Sprintf(" %.0fmin", t)
				}
			}
			detail += ")"
		}
		parts = append(parts, detail)
	}
	return strings.Join(parts, ", ")
}

func formatYeastSummary(items []any) string {
	maps := mapSliceToMaps(items)
	if len(maps) == 0 {
		return ""
	}
	parts := make([]string, 0, len(maps))
	for _, m := range maps {
		name := safeString(m, "name")
		if name == "" {
			continue
		}
		detail := name
		form := safeString(m, "form")
		if form != "" {
			detail += " (" + form + ")"
		}
		parts = append(parts, detail)
	}
	return strings.Join(parts, ", ")
}
