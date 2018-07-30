// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package value // import "robpike.io/ivy/value"

import "math/big"

// Unary operators.

// To avoid initialization cycles when we refer to the ops from inside
// themselves, we use an init function to initialize the ops.

// unaryBigIntOp applies the op to a BigInt.
func unaryBigIntOp(op func(*big.Int, *big.Int) *big.Int, v Value) Value {
	i := v.(BigInt)
	z := bigInt64(0)
	op(z.Int, i.Int)
	return z.shrink()
}

// unaryBigRatOp applies the op to a BigRat.
func unaryBigRatOp(op func(*big.Rat, *big.Rat) *big.Rat, v Value) Value {
	i := v.(BigRat)
	z := bigRatInt64(0)
	op(z.Rat, i.Rat)
	return z.shrink()
}

var (
	unaryRoll                         *unaryOp
	unaryPlus, unaryMinus, unaryRecip *unaryOp
	unaryAbs, unarySignum             *unaryOp
	unaryBitwiseNot, unaryLogicalNot  *unaryOp
	unaryIota, unaryRho, unaryRavel   *unaryOp
	gradeUp, gradeDown                *unaryOp
	reverse, flip                     *unaryOp
	floor, ceil                       *unaryOp
	unaryOps                          map[string]*unaryOp
)

// bigIntRand sets a to a random number in [origin, origin+b].
func bigIntRand(a, b *big.Int) *big.Int {
	a.Rand(conf.Random(), b)
	return a.Add(a, conf.BigOrigin())
}

func self(v Value) Value {
	return v
}

func init() {
	unaryRoll = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				i := int64(v.(Int))
				if i <= 0 {
					Errorf("illegal roll value %v", v)
				}
				return Int(conf.Origin()) + Int(conf.Random().Int63n(i))
			},
			bigIntType: func(v Value) Value {
				if v.(BigInt).Sign() <= 0 {
					Errorf("illegal roll value %v", v)
				}
				return unaryBigIntOp(bigIntRand, v)
			},
		},
	}

	unaryPlus = &unaryOp{
		fn: [numType]unaryFn{
			intType:    self,
			bigIntType: self,
			bigRatType: self,
			vectorType: self,
			matrixType: self,
		},
	}

	unaryMinus = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				return -v.(Int)
			},
			bigIntType: func(v Value) Value {
				return unaryBigIntOp((*big.Int).Neg, v)
			},
			bigRatType: func(v Value) Value {
				return unaryBigRatOp((*big.Rat).Neg, v)
			},
		},
	}

	unaryRecip = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				i := int64(v.(Int))
				if i == 0 {
					Errorf("division by zero")
				}
				return BigRat{
					Rat: big.NewRat(0, 1).SetFrac64(1, i),
				}.shrink()
			},
			bigIntType: func(v Value) Value {
				// Zero division cannot happen for unary.
				return BigRat{
					Rat: big.NewRat(0, 1).SetFrac(bigOne.Int, v.(BigInt).Int),
				}.shrink()
			},
			bigRatType: func(v Value) Value {
				// Zero division cannot happen for unary.
				r := v.(BigRat)
				return BigRat{
					Rat: big.NewRat(0, 1).SetFrac(r.Denom(), r.Num()),
				}.shrink()
			},
		},
	}

	unarySignum = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				i := int64(v.(Int))
				if i > 0 {
					return one
				}
				if i < 0 {
					return minusOne
				}
				return zero
			},
			bigIntType: func(v Value) Value {
				return Int(v.(BigInt).Sign())
			},
			bigRatType: func(v Value) Value {
				return Int(v.(BigRat).Sign())
			},
		},
	}

	unaryBitwiseNot = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				return ^v.(Int)
			},
			bigIntType: func(v Value) Value {
				// Lots of ways to do this, here's one.
				return BigInt{Int: bigInt64(0).Xor(v.(BigInt).Int, bigMinusOne.Int)}
			},
		},
	}

	unaryLogicalNot = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				if v.(Int) == 0 {
					return one
				}
				return zero
			},
			bigIntType: func(v Value) Value {
				if v.(BigInt).Sign() == 0 {
					return one
				}
				return zero
			},
			bigRatType: func(v Value) Value {
				if v.(BigRat).Sign() == 0 {
					return one
				}
				return zero
			},
		},
	}

	unaryAbs = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				i := v.(Int)
				if i < 0 {
					i = -i
				}
				return i
			},
			bigIntType: func(v Value) Value {
				return unaryBigIntOp((*big.Int).Abs, v)
			},
			bigRatType: func(v Value) Value {
				return unaryBigRatOp((*big.Rat).Abs, v)
			},
		},
	}

	floor = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType:    func(v Value) Value { return v },
			bigIntType: func(v Value) Value { return v },
			bigRatType: func(v Value) Value {
				i := v.(BigRat)
				if i.IsInt() {
					// It can't be an integer, which means we must move up or down.
					panic("min: is int")
				}
				positive := i.Sign() >= 0
				if !positive {
					j := bigRatInt64(0)
					j.Abs(i.Rat)
					i = j
				}
				z := bigInt64(0)
				z.Quo(i.Num(), i.Denom())
				if !positive {
					z.Add(z.Int, bigOne.Int)
					z.Neg(z.Int)
				}
				return z
			},
		},
	}

	ceil = &unaryOp{
		elementwise: true,
		fn: [numType]unaryFn{
			intType:    func(v Value) Value { return v },
			bigIntType: func(v Value) Value { return v },
			bigRatType: func(v Value) Value {
				i := v.(BigRat)
				if i.IsInt() {
					// It can't be an integer, which means we must move up or down.
					panic("max: is int")
				}
				positive := i.Sign() >= 0
				if !positive {
					j := bigRatInt64(0)
					j.Abs(i.Rat)
					i = j
				}
				z := bigInt64(0)
				z.Quo(i.Num(), i.Denom())
				if positive {
					z.Add(z.Int, bigOne.Int)
				} else {
					z.Neg(z.Int)
				}
				return z
			},
		},
	}

	unaryIota = &unaryOp{
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				i := v.(Int)
				if i < 0 || maxInt < i {
					Errorf("bad iota %d", i)
				}
				if i == 0 {
					return Vector{}
				}
				n := make([]Value, i)
				for k := range n {
					n[k] = Int(k + conf.Origin())
				}
				return NewVector(n)
			},
		},
	}

	unaryRho = &unaryOp{
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				return Vector{}
			},
			bigIntType: func(v Value) Value {
				return Vector{}
			},
			bigRatType: func(v Value) Value {
				return Vector{}
			},
			vectorType: func(v Value) Value {
				return Int(len(v.(Vector)))
			},
			matrixType: func(v Value) Value {
				return v.(Matrix).shape
			},
		},
	}

	unaryRavel = &unaryOp{
		fn: [numType]unaryFn{
			intType: func(v Value) Value {
				return NewVector([]Value{v})
			},
			bigIntType: func(v Value) Value {
				return NewVector([]Value{v})
			},
			bigRatType: func(v Value) Value {
				return NewVector([]Value{v})
			},
			vectorType: self,
			matrixType: func(v Value) Value {
				return v.(Matrix).data
			},
		},
	}

	gradeUp = &unaryOp{
		fn: [numType]unaryFn{
			intType:    self,
			bigIntType: self,
			bigRatType: self,
			vectorType: func(v Value) Value {
				return v.(Vector).grade()
			},
		},
	}

	gradeDown = &unaryOp{
		fn: [numType]unaryFn{
			intType:    self,
			bigIntType: self,
			bigRatType: self,
			vectorType: func(v Value) Value {
				x := v.(Vector).grade()
				for i, j := 0, len(x)-1; i < j; i, j = i+1, j-1 {
					x[i], x[j] = x[j], x[i]
				}
				return x
			},
		},
	}

	reverse = &unaryOp{
		fn: [numType]unaryFn{
			intType:    self,
			bigIntType: self,
			bigRatType: self,
			vectorType: func(v Value) Value {
				x := v.(Vector)
				for i, j := 0, len(x)-1; i < j; i, j = i+1, j-1 {
					x[i], x[j] = x[j], x[i]
				}
				return x
			},
			matrixType: func(v Value) Value {
				m := v.(Matrix)
				if len(m.shape) == 0 {
					return m
				}
				if len(m.shape) == 1 {
					Errorf("rev: matrix is vector")
				}
				size := m.size()
				ncols := int(m.shape[len(m.shape)-1].(Int))
				x := m.data
				for index := 0; index <= size-ncols; index += ncols {
					for i, j := 0, ncols-1; i < j; i, j = i+1, j-1 {
						x[index+i], x[index+j] = x[index+j], x[index+i]
					}
				}
				return m
			},
		},
	}

	flip = &unaryOp{
		fn: [numType]unaryFn{
			intType:    self,
			bigIntType: self,
			bigRatType: self,
			vectorType: func(v Value) Value {
				return Unary("rev", v)
			},
			matrixType: func(v Value) Value {
				m := v.(Matrix)
				if len(m.shape) == 0 {
					return m
				}
				if len(m.shape) == 1 {
					Errorf("flip: matrix is vector")
				}
				elemSize := m.elemSize()
				size := m.size()
				x := m.data
				lo := 0
				hi := size - elemSize
				for lo < hi {
					for i := 0; i < elemSize; i++ {
						x[lo+i], x[hi+i] = x[hi+i], x[lo+i]
					}
					lo += elemSize
					hi -= elemSize
				}
				return m
			},
		},
	}

	unaryOps = map[string]*unaryOp{
		"+":     unaryPlus,
		",":     unaryRavel,
		"-":     unaryMinus,
		"/":     unaryRecip,
		"?":     unaryRoll,
		"^":     unaryBitwiseNot,
		"abs":   unaryAbs,
		"ceil":  ceil,
		"down":  gradeDown,
		"flip":  flip,
		"floor": floor,
		"iota":  unaryIota,
		"rev":   reverse,
		"rho":   unaryRho,
		"sgn":   unarySignum,
		"up":    gradeUp,
		"~":     unaryLogicalNot,
	}
}
