/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ipam

import (
	"fmt"
	"math/big"
)

type Int big.Int

var IntOne = Int64(1)
var IntZero = Int64(0)
var IntTwoFiveFive = Int64(255)

func ParseInt(s string) (Int, error) {
	n := new(big.Int)
	n, ok := n.SetString(s, 10)
	if !ok {
		return IntZero, fmt.Errorf("invalid number")
	}
	return Int(*n), nil
}

func Int64(i int64) Int {
	return Int(*big.NewInt(i))
}

func (this Int) String() string {
	return (*big.Int)(&this).String()
}

func (this Int) LShift(n uint) Int {
	r := big.Int{}
	return Int(*r.Lsh((*big.Int)(&this), n))
}

func (this Int) RShift(n uint) Int {
	r := big.Int{}
	return Int(*r.Rsh((*big.Int)(&this), n))
}

func (this Int) Add(a Int) Int {
	r := big.Int{}
	return Int(*r.Add((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) Sub(a Int) Int {
	r := big.Int{}
	return Int(*r.Sub((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) Mul(a Int) Int {
	r := big.Int{}
	return Int(*r.Mul((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) Div(a Int) Int {
	r := big.Int{}
	return Int(*r.Div((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) Cmp(a Int) int {
	return (*big.Int)(&this).Cmp((*big.Int)(&a))
}

func (this Int) Mod(a Int) Int {
	r := big.Int{}
	return Int(*r.Mod((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) DivMod(a Int) (Int, Int) {
	r := big.Int{}
	m := big.Int{}
	r.DivMod((*big.Int)(&this), (*big.Int)(&a), &m)
	return Int(r), Int(m)
}

func (this Int) Sgn() int {
	return (*big.Int)(&this).Sign()
}

func (this Int) And(a Int) Int {
	r := big.Int{}
	return Int(*r.And((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) Or(a Int) Int {
	r := big.Int{}
	return Int(*r.Or((*big.Int)(&this), (*big.Int)(&a)))
}

func (this Int) IsInt64() bool {
	return (*big.Int)(&this).IsInt64()
}

func (this Int) Int64() int64 {
	return (*big.Int)(&this).Int64()
}

func (this Int) IsUint64() bool {
	return (*big.Int)(&this).IsUint64()
}

func (this Int) Uint64() uint64 {
	return (*big.Int)(&this).Uint64()
}
