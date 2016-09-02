// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
)

// UniformGenerator is a random number generator for uniform distribution.
// The zero value is invalid, use NewBernoulliGenerator to create a generator
type BernoulliGenerator struct {
	uniform *UniformGenerator
}

// NewBernoulliGenerator returns a bernoulli-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// beng := rng.NewBernoulliGenerator(time.Now().UnixNano())
func NewBernoulliGenerator(seed int64) *BernoulliGenerator {
	urng := NewUniformGenerator(seed)
	return &BernoulliGenerator{urng}
}

// Bernoulli returns a bool, which is true with probablity 0.5
func (beng BernoulliGenerator) Bernoulli() bool {
	return beng.Bernoulli_P(0.5)
}

// Bernoulli_P returns a bool, which is true with probablity p
func (beng BernoulliGenerator) Bernoulli_P(p float64) bool {
	if !(0.0 <= p && p <= 1.0) {
		panic(fmt.Sprintf("Invalid probability: %.2f", p))
	}
	return beng.uniform.Float64() < p
}
