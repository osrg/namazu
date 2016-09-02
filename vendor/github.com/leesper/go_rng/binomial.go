// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
)

// BinomialGenerator is a random number generator for binomial distribution.
// The zero value is invalid, use NewBinomialGenerator to create a generator
type BinomialGenerator struct {
	uniform *UniformGenerator
}

// NewBinomialGenerator returns a binomial-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// bing := rng.NewBinomialGenerator(time.Now().UnixNano())
func NewBinomialGenerator(seed int64) *BinomialGenerator {
	urng := NewUniformGenerator(seed)
	return &BinomialGenerator{urng}
}

// Binomial returns a random number X ~ binomial(n, p)
func (bing BinomialGenerator) Binomial(n int64, p float64) int64 {
	if !(0.0 <= p && p <= 1.0) {
		panic(fmt.Sprintf("Invalid probability p: %f", p))
	}
	if n <= 0 {
		panic(fmt.Sprintf("Invalid parameter n: %d", n))
	}

	if n > 1000 {
		workers := 0
		resChan := make(chan int64)
		for n > 0 {
			go func() {
				res := bing.binomial(1000, p)
				resChan <- res
			}()
			n -= 1000
			workers++
		}
		var result int64
		for i := 0; i < workers; i++ {
			result += <-resChan
		}
		return result
	} else {
		return bing.binomial(n, p)
	}
}

func (bing BinomialGenerator) binomial(n int64, p float64) int64 {
	if !(0.0 <= p && p <= 1.0) {
		panic(fmt.Sprintf("Invalid probability p: %.2f", p))
	}
	if n <= 0 {
		panic(fmt.Sprintf("Invalid parameter n: %d", n))
	}
	var i, result int64
	for i = 0; i < n; i++ {
		if bing.uniform.Float64() < p {
			result++
		}
	}
	return result
}
