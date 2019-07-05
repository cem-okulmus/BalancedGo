package test

import "testing"

func FibFast(n int) int {
	return Fib2(1, 1, 1, n)
}

func Fib2(val1 int, val2 int, i int, until int) int {
	if i >= until {
		return val1
	}

	return Fib2(val2, val1+val2, i+1, until)

}

func Fib(n int) int {
	if n <= 2 {
		return 1
	}

	return Fib(n-1) + Fib(n-2)
}

func BenchmarkFib10(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		FibFast(10)
	}
}
