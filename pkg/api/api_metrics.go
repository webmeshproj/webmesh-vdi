package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

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
type apiResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (a *apiResponseWriter) WriteHeader(code int) {
	a.statusCode = code
	a.ResponseWriter.WriteHeader(code)
}

// prometheusMiddleware implements mux.MiddlewareFunc and tracks request metrics.s
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the path for this request
		path := strings.TrimSuffix(apiutil.GetGorillaPath(r), "/")

		// determine if this is a websocket path
		isWebsocket := strings.HasSuffix(path, "websockify") || strings.HasSuffix(path, "wsaudio")

		var timer *prometheus.Timer
		var aw http.ResponseWriter

		if !isWebsocket {
			// start a timer for non-websocket endpoints
			timer = prometheus.NewTimer(requestDuration.With(prometheus.Labels{
				"path":   path,
				"method": r.Method,
			}))
			// wrap the response writer so we can retrieve the status code.
			aw = &apiResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		} else {
			// Track active websocket connections
			if strings.HasSuffix(path, "websockify") {
				// this is a display connection
				activeDisplayStreams.Inc()
			} else if strings.HasSuffix(path, "wsaudio") {
				// this is an audio connection
				activeAudioStreams.Inc()
			}
			// use the native response writer
			// NOTE: To override the writer for websocket connections an http.Hijacker
			// will need to be implemented.
			aw = w
		}

		// run the request flow
		next.ServeHTTP(aw, r)

		if !isWebsocket {
			// get the apiResponseWriter from the writer interface
			awr := aw.(*apiResponseWriter)
			// incremement the requestsTotal metric
			requestsTotal.With(prometheus.Labels{
				"path":   path,
				"method": r.Method,
				"code":   strconv.Itoa(awr.statusCode),
			}).Inc()
			// record the duration of the request
			timer.ObserveDuration()
		} else {
			if strings.HasSuffix(path, "websockify") {
				// this was a display connection
				activeDisplayStreams.Dec()
			} else if strings.HasSuffix(path, "wsaudio") {
				// this was an audio connection
				activeAudioStreams.Dec()
			}
		}

	})
}
