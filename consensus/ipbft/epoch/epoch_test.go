package epoch

import (
	"testing"
	"time"
)

func TestEstimateEpoch(t *testing.T) {
	timeA := time.Now()
	timeB := timeA.Unix()
	timeStr := timeA.String()
	t.Logf("now: %v, %v", timeB, timeStr)

	formatTimeStr := "2020-06-23 09:51:38.397502 +0800 CST m=+10323.270024761"
	parse, e := time.Parse("", formatTimeStr)
	if e == nil {
		t.Logf("time: %v", parse)
	} else {
		t.Errorf("parse error: %v", e)
	}

	timeC := time.Now().UnixNano()
	t.Logf("time c %v", timeC)
	t.Logf("time c %v", timeC)
}
