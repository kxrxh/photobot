import { handleApiResponse } from "../helpers"
import type { ApiResponse, PaginatedResponse } from "../types"
import { catalogClient as client } from "./catalogClient"
import type {
	CreateProposalParams,
	ListProposalsParams,
	ListWeedParams,
	Proposal,
	ProposalListItem,
	UpdateProposalDraftParams,
	UpdateWeedParams,
	Weed,
	WeedAnalysisObjects,
	WeedDetails,
	WeedListItem,
} from "./types"

export const fetchWeedList = async (
	params?: ListWeedParams
): Promise<PaginatedResponse<WeedListItem>> => {
	const searchParams = new URLSearchParams()

	if (params?.limit) searchParams.set("limit", params.limit.toString())
	if (params?.offset) searchParams.set("offset", params.offset.toString())
	if (params?.name) searchParams.set("name", params.name)

	if (params?.main_group !== undefined) {
		searchParams.set("main_group", params.main_group.toString())
	}
	if (params?.main_subgroup !== undefined) {
		searchParams.set("main_subgroup", params.main_subgroup.toString())
	}
	if (params?.subgroup !== undefined) {
		searchParams.set("subgroup", params.subgroup.toString())
	}
	if (params?.is_quarantine !== undefined) {
		searchParams.set("is_quarantine", params.is_quarantine.toString())
	}
	if (params?.sort_order) searchParams.set("sort_order", params.sort_order)

	const setNum = (k: keyof ListWeedParams) => {
		const v = params?.[k]
		if (v !== undefined && v !== null && !Number.isNaN(v)) {
			searchParams.set(String(k), String(v))
		}
	}

	;(
		[
			"l_min",
			"l_max",
			"w_min",
			"w_max",
			"lw_min",
			"lw_max",
			"h_min",
			"h_max",
			"s_min",
			"s_max",
			"v_min",
			"v_max",
			"r_min",
			"r_max",
			"g_min",
			"g_max",
			"b_min",
			"b_max",
			"brt_min",
			"brt_max",
			"sq_sqcrl_min",
			"sq_sqcrl_max",
		] as (keyof ListWeedParams)[]
	).forEach(setNum)

	const response = await client
		.get(`weeds?${searchParams.toString()}`)
		.json<ApiResponse<PaginatedResponse<WeedListItem>>>()

	return handleApiResponse(response)
}

export const fetchWeedDetails = async (id: number): Promise<WeedDetails> => {
	const response = await client.get(`weeds/${id}/details`).json<ApiResponse<WeedDetails>>()

	return handleApiResponse(response)
}

export const updateWeed = async (id: number, params: UpdateWeedParams): Promise<Weed> => {
	const response = await client.put(`weeds/${id}`, { json: params }).json<ApiResponse<Weed>>()

	return handleApiResponse(response)
}

export const deleteWeed = async (id: number): Promise<void> => {
	await client.delete(`weeds/${id}`)
}

export const fetchWeedAnalysisObjects = async (weedId: number): Promise<WeedAnalysisObjects> => {
	const response = await client
		.get(`weeds/${weedId}/analysis-objects`)
		.json<ApiResponse<WeedAnalysisObjects>>()

	return handleApiResponse(response)
}

export const createProposal = async (params: CreateProposalParams): Promise<Proposal> => {
	const response = await client.post("proposals", { json: params }).json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const uploadProposalImage = async (
	proposalId: number,
	file: File
): Promise<{
	id: number
	pending_weed_id: number
	url: string
	is_primary: boolean
}> => {
	const formData = new FormData()
	formData.append("file", file)

	const response = await client.post(`proposals/${proposalId}/images`, { body: formData }).json<
		ApiResponse<{
			id: number
			pending_weed_id: number
			url: string
			is_primary: boolean
		}>
	>()

	return handleApiResponse(response)
}

export const fetchProposals = async (
	params?: ListProposalsParams
): Promise<PaginatedResponse<ProposalListItem>> => {
	const searchParams = new URLSearchParams()

	if (params?.limit) searchParams.set("limit", params.limit.toString())
	if (params?.offset) searchParams.set("offset", params.offset.toString())
	if (params?.status) searchParams.set("status", params.status)
	if (params?.request_by !== undefined) searchParams.set("request_by", params.request_by.toString())
	if (params?.reviewed_by !== undefined)
		searchParams.set("reviewed_by", params.reviewed_by.toString())
	if (params?.sort_order) searchParams.set("sort_order", params.sort_order)

	const queryString = searchParams.toString()
	const response = await client
		.get(`proposals${queryString ? `?${queryString}` : ""}`)
		.json<ApiResponse<PaginatedResponse<ProposalListItem>>>()

	return handleApiResponse(response)
}

export const fetchProposal = async (id: number): Promise<Proposal> => {
	const response = await client.get(`proposals/${id}`).json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const updateProposalDraft = async (
	id: number,
	params: UpdateProposalDraftParams
): Promise<Proposal> => {
	const response = await client
		.patch(`proposals/${id}`, { json: params })
		.json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const submitProposal = async (id: number): Promise<Proposal> => {
	const response = await client.post(`proposals/${id}/submit`).json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const requestChanges = async (id: number, message: string): Promise<Proposal> => {
	const response = await client
		.post(`proposals/${id}/request-changes`, { json: { message } })
		.json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const rejectProposal = async (id: number, reason: string): Promise<Proposal> => {
	const response = await client
		.post(`proposals/${id}/reject`, { json: { message: reason } })
		.json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const applyProposal = async (id: number, note?: string): Promise<Proposal> => {
	const response = await client
		.post(`proposals/${id}/apply`, { json: { note } })
		.json<ApiResponse<Proposal>>()

	return handleApiResponse(response)
}

export const deleteProposalImage = async (proposalId: number, imageId: number): Promise<void> => {
	await client.delete(`proposals/${proposalId}/images/${imageId}`)
}
