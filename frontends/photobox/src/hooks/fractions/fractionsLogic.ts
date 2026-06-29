import type { KalibriObject } from "@/api/analysis/types"
import type {
	ClassificationRule,
	ClassificationRules,
	FractionType,
	KalibriObjectWithSource,
} from "@/hooks/fractions/types"
import { getObjectImageUrl } from "@/utils/image"

export const DEFAULT_FRACTION_ID = "0"

export const DEFAULT_FRACTION: FractionType = {
	id: DEFAULT_FRACTION_ID,
	name: "Прочее",
	objects: [],
	classificationRules: undefined,
}

export const createNewFraction = (): FractionType => ({
	id: crypto.randomUUID(),
	name: "Новая фракция",
	objects: [],
})

export const findObjectInFraction = (
	fraction: FractionType,
	objectId: number
): KalibriObjectWithSource | undefined => fraction.objects.find((obj) => obj.id === objectId)

export const removeObjectsFromFraction = (
	fraction: FractionType,
	objectIdsToRemove: Set<number>
): FractionType => ({
	...fraction,
	objects: fraction.objects.filter((obj) => !objectIdsToRemove.has(obj.id)),
})

export const addUniqueObjectsToFraction = (
	fraction: FractionType,
	newObjects: KalibriObjectWithSource[]
): FractionType => {
	const existingIds = new Set(fraction.objects.map((obj) => obj.id))
	const uniqueObjects = newObjects.filter((obj) => !existingIds.has(obj.id))
	return { ...fraction, objects: [...fraction.objects, ...uniqueObjects] }
}

const matchesRule = (obj: KalibriObject, rule: ClassificationRule): boolean => {
	const { parameter, operator, value } = rule
	const objValue = obj[parameter as keyof KalibriObject] as number
	if (objValue === undefined || objValue === null) return false
	switch (operator) {
		case "<":
			return objValue < value
		case "<=":
			return objValue <= value
		case "==":
			return objValue === value
		case ">=":
			return objValue >= value
		case ">":
			return objValue > value
		case "!=":
			return objValue !== value
		default:
			return false
	}
}

const matchesClassificationRules = (obj: KalibriObject, rules: ClassificationRules): boolean => {
	if (rules.rules.length === 0) return false
	if (rules.logic === "AND") {
		return rules.rules.every((rule) => matchesRule(obj, rule))
	}
	if (rules.logic === "OR") {
		return rules.rules.some((rule) => matchesRule(obj, rule))
	}
	return false
}

export const classifyObjects = (fractions: FractionType[]): FractionType[] => {
	const hasClassificationRules = fractions.some((f) => f.classificationRules)
	if (!hasClassificationRules) return fractions

	const allObjects = fractions.flatMap((f) => f.objects)
	const classifiedFractions = fractions.map((fraction) => ({
		...fraction,
		objects: [] as KalibriObjectWithSource[],
	}))
	const fractionsWithRules = fractions.filter((f) => f.classificationRules)
	const objectsAssigned = new Set<number>()

	for (const fraction of fractionsWithRules) {
		const matchingObjects = allObjects.filter((obj) =>
			fraction.classificationRules
				? matchesClassificationRules(obj, fraction.classificationRules)
				: false
		)
		const targetFraction = classifiedFractions.find((f) => f.id === fraction.id)
		if (targetFraction) {
			targetFraction.objects = matchingObjects
			for (const obj of matchingObjects) {
				objectsAssigned.add(obj.id)
			}
		}
	}

	const otherFraction = classifiedFractions.find((f) => f.id === DEFAULT_FRACTION_ID)
	if (otherFraction) {
		otherFraction.objects = allObjects.filter((obj) => !objectsAssigned.has(obj.id))
	}

	return classifiedFractions
}

export const classifyObjectsSelective = (
	fractions: FractionType[],
	newObjects?: KalibriObjectWithSource[]
): FractionType[] => {
	const hasClassificationRules = fractions.some((f) => f.classificationRules)
	if (!hasClassificationRules) return fractions

	const allObjects = fractions.flatMap((f) => f.objects)
	const manualObjects = allObjects.filter((obj) => obj.classificationState === "manual")
	const autoObjects = allObjects.filter((obj) => obj.classificationState === "auto")
	const objectsToClassify = [...autoObjects, ...(newObjects ?? [])]

	const classifiedFractions = fractions.map((fraction) => ({
		...fraction,
		objects: [] as KalibriObjectWithSource[],
	}))
	const fractionsWithRules = fractions.filter((f) => f.classificationRules)
	const objectsAssigned = new Set<number>()

	for (const fraction of fractionsWithRules) {
		const matchingObjects = objectsToClassify.filter((obj) =>
			fraction.classificationRules
				? matchesClassificationRules(obj, fraction.classificationRules)
				: false
		)
		const targetFraction = classifiedFractions.find((f) => f.id === fraction.id)
		if (targetFraction) {
			targetFraction.objects = matchingObjects
			for (const obj of matchingObjects) {
				objectsAssigned.add(obj.id)
			}
		}
	}

	const otherFraction = classifiedFractions.find((f) => f.id === DEFAULT_FRACTION_ID)
	if (otherFraction) {
		const remainingObjects = objectsToClassify.filter((obj) => !objectsAssigned.has(obj.id))
		otherFraction.objects = [...otherFraction.objects, ...remainingObjects]
	}

	for (const manualObj of manualObjects) {
		const originalFraction = fractions.find((f) => f.objects.some((obj) => obj.id === manualObj.id))
		if (originalFraction) {
			const targetFraction = classifiedFractions.find((f) => f.id === originalFraction.id)
			targetFraction?.objects.push(manualObj)
		} else {
			otherFraction?.objects.push(manualObj)
		}
	}

	return classifiedFractions
}

export const toObjectsWithSource = (
	newObjects: KalibriObject[],
	source: { type: "analysis" | "catalog"; sourceId: string | number }
): KalibriObjectWithSource[] =>
	newObjects.map((obj) => ({
		...obj,
		file: getObjectImageUrl(obj) ?? obj.file ?? "",
		source,
		classificationState: "auto" as const,
	}))

export const countManualObjects = (fractions: FractionType[]): number =>
	fractions.flatMap((f) => f.objects).filter((obj) => obj.classificationState === "manual").length

export const hasManualObjects = (fractions: FractionType[]): boolean =>
	fractions.some((f) => f.objects.some((obj) => obj.classificationState === "manual"))
