// Package rng implements a series of pseudo-random number generator
// based on a variety of common probability distributions
// Author: Leesper
// Email: pascal7718@gmail.com 394683518@qq.com

package rng

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// UniformGenerator is a random number generator for uniform distribution.
// The zero value is invalid, use NewUniformGenerator to create a generator
type UniformGenerator struct {
	mu *sync.Mutex
	rd *rand.Rand
}

// NewUniformGenerator returns a uniform-distribution generator
// it is recommended using time.Now().UnixNano() as the seed, for example:
// urng := rng.NewUniformGenerator(time.Now().UnixNano())
func NewUniformGenerator(seed int64) *UniformGenerator {
	return &UniformGenerator{mu: new(sync.Mutex), rd: rand.New(rand.NewSource(seed))}
}

// Int32 returns a random uint32
func (ung UniformGenerator) Int32() int32 {
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Int31()
}

// Int64 returns a random uint64
func (ung UniformGenerator) Int64() int64 {
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Int63()
}

// Int32n returns a random uint32 in [0, n)
func (ung UniformGenerator) Int32n(n int32) int32 {
	if n <= 0 {
		panic(fmt.Sprintf("Illegal parameter n: %d", n))
	}
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Int31n(n)
}

// Int64n returns a random uint64 in [0, n)
func (ung UniformGenerator) Int64n(n int64) int64 {
	if n <= 0 {
		panic(fmt.Sprintf("Illegal parameter n: %d", n))
	}
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Int63n(n)
}

// Int32Range returns a random uint32 in [a, b)
func (ung UniformGenerator) Int32Range(a, b int32) int32 {
	if b <= a {
		panic(fmt.Sprintf("Illegal parameter a, b: %d, %d", a, b))
	}
	if b-a > math.MaxInt32 {
		panic(fmt.Sprintf("Illegal parameter a, b: %d, %d", a, b))
	}
	return a + ung.Int32n(b-a)
}

// Int64Range returns a random uint64 in [a, b)
func (ung UniformGenerator) Int64Range(a, b int64) int64 {
	if b <= a {
		panic(fmt.Sprintf("Illegal parameter a, b: %d, %d", a, b))
	}
	if b-a > math.MaxInt32 {
		panic(fmt.Sprintf("Illegal parameter a, b: %d, %d", a, b))
	}
	return a + ung.Int64n(b-a)
}

// Float32 returns a random float32 in [0.0, 1.0)
func (ung UniformGenerator) Float32() float32 {
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Float32()
}

// Float64 returns a random float64 in [0.0, 1.0)
func (ung UniformGenerator) Float64() float64 {
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Float64()
}

// Float32Range returns a random float32 in [a, b)
func (ung UniformGenerator) Float32Range(a, b float32) float32 {
	if !(a < b) {
		panic(fmt.Sprintf("Invalid range: %.2f ~ %.2f", a, b))
	}
	return a + ung.Float32()*(b-a)
}

// Float32Range returns a random float32 in [a, b)
func (ung UniformGenerator) Float64Range(a, b float64) float64 {
	if !(a < b) {
		panic(fmt.Sprintf("Invalid range: %.2f ~ %.2f", a, b))
	}
	return a + ung.Float64()*(b-a)
}

// Float32n returns a random float32 in [0.0, n)
func (ung UniformGenerator) Float32n(n float32) float32 {
	return ung.Float32Range(0.0, n)
}

// Float64n returns a random float64 in [0.0, n)
func (ung UniformGenerator) Float64n(n float64) float64 {
	return ung.Float64Range(0.0, n)
}

// Shuffle rearrange the elements of an array in random order
func (ung UniformGenerator) Shuffle(arr []interface{}) {
	N := len(arr)
	for i := range arr {
		r := int32(i) + ung.Int32n(int32(N-i))
		arr[i], arr[r] = arr[r], arr[i]
	}
}

// Shuffle rearrange the elements of the subarray[low..high] in random order
func (ung UniformGenerator) ShuffleRange(arr []interface{}, low, high int) {
	if low < 0 || low > high || high >= len(arr) {
		panic(fmt.Sprintf("Illegal subarray range %d ~ %d", low, high))
	}
	for i := low; i <= high; i++ {
		r := int32(i) + ung.Int32n(int32(high-i+1))
		arr[i], arr[r] = arr[r], arr[i]
	}
}
