package analysis

import "encoding/json"

type AnalysisAPIResponse struct {
	Success bool           `json:"success"`
	Result  AnalysisResult `json:"result"`
}

type AnalysisResult struct {
	ID              string          `json:"id"`
	DateTime        string          `json:"date_time"`
	Product         *string         `json:"product,omitempty"`
	UserID          int64           `json:"user_id"`
	Source          *string         `json:"source,omitempty"`
	BotMessage      *string         `json:"bot_message,omitempty"`
	FilesSource     []string        `json:"files_source"`
	FilesOutput     []string        `json:"files_output"`
	FilesSourceURLs []string        `json:"files_source_urls,omitempty"`
	FilesOutputURLs []string        `json:"files_output_urls,omitempty"`
	ScaleMmPixel    *float64        `json:"scale_mm_pixel,omitempty"`
	AnalysisParams  *AnalysisParams `json:"analysis_params,omitempty"`
	Objects         []Object        `json:"objects,omitempty"`

	// TODO: Add global support for these fields.
	Company      *string  `json:"company,omitempty"`
	Sort         *string  `json:"sort,omitempty"`
	Reproduction *string  `json:"reproduction,omitempty"`
	Germination  *float64 `json:"germination,omitempty"`
}

type ChannelStats struct {
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Median *float64 `json:"median,omitempty"`
	Med    *float64 `json:"med,omitempty"`
	Avg    *float64 `json:"avg,omitempty"`
	Stddev *float64 `json:"stddev,omitempty"`
}

// UnmarshalJSON merges "med" into Median when "median" is absent (upstream uses either key).
func (c *ChannelStats) UnmarshalJSON(data []byte) error {
	type raw struct {
		Min    *float64 `json:"min,omitempty"`
		Max    *float64 `json:"max,omitempty"`
		Median *float64 `json:"median,omitempty"`
		Med    *float64 `json:"med,omitempty"`
		Avg    *float64 `json:"avg,omitempty"`
		Stddev *float64 `json:"stddev,omitempty"`
	}
	var x raw
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	c.Min = x.Min
	c.Max = x.Max
	c.Avg = x.Avg
	c.Stddev = x.Stddev
	switch {
	case x.Median != nil:
		c.Median = x.Median
	case x.Med != nil:
		c.Median = x.Med
	default:
		c.Median = nil
	}
	c.Med = nil
	return nil
}

// CoalesceMedian returns the median whether set as Median or legacy Med (e.g. hand-built structs in tests).
func CoalesceMedian(c *ChannelStats) *float64 {
	if c == nil {
		return nil
	}
	if c.Median != nil {
		return c.Median
	}
	return c.Med
}

type AnalysisParams struct {
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
