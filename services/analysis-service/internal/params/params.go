package params

type ChannelStats struct {
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Median *float64 `json:"median,omitempty"`
	Avg    *float64 `json:"avg,omitempty"`
}

type Params struct {
	R              *ChannelStats `json:"r,omitempty"`
	G              *ChannelStats `json:"g,omitempty"`
	B              *ChannelStats `json:"b,omitempty"`
	H              *ChannelStats `json:"h,omitempty"`
	S              *ChannelStats `json:"s,omitempty"`
	V              *ChannelStats `json:"v,omitempty"`
	LabL           *ChannelStats `json:"lab_l,omitempty"`
	LabA           *ChannelStats `json:"lab_a,omitempty"`
	LabB           *ChannelStats `json:"lab_b,omitempty"`
	W              *ChannelStats `json:"w,omitempty"`
	L              *ChannelStats `json:"l,omitempty"`
	T              *ChannelStats `json:"t,omitempty"`
	ColorRhs       *string       `json:"color_rhs,omitempty"`
	MassLiter      *float64      `json:"mass_liter,omitempty"`
	Mass1000       *float64      `json:"mass_1000,omitempty"`
	Location       *string       `json:"location,omitempty"`
	BrokenPercent  *float64      `json:"broken_percent,omitempty"`
	MassPercent    *float64      `json:"mass_percent,omitempty"`
	Mass           *float64      `json:"mass,omitempty"`
	Count50        *float64      `json:"count_50,omitempty"`
	PhotoTimestamp *string       `json:"photo_timestamp,omitempty"`
}
