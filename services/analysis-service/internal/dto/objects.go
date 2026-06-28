package dto

type ObjectRef struct {
	ID           string `json:"id"`
	File         string `json:"file,omitempty"` // deprecated; use PresignedURL
	PresignedURL string `json:"presigned_url,omitempty"`
}

type SearchObjectsRequest struct {
	AnalysisID string  `json:"analysis_id"`
	Objects    []int32 `json:"objects"` // 0-based object indices
}

type ObjectResponse struct {
	ID       int32    `json:"id"`
	IDImage  *string  `json:"id_image,omitempty"`
	Class    *string  `json:"class,omitempty"`
	Geometry *string  `json:"geometry,omitempty"`
	File     *string  `json:"file,omitempty"`
	ImageURL *string  `json:"image_url,omitempty"`
	MH       *float64 `json:"m_h,omitempty"`
	MS       *float64 `json:"m_s,omitempty"`
	MV       *float64 `json:"m_v,omitempty"`
	MR       *float64 `json:"m_r,omitempty"`
	MG       *float64 `json:"m_g,omitempty"`
	MB       *float64 `json:"m_b,omitempty"`
	LAvg     *float64 `json:"l_avg,omitempty"`
	WAvg     *float64 `json:"w_avg,omitempty"`
	BrtAvg   *float64 `json:"brt_avg,omitempty"`
	RAvg     *float64 `json:"r_avg,omitempty"`
	GAvg     *float64 `json:"g_avg,omitempty"`
	BAvg     *float64 `json:"b_avg,omitempty"`
	HAvg     *float64 `json:"h_avg,omitempty"`
	SAvg     *float64 `json:"s_avg,omitempty"`
	VAvg     *float64 `json:"v_avg,omitempty"`
	H        *float64 `json:"h,omitempty"`
	S        *float64 `json:"s,omitempty"`
	V        *float64 `json:"v,omitempty"`
	HM       *float64 `json:"h_m,omitempty"`
	SM       *float64 `json:"s_m,omitempty"`
	VM       *float64 `json:"v_m,omitempty"`
	RM       *float64 `json:"r_m,omitempty"`
	GM       *float64 `json:"g_m,omitempty"`
	BM       *float64 `json:"b_m,omitempty"`
	BrtM     *float64 `json:"brt_m,omitempty"`
	WM       *float64 `json:"w_m,omitempty"`
	LM       *float64 `json:"l_m,omitempty"`
	L        *float64 `json:"l,omitempty"`
	W        *float64 `json:"w,omitempty"`
	LW       *float64 `json:"l_w,omitempty"`
	Pr       *float64 `json:"pr,omitempty"`
	R        *float64 `json:"r,omitempty"`
	G        *float64 `json:"g,omitempty"`
	B        *float64 `json:"b,omitempty"`
	Sq       *float64 `json:"sq,omitempty"`
	Brt      *float64 `json:"brt,omitempty"`
	MinH     *float64 `json:"min_h,omitempty"`
	MinS     *float64 `json:"min_s,omitempty"`
	MinV     *float64 `json:"min_v,omitempty"`
	MaxH     *float64 `json:"max_h,omitempty"`
	MaxS     *float64 `json:"max_s,omitempty"`
	MaxV     *float64 `json:"max_v,omitempty"`
	ColorRhs *string  `json:"color_rhs,omitempty"`
	Solid    *float64 `json:"solid,omitempty"`
	Entropy  *float64 `json:"entropy,omitempty"`
	SqSqcrl  *float64 `json:"sq_sqcrl,omitempty"`
	Hu1      *float64 `json:"hu1,omitempty"`
	Hu2      *float64 `json:"hu2,omitempty"`
	Hu3      *float64 `json:"hu3,omitempty"`
	Hu4      *float64 `json:"hu4,omitempty"`
	Hu5      *float64 `json:"hu5,omitempty"`
	Hu6      *float64 `json:"hu6,omitempty"`
	Mass1000 *float64 `json:"mass_1000,omitempty"`
	Mass     *float64 `json:"mass,omitempty"`
}
