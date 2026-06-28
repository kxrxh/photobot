import type { KalibriObject } from "@/api/analysis/types"

/** Default numeric fields for stats/markup tests; override per case. */
export function kalibriObject(id: number, overrides: Partial<KalibriObject> = {}): KalibriObject {
	return {
		id,
		l: 10,
		w: 5,
		sq: 50,
		r: 100,
		g: 150,
		b: 200,
		h: 180,
		s: 0.5,
		v: 0.8,
		...overrides,
	}
}
