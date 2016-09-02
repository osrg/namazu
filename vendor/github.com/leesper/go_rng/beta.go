// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
)

// BetaGenerator is a random number generator for beta distribution.
// The zero value is invalid, use NewBetaGenerator to create a generator
type BetaGenerator struct {
	gamma1 *GammaGenerator
	gamma2 *GammaGenerator
}

// NewBetaGenerator returns a beta distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// brng := rng.NewBetaGenerator(time.Now().UnixNano())
func NewBetaGenerator(seed int64) *BetaGenerator {
	grng1 := NewGammaGenerator(seed)
	// according to "Numerical Recipes", all distinct
	//gamma variates must have different random seeds
	grng2 := NewGammaGenerator(2*seed + 1)
	return &BetaGenerator{grng1, grng2}
}

// Beta returns a random number of beta distribution (alpha > 0.0 and beta > 0.0)
func (brng BetaGenerator) Beta(alpha, beta float64) float64 {
	if !(alpha > 0.0) {
		panic(fmt.Sprintf("Invalid parameter alpha: %.2f", alpha))
	}
	if !(beta > 0.0) {
		panic(fmt.Sprintf("Invalid parameter beta: %.2f", beta))
	}
	return brng.beta(alpha, beta)
}

func (brng BetaGenerator) beta(alpha, beta float64) float64 {
	X := brng.gamma1.Gamma(alpha, 1)
	Y := brng.gamma2.Gamma(beta, 1)
	return X / (X + Y)
}
