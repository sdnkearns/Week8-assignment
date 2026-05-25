package stats_test

import (
	"math"
	"testing"

	"github.com/example/bootstrap_sim/internal/stats"
)

func TestMean(t *testing.T) {
	got := stats.Mean([]float64{1, 2, 3, 4, 5})
	if math.Abs(got-3.0) > 1e-12 {
		t.Fatalf("Mean: want 3.0, got %v", got)
	}
}

func TestMedianOdd(t *testing.T) {
	got := stats.Median([]float64{5, 1, 3})
	if math.Abs(got-3.0) > 1e-12 {
		t.Fatalf("Median (odd): want 3.0, got %v", got)
	}
}

func TestMedianEven(t *testing.T) {
	got := stats.Median([]float64{1, 2, 3, 4})
	if math.Abs(got-2.5) > 1e-12 {
		t.Fatalf("Median (even): want 2.5, got %v", got)
	}
}

func TestVariance(t *testing.T) {
	// Var of {2,4,4,4,5,5,7,9} = 4.571...
	xs  := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	got := stats.Variance(xs)
	want := 4.571428571428571
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("Variance: want %.6f, got %.6f", want, got)
	}
}

func TestQuantileEdges(t *testing.T) {
	xs := []float64{0, 1, 2, 3, 4}
	if got := stats.Quantile(xs, 0.0); got != 0 {
		t.Errorf("p=0: want 0, got %v", got)
	}
	if got := stats.Quantile(xs, 1.0); got != 4 {
		t.Errorf("p=1: want 4, got %v", got)
	}
}

// ── Benchmarks ────────────────────────────────────────────────────────────────

func BenchmarkMean1000(b *testing.B) {
	xs := make([]float64, 1000)
	for i := range xs {
		xs[i] = float64(i)
	}
	b.ResetTimer()
	for range b.N {
		stats.Mean(xs)
	}
}

func BenchmarkMedian1000(b *testing.B) {
	xs := make([]float64, 1000)
	for i := range xs {
		xs[i] = float64(i % 100)
	}
	b.ResetTimer()
	for range b.N {
		tmp := make([]float64, len(xs))
		copy(tmp, xs)
		stats.Median(tmp)
	}
}

func BenchmarkStdDev1000(b *testing.B) {
	xs := make([]float64, 1000)
	for i := range xs {
		xs[i] = float64(i)
	}
	b.ResetTimer()
	for range b.N {
		stats.StdDev(xs)
	}
}
