// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
)

// ChiSquaredGenerator is a random number generator for chi-squared distribution.
// The zero value is invalid, use NewChiSquaredGenerator to create a generator
type ChiSquaredGenerator struct {
	gamma *GammaGenerator
}

// NewChiSquaredGenerator returns a chi-squared distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// crng := rng.NewChiSquaredGenerator(time.Now().UnixNano())
func NewChiSquaredGenerator(seed int64) *ChiSquaredGenerator {
	grng := NewGammaGenerator(seed)
	return &ChiSquaredGenerator{grng}
}

// ChiSquared returns a random number of chi-squared distribution (freedom > 0)
func (crng ChiSquaredGenerator) ChiSquared(freedom int64) float64 {
	if freedom <= 0 {
		panic(fmt.Sprintf("Invalid parameter freedom: %d (freedom > 0)", freedom))
	}
	return crng.chisquared(freedom)
}

func (crng ChiSquaredGenerator) chisquared(freedom int64) float64 {
	real_val := float64(freedom)
	return crng.gamma.Gamma(real_val/2.0, 2.0)
}
