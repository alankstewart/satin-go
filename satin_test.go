package main

import (
	"runtime"
	"testing"
)

func BenchmarkSingleThread(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Calculate(false)
	}
}

func BenchmarkGoroutines(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ci = make(chan int, N)
	for i := 0; i < b.N; i++ {
		Calculate(true)
	}
}
