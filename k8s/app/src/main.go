package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

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

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "alloy.monitoring.svc.cluster.local:4317"
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return nil, err
	}

	serviceName := os.Getenv("APP_NAME")
	if serviceName == "" {
		serviceName = "grafana-demo-app"
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func logJSON(ctx context.Context, level, msg string, extra map[string]interface{}) {
	entry := map[string]interface{}{
		"ts":    time.Now().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}

	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		if span != nil {
			sc := span.SpanContext()
			if sc.IsValid() {
				entry["trace_id"] = sc.TraceID().String()
				entry["span_id"] = sc.SpanID().String()
			}
		}
	}

	for k, v := range extra {
		entry[k] = v
	}

	b, _ := json.Marshal(entry)
	log.Println(string(b))
}

func doBusinessLogic(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(ctx, "service_B_business_logic")
	defer span.End()

	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	if os.Getenv("CHAOS_ERROR") == "true" && rand.Intn(5) == 0 {
		err := errors.New("simulated business logic failure")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("chaos", "random_error"))
		return err
	}

	return nil
}

func callDatabase(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "service_C_db_call")
	defer span.End()

	if os.Getenv("CHAOS_DB_FAILURE") == "true" {
		err := errors.New("db timeout")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("chaos", "db_failure"))
		time.Sleep(2 * time.Second)
		return
	}

	if os.Getenv("CHAOS_SLOW_DB") == "true" {
		delay := time.Duration(800+rand.Intn(1200)) * time.Millisecond
		time.Sleep(delay)
		span.SetAttributes(attribute.String("chaos", "slow_db"))
		return
	}

	time.Sleep(time.Duration(20+rand.Intn(50)) * time.Millisecond)
}

func renderTemplate(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "template_rendering")
	defer span.End()

	time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)
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

	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		logJSON(ctx, "error", "failed to initialize tracer", map[string]interface{}{"error": err.Error()})
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
		ctx, span := tracer.Start(r.Context(), "handle_root")
		defer span.End()

		path := r.URL.Path
		method := r.Method
		status := http.StatusOK

		defer func() {
			requestCounter.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
			requestDuration.WithLabelValues(path, method).Observe(time.Since(start).Seconds())
		}()

		span.SetAttributes(
			semconv.HTTPRequestMethodKey.String(method),
			semconv.URLPathKey.String(path),
		)

		if err := doBusinessLogic(ctx, tracer); err != nil {
			span.SetStatus(codes.Error, "business logic failed")
			status = http.StatusInternalServerError

			logJSON(ctx, "error", "business logic failed", map[string]interface{}{
				"path":       path,
				"method":     method,
				"status":     status,
				"latency_ms": time.Since(start).Milliseconds(),
				"error":      err.Error(),
			})

			http.Error(w, "internal failure", status)
			return
		}

		callDatabase(ctx, tracer)
		renderTemplate(ctx, tracer)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"hello from grafana demo"}`))

		logJSON(ctx, "info", "handled request", map[string]interface{}{
			"path":       path,
			"method":     method,
			"status":     status,
			"latency_ms": time.Since(start).Milliseconds(),
		})
	})

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	logJSON(ctx, "info", "starting server", map[string]interface{}{"port": port})
	handler := otelhttp.NewHandler(mux, "http-server")
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logJSON(ctx, "error", "server failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}
