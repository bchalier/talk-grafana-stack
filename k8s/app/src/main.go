package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	prometheus.MustRegister(requestCounter, requestDuration)
	rand.Seed(time.Now().UnixNano())
}

func initTracer(ctx context.Context, serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func logJSON(level, msg string, extra map[string]interface{}) {
	entry := map[string]interface{}{
		"ts":    time.Now().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}

	for k, v := range extra {
		entry[k] = v
	}

	b, _ := json.Marshal(entry)
	log.Println(string(b))
}

func main() {
	log.SetFlags(0)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "grafana-demo-app"
	}

	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "alloy.default.svc.cluster.local:4317"
	}

	ctx := context.Background()
	tp, err := initTracer(ctx, appName, otlpEndpoint)
	if err != nil {
		logJSON("error", "failed to initialize tracer", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	tracer := otel.Tracer(appName)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		_, span := tracer.Start(r.Context(), "handle_request")
		defer span.End()

		// Simulate variable latency
		delay := time.Duration(50+rand.Intn(500)) * time.Millisecond
		time.Sleep(delay)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"hello from grafana demo"}`))

		status := http.StatusOK
		path := r.URL.Path
		method := r.Method

		requestCounter.WithLabelValues(path, method, "200").Inc()
		requestDuration.WithLabelValues(path, method).Observe(time.Since(start).Seconds())

		logJSON("info", "handled request", map[string]interface{}{
			"path":       path,
			"method":     method,
			"status":     status,
			"latency_ms": time.Since(start).Milliseconds(),
			"trace_id":   span.SpanContext().TraceID().String(),
			"span_id":    span.SpanContext().SpanID().String(),
		})
	})

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	logJSON("info", "starting server", map[string]interface{}{"port": port})
	handler := otelhttp.NewHandler(mux, "http-server")
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logJSON("error", "server failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}
