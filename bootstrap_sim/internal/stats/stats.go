// Package stats provides basic descriptive-statistics helpers used across the
// bootstrap simulation study.
package stats

import (
	"math"
	"sort"
)

// Mean returns the arithmetic mean of xs.  Panics on empty slice.
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		panic("stats.Mean: empty slice")
	}
	var s float64
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}

// Median returns the median of xs.  The input slice is sorted in place.
func Median(xs []float64) float64 {
	n := len(xs)
	if n == 0 {
		panic("stats.Median: empty slice")
	}
	sort.Float64s(xs)
	mid := n / 2
	if n%2 == 1 {
		return xs[mid]
	}
	return (xs[mid-1] + xs[mid]) / 2
}

// StdDev returns the sample standard deviation (Bessel-corrected, ddof=1).
func StdDev(xs []float64) float64 {
	return math.Sqrt(Variance(xs))
}

// Variance returns the sample variance (Bessel-corrected).
func Variance(xs []float64) float64 {
	n := len(xs)
	if n < 2 {
		panic("stats.Variance: need at least 2 observations")
	}
	mu := Mean(xs)
	var ss float64
	for _, v := range xs {
		d := v - mu
		ss += d * d
	}
	return ss / float64(n-1)
}

// Quantile returns the p-th quantile of xs using linear interpolation
// (equivalent to R's type 7).  xs must be sorted ascending.
func Quantile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		panic("stats.Quantile: empty slice")
	}
	h := p * float64(n-1)
	lo := int(math.Floor(h))
	hi := int(math.Ceil(h))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (h-float64(lo))*(sorted[hi]-sorted[lo])
}
