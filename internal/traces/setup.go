package traces

//
import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// var tracer trace.Tracer
func newExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	endpoint := os.Getenv("OTEL_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}
	return otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("supermarkethelper"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

type contextTPKey int

const (
	traceProviderKey contextTPKey = iota
)

func SetupTracer() (trace.Tracer, func(context.Context) error, error) {
	ctx := context.Background()
	// traceProvider := ctx.Value(traces.TraceProviderKey)

	exp, err := newExporter(ctx)

	if err != nil {
		return nil, nil, err
	}

	// Create a new tracer provider with a batch span processor and the given exporter.
	tp := newTraceProvider(exp)

	// otel.SetTracerProvider(tp)

	// Finally, set the tracer that can be used for this package.
	tracer := tp.Tracer("github.com/NachoxMacho/supermarkethelper")
	return tracer, tp.Shutdown, nil

	// ctx, span := tracer.Start(ctx, "test")
	//
	// span.End()
}

func GetTraceProviderFromContext(ctx context.Context) trace.Tracer {
	traceProvider := ctx.Value(traceProviderKey).(trace.Tracer)
	return traceProvider
}

func AddTraceProviderToContext(ctx context.Context, tp trace.Tracer) context.Context {
	return context.WithValue(ctx, traceProviderKey, tp)
}
