// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"math"
)

// GaussianGenerator is a random number generator for gaussian distribution.
// The zero value is invalid, use NewGaussianGenerator to create a generator
type GaussianGenerator struct {
	uniform *UniformGenerator
}

// NewGaussianGenerator returns a gaussian-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// crng := rng.NewGaussianGenerator(time.Now().UnixNano())
func NewGaussianGenerator(seed int64) *GaussianGenerator {
	urng := NewUniformGenerator(seed)
	return &GaussianGenerator{urng}
}

// StdGaussian returns a random number of standard gaussian distribution
func (grng GaussianGenerator) StdGaussian() float64 {
	return grng.gaussian()
}

// Gaussian returns a random number of gaussian distribution Gauss(mean, stddev^2)
func (grng GaussianGenerator) Gaussian(mean, stddev float64) float64 {
	return mean + stddev*grng.gaussian()
}

func (grng GaussianGenerator) gaussian() float64 {
	// Box-Muller Transform
	var r, x, y float64
	for r >= 1 || r == 0 {
		x = grng.uniform.Float64Range(-1.0, 1.0)
		y = grng.uniform.Float64Range(-1.0, 1.0)
		r = x*x + y*y
	}
	return x * math.Sqrt(-2*math.Log(r)/r)
}
