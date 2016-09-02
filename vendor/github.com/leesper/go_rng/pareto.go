// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
	"math"
)

// ParetoGenerator is a random number generator for type I pareto distribution.
// The zero value is invalid, use NewParetoGenerator to create a generator
type ParetoGenerator struct {
	uniform *UniformGenerator
}

// NewParetoGenerator returns a type I pareto-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// crng := rng.NewParetoGenerator(time.Now().UnixNano())
func NewParetoGenerator(seed int64) *ParetoGenerator {
	urng := NewUniformGenerator(seed)
	return &ParetoGenerator{urng}
}

// Pareto returns a random number of type I pareto distribution (alpha > 0,0)
func (prng ParetoGenerator) Pareto(alpha float64) float64 {
	if !(alpha > 0.0) {
		panic(fmt.Sprintf("Invalid parameter alpha: %.2f", alpha))
	}

	return prng.pareto(alpha)
}

func (prng ParetoGenerator) pareto(alpha float64) float64 {
	return math.Pow(1-prng.uniform.Float64(), -1.0/alpha) - 1.0
}
