package ratecounter

import (
	"testing"
	"time"
)

func TestAvgRateCounter(t *testing.T) {
	interval := 500 * time.Millisecond
	r := NewAvgRateCounter(interval)

	check := func(expectedRate float64, expectedHits int64) {
		rate, hits := r.Rate(), r.Hits()
		if rate != expectedRate {
			t.Error("Expected rate ", rate, " to equal ", expectedRate)
		}
		if hits != expectedHits {
			t.Error("Expected hits ", hits, " to equal ", expectedHits)
		}
	}

	check(0, 0)
	r.Incr(1) // counter = 1, hits = 1
	check(1.0, 1)
	r.Incr(3) // counter = 4, hits = 2
	check(2.0, 2)
	time.Sleep(2 * interval)
	check(0, 0)
}

func TestAvgRateCounterAdvanced(t *testing.T) {
	interval := 500 * time.Millisecond
	almost := 450 * time.Millisecond
	r := NewAvgRateCounter(interval)

	check := func(expectedRate float64, expectedHits int64) {
		rate, hits := r.Rate(), r.Hits()
		if rate != expectedRate {
			t.Error("Expected rate ", rate, " to equal ", expectedRate)
		}
		if hits != expectedHits {
			t.Error("Expected hits ", hits, " to equal ", expectedHits)
		}
	}

	check(0, 0)
	r.Incr(1) // counter = 1, hits = 1
	check(1.0, 1)
	time.Sleep(interval - almost)
	r.Incr(3) // counter = 4, hits = 2
	check(2.0, 2)
	time.Sleep(almost)
	check(3.0, 1) // counter = 3, hits = 1
	time.Sleep(2 * interval)
	check(0, 0)
}

func TestAvgRateCounterMinResolution(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Resolution < 1 did not panic")
		}
	}()

	NewAvgRateCounter(500 * time.Millisecond).WithResolution(0)
}

func TestAvgRateCounterNoResolution(t *testing.T) {
	interval := 500 * time.Millisecond
	almost := 450 * time.Millisecond
	r := NewAvgRateCounter(interval).WithResolution(1)

	check := func(expectedRate float64, expectedHits int64) {
		rate, hits := r.Rate(), r.Hits()
		if rate != expectedRate {
			t.Error("Expected rate ", rate, " to equal ", expectedRate)
		}
		if hits != expectedHits {
			t.Error("Expected hits ", hits, " to equal ", expectedHits)
		}
	}

	check(0, 0)
	r.Incr(1) // counter = 1, hits = 1
	check(1.0, 1)
	time.Sleep(interval - almost)
	r.Incr(3) // counter = 4, hits = 2
	check(2.0, 2)
	time.Sleep(almost)
	check(0, 0) // counter = 0, hits = 0
	time.Sleep(2 * interval)
	check(0, 0)
}

func TestAvgRateCounter_String(t *testing.T) {
	r := NewAvgRateCounter(1 * time.Second)
	if r.String() != "0.00000e+00" {
		t.Error("Expected ", r.String(), " to equal ", "0.00000e+00")
	}

	r.Incr(1)
	if r.String() != "1.00000e+00" {
		t.Error("Expected ", r.String(), " to equal ", "1.00000e+00")
	}
}

func TestAvgRateCounter_Incr_ReturnsImmediately(t *testing.T) {
	interval := 1 * time.Second
	r := NewAvgRateCounter(interval)

	start := time.Now()
	r.Incr(-1)
	duration := time.Since(start)

	if duration >= 1*time.Second {
		t.Error("incr took", duration, "to return")
	}
}

func BenchmarkAvgRateCounter(b *testing.B) {
	interval := 0 * time.Millisecond
	r := NewAvgRateCounter(interval)

	for i := 0; i < b.N; i++ {
		r.Incr(1)
		r.Rate()
	}
}
