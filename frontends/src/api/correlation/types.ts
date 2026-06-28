export type ParameterGroup = "color" | "geometry" | "median" | "all"

export interface ObjectGroup {
	name: string
	object_ids: number[]
}

export interface CorrelationRequest {
	fractions: ObjectGroup[]
	parameter_groups: ParameterGroup[]
}

export interface Condition {
	attribute: string
	operator: string
	value: number
}

export interface CorrelationBase {
	name: string
	conditions: Condition[]
}

export interface CorrelationWithTest extends CorrelationBase {
	true_positives: number
	false_positives: number
	true_negatives: number
	false_negatives: number
	precision: number
	recall: number
	accuracy: number
	f1_score: number
}
