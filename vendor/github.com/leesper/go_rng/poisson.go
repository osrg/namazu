// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
	"time"
)

// PoissonGenerator is a random number generator for possion distribution.
// The zero value is invalid, use NewPoissonGenerator to create a generator
type PoissonGenerator struct {
	uniform *UniformGenerator
}

// NewPoissonGenerator returns a possion-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// prng := rng.NewPoissonGenerator(time.Now().UnixNano())
func NewPoissonGenerator(seed int64) *PoissonGenerator {
	urng := NewUniformGenerator(time.Now().UnixNano())
	return &PoissonGenerator{urng}
}

// Poisson returns a random number of possion distribution
func (prng PoissonGenerator) Poisson(lambda float64) int64 {
	if !(lambda > 0.0) {
		panic(fmt.Sprintf("Invalid lambda: %.2f", lambda))
	}
	return prng.poisson(lambda)
}

func (prng PoissonGenerator) poisson(lambda float64) int64 {
	// algorithm given by Knuth
	L := math.Pow(math.E, -lambda)
	var k int64 = 0
	var p float64 = 1.0

	for p > L {
		k++
		p *= prng.uniform.Float64()
	}
	return (k - 1)
}
