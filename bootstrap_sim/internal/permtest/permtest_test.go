package permtest_test

import (
	"math/rand"
	"testing"

	"github.com/example/bootstrap_sim/internal/permtest"
	"github.com/seehuhn/mt19937"
)

func newRNG(seed int64) *rand.Rand {
	mt := mt19937.New()
	mt.Seed(seed)
	return rand.New(mt)
}

// TestPermTestSameDistribution verifies that when both samples come from the
// same distribution the p-value is not systematically small.
func TestPermTestSameDistribution(t *testing.T) {
	rng := newRNG(7)
	rejectCount := 0
	iters := 200
	for range iters {
		x := make([]float64, 50)
		y := make([]float64, 50)
		for i := range x {
			x[i] = rng.NormFloat64()*10 + 100
			y[i] = rng.NormFloat64()*10 + 100
		}
		r := permtest.Run(x, y, 500, rng)
		if r.PValue < 0.05 {
			rejectCount++
		}
	}
	rate := float64(rejectCount) / float64(iters)
	// Under H₀ expect ~5% rejection; allow wide tolerance.
	if rate > 0.15 {
		t.Errorf("rejection rate %.2f%% too high under H₀ (same distributions)", rate*100)
	}
}

// ── Benchmarks ────────────────────────────────────────────────────────────────

func BenchmarkPermTest_n50_reps500(b *testing.B) {
	rng := newRNG(11)
	x   := make([]float64, 50)
	y   := make([]float64, 50)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
		y[i] = rng.NormFloat64()*10 + 102
	}
	b.ResetTimer()
	for range b.N {
		permtest.Run(x, y, 500, rng)
	}
}

func BenchmarkPermTest_n200_reps1000(b *testing.B) {
	rng := newRNG(11)
	x   := make([]float64, 200)
	y   := make([]float64, 200)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
		y[i] = rng.NormFloat64()*10 + 102
	}
	b.ResetTimer()
	for range b.N {
		permtest.Run(x, y, 1000, rng)
	}
}
