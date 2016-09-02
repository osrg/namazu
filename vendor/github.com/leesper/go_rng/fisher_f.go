// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
)

// FisherFGenerator is a random number generator for Fisher's F distribution.
// The zero value is invalid, use NewFisherFGenerator to create a generator
type FisherFGenerator struct {
	beta *BetaGenerator
}

// NewFisherFGenerator returns a Fisher's F distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// frng := rng.NewFisherFGenerator(time.Now().UnixNano())
func NewFisherFGenerator(seed int64) *FisherFGenerator {
	brng := NewBetaGenerator(seed)
	return &FisherFGenerator{brng}
}

// Fisher returns a random number of Fisher's F distribution (d1 > 0 and d2 > 0)
func (frng FisherFGenerator) Fisher(d1, d2 int64) float64 {
	if d1 <= 0 {
		panic(fmt.Sprintf("Invalid parameter d1: %d (d1 > 0)", d1))
	}
	if d2 <= 0 {
		panic(fmt.Sprintf("Invalid parameter d2: %d (d2 > 0)", d2))
	}
	return frng.fisher(d1, d2)
}

func (frng FisherFGenerator) fisher(d1, d2 int64) float64 {
	f1 := float64(d1)
	f2 := float64(d2)
	X := frng.beta.Beta(f1/2.0, f2/2.0)
	return f2 * X / (f1 * (1 - X))
}
