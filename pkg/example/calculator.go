// Package example provides basic calculator functions for testing coverage history.
package example

import (
	"errors"
	"math"
)

// ErrDivisionByZero is returned when attempting to divide by zero
var ErrDivisionByZero = errors.New("division by zero")

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Subtract returns the difference between two integers.
func Subtract(a, b int) int {
	return a - b
}

// Multiply returns the product of two integers.
func Multiply(a, b int) int {
	return a * b
}

// Divide returns the quotient of two integers.
// Returns an error if attempting to divide by zero.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Power returns the result of base raised to the power of exponent.
func Power(base, exponent int) int {
	return int(math.Pow(float64(base), float64(exponent)))
}

// IsEven returns true if the number is even.
func IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd returns true if the number is odd.
func IsOdd(n int) bool {
	return n%2 != 0
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value of an integer.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
