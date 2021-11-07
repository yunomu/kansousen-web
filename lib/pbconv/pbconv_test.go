package pbconv

import (
	"testing"

	"time"
)

func TestParseDateTime(t *testing.T) {
	loc := time.UTC

	if _, err := ParseDateTime("12/34/56", loc); err == nil {
		t.Errorf("`12/34/56` is not date")
	}

	ts, err := ParseDateTime("2020/10/11", loc)
	if err != nil {
		t.Fatalf("ParseDateTime error: %v", err)
	}

	tm := time.Date(2020, time.October, 11, 0, 0, 0, 0, loc).Unix()

	if ts != tm {
		t.Errorf("expected=%d actual=%d", tm, ts)
	}
}

func TestParseDateTimeWithWeek(t *testing.T) {
	loc := time.UTC

	ts, err := ParseDateTime("2020/10/09(金)", loc)
	if err != nil {
		t.Fatalf("ParseDateTime error: %v", err)
	}

	tm := time.Date(2020, time.October, 9, 0, 0, 0, 0, loc).Unix()

	if ts != tm {
		t.Errorf("expected=%d actual=%d", tm, ts)
	}
}
