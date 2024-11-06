package traces

import (
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

func SetupSpan(ctx context.Context) (context.Context, trace.Span) {

	tp := GetTraceProviderFromContext(ctx)

	pc, file, lineNumber, ok := runtime.Caller(1)
	if !ok {
		return ctx, nil
	}

	details := runtime.FuncForPC(pc)
	if details == nil {
		return ctx, nil
	}

	ctx, span := tp.Start(ctx, details.Name())
	span.SetAttributes(
		attribute.String("function-name", details.Name()),
		attribute.Int("line-number", lineNumber),
		attribute.String("file-name", file),
	)

	return ctx, span
}
