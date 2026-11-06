package db

import (
	"testing"
	"time"
)

// TestImplicitCalendarUsesLocalDateAndISOYear covers cross-zone and ISO year boundaries.
func TestImplicitCalendarUsesLocalDateAndISOYear(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kathmandu")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		tf   Timeframe
		now  time.Time
		want string
	}{
		{"local month", Timeframe{Month: 1, Timezone: loc}, time.Date(2024, 12, 31, 23, 30, 0, 0, time.UTC), "2025-01-01"},
		{"ISO year", Timeframe{Week: 53, Timezone: time.UTC}, time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC), "2020-12-28"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := TimeframeToTimeRangeAt(tc.tf, tc.now)
			if got.Format("2006-01-02") != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

// TestISOWeekUsesISOYearAtYearEnd covers the opposite side of an ISO rollover.
func TestISOWeekUsesISOYearAtYearEnd(t *testing.T) {
	got, _ := TimeframeToTimeRangeAt(Timeframe{Week: 1, Timezone: time.UTC}, time.Date(2018, 12, 31, 12, 0, 0, 0, time.UTC))
	if got.Format("2006-01-02") != "2018-12-31" {
		t.Fatalf("got %s", got)
	}
}

// TestCalendarMonthStartsOnItsCivilDateWhenMidnightIsSkipped covers historic DST.
func TestCalendarMonthStartsOnItsCivilDateWhenMidnightIsSkipped(t *testing.T) {
	loc, err := time.LoadLocation("America/Asuncion")
	if err != nil {
		t.Fatal(err)
	}
	from, to := TimeframeToTimeRangeAt(Timeframe{Year: 2017, Month: 10, Timezone: loc}, time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC))
	if from.Format("2006-01-02 15:04:05") != "2017-10-01 01:00:00" {
		t.Fatalf("start %v", from)
	}
	if to.Format("2006-01-02 15:04:05") != "2017-10-31 23:59:59" {
		t.Fatalf("end %v", to)
	}
}

// TestCivilDayStartFindsFirstRepeatedMidnight covers an actual historical overlap.
func TestCivilDayStartFindsFirstRepeatedMidnight(t *testing.T) {
	loc, err := time.LoadLocation("Africa/Tunis")
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(1977, 9, 23, 22, 0, 0, 0, time.UTC).In(loc)
	second := time.Date(1977, 9, 23, 23, 0, 0, 0, time.UTC).In(loc)
	if first.Format("2006-01-02 15:04") != "1977-09-24 00:00" || second.Format("2006-01-02 15:04") != "1977-09-24 00:00" {
		t.Fatalf("unexpected timezone fixture: %s %s", first, second)
	}
	for _, at := range []time.Time{first, second} {
		got := CivilDayStart(at)
		if !got.Equal(first) {
			t.Fatalf("input %s got %s want %s", at, got, first)
		}
	}
}
