import type { KalibriObject } from "@/api/analysis/types"

export interface KalibriObjectWithSource extends KalibriObject {
	source: {
		type: "analysis" | "catalog"
		sourceId: string | number
	}
	classificationState: "auto" | "manual" | "pinned"
}

export interface FractionType {
	id: string
	name: string
	objects: KalibriObjectWithSource[]
	classificationRules?: ClassificationRules
}

export interface ClassificationRule {
	parameter: string
	operator: "<" | "<=" | "==" | ">=" | ">" | "!="
	value: number
}

export interface ClassificationRules {
	rules: ClassificationRule[]
	logic: "AND" | "OR"
}
