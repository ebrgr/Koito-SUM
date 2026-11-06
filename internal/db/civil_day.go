package db

import "time"

// CivilDayKey represents a local calendar date as a stable UTC ordinal.
func CivilDayKey(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// CivilDayStart finds the first representable second of t's local calendar date.
func CivilDayStart(t time.Time) time.Time {
	loc := t.Location()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	if CivilDayKey(midnight).Equal(CivilDayKey(t)) && CivilDayKey(midnight.Add(-time.Second)).Before(CivilDayKey(t)) {
		return midnight
	}
	// A DST transition can skip local midnight. Find the date boundary without
	// assuming that the transition occurs on an hour or minute boundary.
	target := CivilDayKey(t)
	lo, hi := t.Add(-48*time.Hour).Unix(), t.Unix()
	for lo < hi {
		mid := lo + (hi-lo)/2
		if CivilDayKey(time.Unix(mid, 0).In(loc)).Before(target) {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return time.Unix(lo, 0).In(loc)
}
