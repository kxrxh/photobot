export type ParameterOption = { value: string; label: string }
export type ParameterGroupDef = { label: string; options: ParameterOption[] }

export const PARAMETER_GROUPS: ParameterGroupDef[] = [
	{
		label: "Размеры",
		options: [
			{ value: "L", label: "Длина (мм)" },
			{ value: "W", label: "Ширина (мм)" },
			{ value: "Pr", label: "Периметр (мм)" },
			{ value: "Sq", label: "Площадь (мм²)" },
			{ value: "corners", label: "Кол-во углов контура" },
		],
	},
	{
		label: "Относительные размеры",
		options: [
			{ value: "L/M", label: "Отн. длина к медиане" },
			{ value: "W/M", label: "Отн. ширина к медиане" },
			{ value: "L/W", label: "Отн. длина к ширине" },
			{ value: "Solid", label: "Плотность" },
			{ value: "L/Avg", label: "Отн. длина к среднему" },
			{ value: "W/Avg", label: "Отн. ширина к среднему" },
		],
	},
	{
		label: "RGB",
		options: [
			{ value: "R", label: "Красный" },
			{ value: "G", label: "Зеленый" },
			{ value: "B", label: "Синий" },
			{ value: "Min_R", label: "Мин. красный" },
			{ value: "Min_G", label: "Мин. зеленый" },
			{ value: "Min_B", label: "Мин. синий" },
			{ value: "Max_R", label: "Макс. красный" },
			{ value: "Max_G", label: "Макс. зеленый" },
			{ value: "Max_B", label: "Макс. синий" },
			{ value: "median_R", label: "Медиана красный" },
			{ value: "median_G", label: "Медиана зеленый" },
			{ value: "median_B", label: "Медиана синий" },
		],
	},
	{
		label: "Относительные RGB",
		options: [
			{ value: "R/M", label: "Отн. красный к медиане" },
			{ value: "G/M", label: "Отн. зеленый к медиане" },
			{ value: "B/M", label: "Отн. синий к медиане" },
			{ value: "R/Avg", label: "Отн. красный к среднему" },
			{ value: "G/Avg", label: "Отн. зеленый к среднему" },
			{ value: "B/Avg", label: "Отн. синий к среднему" },
		],
	},
	{
		label: "Яркость",
		options: [
			{ value: "Brt", label: "Яркость" },
			{ value: "Brt/M", label: "Отн. яркость к медиане" },
			{ value: "Brt/Avg", label: "Отн. яркость к среднему" },
		],
	},
	{
		label: "HSV",
		options: [
			{ value: "H", label: "H" },
			{ value: "S", label: "S" },
			{ value: "V", label: "V" },
			{ value: "Min_H", label: "Минимальный H" },
			{ value: "Min_S", label: "Минимальный S" },
			{ value: "Min_V", label: "Минимальный V" },
			{ value: "Max_H", label: "Максимальный H" },
			{ value: "Max_S", label: "Максимальный S" },
			{ value: "Max_V", label: "Максимальный V" },
			{ value: "median_H", label: "Медиана H" },
			{ value: "median_S", label: "Медиана S" },
			{ value: "median_V", label: "Медиана V" },
		],
	},
	{
		label: "Относительные HSV",
		options: [
			{ value: "H/M", label: "Отн. H к медиане" },
			{ value: "S/M", label: "Отн. S к медиане" },
			{ value: "V/M", label: "Отн. V к медиане" },
			{ value: "H/Avg", label: "Отн. H к среднему" },
			{ value: "S/Avg", label: "Отн. S к среднему" },
			{ value: "V/Avg", label: "Отн. V к среднему" },
		],
	},
	{
		label: "Относительная площадь",
		options: [{ value: "Sq/SqCrl", label: "Отн. площадь к кругу" }],
	},
	{
		label: "Инвариантные моменты",
		options: [
			{ value: "Hu1", label: "Hu1" },
			{ value: "Hu2", label: "Hu2" },
			{ value: "Hu3", label: "Hu3" },
			{ value: "Hu4", label: "Hu4" },
			{ value: "Hu5", label: "Hu5" },
			{ value: "Hu6", label: "Hu6" },
		],
	},
]
