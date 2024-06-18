package overleash

import "time"

type ticker struct {
	period time.Duration
	ticker *time.Ticker
}

func createTicker(period time.Duration) ticker {
	return ticker{period, time.NewTicker(period)}
}

func (t *ticker) resetTicker() {
	t.ticker.Reset(t.period)
}
