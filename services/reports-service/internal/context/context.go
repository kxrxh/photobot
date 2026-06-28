package reportcontext

import (
	"strconv"

	"csort.ru/reports-service/internal/charts"
)

// UserIDUnavailable is used when there is no analysis user (empty report skeleton).
const UserIDUnavailable int64 = -1

type ObjectStatSeries5 struct {
	Min, Max, Avg, Med, Asym string
}

type ObjectStatSeries4 struct {
	Min, Max, Avg, Med string
}

type ReportObjectStats struct {
	L, W, T, Sq                                            ObjectStatSeries5
	Lw                                                     ObjectStatSeries4
	R, G, B, H, S, V                                       ObjectStatSeries4
	ChartLenMin, ChartLenMax, ChartWidthMin, ChartWidthMax string
	ChartLenAvg, ChartLenMed, ChartWidthAvg, ChartWidthMed string
}

type ReportContext struct {
	Product              string
	UserID               int64
	Company              string
	Sort                 string
	Reproduction         string
	Germination          string
	Date                 string
	DateTime             string
	AnalysisID           string
	BotLink              string
	ColorRHS             string
	LenAvg, LenStd       string
	WidthAvg, WidthStd   string
	ThickAvg, ThickStd   string
	StatThickMin         string
	StatThickMax         string
	StatThickAvg         string
	StatThickMed         string
	Weight1000Grains     string
	SampleMass           string
	MassLiter            string
	Weight1000GrainsCalc string
	SampleMassCalc       string
	MassLiterCalc        string
	EntitiesFor50Gramm   string
	Count50              string
	BrokenPers           string
	ObjectCount          string
	Stats                ReportObjectStats
	Img2                 []string
	Img2DownloadIndices  []int
}

func (st *ReportObjectStats) ChartRanges() charts.ChartRanges {
	return charts.ChartRanges{
		LenMin: st.ChartLenMin, LenMax: st.ChartLenMax,
		WidthMin: st.ChartWidthMin, WidthMax: st.ChartWidthMax,
	}
}

func (c *ReportContext) ToCSVMap() map[string]any {
	out := make(map[string]any)
	putNE := func(k, v string) {
		if v != "" {
			out[k] = v
		}
	}
	putDash := func(k, v string) {
		if v == "" {
			out[k] = "-"
		} else {
			out[k] = v
		}
	}

	putDash("product", c.Product)
	if c.UserID == UserIDUnavailable {
		putDash("UserID", "-")
	} else {
		putDash("UserID", strconv.FormatInt(c.UserID, 10))
	}
	putDash("company", c.Company)
	putDash("sort", c.Sort)
	putDash("reproduction", c.Reproduction)
	putDash("germination", c.Germination)
	putDash("date", c.Date)
	putDash("dateTime", c.DateTime)
	out["analize"] = c.AnalysisID
	if c.BotLink != "" {
		out["bot_link"] = c.BotLink
	}
	putDash("color_rhs", c.ColorRHS)
	putDash("len", c.LenAvg)
	putNE("len1", c.LenStd)
	putDash("width", c.WidthAvg)
	putNE("width1", c.WidthStd)
	putDash("thick", c.ThickAvg)
	putNE("thick1", c.ThickStd)
	putDash("weight_1000_grains", c.Weight1000Grains)
	putDash("sample_mass", c.SampleMass)
	putDash("mass_liter", c.MassLiter)
	putNE("weight_1000_grains_calculated", c.Weight1000GrainsCalc)
	putNE("sample_mass_calculated", c.SampleMassCalc)
	putNE("mass_liter_calculated", c.MassLiterCalc)
	putDash("entities_for_50_gramm", c.EntitiesFor50Gramm)
	putDash("count_50", c.Count50)
	putNE("broken_pers", c.BrokenPers)
	if c.ObjectCount != "" {
		out["object_"] = c.ObjectCount
	}

	flat5 := func(prefix string, s ObjectStatSeries5) {
		putNE(prefix+"_min", s.Min)
		putNE(prefix+"_max", s.Max)
		putNE(prefix+"_avg", s.Avg)
		putNE(prefix+"_med", s.Med)
		putNE(prefix+"_asym", s.Asym)
	}
	flat4 := func(prefix string, s ObjectStatSeries4) {
		putNE(prefix+"_min", s.Min)
		putNE(prefix+"_max", s.Max)
		putNE(prefix+"_avg", s.Avg)
		putNE(prefix+"_med", s.Med)
	}
	st := c.Stats
	flat5("stat_l", st.L)
	flat5("stat_w", st.W)
	flat5("stat_t", st.T)
	flat5("stat_sq", st.Sq)
	flat4("stat_l_w", st.Lw)
	flat4("stat_r", st.R)
	flat4("stat_g", st.G)
	flat4("stat_b", st.B)
	flat4("stat_h", st.H)
	flat4("stat_s", st.S)
	flat4("stat_v", st.V)
	putNE("stat_len_min", st.ChartLenMin)
	putNE("stat_len_max", st.ChartLenMax)
	putNE("stat_len_avg", st.ChartLenAvg)
	putNE("stat_len_med", st.ChartLenMed)
	putNE("stat_width_min", st.ChartWidthMin)
	putNE("stat_width_max", st.ChartWidthMax)
	putNE("stat_width_avg", st.ChartWidthAvg)
	putNE("stat_width_med", st.ChartWidthMed)

	return out
}
