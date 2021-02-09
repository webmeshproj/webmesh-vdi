/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package apiutil

import (
	"bufio"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// WebsocketWatcher implements a wrapper around websocket connections,
// primarily for tracking metrics.
type WebsocketWatcher struct {
	net.Conn

	rsize int
	wsize int

	labels                   map[string]string
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

// WithLabels adds the given labels to the prometheus metrics.
func (w *WebsocketWatcher) WithLabels(labels map[string]string) *WebsocketWatcher {
	w.labels = labels
	return w
}

// Hijack will hijack the given ResponseWriter. Use `nil` for NewWebsocketWatcher when intending to
// call this method.
func (w *WebsocketWatcher) Hijack(writer http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	h, ok := writer.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("Attempted to call Hijack on a non http.Hijacker")
	}
	conn, rw, err := h.Hijack()
	w.Conn = conn
	return w, rw, err
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

// BytesRecvdCount returns the total number of bytes read on the connection so far.
func (w *WebsocketWatcher) BytesRecvdCount() int { return w.rsize }

// BytesSentCount returns the total number of bytes written to the connection so far.
func (w *WebsocketWatcher) BytesSentCount() int { return w.wsize }

// prometheusLabels returns the labels to apply to the prometheus counters.
func (w *WebsocketWatcher) prometheusLabels() prometheus.Labels {
	if w.labels == nil {
		return prometheus.Labels{}
	}
	return prometheus.Labels(w.labels)
}
