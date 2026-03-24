package reflection

import (
	"testing"
	"time"
)

func TestUntilNextHour_Daily(t *testing.T) {
	d := untilNextHour(2)
	if d <= 0 || d > 24*time.Hour {
		t.Errorf("unexpected duration: %v", d)
	}
}

func TestUntilNextHour_Hourly(t *testing.T) {
	d := untilNextHour(-1)
	if d <= 0 || d > time.Hour {
		t.Errorf("unexpected duration: %v", d)
	}
}
