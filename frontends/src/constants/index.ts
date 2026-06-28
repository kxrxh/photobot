const rawBase = import.meta.env.VITE_BASE_PATH ?? "/"
export const BASE_PATH = rawBase.endsWith("/") ? rawBase : `${rawBase}/`

export {
	PARAMETER_GROUPS,
	type ParameterGroupDef,
	type ParameterOption,
} from "./parameterGroups"
