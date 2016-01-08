package cmd

import (
	"testing"
)

func TestSendApacheRSSMetrics(t *testing.T) {
	matches := GetMatches("go")
	if matches != nil {
		result := SendApacheRSSMetrics(matches)
		if !result {
			t.Error("Did not send Go RSS Metrics.")
		}
	} else {
		t.Error("Didn't find any matches.")
	}
}
