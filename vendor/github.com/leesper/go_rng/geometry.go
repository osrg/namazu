// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
)

// GeometricGenerator is a random number generator for geometric distribution.
// The zero value is invalid, use NewGeometryGenerator to create a generator
type GeometricGenerator struct {
	uniform *UniformGenerator
}

// NewGeometricGenerator returns a geometric-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// grng := rng.NewGeometricGenerator(time.Now().UnixNano())
func NewGeometricGenerator(seed int64) *GeometricGenerator {
	urng := NewUniformGenerator(seed)
	return &GeometricGenerator{urng}
}

// Geometric returns a random number X ~ binomial(n, p)
func (grng GeometricGenerator) Geometric(p float64) int64 {
	if !(0.0 <= p && p <= 1.0) {
		panic(fmt.Sprintf("Invalid probability p: %.2f", p))
	}
	return grng.geometric(p)
}

func (grng GeometricGenerator) geometric(p float64) int64 {
	// ceil( ln(U)/ln(1-p) ), where U ~ Uniform(0, 1)
	return int64(math.Floor(math.Log(grng.uniform.Float64()) / math.Log(1.0-p)))
}
