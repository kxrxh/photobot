package analysis

import (
	"encoding/json"
	"testing"
)

func TestAnalysisAPIResponseJSONRoundTrip(t *testing.T) {
	raw := `{"success":true,"result":{"id":"a1","date_time":"2026-01-01T00:00:00","user_id":1,"files_source":[],"files_output":[],"objects":[{"id":1,"l":5.5,"mass_1000":1000}]}}`
	var env AnalysisAPIResponse
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !env.Success || env.Result.ID != "a1" || env.Result.UserID != 1 {
		t.Fatalf("result: %+v", env.Result)
	}
	if len(env.Result.Objects) != 1 || env.Result.Objects[0].ID != 1 {
		t.Fatalf("objects: %+v", env.Result.Objects)
	}
	if env.Result.Objects[0].L == nil || *env.Result.Objects[0].L != 5.5 {
		t.Fatalf("object L: %+v", env.Result.Objects[0].L)
	}
}

func TestChannelStatsUnmarshalJSON_medianKey(t *testing.T) {
	raw := `{"min":1,"max":3,"median":2,"avg":2,"stddev":0.5}`
	var cs ChannelStats
	if err := json.Unmarshal([]byte(raw), &cs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cs.Med != nil {
		t.Fatalf("expected med cleared after unmarshal, got %v", cs.Med)
	}
	if cs.Median == nil || *cs.Median != 2 {
		t.Fatalf("median: %+v", cs.Median)
	}
}

func TestChannelStatsUnmarshalJSON_medKeyCoalescesToMedian(t *testing.T) {
	raw := `{"min":1,"max":3,"med":2.5,"avg":2}`
	var cs ChannelStats
	if err := json.Unmarshal([]byte(raw), &cs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if cs.Median == nil || *cs.Median != 2.5 {
		t.Fatalf("want med merged into Median, got %+v", cs.Median)
	}
	if CoalesceMedian(&cs) == nil || *CoalesceMedian(&cs) != 2.5 {
		t.Fatalf("CoalesceMedian: %+v", CoalesceMedian(&cs))
	}
}
