package view

import "strings"

func TranslateClassName(className string) string {
	m := map[string]string{
		"whole":  "Целые",
		"broken": "Ломанные",
		"large":  "Крупные",
		"medium": "Средние",
		"small":  "Мелкие",
	}
	if t, ok := m[strings.ToLower(strings.TrimSpace(className))]; ok {
		return t
	}
	return className
}

func TranslateProductName(name string) string {
	m := map[string]string{
		"wheat":         "Пшеница",
		"barley":        "Ячмень",
		"oat":           "Овес",
		"corn":          "Кукуруза",
		"peas":          "Горох",
		"rapeseed":      "Рапс",
		"soybeans":      "Соя",
		"coffee":        "Кофе",
		"rice":          "Рис",
		"seeds":         "Подсолнечник",
		"kernels":       "Ядро подсолнечник",
		"general":       "Общее",
		"fertilizers":   "Удобрения",
		"flax":          "Лен",
		"lentils":       "Чечевица",
		"seeds_striped": "Подсолнечник полосатый",
	}
	if t, ok := m[strings.ToLower(strings.TrimSpace(name))]; ok {
		return t
	}
	return name
}
