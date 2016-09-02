// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
)

// DirichletGenerator is a random number generator for dirichlet distribution.
// The zero value is invalid, use NewDirichletGenerator to create a generator
type DirichletGenerator struct {
	gamma *GammaGenerator
}

// NewDirichletGenerator returns a dirichlet-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// drng := rng.NewDirichletGenerator(time.Now().UnixNano())
func NewDirichletGenerator(seed int64) *DirichletGenerator {
	grng := NewGammaGenerator(seed)
	return &DirichletGenerator{grng}
}

// Dirichlet returns random numbers of dirichlet distribution (alpha > 0.0, for alpha in alphas)
func (drng DirichletGenerator) Dirichlet(alphas []float64) []float64 {
	for i, alpha := range alphas {
		if !(alpha > 0.0) {
			panic(fmt.Sprintf("Invalid parameter alpha: %.2f (index: %d)", alpha, i))
		}
	}

	return drng.dirichlet(alphas)
}

// SymmetricDirichlet returns random numbers of symmetric-dirichlet distribution (alpha > 0.0 and n > 0)
func (drng DirichletGenerator) SymmetricDirichlet(alpha float64, n int) []float64 {
	if !(alpha > 0.0) {
		panic(fmt.Sprintf("Invalid parameter alpha: %.2f", alpha))
	}
	if !(n > 0) {
		panic(fmt.Sprintf("Invalid parameter n: %d", n))
	}

	alphas := make([]float64, n)
	for i := 0; i < n; i++ {
		alphas[i] = alpha
	}
	return drng.dirichlet(alphas)
}

// FlatDirichlet returns random numbers of flat-dirichlet distribution (n > 0)
func (drng DirichletGenerator) FlatDirichlet(n int) []float64 {
	if !(n > 0) {
		panic(fmt.Sprintf("Invalid parameter n: %d", n))
	}

	alphas := make([]float64, n)
	for i := 0; i < n; i++ {
		alphas[i] = 1.0
	}
	return drng.dirichlet(alphas)
}

func (drng DirichletGenerator) dirichlet(alphas []float64) []float64 {
	gammas := make([]float64, len(alphas))
	results := make([]float64, len(alphas))
	var gsum float64
	for i, alpha := range alphas {
		gammas[i] = drng.gamma.Gamma(alpha, 1)
		gsum += gammas[i]
	}
	for i, gamma := range gammas {
		results[i] = gamma / gsum
	}
	return results
}
