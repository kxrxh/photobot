export interface ChannelStats {
	min?: number
	max?: number
	median?: number
	avg?: number
	stddev?: number
}

export interface AnalysisParams {
	r?: ChannelStats
	g?: ChannelStats
	b?: ChannelStats
	h?: ChannelStats
	s?: ChannelStats
	v?: ChannelStats
	lab_l?: ChannelStats
	lab_a?: ChannelStats
	lab_b?: ChannelStats
	w?: ChannelStats
	l?: ChannelStats
	t?: ChannelStats
	color_rhs?: string
	mass_liter?: number
	mass_1000?: number
	location?: string
	broken_percent?: number
	mass_percent?: number
	mass?: number
	photo_timestamp?: string
	count_50?: number
}

export interface Analysis {
	id: string
	date_time: string
	product?: string
	user_id: number
	source?: string
	bot_message?: string
	files_source: string[]
	files_output: string[]
	files_source_urls?: string[]
	files_output_urls?: string[]
	scale_mm_pixel?: number
	analysis_params?: AnalysisParams
	// Deprecated: use analysis_params?.l, analysis_params?.w, etc. Kept for backward compat.
	l?: ChannelStats
	w?: ChannelStats
	t?: ChannelStats
	r?: ChannelStats
	g?: ChannelStats
	b?: ChannelStats
	h?: ChannelStats
	s?: ChannelStats
	v?: ChannelStats
	lab_l?: ChannelStats
	lab_a?: ChannelStats
	lab_b?: ChannelStats
}

export interface AnalysisWithObjects extends Analysis {
	objects: KalibriObject[]
}

export interface KalibriObject {
	id: number
	file?: string
	image_url?: string
	class?: string
	geometry?: string
	m_h?: number
	m_s?: number
	m_v?: number
	m_r?: number
	m_g?: number
	m_b?: number
	l_avg?: number
	w_avg?: number
	brt_avg?: number
	r_avg?: number
	g_avg?: number
	b_avg?: number
	h_avg?: number
	s_avg?: number
	v_avg?: number
	h?: number
	s?: number
	v?: number
	h_m?: number
	s_m?: number
	v_m?: number
	r_m?: number
	g_m?: number
	b_m?: number
	brt_m?: number
	w_m?: number
	l_m?: number
	l?: number
	w?: number
	l_w?: number
	pr?: number
	sq?: number
	brt?: number
	r?: number
	g?: number
	b?: number
	solid?: number
	min_h?: number
	min_s?: number
	min_v?: number
	max_h?: number
	max_s?: number
	max_v?: number
	entropy?: number
	id_image?: number
	color_rhs?: string
	sq_sqcrl?: number
	hu1?: number
	hu2?: number
	hu3?: number
	hu4?: number
	hu5?: number
	hu6?: number
	mass_1000?: number
	t?: number
}

export interface GetAnalysesParams {
	limit?: number
	offset?: number
	product?: string
	id?: string
	sort_by?: "date_time" | "id" | "product"
	sort_order?: "asc" | "desc"
}

export interface SimpleObject {
	id: string
	file?: string
	presigned_url?: string
}

export interface CreateAnalysisResponse {
	request_id: string
}

export interface ConfirmRequestRequest {
	request_id: string
	excluded_object_ids: string[]
}

export interface ReportFileResponse {
	success: boolean
	error?: string
	data?: {
		analysisId: string
		fileType: string
		content: Blob
		size: number
	}
}

export interface AnalysisRequest {
	id: string
	user_id: string
	product: string
	status: "created" | "processing" | "waiting_for_confirmation" | "completed" | "failed"
	temp_id?: string
	error_message?: string
	created_at: string
	updated_at: string
}
