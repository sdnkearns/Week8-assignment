// Package bootstrap implements one-sample bootstrap resampling.
//
// Design notes
//   - All resampling uses the caller-supplied *rand.Rand (backed by MT19937)
//     so there is no hidden global state.
//   - Result structs are value types; no heap allocation per bootstrap round.
package bootstrap

import (
	"math/rand"
	"sort"

	"github.com/example/bootstrap_sim/internal/stats"
)

// Result holds the bootstrap estimates for a single sample.
type Result struct {
	SEMean   float64 // bootstrap SE of the mean
	SEMedian float64 // bootstrap SE of the median
	CIMeanLo float64 // lower bound of the percentile CI for the mean
	CIMeanHi float64 // upper bound of the percentile CI for the mean
	CIMedLo  float64 // lower bound of the percentile CI for the median
	CIMedHi  float64 // upper bound of the percentile CI for the median
}

// Run performs B bootstrap resamples of x and returns the aggregate Result.
//
// Parameters:
//   - x     : original sample
//   - B     : number of bootstrap resamples
//   - alpha : significance level for the percentile CI (e.g. 0.05 → 95% CI)
//   - rng   : caller-owned rand.Rand (MT19937-backed)
func Run(x []float64, B int, alpha float64, rng *rand.Rand) Result {
	n := len(x)
	buf := make([]float64, n) // reusable resample buffer

	bootMeans   := make([]float64, B)
	bootMedians := make([]float64, B)

	for b := range B {
		// Sample with replacement into buf.
		for i := range n {
			buf[i] = x[rng.Intn(n)]
		}

		// Copy before median sorts in place.
		tmp := make([]float64, n)
		copy(tmp, buf)

		bootMeans[b]   = stats.Mean(buf)
		bootMedians[b] = stats.Median(tmp) // Median sorts tmp
	}

	sort.Float64s(bootMeans)
	sort.Float64s(bootMedians)

	lo, hi := alpha/2, 1-alpha/2

	return Result{
		SEMean:   stats.StdDev(bootMeans),
		SEMedian: stats.StdDev(bootMedians),
		CIMeanLo: stats.Quantile(bootMeans, lo),
		CIMeanHi: stats.Quantile(bootMeans, hi),
		CIMedLo:  stats.Quantile(bootMedians, lo),
		CIMedHi:  stats.Quantile(bootMedians, hi),
	}
}
