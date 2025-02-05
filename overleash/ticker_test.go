package overleash

import (
	"testing"
	"time"
)

func TestCreateTicker(t *testing.T) {
	period := 100 * time.Millisecond
	tk := createTicker(period)
	defer tk.ticker.Stop()

	if tk.period != period {
		t.Errorf("Expected ticker.period %v, got %v", period, tk.period)
	}

	start := time.Now()
	select {
	case <-tk.ticker.C:
		elapsed := time.Since(start)
		if elapsed < period {
			t.Errorf("Tick arrived too early: got %v, expected at least %v", elapsed, period)
		}
	case <-time.After(2 * period):
		t.Fatal("Timed out waiting for the ticker to tick")
	}
}

func TestResetTicker(t *testing.T) {
	period := 100 * time.Millisecond
	tk := createTicker(period)
	defer tk.ticker.Stop()

	select {
	case <-tk.ticker.C:
	case <-time.After(2 * period):
		t.Fatal("Timed out waiting for the initial tick")
	}

	// Reset the ticker.
	start := time.Now()
	tk.resetTicker()

	select {
	case <-tk.ticker.C:
		elapsed := time.Since(start)
		lowerBound := period - 50*time.Millisecond
		upperBound := period + 50*time.Millisecond
		if elapsed < lowerBound || elapsed > upperBound {
			t.Errorf("After reset, tick arrived after %v; expected around %v (±50ms)", elapsed, period)
		}
	case <-time.After(2 * period):
		t.Fatal("Timed out waiting for a tick after reset")
	}
}

func TestTickerNotErroringWhenEmpty(t *testing.T) {
	tk := ticker{}

	tk.resetTicker()
}
