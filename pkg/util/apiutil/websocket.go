package apiutil

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
)

// WebsocketWatcher implements a wrapper around websocket connections,
// primarily for tracking metrics.
type WebsocketWatcher struct {
	net.Conn

	rsize int
	wsize int

	clientAddr, desktopName  string
	sendCounter, recvCounter *prometheus.CounterVec
}

// NewWebsocketWatcher returns a new websocket watcher.
func NewWebsocketWatcher(c net.Conn) *WebsocketWatcher {
	return &WebsocketWatcher{Conn: c}
}

// WithMetrics applies prometheus counters to the read/write events on the websocket.
func (w *WebsocketWatcher) WithMetrics(sendCounter, recvCounter *prometheus.CounterVec) *WebsocketWatcher {
	w.sendCounter = sendCounter
	w.recvCounter = recvCounter
	return w
}

// WithMetadata adds client and desktop metadata to the prometheus metrics.
func (w *WebsocketWatcher) WithMetadata(clientAddr, desktopName string) *WebsocketWatcher {
	w.clientAddr = clientAddr
	w.desktopName = desktopName
	return w
}

// Read implements read on the net.Conn interface.
func (w *WebsocketWatcher) Read(b []byte) (int, error) {
	size, err := w.Conn.Read(b)
	w.rsize += size
	if w.recvCounter != nil {
		w.recvCounter.With(w.prometheusLabels()).Add(float64(size))
	}
	return size, err
}

// Write implements write on the net.Conn interface.
func (w *WebsocketWatcher) Write(b []byte) (int, error) {
	size, err := w.Conn.Write(b)
	w.wsize += size
	if w.sendCounter != nil {
		w.sendCounter.With(w.prometheusLabels()).Add(float64(size))
	}
	return size, err
}

// ReadCount returns the total number of bytes read on the connection so far.
func (w *WebsocketWatcher) ReadCount() int { return w.rsize }

// WriteCount returns the total number of bytes written to the connection so far.
func (w *WebsocketWatcher) WriteCount() int { return w.wsize }

// prometheusLabels returns the labels to apply to the prometheus counters.
func (w *WebsocketWatcher) prometheusLabels() prometheus.Labels {
	return prometheus.Labels{"desktop": w.desktopName, "client": w.clientAddr}
}
