package cbtcategory

// Only domain phrases with a clear, stable meaning are covered globally.
// Ambiguous labels use real parent context and full-path overrides.json.
var categoryTerms = map[string]struct {
	label string
	bad   string
}{
	"Space Heaters":          {"取暖器", "太空加热器"},
	"Electric Kettles":       {"电水壶", ""},
	"Antique Cash Registers": {"古董收银机", ""},
	"Water Heaters":          {"热水器", ""},
	"Toasters":               {"烤面包机", ""},
	"VCRs":                   {"录像机", ""},
}

func applyCategoryGlossary(staged *snapshot, overrides map[string]string, failed map[string]bool) bool {
	changed := false
	for id, item := range staged.Categories {
		if item.Status != "enabled" || overrides[pathKey(item.Path)] != "" {
			continue
		}
		term, ok := categoryTerms[item.Name]
		if !ok || item.Zh == term.label || (item.Zh != "" && item.Zh != item.Name && item.Zh != term.bad) {
			continue
		}
		changed = true
		item.Zh = term.label
		staged.Categories[id] = item
		delete(failed, id)
		delete(staged.Skipped, id)
	}
	return changed
}
