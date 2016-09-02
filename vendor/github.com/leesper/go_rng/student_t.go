// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com
package rng

import (
	"fmt"
	"math"
)

// StudentTGenerator is a random number generator for student-t distribution.
// The zero value is invalid, use NewStudentTGenerator to create a generator
type StudentTGenerator struct {
	gaussian   *GaussianGenerator
	chisquared *ChiSquaredGenerator
}

// NewStudentTGenerator returns a student-t distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// stng := rng.NewStudentTGenerator(time.Now().UnixNano())
func NewStudentTGenerator(seed int64) *StudentTGenerator {
	// make sure they have different random seeds
	grng := NewGaussianGenerator(seed)
	crng := NewChiSquaredGenerator(2*seed + 1)
	return &StudentTGenerator{grng, crng}
}

// Student returns a random number of student-t distribution (freedom > 0.0)
func (stng StudentTGenerator) Student(freedom int64) float64 {
	if !(freedom > 0) {
		panic(fmt.Sprintf("Invalid parameter freedom: %d (freedom > 0)", freedom))
	}
	return stng.student(freedom)
}

// inspired by http://www.xycoon.com/stt_random.htm
func (stng StudentTGenerator) student(freedom int64) float64 {
	X := stng.gaussian.StdGaussian() / math.Sqrt(stng.chisquared.ChiSquared(freedom)/float64(freedom))
	return X
}
