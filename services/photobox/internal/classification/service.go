package classification

var (
	mainGroupMap = map[string]string{
		"MONOCOT": "Однодольные",
		"DICOT":   "Двудольные",
		"SPORE":   "Споровые",
	}

	mainSubgroupMap = map[string]string{
		"ANNUAL":    "Малолетние",
		"PERENNIAL": "Многолетние",
	}

	subgroupMap = map[string]string{
		"ANNUAL_SPRING_EARLY":     "Однолетние яровые ранние",
		"EPHEMERAL":               "Эфемеры",
		"ANNUAL_SPRING_LATE":      "Однолетние яровые поздние",
		"ANNUAL_WINTERING":        "Однолетние зимующие",
		"ANNUAL_WINTER":           "Однолетние озимые",
		"BIENNIAL":                "Двулетние",
		"STEM_PARASITE":           "Паразиты стеблевые",
		"ROOT_PARASITE":           "Паразиты корневые",
		"HEMIPARASITE":            "Полупаразиты",
		"FIBROUS_ROOT":            "Кистекорневые",
		"TAPROOT":                 "Стержнекорневые",
		"RHIZOME":                 "Корневищные",
		"TUBEROUS":                "Клубневые",
		"BULBOUS":                 "Луковичные",
		"CREEPING_SHOOTS":         "С ползучими побегами",
		"ROOT_SUCKER":             "Корнеотпрысковые",
		"SPORE_PERENNIAL_RHIZOME": "Корневищные (Споровые)",
	}

	hierarchy = map[string]map[string][]string{
		"MONOCOT": {
			"ANNUAL": {
				"ANNUAL_SPRING_EARLY", "EPHEMERAL", "ANNUAL_SPRING_LATE", "ANNUAL_WINTERING",
				"ANNUAL_WINTER", "BIENNIAL", "STEM_PARASITE", "ROOT_PARASITE", "HEMIPARASITE",
			},
			"PERENNIAL": {
				"FIBROUS_ROOT", "RHIZOME", "TUBEROUS", "BULBOUS", "CREEPING_SHOOTS",
			},
		},
		"DICOT": {
			"ANNUAL": {
				"ANNUAL_SPRING_EARLY", "EPHEMERAL", "ANNUAL_SPRING_LATE", "ANNUAL_WINTERING",
				"ANNUAL_WINTER", "BIENNIAL", "STEM_PARASITE", "ROOT_PARASITE", "HEMIPARASITE",
			},
			"PERENNIAL": {
				"FIBROUS_ROOT", "TAPROOT", "RHIZOME", "ROOT_SUCKER", "TUBEROUS", "BULBOUS",
				"CREEPING_SHOOTS",
			},
		},
		"SPORE": {
			"PERENNIAL": {"SPORE_PERENNIAL_RHIZOME"},
		},
	}
)

func MainGroupMap() map[string]string              { return mainGroupMap }
func MainSubgroupMap() map[string]string           { return mainSubgroupMap }
func SubgroupMap() map[string]string               { return subgroupMap }
func HierarchyMap() map[string]map[string][]string { return hierarchy }

// MapMainGroup, MapMainSubgroup, and MapSubgroup return labels for stored codes; unknown codes are unchanged.
func MapMainGroup(code string) string {
	if v, ok := mainGroupMap[code]; ok {
		return v
	}
	return code
}

func MapMainSubgroup(code string) string {
	if v, ok := mainSubgroupMap[code]; ok {
		return v
	}
	return code
}

func MapSubgroup(code string) string {
	if v, ok := subgroupMap[code]; ok {
		return v
	}
	return code
}
