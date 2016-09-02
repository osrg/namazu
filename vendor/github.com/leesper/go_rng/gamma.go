// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
	"math"
)

// GammaGenerator is a random number generator for gamma distribution.
// The zero value is invalid, use NewGammaGenerator to create a generator
type GammaGenerator struct {
	uniform *UniformGenerator
}

// NewGammaGenerator returns a gamma distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// grng := rng.NewGammaGenerator(time.Now().UnixNano())
func NewGammaGenerator(seed int64) *GammaGenerator {
	urng := NewUniformGenerator(seed)
	return &GammaGenerator{urng}
}

// Gamma returns a random number of gamma distribution (alpha > 0.0 and beta > 0.0)
func (grng GammaGenerator) Gamma(alpha, beta float64) float64 {
	if !(alpha > 0.0) || !(beta > 0.0) {
		panic(fmt.Sprintf("Invalid parameter alpha %.2f beta %.2f", alpha, beta))
	}

	return grng.gamma(alpha, beta)
}

// inspired by random.py
func (grng GammaGenerator) gamma(alpha, beta float64) float64 {
	var MAGIC_CONST float64 = 4 * math.Exp(-0.5) / math.Sqrt(2.0)
	if alpha > 1.0 {
		// Use R.C.H Cheng "The generation of Gamma variables with
		// non-integral shape parameters", Applied Statistics, (1977), 26, No. 1, p71-74

		ainv := math.Sqrt(2.0*alpha - 1.0)
		bbb := alpha - math.Log(4.0)
		ccc := alpha + ainv

		for {
			u1 := grng.uniform.Float64()
			if !(1e-7 < u1 && u1 < .9999999) {
				continue
			}
			u2 := 1.0 - grng.uniform.Float64()
			v := math.Log(u1/(1.0-u1)) / ainv
			x := alpha * math.Exp(v)
			z := u1 * u1 * u2
			r := bbb + ccc*v - x
			if r+MAGIC_CONST-4.5*z >= 0.0 || r >= math.Log(z) {
				return x * beta
			}
		}
	} else if alpha == 1.0 {
		u := grng.uniform.Float64()
		for u <= 1e-7 {
			u = grng.uniform.Float64()
		}
		return -math.Log(u) * beta
	} else { // alpha between 0.0 and 1.0 (exclusive)
		// Uses Algorithm of Statistical Computing - kennedy & Gentle
		var x float64
		for {
			u := grng.uniform.Float64()
			b := (math.E + alpha) / math.E
			p := b * u
			if p <= 1.0 {
				x = math.Pow(p, 1.0/alpha)
			} else {
				x = -math.Log((b - p) / alpha)
			}
			u1 := grng.uniform.Float64()
			if p > 1.0 {
				if u1 <= math.Pow(x, alpha-1.0) {
					break
				}
			} else if u1 <= math.Exp(-x) {
				break
			}
		}
		return x * beta
	}
}
