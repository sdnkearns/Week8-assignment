// Package distrib provides sample generators for three distribution shapes:
// symmetric (normal), positively skewed (log-normal), and negatively skewed
// (reflected log-normal).  All three share the same theoretical mean and
// standard deviation so that results across shapes are directly comparable.
//
// Every generator accepts a *rand.Rand seeded with the MT19937 source so
// that callers control reproducibility.
package distrib

import (
	"math"
	"math/rand"
)

// Params holds the target mean and standard deviation shared across all
// distribution shapes.
type Params struct {
	Mean float64
	SD   float64
}

// LogNormalParams derives the log-normal meanlog / sdlog parameters that
// reproduce the target mean and standard deviation.
type LogNormalParams struct {
	MeanLog float64
	SDLog   float64
}

// DeriveLogNormal computes the log-normal parameterisation that gives
// E[X] = p.Mean and Std[X] = p.SD.
func DeriveLogNormal(p Params) LogNormalParams {
	sigSq := math.Log(1 + (p.SD/p.Mean)*(p.SD/p.Mean))
	return LogNormalParams{
		MeanLog: math.Log(p.Mean) - sigSq/2,
		SDLog:   math.Sqrt(sigSq),
	}
}

// Shape identifies the distribution family.
type Shape string

const (
	Symmetric      Shape = "Symmetric (Normal)"
	PosSkewed      Shape = "Positively Skewed (Log-Normal)"
	NegSkewed      Shape = "Negatively Skewed (Reflected Log-Normal)"
)

// Generator holds everything needed to produce samples from one distribution.
type Generator struct {
	Shape  Shape
	params Params
	lnp    LogNormalParams // valid only for skewed shapes
}

// NewGenerator constructs a Generator for the given shape.
func NewGenerator(s Shape, p Params) Generator {
	return Generator{
		Shape:  s,
		params: p,
		lnp:    DeriveLogNormal(p),
	}
}

// Sample draws n observations from the generator using rng.
func (g Generator) Sample(rng *rand.Rand, n int) []float64 {
	out := make([]float64, n)
	switch g.Shape {
	case Symmetric:
		for i := range out {
			out[i] = rng.NormFloat64()*g.params.SD + g.params.Mean
		}
	case PosSkewed:
		for i := range out {
			out[i] = math.Exp(rng.NormFloat64()*g.lnp.SDLog + g.lnp.MeanLog)
		}
	case NegSkewed:
		for i := range out {
			raw := math.Exp(rng.NormFloat64()*g.lnp.SDLog + g.lnp.MeanLog)
			out[i] = 2*g.params.Mean - raw
		}
	}
	return out
}

// TheoreticalMean returns the population mean (same for all shapes by design).
func (g Generator) TheoreticalMean() float64 { return g.params.Mean }

// TheoreticalSD returns the population standard deviation.
func (g Generator) TheoreticalSD() float64 { return g.params.SD }
