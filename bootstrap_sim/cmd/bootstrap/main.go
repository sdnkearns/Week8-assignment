// Command bootstrap runs the full bootstrap and simulation-based inference
// study across three distribution shapes (symmetric, positively skewed, and
// negatively skewed).
//
// Usage:
//
//	go run ./cmd/bootstrap [flags]
//
// Flags:
//
//	-B         int   Bootstrap resamples per sample      (default 500)
//	-iter      int   Simulation iterations per cell      (default 200)
//	-permReps  int   Permutation test replicates         (default 1000)
//	-seed      int   MT19937 global seed                 (default 42)
//	-ci        float Confidence level for bootstrap CIs  (default 0.95)
//	-json      bool  Emit JSON-formatted log lines       (default false)
//	-debug     bool  Enable debug-level logging          (default false)
//	-cpuprofile file Write CPU profile to file
//	-memprofile file Write memory profile to file
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"text/tabwriter"
	"time"

	"github.com/example/bootstrap_sim/internal/bootstrap"
	"github.com/example/bootstrap_sim/internal/distrib"
	"github.com/example/bootstrap_sim/internal/logging"
	"github.com/example/bootstrap_sim/internal/permtest"
	"github.com/example/bootstrap_sim/internal/stats"
	"github.com/seehuhn/mt19937"
)

// ── CLI flags ─────────────────────────────────────────────────────────────────

var (
	fB          = flag.Int("B", 500, "bootstrap resamples per sample")
	fIter       = flag.Int("iter", 200, "simulation iterations per (distribution, n) cell")
	fPermReps   = flag.Int("permReps", 1000, "permutation test replicates")
	fSeed       = flag.Int64("seed", 42, "MT19937 seed")
	fCI         = flag.Float64("ci", 0.95, "confidence level for bootstrap CIs")
	fJSON       = flag.Bool("json", false, "emit JSON-formatted log lines")
	fDebug      = flag.Bool("debug", false, "enable debug-level logging")
	fCPUProfile = flag.String("cpuprofile", "", "write CPU profile to this file")
	fMemProfile = flag.String("memprofile", "", "write memory profile to this file")
)

// ── Study parameters ──────────────────────────────────────────────────────────

var studySampleSizes = []int{25, 100, 225, 400}

const (
	targetMean = 100.0
	targetSD   = 10.0
)

// ── cellResult accumulates per-(distribution, n) statistics ──────────────────

type cellResult struct {
	distribution string
	n            int
	// SE estimates
	cltSEMean     float64
	simSEMean     float64 // sd of sample means across iterations
	avgBootSEMean float64
	avgBootSEMed  float64
	// CI coverage
	ciMeanCoverage float64
	ciMedCoverage  float64
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	// ── Logging setup ─────────────────────────────────────────────────────────
	level := logging.LevelInfo
	if *fDebug {
		level = logging.LevelDebug
	}
	logging.Init(level, os.Stdout, *fJSON)

	logging.Info("Study starting",
		slog.Int("B", *fB),
		slog.Int("iter", *fIter),
		slog.Int("permReps", *fPermReps),
		slog.Int64("seed", *fSeed),
		slog.Float64("ci", *fCI),
	)

	// ── CPU profiling ─────────────────────────────────────────────────────────
	if *fCPUProfile != "" {
		f, err := os.Create(*fCPUProfile)
		if err != nil {
			logging.Error("cannot create CPU profile file", slog.Any("err", err))
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			logging.Error("cannot start CPU profile", slog.Any("err", err))
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
		logging.Info("CPU profiling enabled", slog.String("file", *fCPUProfile))
	}

	// ── RNG: MT19937 ─────────────────────────────────────────────────────────
	mt  := mt19937.New()
	mt.Seed(*fSeed)
	rng := rand.New(mt)

	alpha := 1.0 - *fCI

	// Distribution generators (all share same theoretical mean/SD)
	p := distrib.Params{Mean: targetMean, SD: targetSD}
	generators := []distrib.Generator{
		distrib.NewGenerator(distrib.PosSkewed,  p),
		distrib.NewGenerator(distrib.Symmetric,  p),
		distrib.NewGenerator(distrib.NegSkewed,  p),
	}

	// ── Print header ──────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("  Bootstrap & Simulation-Based Inference Study (Go)")
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Printf("  Bootstrap resamples  B   : %d\n", *fB)
	fmt.Printf("  Simulation iterations   : %d\n", *fIter)
	fmt.Printf("  Sample sizes            : %v\n", studySampleSizes)
	fmt.Printf("  CI level                : %.0f%%\n", *fCI*100)
	fmt.Printf("  Permutation replicates  : %d\n", *fPermReps)
	fmt.Printf("  MT19937 seed            : %d\n", *fSeed)
	fmt.Println()

	// ── Main study loop ───────────────────────────────────────────────────────
	var allResults []cellResult
	studyStart := time.Now()

	for _, gen := range generators {
		fmt.Printf("──────────────────────────────────────────────────────────\n")
		fmt.Printf("  Distribution: %s\n", gen.Shape)
		fmt.Printf("──────────────────────────────────────────────────────────\n")

		logging.Info("Starting distribution",
			slog.String("distribution", string(gen.Shape)))

		for _, n := range studySampleSizes {
			cellStart := time.Now()

			sampleMeans   := make([]float64, *fIter)
			bootSEMeans   := make([]float64, *fIter)
			bootSEMeds    := make([]float64, *fIter)
			ciMeanHits    := 0
			ciMedHits     := 0

			for iter := range *fIter {
				x    := gen.Sample(rng, n)
				sampleMeans[iter] = stats.Mean(x)

				res := bootstrap.Run(x, *fB, alpha, rng)
				bootSEMeans[iter] = res.SEMean
				bootSEMeds[iter]  = res.SEMedian

				if res.CIMeanLo <= targetMean && targetMean <= res.CIMeanHi {
					ciMeanHits++
				}
				if res.CIMedLo <= targetMean && targetMean <= res.CIMedHi {
					ciMedHits++
				}

				logging.Debug("iteration complete",
					slog.String("dist", string(gen.Shape)),
					slog.Int("n", n),
					slog.Int("iter", iter),
					slog.Float64("sampleMean", sampleMeans[iter]),
					slog.Float64("bootSEMean", res.SEMean),
				)
			}

			cltSE     := targetSD / math.Sqrt(float64(n))
			simSE     := stats.StdDev(sampleMeans)
			avgBSEMean := stats.Mean(bootSEMeans)
			avgBSEMed  := stats.Mean(bootSEMeds)
			covMean   := float64(ciMeanHits) / float64(*fIter)
			covMed    := float64(ciMedHits)  / float64(*fIter)

			elapsed := time.Since(cellStart)
			logging.Info("cell complete",
				slog.String("dist", string(gen.Shape)),
				slog.Int("n", n),
				slog.Float64("cltSE", cltSE),
				slog.Float64("simSE", simSE),
				slog.Float64("avgBootSEMean", avgBSEMean),
				slog.Float64("avgBootSEMed", avgBSEMed),
				slog.Float64("ciMeanCoverage", covMean),
				slog.Float64("ciMedCoverage", covMed),
				slog.Duration("elapsed", elapsed),
			)

			fmt.Printf("\n  n = %d\n", n)
			fmt.Printf("    CLT SE (mean)               : %.3f\n", cltSE)
			fmt.Printf("    Simulation SE (mean)         : %.3f\n", simSE)
			fmt.Printf("    Avg bootstrap SE (mean)      : %.3f\n", avgBSEMean)
			fmt.Printf("    Avg bootstrap SE (median)    : %.3f\n", avgBSEMed)
			fmt.Printf("    Bootstrap CI coverage (mean)  : %.1f%%\n", covMean*100)
			fmt.Printf("    Bootstrap CI coverage (median): %.1f%%\n", covMed*100)
			fmt.Printf("    Wall time                    : %s\n", elapsed.Round(time.Millisecond))

			allResults = append(allResults, cellResult{
				distribution:   string(gen.Shape),
				n:              n,
				cltSEMean:      cltSE,
				simSEMean:      simSE,
				avgBootSEMean:  avgBSEMean,
				avgBootSEMed:   avgBSEMed,
				ciMeanCoverage: covMean,
				ciMedCoverage:  covMed,
			})
		}
		fmt.Println()
	}

	// ── Permutation tests ─────────────────────────────────────────────────────
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("  Simulation-Based Permutation Tests")
	fmt.Println("  H₀: PosSkewed mean == NegSkewed mean  (true by construction)")
	fmt.Printf("  Expected rejection rate ≈ %.0f%% under H₀\n", alpha*100)
	fmt.Println("════════════════════════════════════════════════════════════")

	posGen := distrib.NewGenerator(distrib.PosSkewed, p)
	negGen := distrib.NewGenerator(distrib.NegSkewed, p)

	for _, n := range []int{25, 100} {
		rejectCount := 0
		for range *fIter {
			x := posGen.Sample(rng, n)
			y := negGen.Sample(rng, n)
			r := permtest.Run(x, y, *fPermReps, rng)
			if r.PValue < alpha {
				rejectCount++
			}
		}
		rate := float64(rejectCount) / float64(*fIter)
		fmt.Printf("\n  n = %3d  |  Rejection rate: %.1f%%  (expected ≈ %.0f%% under H₀)\n",
			n, rate*100, alpha*100)
		logging.Info("permutation test summary",
			slog.Int("n", n),
			slog.Float64("rejectionRate", rate),
			slog.Float64("alpha", alpha))
	}

	// Illustrative single large-sample test
	fmt.Println()
	x400 := posGen.Sample(rng, 400)
	y400 := negGen.Sample(rng, 400)
	bigTest := permtest.Run(x400, y400, *fPermReps, rng)
	fmt.Printf("  Illustrative single test (n=400):\n")
	fmt.Printf("    Observed diff in means : %.3f\n", bigTest.ObsDiff)
	fmt.Printf("    Two-sided p-value       : %.4f\n", bigTest.PValue)

	// ── Summary tables ────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println("  Summary: SE Estimates Across Shapes and Sample Sizes")
	fmt.Println("════════════════════════════════════════════════════════════")

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Distribution\tn\tCLT_SE\tSim_SE\tBoot_SE\tMed_SE")
	fmt.Fprintln(tw, "────────────────────────────────────────\t────\t──────\t──────\t───────\t──────")
	for _, r := range allResults {
		label := r.distribution
		if len(label) > 40 {
			label = label[:40]
		}
		fmt.Fprintf(tw, "%s\t%d\t%.3f\t%.3f\t%.3f\t%.3f\n",
			label, r.n, r.cltSEMean, r.simSEMean, r.avgBootSEMean, r.avgBootSEMed)
	}
	tw.Flush()

	fmt.Println()
	fmt.Printf("  Bootstrap CI Coverage (target: %.0f%%)\n", *fCI*100)
	fmt.Println("════════════════════════════════════════════════════════════")
	tw2 := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw2, "Distribution\tn\tMean CI%\tMedian CI%")
	fmt.Fprintln(tw2, "────────────────────────────────────────\t────\t────────\t──────────")
	for _, r := range allResults {
		label := r.distribution
		if len(label) > 40 {
			label = label[:40]
		}
		fmt.Fprintf(tw2, "%s\t%d\t%.1f%%\t%.1f%%\n",
			label, r.n, r.ciMeanCoverage*100, r.ciMedCoverage*100)
	}
	tw2.Flush()

	totalElapsed := time.Since(studyStart)
	fmt.Printf("\nTotal wall time: %s\n", totalElapsed.Round(time.Millisecond))
	logging.Info("Study complete", slog.Duration("totalElapsed", totalElapsed))

	// ── Memory profiling ──────────────────────────────────────────────────────
	if *fMemProfile != "" {
		f, err := os.Create(*fMemProfile)
		if err != nil {
			logging.Error("cannot create memory profile file", slog.Any("err", err))
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			logging.Error("cannot write memory profile", slog.Any("err", err))
			os.Exit(1)
		}
		logging.Info("Memory profile written", slog.String("file", *fMemProfile))
	}

	fmt.Println("\n----- Run Complete -----")
}
