package distrib_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/example/bootstrap_sim/internal/distrib"
	"github.com/example/bootstrap_sim/internal/stats"
	"github.com/seehuhn/mt19937"
)

const (
	targetMean = 100.0
	targetSD   = 10.0
	tolerance  = 0.10 // 10% relative tolerance for empirical checks
)

func newRNG(seed int64) *rand.Rand {
	mt := mt19937.New()
	mt.Seed(seed)
	return rand.New(mt)
}

func checkMeanSD(t *testing.T, name string, xs []float64) {
	t.Helper()
	m := stats.Mean(xs)
	s := stats.StdDev(xs)
	if math.Abs(m-targetMean)/targetMean > tolerance {
		t.Errorf("%s: empirical mean %.2f too far from %.2f", name, m, targetMean)
	}
	if math.Abs(s-targetSD)/targetSD > tolerance {
		t.Errorf("%s: empirical SD %.2f too far from %.2f", name, s, targetSD)
	}
}

func TestSymmetric(t *testing.T) {
	rng := newRNG(1)
	p   := distrib.Params{Mean: targetMean, SD: targetSD}
	gen := distrib.NewGenerator(distrib.Symmetric, p)
	xs  := gen.Sample(rng, 10_000)
	checkMeanSD(t, "Symmetric", xs)
}

func TestPosSkewed(t *testing.T) {
	rng := newRNG(2)
	p   := distrib.Params{Mean: targetMean, SD: targetSD}
	gen := distrib.NewGenerator(distrib.PosSkewed, p)
	xs  := gen.Sample(rng, 10_000)
	checkMeanSD(t, "PosSkewed", xs)
}

func TestNegSkewed(t *testing.T) {
	rng := newRNG(3)
	p   := distrib.Params{Mean: targetMean, SD: targetSD}
	gen := distrib.NewGenerator(distrib.NegSkewed, p)
	xs  := gen.Sample(rng, 10_000)
	checkMeanSD(t, "NegSkewed", xs)
}

// ── Benchmarks ────────────────────────────────────────────────────────────────

func benchSample(b *testing.B, shape distrib.Shape, n int) {
	b.Helper()
	rng := newRNG(42)
	p   := distrib.Params{Mean: targetMean, SD: targetSD}
	gen := distrib.NewGenerator(shape, p)
	b.ResetTimer()
	for range b.N {
		gen.Sample(rng, n)
	}
}

func BenchmarkSymmetric_n400(b *testing.B)  { benchSample(b, distrib.Symmetric, 400) }
func BenchmarkPosSkewed_n400(b *testing.B)  { benchSample(b, distrib.PosSkewed, 400) }
func BenchmarkNegSkewed_n400(b *testing.B)  { benchSample(b, distrib.NegSkewed, 400) }
