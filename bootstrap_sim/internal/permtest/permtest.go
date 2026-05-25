// Package permtest implements a two-sample permutation test for equality of
// means.  The test statistic is the observed difference in group means; the
// p-value is two-sided.
package permtest

import (
	"math"
	"math/rand"

	"github.com/example/bootstrap_sim/internal/stats"
)

// Result holds the outcome of a single permutation test.
type Result struct {
	ObsDiff float64 // observed difference in means (x̄ - ȳ)
	PValue  float64 // two-sided p-value
	Reps    int     // number of permutation replicates used
}

// Run performs a two-sample permutation test.
//
// Parameters:
//   - x, y : observed samples from the two groups
//   - reps  : number of permutation replicates
//   - rng   : caller-owned rand.Rand (MT19937-backed)
func Run(x, y []float64, reps int, rng *rand.Rand) Result {
	nx := len(x)
	combined := make([]float64, nx+len(y))
	copy(combined, x)
	copy(combined[nx:], y)
	nTotal := len(combined)

	obsDiff := stats.Mean(x) - stats.Mean(y)

	exceeds := 0
	perm := make([]float64, nTotal)
	copy(perm, combined)

	for range reps {
		// Fisher-Yates shuffle of perm.
		for i := nTotal - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			perm[i], perm[j] = perm[j], perm[i]
		}
		permDiff := stats.Mean(perm[:nx]) - stats.Mean(perm[nx:])
		if math.Abs(permDiff) >= math.Abs(obsDiff) {
			exceeds++
		}
	}

	return Result{
		ObsDiff: obsDiff,
		PValue:  float64(exceeds) / float64(reps),
		Reps:    reps,
	}
}
