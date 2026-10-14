package overleash

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
)

type metrics struct {
	metricChannel     chan *MetricsData
	clientDataChannel chan *ClientData

	metrics    []*MetricsData
	clientData []*ClientData
}

func (o *OverleashContext) startMetrics(ctx context.Context) {
	o.metrics = &metrics{
		metricChannel:     make(chan *MetricsData),
		clientDataChannel: make(chan *ClientData),
		metrics:           make([]*MetricsData, 0),
		clientData:        make([]*ClientData, 0),
	}

	t := createTicker(time.Minute)

	go func() {
		defer t.ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Preserve the best-effort final flush, but do not let shutdown
				// wait indefinitely or inherit an already-canceled context.
				flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), upstreamRequestTimeout)
				o.sendMetrics(flushCtx)
				cancel()
				return

			case val := <-o.metrics.metricChannel:
				o.metrics.metrics = append(o.metrics.metrics, val)
				break

			case val := <-o.metrics.clientDataChannel:
				o.metrics.clientData = append(o.metrics.clientData, val)
				break

			case <-t.ticker.C:
				o.sendMetrics(ctx)
			}
		}
	}()
}

func (m *metrics) reset() {
	m.clientData = m.clientData[:0]
	m.metrics = m.metrics[:0]
}

func (o *OverleashContext) sendMetrics(ctx context.Context) {
	if len(o.metrics.metrics) == 0 && len(o.metrics.clientData) == 0 {
		log.Debug("No metrics to send")
		return
	}

	log.Debug("Sending metrics")
	err := o.client.bulkMetrics(ctx, o.ActiveFeatureEnvironment().token, o.metrics.clientData, o.metrics.metrics)

	if err != nil {
		log.Errorf("Failed to send metrics to upstream: %v", err)

		return
	}

	o.metrics.reset()
}
