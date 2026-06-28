package calc

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/numutil"
	"csort.ru/reports-service/internal/statutil"
)

type RepresentativeCard struct {
	ObjectID     string
	Value        float64
	ImageDataURL string
}

type RepresentativeGroup struct {
	ClassName       string
	TotalObjects    int
	Representatives []RepresentativeCard
}

type RepresentativeOptions struct {
	PerClassLimit   int
	LowerPercentile float64
	UpperPercentile float64
}

type repScored struct {
	id   string
	dist float64
}

func objectIDFromObject(obj analysis.Object) string {
	return strconv.FormatInt(int64(obj.ID), 10)
}

func CalculateTypicalRepresentativesByClass(
	objects []analysis.Object,
	opt RepresentativeOptions,
) []RepresentativeGroup {
	if len(objects) == 0 {
		return nil
	}
	perClass := opt.PerClassLimit
	if perClass <= 0 {
		perClass = 3
	}
	lo := opt.LowerPercentile
	if lo == 0 {
		lo = 0.25
	}
	hi := opt.UpperPercentile
	if hi == 0 {
		hi = 0.75
	}
	byClass := map[string][]analysis.Object{}
	for _, item := range objects {
		cn := "unclassified"
		if item.Class != nil {
			if s := strings.TrimSpace(*item.Class); s != "" {
				cn = s
			}
		}
		byClass[cn] = append(byClass[cn], item)
	}
	var groups []RepresentativeGroup
	for className, classObjects := range byClass {
		if len(classObjects) == 0 {
			continue
		}
		center := map[string]float64{}
		dims := []string{"l", "w", "sq", "l_w", "mass_1000"}
		for _, dim := range dims {
			var vals []float64
			for _, it := range classObjects {
				v := NumericFieldValue(it, dim)
				if v != nil {
					vals = append(vals, *v)
				}
			}
			if len(vals) == 0 {
				continue
			}
			sort.Float64s(vals)
			center[dim] = statutil.MedianSorted(vals)
		}
		var scoredList []repScored
		for _, it := range classObjects {
			id := objectIDFromObject(it)
			if id == "" {
				continue
			}
			d := repDistanceToCenter(it, center)
			if !numutil.IsFinite(d) {
				continue
			}
			scoredList = append(scoredList, repScored{id: id, dist: d})
		}
		if len(scoredList) == 0 {
			continue
		}
		sort.Slice(
			scoredList,
			func(i, j int) bool { return scoredList[i].dist < scoredList[j].dist },
		)
		distVals := make([]float64, len(scoredList))
		for i := range scoredList {
			distVals[i] = scoredList[i].dist
		}
		qStart := LinearPercentileInSorted(distVals, lo)
		qEnd := LinearPercentileInSorted(distVals, hi)
		var pool []repScored
		for _, s := range scoredList {
			if s.dist >= qStart && s.dist <= qEnd {
				pool = append(pool, s)
			}
		}
		if len(pool) == 0 {
			pool = scoredList
		}
		anchors := []float64{0.25, 0.5, 0.75}
		used := map[string]bool{}
		var chosen []RepresentativeCard
		for _, a := range anchors {
			if idx := indexAtProportionInPool(pool, a); idx >= 0 && idx < len(pool) {
				s := pool[idx]
				if used[s.id] {
					continue
				}
				used[s.id] = true
				chosen = append(
					chosen,
					RepresentativeCard{ObjectID: s.id, Value: numutil.Round3(s.dist)},
				)
				if len(chosen) >= perClass {
					break
				}
			}
		}
		for _, s := range pool {
			if len(chosen) >= perClass {
				break
			}
			if used[s.id] {
				continue
			}
			used[s.id] = true
			chosen = append(
				chosen,
				RepresentativeCard{ObjectID: s.id, Value: numutil.Round3(s.dist)},
			)
		}
		if len(chosen) == 0 {
			continue
		}
		groups = append(groups, RepresentativeGroup{
			ClassName:       className,
			TotalObjects:    len(classObjects),
			Representatives: chosen,
		})
	}
	return groups
}

func LinearPercentileInSorted(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	p = math.Min(1, math.Max(0, p))
	idx := p * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (sorted[hi]-sorted[lo])*(idx-float64(lo))
}

func indexAtProportionInPool(pool []repScored, p float64) int {
	if len(pool) == 0 {
		return -1
	}
	p = math.Min(1, math.Max(0, p))
	return int(math.Round(p * float64(len(pool)-1)))
}

func repDistanceToCenter(item analysis.Object, center map[string]float64) float64 {
	dims := []string{"l", "w", "sq", "l_w", "mass_1000"}
	var acc float64
	var used int
	for _, dim := range dims {
		v := NumericFieldValue(item, dim)
		t, ok := center[dim]
		if v == nil || !ok || math.Abs(t) < 1e-12 {
			continue
		}
		d := (*v - t) / math.Abs(t)
		acc += d * d
		used++
	}
	if used == 0 {
		return math.Inf(1)
	}
	return math.Sqrt(acc / float64(used))
}
