// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"math"
)

// LognormalGenerator is a random number generator for lognormal distribution.
// The zero value is invalid, use NewLognormalGenerator to create a generator
type LognormalGenerator struct {
	gauss *GaussianGenerator
}

// NewLognormalGenerator returns a lognormal-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// crng := rng.NewLognormalGenerator(time.Now().UnixNano())
func NewLognormalGenerator(seed int64) *LognormalGenerator {
	grng := NewGaussianGenerator(seed)
	return &LognormalGenerator{grng}
}

// Lognormal return a random number of lognormal distribution
func (lnng LognormalGenerator) Lognormal(mean, stddev float64) float64 {
	return math.Exp(mean + stddev*lnng.gauss.StdGaussian())
}
