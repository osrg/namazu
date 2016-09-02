go_rng --- 伪随机数生成器库的Go语言实现
=======================================

[![Build Status](https://travis-ci.org/leesper/go_rng.svg?branch=master)](https://travis-ci.org/leesper/go_rng)

### A pseudo-random number generator written in Golang v1.1

### Author: 
1.[Leesper](pascal7718@gmail.com) 
2.[Danny Patrie](https://github.com/dpatrie)
3.[Akihiro Suda](https://github.com/AkihiroSuda)

Inspired by:
------------

1.[StdRandom.java](http://introcs.cs.princeton.edu/java/stdlib/StdRandom.java.html)

2.[Numerical Recipes](http://www.nr.com/)

3.[Random number generation](http://en.wikipedia.org/wiki/Random_number_generation)

4.[Quantile function](http://en.wikipedia.org/wiki/Quantile_function)

5.[Monte Carlo method](http://en.wikipedia.org/wiki/Monte_Carlo_method)

6.[Pseudo-random number sampling](http://en.wikipedia.org/wiki/Pseudo-random_number_sampling)

7.[Inverse transform sampling](http://en.wikipedia.org/wiki/Inverse_transform_sampling)

Supported Distributions and Functionalities:
-------------------------------------------
均匀分布      Uniform Distribution <br />
伯努利分布    Bernoulli Distribution <br />
卡方分布      Chi-Squared Distribution <br />
Gamma分布     Gamma Distribution <br />
Beta分布      Beta Distribution <br />
费舍尔F分布   Fisher's F Distribution <br />
柯西分布      Cauchy Distribution <br />
韦伯分布      Weibull Distribution <br />
Pareto分布    Pareto Distribution <br />
对数高斯分布  Log Normal Distribution <br />
指数分布      Exponential Distribution <br />
学生T分布     Student's t-Distribution <br />
二项分布      Binomial Distribution <br />
泊松分布      Poisson Distribution <br />
几何分布      Geometric Distribution <br />
高斯分布      Gaussian Distribution <br />
逻辑分布      Logistic Distribution <br />
狄利克雷分布  Dirichlet Distribution <br />

API Documentation
-----------------

        package rng
        import "rng"

### Uniform Distribution

        1) struct UniformGenerator
        UniformGenerator is a random number generator for uniform
        distribution. The zero value is invalid, use NewUniformGenerator to
        create a generator
        
        2) func NewUniformGenerator(seed int64) *UniformGenerator
        NewUniformGenerator returns a uniform-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: urng
        := rng.NewUniformGenerator(time.Now().UnixNano())
        
        3) func (ung UniformGenerator) Float32() float32
        Float32 returns a random float32 in [0.0, 1.0)
        
        4) func (ung UniformGenerator) Float32Range(a, b float32) float32
        Float32Range returns a random float32 in [a, b)
        
        5) func (ung UniformGenerator) Float32n(n float32) float32
        Float32n returns a random float32 in [0.0, n)
        
        6) func (ung UniformGenerator) Float64() float64
        Float64 returns a random float64 in [0.0, 1.0)
        
        7) func (ung UniformGenerator) Float64Range(a, b float64) float64
        Float32Range returns a random float32 in [a, b)
        
        8) func (ung UniformGenerator) Float64n(n float64) float64
        Float64n returns a random float64 in [0.0, n)
        
        9) func (ung UniformGenerator) Int32() int32
        Int32 returns a random uint32
        
        10) func (ung UniformGenerator) Int32Range(a, b int32) int32
        Int32Range returns a random uint32 in [a, b)
        
        11) func (ung UniformGenerator) Int32n(n int32) int32
        Int32n returns a random uint32 in [0, n)
        
        12) func (ung UniformGenerator) Int64() int64
        Int64 returns a random uint64
        
        13) func (ung UniformGenerator) Int64Range(a, b int64) int64
        Int64Range returns a random uint64 in [a, b)
        
        14) func (ung UniformGenerator) Int64n(n int64) int64
        Int64n returns a random uint64 in [0, n)
        
        15) func (ung UniformGenerator) Shuffle(arr []interface{})
        Shuffle rearrange the elements of an array in random order
        
        16) func (ung UniformGenerator) ShuffleRange(arr []interface{}, low, high int)
        Shuffle rearrange the elements of the subarray[low..high] in random order

### Bernoulli Distribution

        1) struct BernoulliGenerator
        UniformGenerator is a random number generator for uniform distribution.
        The zero value is invalid, use NewBernoulliGenerator to create a
        generator
        
        2) func NewBernoulliGenerator(seed int64) *BernoulliGenerator
        NewBernoulliGenerator returns a bernoulli-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: urng
        := rng.NewBernoulliGenerator(time.Now().UnixNano())
        
        3) func (beng BernoulliGenerator) Bernoulli() bool
        bernoulli returns a bool, which is true with probablity 0.5
        
        4) func (beng BernoulliGenerator) Bernoulli_P(p float32) bool
        bernoulli_P returns a bool, which is true with probablity p

### Binomial Distribution

        1) struct BinomialGenerator
        BinomialGenerator is a random number generator for binomial
        distribution. The zero value is invalid, use NewBinomialGenerator to
        create a generator
        
        2) func NewBinomialGenerator(seed int64) *BinomialGenerator
        NewBinomialGenerator returns a binomial-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: urng
        := rng.NewBinomialGenerator(time.Now().UnixNano())
        
        3) func (bing BinomialGenerator) Binomial(n int64, p float32) int64
        Binomial returns a random number X ~ binomial(n, p)
        
### Geometric Distribution
        
        1) struct GeometricGenerator
        GeometricGenerator is a random number generator for geometric
        distribution. The zero value is invalid, use NewGeometryGenerator to
        create a generator
        
        2) func NewGeometricGenerator(seed int64) *GeometricGenerator
        NewGeometricGenerator returns a geometric-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: urng
        := rng.NewGeometricGenerator(time.Now().UnixNano())
        
        3) func (grng GeometricGenerator) Geometric(p float64) int64
        Geometric returns a random number X ~ binomial(n, p)

### Poisson Distribution

        1) struct PoissonGenerator
        PoissonGenerator is a random number generator for possion distribution.
        The zero value is invalid, use NewPoissonGenerator to create a generator
        
        2) func NewPoissonGenerator(seed int64) *PoissonGenerator
        NewPoissonGenerator returns a possion-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: prng
        := rng.NewPoissonGenerator(time.Now().UnixNano())
        
        3) func (prng PoissonGenerator) Possion(lambda float64) int64
        Poisson returns a random number of possion distribution
        
### Exponential Distribution

        1) struct ExpGenerator
        ExpGenerator is a random number generator for exponential distribution.
        The zero value is invalid, use NewExpGenerator to create a generator
        
        2) func NewExpGenerator(seed int64) *ExpGenerator
        NewExpGenerator returns a exponential-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: erng
        := rng.NewExpGenerator(time.Now().UnixNano())
        
        3) func (erng ExpGenerator) Exp(lambda float64) float64
        Exp returns a random number of exponential distribution

### Cauchy Distribution

        1) struct CauchyGenerator
        CauchyGenerator is a random number generator for cauchy distribution.
        The zero value is invalid, use NewCauchyGenerator to create a generator
        
        2) func NewCauchyGenerator(seed int64) *CauchyGenerator
        NewCauchyGenerator returns a cauchy-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: crng
        := rng.NewCauchyGenerator(time.Now().UnixNano())
        
        3) func (crng CauchyGenerator) Cauchy(x0, gamma float64) float64
        Cauchy returns a random number of cauchy distribution
        
        4) func (crng CauchyGenerator) StandardCauchy() float64
        StandardCauchy() returns a random number of standard cauchy distribution (x0 = 0.0, gamma = 1.0)
        
### Logistic Distribution

        1) struct LogisticGenerator
        LogisticGenerator is a random number generator for cauchy distribution.
        The zero value is invalid, use NewLogisticGenerator to create a generator
        
        2) func NewLogisticGenerator(seed int64) *LogisticGenerator
        NewLogisticGenerator returns a logistic-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: lrng
        := rng.NewLogisticGenerator(time.Now().UnixNano())
        
        3) func (lrng LogisticGenerator) Logistic(mu, s float64) float64
        Logistic returns a random number of logistic distribution

### Gaussian Distribution

        1) struct GaussianGenerator
        GaussianGenerator is a random number generator for gaussian
        distribution. The zero value is invalid, use NewGaussianGenerator to
        create a generator
        
        2) func NewGaussianGenerator(seed int64) *GaussianGenerator
        NewGaussianGenerator returns a gaussian-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: crng
        := rng.NewGaussianGenerator(time.Now().UnixNano())
        
        3) func (grng GaussianGenerator) Gaussian(mean, stddev float64) float64
        Gaussian returns a random number of gaussian distribution Gauss(mean, stddev^2)
        
        4)func (grng GaussianGenerator) StdGaussian() float64
        StdGaussian returns a random number of standard gaussian distribution

### Pareto Distribution (type I)

        1) struct ParetoGenerator
        ParetoGenerator is a random number generator for type I pareto
        distribution. The zero value is invalid, use NewParetoGenerator to
        create a generator
        
        2) func NewParetoGenerator(seed int64) *ParetoGenerator
        NewParetoGenerator returns a type I pareto-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: crng
        := rng.NewParetoGenerator(time.Now().UnixNano())
        
        3) func (prng ParetoGenerator) Pareto(alpha float64) float64
        Pareto returns a random number of type I pareto distribution (alpha > 0,0)

### Weibull Distribution
        
        1) struct WeibullGenerator
        WeibullGenerator is a random number generator for weibull
        distribution. The zero value is invalid, use NewWeibullGenerator to
        create a generator
        
        2) func NewWeibullGenerator(seed int64) *WeibullGenerator
        NewWeibullGenerator returns a weibull-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: wrng
        := rng.NewWeibullGenerator(time.Now().UnixNano())

        3) func (wrng WeibullGenerator) Weibull(lambda, k float64) float64
        Weibull returns a random number of weibull distribution (lambda > 0.0 and k > 0.0)

### Gamma Distribution
        1) struct GammaGenerator
        GammaGenerator is a random number generator for gamma distribution. The
        zero value is invalid, use NewGammaGenerator to create a generator

        2) func NewGammaGenerator(seed int64) *GammaGenerator
        NewGammaGenerator returns a gamma distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: grng
        := rng.NewGammaGenerator(time.Now().UnixNano())
        
        3) func (grng GammaGenerator) Gamma(alpha, beta float64) float64
        Gamma returns a random number of gamma distribution (alpha > 0.0 and beta > 0.0)

### Lognormal Distribution

        1) struct LognormalGenerator
        LognormalGenerator is a random number generator for lognormal
        distribution. The zero value is invalid, use NewLognormalGenerator to
        create a generator
        
        2) func NewLognormalGenerator(seed int64) *LognormalGenerator
        NewLognormalGenerator returns a lognormal-distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: crng
        := rng.NewLognormalGenerator(time.Now().UnixNano())
        
        3) func (lnng LognormalGenerator) Lognormal(mean, stddev float64) float64
        Lognormal return a random number of lognormal distribution

### Beta Distribution

        1) struct BetaGenerator struct
        BetaGenerator is a random number generator for beta distribution. The
        zero value is invalid, use NewBetaGenerator to create a generator
        
        2) func NewBetaGenerator(seed int64) *BetaGenerator
        NewBetaGenerator returns a beta distribution generator it is recommended
        using time.Now().UnixNano() as the seed, for example: brng := 
        rng.NewBetaGenerator(time.Now().UnixNano())
        
        3) func (brng BetaGenerator) Beta(alpha, beta float64) float64
        Beta returns a random number of beta distribution (alpha > 0.0 and beta > 0.0)

### Chi-Squared Distribution

        1) struct ChiSquaredGenerator
        ChiSquaredGenerator is a random number generator for chi-squared
        distribution. The zero value is invalid, use NewChiSquaredGenerator to
        create a generator
        
        2) func NewChiSquaredGenerator(seed int64) *ChiSquaredGenerator
        NewChiSquaredGenerator returns a chi-squared distribution generator it
        is recommended using time.Now().UnixNano() as the seed, for example:
        crng := rng.NewChiSquaredGenerator(time.Now().UnixNano())
        
        3) func (crng ChiSquaredGenerator) ChiSquared(freedom int64) float64
        ChiSquared returns a random number of chi-squared distribution

### Student's t-distribution

        1) struct StudentTGenerator
        StudentTGenerator is a random number generator for student-t
        distribution. The zero value is invalid, use NewStudentTGenerator to
        create a generator
        
        2) func NewStudentTGenerator(seed int64) *StudentTGenerator
        NewStudentTGenerator returns a student-t distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: stng
        := rng.NewStudentTGenerator(time.Now().UnixNano())
        
        3) func (stng StudentTGenerator) Student(freedom int64) float64
        Student returns a random number of student-t distribution (freedom > 0.0)

### Fisher's F Distribution

        1) struct FisherFGenerator
        FisherFGenerator is a random number generator for Fisher's F
        distribution. The zero value is invalid, use NewFisherFGenerator to
        create a generator
        
        2) func NewFisherFGenerator(seed int64) *FisherFGenerator
        NewFisherFGenerator returns a Fisher's F distribution generator it is
        recommended using time.Now().UnixNano() as the seed, for example: frng
        := rng.NewFisherFGenerator(time.Now().UnixNano())
        
        3) func (frng FisherFGenerator) Fisher(d1, d2 int64) float64
        Fisher returns a random number of Fisher's F distribution (d1 > 0 and d2 > 0)

### Dirichlet Distribution

        1) struct DirichletGenerator
        DirichletGenerator is a random number generator for dirichlet
	distribution. The zero value is invalid, use NewDirichletGenerator to
	create a generator
        
        2) func NewDirichletGenerator(seed int64) *DirichletGenerator
	NewDirichletGenerator returns a dirichlet-distribution generator it is
	recommended using time.Now().UnixNano() as the seed, for example: drng
	:= rng.NewDirichletGenerator(time.Now().UnixNano())
        
        3) func (drng DirichletGenerator) Dirichlet(alphas []float64) []float64
	Dirichlet returns random numbers of dirichlet distribution (alpha > 0.0, for alpha in alphas)

        4) func (drng DirichletGenerator) SymmetricDirichlet(alpha float64, n int) []float64
        SymmetricDirichlet returns random numbers of symmetric-dirichlet distribution (alpha > 0.0 and n > 0)

        5) func (drng DirichletGenerator) FlatDirichlet(n int) []float64
        FlatDirichlet returns random numbers of flat-dirichlet distribution (n > 0)
        
