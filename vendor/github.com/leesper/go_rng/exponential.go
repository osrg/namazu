// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
)

// ExpGenerator is a random number generator for exponential distribution.
// The zero value is invalid, use NewExpGenerator to create a generator
type ExpGenerator struct {
	uniform *UniformGenerator
}

// NewExpGenerator returns a exponential-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// erng := rng.NewExpGenerator(time.Now().UnixNano())
func NewExpGenerator(seed int64) *ExpGenerator {
	urng := NewUniformGenerator(seed)
	return &ExpGenerator{urng}
}

// Exp returns a random number of exponential distribution
func (erng ExpGenerator) Exp(lambda float64) float64 {
	if !(lambda > 0.0) {
		panic(fmt.Sprintf("Invalid lambda: %.2f", lambda))
	}
	return erng.exp(lambda)
}

func (erng ExpGenerator) exp(lambda float64) float64 {
	return -math.Log(1-erng.uniform.Float64()) / lambda
}
