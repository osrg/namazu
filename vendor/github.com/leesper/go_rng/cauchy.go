// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
)

// CauchyGenerator is a random number generator for cauchy distribution.
// The zero value is invalid, use NewCauchyGenerator to create a generator
type CauchyGenerator struct {
	uniform *UniformGenerator
}

// NewCauchyGenerator returns a cauchy-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// crng := rng.NewCauchyGenerator(time.Now().UnixNano())
func NewCauchyGenerator(seed int64) *CauchyGenerator {
	urng := NewUniformGenerator(seed)
	return &CauchyGenerator{urng}
}

// Cauchy returns a random number of cauchy distribution
func (crng CauchyGenerator) Cauchy(x0, gamma float64) float64 {
	if !(gamma > 0.0) {
		panic(fmt.Sprintf("Invalid parameter gamma: %.2f", gamma))
	}
	return crng.cauchy(x0, gamma)
}

// StandardCauchy() returns a random number of standard cauchy distribution (x0 = 0.0, gamma = 1.0)
func (crng CauchyGenerator) StandardCauchy() float64 {
	return crng.cauchy(0.0, 1.0)
}

func (crng CauchyGenerator) cauchy(x0, gamma float64) float64 {
	return x0 + gamma*math.Tan(math.Pi*(crng.uniform.Float64()-0.5))
}
