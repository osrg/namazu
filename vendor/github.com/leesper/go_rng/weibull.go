// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
)

// WeibullGenerator is a random number generator for weibull distribution.
// The zero value is invalid, use NewWeibullGenerator to create a generator
type WeibullGenerator struct {
	uniform *UniformGenerator
}

// NewWeibullGenerator returns a weibull-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// wrng := rng.NewWeibullGenerator(time.Now().UnixNano())
func NewWeibullGenerator(seed int64) *WeibullGenerator {
	urng := NewUniformGenerator(seed)
	return &WeibullGenerator{urng}
}

// Weibull returns a random number of weibull distribution (lambda > 0.0 and k > 0.0)
func (wrng WeibullGenerator) Weibull(lambda, k float64) float64 {
	if !(lambda > 0.0) {
		panic(fmt.Sprintf("Invalid parameter lambda: %.2f", lambda))
	}

	if !(k > 0.0) {
		panic(fmt.Sprintf("Invalid parameter k: %.2f", k))
	}

	return wrng.weibull(lambda, k)
}

func (wrng WeibullGenerator) weibull(lambda, k float64) float64 {
	return lambda * math.Pow(-math.Log(1-wrng.uniform.Float64()), 1/k)
}
