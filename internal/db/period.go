package db

import (
	"time"
)

// should this be in db package ???

type Period string

const (
	PeriodDay     Period = "day"
	PeriodWeek    Period = "week"
	PeriodMonth   Period = "month"
	PeriodYear    Period = "year"
	PeriodAllTime Period = "all_time"
)

func (p Period) IsZero() bool {
	return p == ""
}

func StartTimeFromPeriod(p Period) time.Time {
	return StartTimeFromPeriodAt(p, time.Now())
}

func StartTimeFromPeriodAt(p Period, now time.Time) time.Time {
	switch p {
	case "day":
		return now.AddDate(0, 0, -1)
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	case "all_time":
		return time.Time{}
	default:
		// default 1 day
		return now.AddDate(0, 0, -1)
	}
}

type StepInterval string

const (
	StepDay     StepInterval = "day"
	StepWeek    StepInterval = "week"
	StepMonth   StepInterval = "month"
	StepYear    StepInterval = "year"
	StepDefault StepInterval = "day"

	DefaultRange int = 12
)

// start is the time of 00:00 at the beginning of opts.Range opts.Steps ago,
// end is the end time of the current opts.Step.
// ISO weeks begin on Monday; range includes the current week and earlier weeks.
// If opts.Year (or opts.Year + opts.Month) is provided, start and end will simply by the start and end times of that year/month.
func ListenActivityOptsToTimes(opts ListenActivityOpts) (start, end time.Time) {
	// An explicit clock keeps one activity response consistent across midnight.
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	return ListenActivityOptsToTimesAt(opts, now)
}

// ListenActivityOptsToTimesAt uses local civil boundaries and an inclusive end.
func ListenActivityOptsToTimesAt(opts ListenActivityOpts, now time.Time) (start, end time.Time) {
	loc := opts.Timezone
	if loc == nil {
		loc = time.UTC
	}
	now = now.In(loc)
	if opts.Range == 0 {
		opts.Range = DefaultRange
	}
	// Noon exists on the DST transition days supported by the activity view;
	// CivilDayStart finds their first representable local instant.
	boundary := func(year int, month time.Month, day int) time.Time {
		return CivilDayStart(time.Date(year, month, day, 12, 0, 0, 0, loc))
	}
	if opts.Year != 0 {
		if opts.Month != 0 {
			return boundary(opts.Year, time.Month(opts.Month), 1), boundary(opts.Year, time.Month(opts.Month)+1, 1).Add(-time.Nanosecond)
		}
		return boundary(opts.Year, 1, 1), boundary(opts.Year+1, 1, 1).Add(-time.Nanosecond)
	}
	switch opts.Step {
	case StepWeek:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 12, 0, 0, 0, loc)
		return CivilDayStart(monday.AddDate(0, 0, -7*(opts.Range-1))), CivilDayStart(monday.AddDate(0, 0, 7)).Add(-time.Nanosecond)
	case StepMonth:
		first := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, loc)
		return CivilDayStart(first.AddDate(0, -(opts.Range - 1), 0)), CivilDayStart(first.AddDate(0, 1, 0)).Add(-time.Nanosecond)
	case StepYear:
		first := time.Date(now.Year(), 1, 1, 12, 0, 0, 0, loc)
		return CivilDayStart(first.AddDate(-(opts.Range - 1), 0, 0)), CivilDayStart(first.AddDate(1, 0, 0)).Add(-time.Nanosecond)
	default:
		today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
		return CivilDayStart(today.AddDate(0, 0, -(opts.Range - 1))), CivilDayStart(today.AddDate(0, 0, 1)).Add(-time.Nanosecond)
	}
}
