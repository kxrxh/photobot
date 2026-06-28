export function translateClassName(className: string): string {
	const translations: Record<string, string> = {
		whole: "Целые",
		broken: "Ломанные",
	}
	const normalized = className.toLowerCase().trim()
	return translations[normalized] || className
}

export function translateProductName(productName: string): string {
	const translations: Record<string, string> = {
		wheat: "Пшеница",
		barley: "Ячмень",
		oat: "Овес",
		corn: "Кукуруза",
		peas: "Горох",
		rapeseed: "Рапс",
		soybeans: "Соя",
		coffee: "Кофе",
		rice: "Рис",
		seeds: "Подсолнечник",
		seeds_striped: "Подсолнечник (полосатый)",
		kernels: "Ядро подсолнечник",
		general: "Общее",
		fertilizers: "Удобрения",
		flax: "Лен",
		lentils: "Чечевица",
	}
	const normalized = productName.toLowerCase().trim()
	return translations[normalized] || productName
}
