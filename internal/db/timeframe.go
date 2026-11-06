package db

import (
	"time"
)

type Timeframe struct {
	Period   Period
	Year     int
	Month    int
	Week     int
	FromUnix int64
	ToUnix   int64
	From     time.Time
	To       time.Time
	Timezone *time.Location
	// Resolved preserves both exact bounds, including a zero start for all_time.
	Resolved bool
}

func TimeframeToTimeRange(tf Timeframe) (t1, t2 time.Time) {
	return TimeframeToTimeRangeAt(tf, time.Now())
}

// TimeframeToTimeRangeAt resolves a timeframe against one captured clock value.
func TimeframeToTimeRangeAt(tf Timeframe, now time.Time) (t1, t2 time.Time) {
	if tf.Resolved {
		return tf.From, tf.To
	}
	loc := tf.Timezone
	if loc == nil {
		loc, _ = time.LoadLocation("UTC")
	}
	// Infer omitted calendar fields using the requested zone; rolling periods keep their existing clock.
	calendarNow := now.In(loc)

	// ---------------------------------------------------------------------
	// 1. Explicit From / To (time.Time) — highest precedence
	// ---------------------------------------------------------------------
	if !tf.From.IsZero() {
		if tf.To.IsZero() {
			return tf.From, now
		}
		return tf.From, tf.To
	}

	// ---------------------------------------------------------------------
	// 2. Unix timestamps
	// ---------------------------------------------------------------------
	if tf.FromUnix != 0 {
		t1 = time.Unix(tf.FromUnix, 0).In(loc)
		if tf.ToUnix == 0 {
			return t1, now
		}
		t2 = time.Unix(tf.ToUnix, 0).In(loc)
		return t1, t2
	}

	// ---------------------------------------------------------------------
	// 3. Derived ranges (Year / Month / Week)
	// ---------------------------------------------------------------------

	// YEAR only
	if tf.Year != 0 && tf.Month == 0 && tf.Week == 0 {
		start := CivilDayStart(time.Date(tf.Year, 1, 1, 12, 0, 0, 0, loc))
		end := CivilDayStart(time.Date(tf.Year+1, 1, 1, 12, 0, 0, 0, loc)).Add(-time.Second)
		return start, end
	}

	// MONTH (+ optional year)
	if tf.Month != 0 {
		year := tf.Year
		if year == 0 {
			year = calendarNow.Year()
			if int(calendarNow.Month()) < tf.Month {
				year--
			}
		}

		start := CivilDayStart(time.Date(year, time.Month(tf.Month), 1, 12, 0, 0, 0, loc))
		end := endOfMonth(year, time.Month(tf.Month), loc)
		return start, end
	}

	// WEEK (+ optional year)
	if tf.Week != 0 {
		year := tf.Year
		if year == 0 {
			var currentWeek int
			year, currentWeek = calendarNow.ISOWeek()
			if currentWeek < tf.Week {
				year--
			}
		}

		// ISO week 1 contains Jan 4
		jan4 := time.Date(year, 1, 4, 12, 0, 0, 0, loc)
		week1Start := startOfWeek(jan4)

		start := CivilDayStart(time.Date(week1Start.Year(), week1Start.Month(), week1Start.Day()+(tf.Week-1)*7, 12, 0, 0, 0, loc))
		end := endOfWeek(start)
		return start, end
	}

	// ---------------------------------------------------------------------
	// 4. Period
	// ---------------------------------------------------------------------

	if !tf.Period.IsZero() {
		return StartTimeFromPeriodAt(tf.Period, now), now
	}

	// ---------------------------------------------------------------------
	// 5. Fallback: empty timeframe → zero values
	// ---------------------------------------------------------------------
	return time.Time{}, time.Time{}
}

func startOfWeek(t time.Time) time.Time {
	// Resolve the first existing instant on the ISO Monday.
	// ISO week: Monday = 1
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return CivilDayStart(time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 12, 0, 0, 0, t.Location()))
}
func endOfWeek(t time.Time) time.Time {
	// The next week's local boundary is not always 168 elapsed hours away.
	monday := startOfWeek(t)
	return CivilDayStart(time.Date(monday.Year(), monday.Month(), monday.Day()+7, 12, 0, 0, 0, monday.Location())).Add(-time.Second)
}
func endOfMonth(year int, month time.Month, loc *time.Location) time.Time {
	// Use the first real instant of the next month, even if midnight is skipped.
	startNextMonth := CivilDayStart(time.Date(year, month+1, 1, 12, 0, 0, 0, loc))
	return startNextMonth.Add(-time.Second)
}
