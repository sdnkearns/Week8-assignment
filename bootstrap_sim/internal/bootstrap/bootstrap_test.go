package bootstrap_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/example/bootstrap_sim/internal/bootstrap"
	"github.com/seehuhn/mt19937"
)

func newRNG(seed int64) *rand.Rand {
	mt := mt19937.New()
	mt.Seed(seed)
	return rand.New(mt)
}

// TestBootstrapSEMeanNormal checks that the bootstrap SE of the mean
// is close to the CLT value σ/√n for a normal population.
func TestBootstrapSEMeanNormal(t *testing.T) {
	rng := newRNG(1)
	n   := 200
	// Draw a large sample from N(100, 10²)
	x := make([]float64, n)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
	}

	res := bootstrap.Run(x, 2000, 0.05, rng)
	cltSE := 10.0 / math.Sqrt(float64(n))

	// Expect the bootstrap SE within 25% of the CLT value
	ratio := res.SEMean / cltSE
	if ratio < 0.75 || ratio > 1.25 {
		t.Errorf("SEMean %.4f is too far from CLT SE %.4f (ratio %.2f)", res.SEMean, cltSE, ratio)
	}
}

// TestCIContainsTrue checks that the 95% bootstrap CI contains the true mean
// for a moderately large sample (probabilistic, uses a generous tolerance).
func TestCIContainsTrue(t *testing.T) {
	rng  := newRNG(2)
	trueMean := 100.0
	n   := 500
	x   := make([]float64, n)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + trueMean
	}
	res := bootstrap.Run(x, 1000, 0.05, rng)
	if !(res.CIMeanLo <= trueMean && trueMean <= res.CIMeanHi) {
		t.Errorf("95%% CI [%.2f, %.2f] does not contain true mean %.2f",
			res.CIMeanLo, res.CIMeanHi, trueMean)
	}
}

// ── Benchmarks ────────────────────────────────────────────────────────────────

// BenchmarkBootstrap_n25_B500 benchmarks one bootstrap run with n=25, B=500.
func BenchmarkBootstrap_n25_B500(b *testing.B) {
	rng := newRNG(99)
	x   := make([]float64, 25)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
	}
	b.ResetTimer()
	for range b.N {
		bootstrap.Run(x, 500, 0.05, rng)
	}
}

// BenchmarkBootstrap_n100_B500 benchmarks one bootstrap run with n=100, B=500.
func BenchmarkBootstrap_n100_B500(b *testing.B) {
	rng := newRNG(99)
	x   := make([]float64, 100)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
	}
	b.ResetTimer()
	for range b.N {
		bootstrap.Run(x, 500, 0.05, rng)
	}
}

// BenchmarkBootstrap_n400_B500 benchmarks one bootstrap run with n=400, B=500.
func BenchmarkBootstrap_n400_B500(b *testing.B) {
	rng := newRNG(99)
	x   := make([]float64, 400)
	for i := range x {
		x[i] = rng.NormFloat64()*10 + 100
	}
	b.ResetTimer()
	for range b.N {
		bootstrap.Run(x, 500, 0.05, rng)
	}
}
