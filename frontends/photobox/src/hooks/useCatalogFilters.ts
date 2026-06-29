import { useState } from "react"

export interface CatalogFilters {
	sort_order: "asc" | "desc"
	main_group: string | undefined
	main_subgroup: string | undefined
	subgroup: string | undefined
	is_quarantine: boolean | undefined
	l_min: number | undefined
	l_max: number | undefined
	w_min: number | undefined
	w_max: number | undefined
	lw_min: number | undefined
	lw_max: number | undefined
	h_min: number | undefined
	h_max: number | undefined
	s_min: number | undefined
	s_max: number | undefined
	v_min: number | undefined
	v_max: number | undefined
	r_min: number | undefined
	r_max: number | undefined
	g_min: number | undefined
	g_max: number | undefined
	b_min: number | undefined
	b_max: number | undefined
	brt_min: number | undefined
	brt_max: number | undefined
	sq_sqcrl_min: number | undefined
	sq_sqcrl_max: number | undefined
}

export const defaultFilters: CatalogFilters = {
	sort_order: "desc",
	main_group: undefined,
	main_subgroup: undefined,
	subgroup: undefined,
	is_quarantine: undefined,
	l_min: undefined,
	l_max: undefined,
	w_min: undefined,
	w_max: undefined,
	lw_min: undefined,
	lw_max: undefined,
	h_min: undefined,
	h_max: undefined,
	s_min: undefined,
	s_max: undefined,
	v_min: undefined,
	v_max: undefined,
	r_min: undefined,
	r_max: undefined,
	g_min: undefined,
	g_max: undefined,
	b_min: undefined,
	b_max: undefined,
	brt_min: undefined,
	brt_max: undefined,
	sq_sqcrl_min: undefined,
	sq_sqcrl_max: undefined,
}

export const useCatalogFilters = () => {
	const [filters, setFilters] = useState<CatalogFilters>(defaultFilters)
	const [searchName, setSearchName] = useState("")

	const updateFilter = (name: string, value: string | boolean | number | undefined) => {
		setFilters((prevFilters) => ({
			...prevFilters,
			[name]: value,
		}))
	}

	const clearFilters = () => {
		setFilters(defaultFilters)
		setSearchName("")
	}

	const updateSearchName = (value: string) => {
		setSearchName(value)
	}

	return {
		filters,
		searchName,
		updateFilter,
		clearFilters,
		updateSearchName,
	}
}
