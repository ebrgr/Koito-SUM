// Package worker measures complete image operations in a disposable process.
package worker

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"
)

type Inspect func([]byte) (int, int, error)
type Transform func([]byte, int, int) ([]byte, error)
type Result struct {
	Error     string    `json:"error,omitempty"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Bytes     int       `json:"bytes"`
	Times     []float64 `json:"times_ms"`
	RSS       int64     `json:"peak_rss_bytes"`
	Wall      float64   `json:"wall_ms"`
	Workers   int       `json:"workers"`
	Completed int       `json:"completed"`
}

// Run includes all concurrent operations in the same process and measures native RSS.
func Run(inspect Inspect, transform Transform) {
	input := flag.String("input", "", "fixture")
	output := flag.String("output", "", "output")
	size := flag.Int("size", 300, "zero means source")
	effort := flag.Int("effort", -1, "candidate effort")
	workers := flag.Int("workers", 1, "parallel operations")
	reps := flag.Int("reps", 3, "operations per worker")
	probe := flag.Bool("probe", false, "header admission only")
	flag.Parse()
	result := Result{Workers: *workers}
	defer func() {
		if x := recover(); x != nil {
			result.Error = fmt.Sprint(x)
		}
		var r syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &r)
		result.RSS = r.Maxrss
		if runtime.GOOS != "darwin" {
			result.RSS *= 1024
		}
		json.NewEncoder(os.Stdout).Encode(result)
	}()
	data, err := os.ReadFile(*input)
	if err != nil {
		result.Error = "operation: " + err.Error()
		return
	}
	if len(data) > 10<<20 {
		result.Error = "byte-limit"
		return
	}
	w, h, err := inspect(data)
	if err != nil {
		result.Error = "operation: " + err.Error()
		return
	}
	result.Width = w
	result.Height = h
	if w <= 0 || h <= 0 || w > 10000 || h > 10000 || int64(w)*int64(h) > 40000000 {
		result.Error = "dimension-limit"
		return
	}
	if *probe {
		return
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var out []byte
	start := time.Now()
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < *reps; n++ {
				t := time.Now()
				value, e := transform(data, *size, *effort)
				duration := float64(time.Since(t).Microseconds()) / 1000
				mu.Lock()
				if e != nil {
					result.Error = "transform: " + e.Error()
				} else {
					result.Times = append(result.Times, duration)
					result.Completed++
					out = value
				}
				mu.Unlock()
				if e != nil {
					return
				}
			}
		}()
	}
	wg.Wait()
	result.Wall = float64(time.Since(start).Microseconds()) / 1000
	sort.Float64s(result.Times)
	if result.Error != "" {
		return
	}
	result.Bytes = len(out)
	result.Width, result.Height, err = inspect(out)
	if err != nil {
		result.Error = "operation: " + err.Error()
		return
	}
	if *output != "" {
		if err = os.WriteFile(*output, out, 0600); err != nil {
			result.Error = "operation: " + err.Error()
		}
	}
}
