import type { PaginationParams } from "../types"

export interface ListWeedParams {
	limit?: number
	offset?: number
	name?: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine?: boolean
	sort_order?: "asc" | "desc"
	l_min?: number
	l_max?: number
	w_min?: number
	w_max?: number
	lw_min?: number
	lw_max?: number
	h_min?: number
	h_max?: number
	s_min?: number
	s_max?: number
	v_min?: number
	v_max?: number
	r_min?: number
	r_max?: number
	g_min?: number
	g_max?: number
	b_min?: number
	b_max?: number
	brt_min?: number
	brt_max?: number
	sq_sqcrl_min?: number
	sq_sqcrl_max?: number
}

export interface WeedListItem {
	id: number
	name: string
	length: number
	width: number
	primary_image_url?: string
	created_at: string
	updated_at: string
	latin_name: string
	main_group: string
	main_subgroup?: string
	subgroup?: string
}

export interface CreateWeedParams {
	name: string
	description: string
	statistics?: WeedStatistics
	analysis_ids?: string[]
	excluded_objects?: number[]
	latin_name: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine: boolean
	harmfulness?: string
}

export interface CreatePendingWeedParams extends CreateWeedParams {
	request_id: number
}

export interface UpdateWeedParams {
	name: string
	description: string
	statistics?: WeedStatistics
	analysis_ids?: string[]
	excluded_objects?: number[]
	latin_name: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine: boolean
	harmfulness?: string
}

export interface Weed {
	id: number
	name: string
	description: string
	created_at: string
	updated_at: string
	primary_image_id?: number
	length: number
	width: number
	latin_name?: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine?: boolean
	harmfulness?: string
}

export interface PendingWeed {
	id: number
	request_id: number
	name: string
	description?: string
	created_at: string
	updated_at: string
	latin_name?: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine?: boolean
	harmfulness?: string
}

export interface WeedDetails {
	id: number
	name: string
	description: string
	primary_image_id?: number
	length: number
	width: number
	images: WeedImage[]
	analyses: string[]
	statistics?: WeedStatistics
	created_at: string
	updated_at: string
	latin_name: string
	plant_type: string
	main_group: string
	main_subgroup: string
	subgroup: string
	is_quarantine: boolean
	harmfulness?: string
}

export interface WeedImage {
	id: number
	weed_id: number
	url: string
	is_primary: boolean
}

export interface PendingWeedImage {
	id: number
	pending_weed_id: number
	url: string
	is_primary: boolean
}

export interface WeedStatistics {
	w_avg: number
	w_median: number
	w_min: number
	w_max: number
	l_avg: number
	l_median: number
	l_min: number
	l_max: number
	sq_avg: number
	sq_median: number
	sq_min: number
	sq_max: number
	r_avg: number
	r_median: number
	r_min: number
	r_max: number
	g_avg: number
	g_median: number
	g_min: number
	g_max: number
	b_avg: number
	b_median: number
	b_min: number
	b_max: number
	h_avg: number
	h_median: number
	h_min: number
	h_max: number
	s_avg: number
	s_median: number
	s_min: number
	s_max: number
	v_avg: number
	v_median: number
	v_min: number
	v_max: number
	lw_avg: number
	lw_median: number
	lw_min: number
	lw_max: number
	brt_avg: number
	brt_median: number
	brt_min: number
	brt_max: number
	solid_avg: number
	solid_median: number
	solid_min: number
	solid_max: number
	sq_sqcrl_avg: number
	sq_sqcrl_median: number
	sq_sqcrl_min: number
	sq_sqcrl_max: number
	excluded_objects?: number[]
}

export type WeedAnalysisObjects = {
	id: number
	weed_id: number
	analyses_ids: string[]
	excluded_objects: number[]
}

export type ProposalStatus =
	| "submitted"
	| "changes_requested"
	| "applied"
	| "rejected"
	| "cancelled"

export interface CreateProposalParams {
	target_weed_id?: number

	name?: string
	description?: string
	harmfulness?: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string

	analysis_ids?: string[]
	excluded_objects?: number[]
	statistics?: WeedStatistics
}

export interface UpdateProposalDraftParams {
	name?: string
	description?: string
	harmfulness?: string
	main_group?: string
	main_subgroup?: string
	subgroup?: string

	analysis_ids?: string[]
	excluded_objects?: number[]
	statistics?: WeedStatistics
}

export interface ProposalActionMessageParams {
	message: string
}

export interface ProposalApplyParams {
	note?: string
}

export interface PendingWeedDraft {
	id: number
	name: string
	latin_name?: string
	description?: string
	length?: number
	width?: number
	main_group?: string
	main_subgroup?: string
	subgroup?: string
	is_quarantine: boolean
	harmfulness?: string
}

export interface PendingWeedImageURL {
	id: number
	pending_weed_id: number
	url: string
	is_primary: boolean
	image_key?: string
}

export interface Proposal {
	id: number
	status: ProposalStatus
	/** Internal user_id of the requester */
	request_by: number
	target_weed_id?: number
	reviewed_by?: number
	reviewed_at?: string
	review_notes?: string
	submitted_at?: string
	applied_by?: number
	applied_at?: string
	applied_weed_id?: number
	created_at: string
	updated_at: string

	draft: PendingWeedDraft
	images: PendingWeedImageURL[]
	analyses: string[]
	statistics?: WeedStatistics
}

export interface ListProposalsParams extends PaginationParams {
	status?: ProposalStatus
	request_by?: number
	reviewed_by?: number
	sort_order?: string
}

export interface ProposalListItem {
	id: number
	status: ProposalStatus
	request_by: number
	reviewed_by?: number
	reviewed_at?: string
	review_notes?: string
	submitted_at?: string
	applied_weed_id?: number
	created_at: string
	updated_at: string
	pending_name: string
}
