package webui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
)

type InterestRow struct {
	From, To    time.Time
	Count       int64
	Last, Empty bool
}

type InterestGraph struct {
	ID, Line, Area   string
	Rows             []InterestRow
	Total            int64
	Empty            bool
	SingleX, SingleY string
}

// buildInterest produces only numeric SVG coordinates. Catalog strings are never
// interpreted as markup, SVG paths or CSS. The table keeps the backend intervals.
func buildInterest(items []db.InterestBucket, loc *time.Location, id string) (*InterestGraph, error) {
	if len(items) > db.MaxInterestBuckets {
		return nil, fmt.Errorf("too many interest buckets")
	}
	if loc == nil {
		loc = time.UTC
	}
	g := &InterestGraph{ID: id, Empty: true}
	var peak int64
	for i, item := range items {
		if item.ListenCount < 0 || item.ListenCount > math.MaxInt64-g.Total || item.BucketEnd.Before(item.BucketStart) {
			return nil, fmt.Errorf("invalid interest bucket")
		}
		last := i == len(items)-1
		if i > 0 && !item.BucketStart.Equal(items[i-1].BucketEnd) {
			return nil, fmt.Errorf("discontinuous interest buckets")
		}
		if !last && item.BucketStart.Equal(item.BucketEnd) && item.ListenCount != 0 {
			return nil, fmt.Errorf("nonempty zero-width interest bucket")
		}
		g.Rows = append(g.Rows, InterestRow{item.BucketStart.In(loc), item.BucketEnd.In(loc), item.ListenCount, last, !last && item.BucketStart.Equal(item.BucketEnd)})
		g.Total += item.ListenCount
		peak = max(peak, item.ListenCount)
	}
	if g.Total == 0 {
		return g, nil
	}
	g.Empty = false
	const left, right, top, bottom = 8.0, 442.0, 12.0, 108.0
	xs, ys := make([]float64, len(items)), make([]float64, len(items))
	for i, item := range items {
		xs[i] = (left + right) / 2
		if len(items) > 1 {
			xs[i] = left + float64(i)*(right-left)/float64(len(items)-1)
		}
		ys[i] = bottom - float64(item.ListenCount)/float64(peak)*(bottom-top)
	}
	g.Line = naturalCurve(xs, ys)
	g.Area = g.Line + fmt.Sprintf(" L%.3f,%.3f L%.3f,%.3f Z", xs[len(xs)-1], bottom, xs[0], bottom)
	if len(items) == 1 {
		g.SingleX = fmt.Sprintf("%.3f", xs[0])
		g.SingleY = fmt.Sprintf("%.3f", ys[0])
	}
	return g, nil
}

// Natural cubic interpolation with zero second derivatives at the endpoints.
// Uniform x spacing yields a tridiagonal system for the interior derivatives.
func naturalCurve(xs, ys []float64) string {
	if len(xs) == 0 {
		return ""
	}
	var path strings.Builder
	fmt.Fprintf(&path, "M%.3f,%.3f", xs[0], ys[0])
	if len(xs) == 1 {
		return path.String()
	}
	n := len(xs)
	step := xs[1] - xs[0]
	second := make([]float64, n)
	if n > 2 {
		diagonal, rhs := make([]float64, n-2), make([]float64, n-2)
		for i := range diagonal {
			diagonal[i] = 4
			rhs[i] = 6 * (ys[i+2] - 2*ys[i+1] + ys[i]) / (step * step)
			if i > 0 {
				factor := 1 / diagonal[i-1]
				diagonal[i] -= factor
				rhs[i] -= factor * rhs[i-1]
			}
		}
		for i := n - 3; i >= 0; i-- {
			second[i+1] = (rhs[i] - second[i+2]) / diagonal[i]
		}
	}
	for i := 0; i < n-1; i++ {
		delta := ys[i+1] - ys[i]
		first := ys[i] + delta/3 - step*step*(2*second[i]+second[i+1])/18
		last := ys[i+1] - delta/3 - step*step*(second[i]+2*second[i+1])/18
		fmt.Fprintf(&path, " C%.3f,%.3f %.3f,%.3f %.3f,%.3f", xs[i]+step/3, first, xs[i+1]-step/3, last, xs[i+1], ys[i+1])
	}
	return path.String()
}
