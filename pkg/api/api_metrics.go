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

package api

import (
	"bufio"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus gatherers

var (
	// requestDuration tracks request latency for all routes
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "kvdi",
		Name:      "http_request_duration_seconds",
		Help:      "The latency of HTTP requests by path and method.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"path", "method"})

	// requestsTotal tracks response codes and methods for all routes
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kvdi",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests by status code, path, and method.",
	}, []string{"path", "code", "method"})

	// displayBytesSentTotal tracks bytes sent over a websocket display stream
	displayBytesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kvdi",
		Name:      "ws_display_bytes_sent_total",
		Help:      "Total bytes sent over websocket display connections by desktop and client.",
	}, []string{"desktop", "client"})

	// audioBytesSentTotal tracks bytes sent over a websocket audio stream
	audioBytesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kvdi",
		Name:      "ws_audio_bytes_sent_total",
		Help:      "Total bytes sent over websocket audio connections by desktop and client.",
	}, []string{"desktop", "client"})

	// displayBytesSentTotal tracks bytes received over a websocket display stream
	displayBytesReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kvdi",
		Name:      "ws_display_bytes_rcvd_total",
		Help:      "Total bytes received over websocket display connections by desktop and client.",
	}, []string{"desktop", "client"})

	// audioBytesSentTotal tracks bytes received over a websocket audio stream
	audioBytesReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kvdi",
		Name:      "ws_audio_bytes_rcvd_total",
		Help:      "Total bytes received over websocket audio connections by desktop and client.",
	}, []string{"desktop", "client"})

	// activeDisplayStreams tracks the number of active display connections
	activeDisplayStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kvdi",
		Name:      "active_display_streams",
		Help:      "The current number of active display streams.",
	})

	// activeDisplayStreams tracks the number of active audio connections
	activeAudioStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "kvdi",
		Name:      "active_audio_streams",
		Help:      "The current number of active audio streams.",
	})
)

// apiResponseWriter extends the regular http.ResponseWriter and stores the
// status code internally to be referenced by the metrics collector.
// When a Hijack is requested for a websocket connection, the net.Conn interface
// is wrapped with an object that sends data transfer metrics to prometheus.
type apiResponseWriter struct {
	http.ResponseWriter
	status int

	isAudio, isDisplay      bool
	clientAddr, desktopName string
}

func (a *apiResponseWriter) WriteHeader(s int) {
	a.ResponseWriter.WriteHeader(s)
	a.status = s
}

func (a *apiResponseWriter) Status() int { return a.status }

func (a *apiResponseWriter) getBytesSentCounter() (counter *prometheus.CounterVec) {
	if a.isAudio {
		counter = audioBytesSentTotal
	}
	if a.isDisplay {
		counter = displayBytesSentTotal
	}
	return
}
func (a *apiResponseWriter) getBytesRcvdCounter() (counter *prometheus.CounterVec) {
	if a.isAudio {
		counter = audioBytesReceivedTotal
	}
	if a.isDisplay {
		counter = displayBytesReceivedTotal
	}
	return
}

func (a *apiResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := apiutil.NewWebsocketWatcher(nil).
		WithLabels(map[string]string{"desktop": a.desktopName, "client": a.clientAddr}).
		WithMetrics(a.getBytesSentCounter(), a.getBytesRcvdCounter()).
		Hijack(a.ResponseWriter)

	if err == nil && a.status == http.StatusOK {
		// The status will be StatusSwitchingProtocols if there was no error and
		// WriteHeader has not been called yet
		a.status = http.StatusSwitchingProtocols
	}

	return conn, rw, err
}

// prometheusMiddleware implements mux.MiddlewareFunc and tracks request metrics.s
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// wrap the response writer so we can intercept request metadata
		aw := &apiResponseWriter{ResponseWriter: w, status: http.StatusOK}

		if isWebsocket(apiutil.GetGorillaPath(r)) {
			doWebsocketMetrics(next, aw, r)
		} else {
			doRequestMetrics(next, aw, r)
		}
	})
}

func doRequestMetrics(next http.Handler, w *apiResponseWriter, r *http.Request) {
	path := apiutil.GetGorillaPath(r)
	// start a timer
	timer := prometheus.NewTimer(requestDuration.With(prometheus.Labels{
		"path":   path,
		"method": r.Method,
	}))

	// run the request flow
	next.ServeHTTP(w, r)

	// incremement the requestsTotal metric
	requestsTotal.With(prometheus.Labels{
		"path":   path,
		"method": r.Method,
		"code":   strconv.Itoa(w.Status()),
	}).Inc()
	// record the duration of the request
	timer.ObserveDuration()
}

func doWebsocketMetrics(next http.Handler, w *apiResponseWriter, r *http.Request) {
	path := apiutil.GetGorillaPath(r)
	w.clientAddr = strings.Split(r.RemoteAddr, ":")[0]
	w.desktopName = apiutil.GetNamespacedNameFromRequest(r).String()
	if isDisplayWebsocket(path) {
		// this is a display connection
		activeDisplayStreams.Inc()
		w.isDisplay = true
	} else if isAudioWebsocket(path) {
		// this is an audio connection
		activeAudioStreams.Inc()
		w.isAudio = true
	}

	// run the request flow
	next.ServeHTTP(w, r)

	if isDisplayWebsocket(path) {
		// this was a display connection
		activeDisplayStreams.Dec()
	} else if isAudioWebsocket(path) {
		// this was an audio connection
		activeAudioStreams.Dec()
	}
}

func isDisplayWebsocket(path string) bool {
	return strings.HasSuffix(strings.TrimSuffix(path, "/"), "display")
}

func isAudioWebsocket(path string) bool {
	return strings.HasSuffix(strings.TrimSuffix(path, "/"), "audio")
}

func isWebsocket(path string) bool { return strings.Contains(path, "/ws/") }
