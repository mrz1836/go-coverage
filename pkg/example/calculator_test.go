package example

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Add(2, 3) = %d; expected %d", result, expected)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	expected := 2
	if result != expected {
		t.Errorf("Subtract(5, 3) = %d; expected %d", result, expected)
	}
}

func TestMultiply(t *testing.T) {
	result := Multiply(3, 4)
	expected := 12
	if result != expected {
		t.Errorf("Multiply(3, 4) = %d; expected %d", result, expected)
	}
}

func TestDivide(t *testing.T) {
	// Test normal division
	result, err := Divide(10, 2)
	if err != nil {
		t.Errorf("Divide(10, 2) returned error: %v", err)
	}
	expected := 5
	if result != expected {
		t.Errorf("Divide(10, 2) = %d; expected %d", result, expected)
	}

	// Test division by zero
	_, err = Divide(10, 0)
	if !errors.Is(err, ErrDivisionByZero) {
		t.Errorf("Divide(10, 0) should return ErrDivisionByZero, got: %v", err)
	}
}

func TestPower(t *testing.T) {
	result := Power(2, 3)
	expected := 8
	if result != expected {
		t.Errorf("Power(2, 3) = %d; expected %d", result, expected)
	}
}

func TestIsEven(t *testing.T) {
	if !IsEven(4) {
		t.Errorf("IsEven(4) should return true")
	}
	if IsEven(5) {
		t.Errorf("IsEven(5) should return false")
	}
}

func TestMax(t *testing.T) {
	result := Max(5, 3)
	expected := 5
	if result != expected {
		t.Errorf("Max(5, 3) = %d; expected %d", result, expected)
	}
}

func TestAbs(t *testing.T) {
	result := Abs(-5)
	expected := 5
	if result != expected {
		t.Errorf("Abs(-5) = %d; expected %d", result, expected)
	}
}

func TestIsOdd(t *testing.T) {
	if IsOdd(4) {
		t.Errorf("IsOdd(4) should return false")
	}
	if !IsOdd(5) {
		t.Errorf("IsOdd(5) should return true")
	}
}

// Min function is intentionally not tested to maintain partial coverage (~78%)
