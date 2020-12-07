package pbconv

import (
	"testing"
)

func TestParseDateTime(t *testing.T) {
	if _, err := ParseDateTime("12/34/56"); err == nil {
		t.Errorf("`12/34/56` is not date")
	}

	d, err := ParseDateTime("2020/10/11")
	if err != nil {
		t.Fatalf("ParseDateTime error: %v", err)
	}
	date := d.GetDate()
	if date == nil {
		t.Fatalf("Date is nil")
	}
	if date.Year != 2020 {
		t.Errorf("Year: expected=2020 actual=%v", date.Year)
	}
	if date.Month != 10 {
		t.Errorf("Month: expected=10 actual=%v", date.Month)
	}
	if date.Day != 11 {
		t.Errorf("Day: expected=11 actual=%v", date.Day)
	}

	d, err = ParseDateTime("2020/10/09(金)")
	if err != nil {
		t.Fatalf("ParseDateTime error: %v", err)
	}

	d, err = ParseDateTime("1999/07/15(木) 19:07:12")
	if err != nil {
		t.Fatalf("ParseDateTime error2: %v", err)
	}
	if d.TimeSec == 0 {
		t.Fatalf("ParseDateTime error: time is empty: %v", d)
	}
}
