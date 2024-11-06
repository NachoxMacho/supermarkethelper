package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/NachoxMacho/supermarkethelper/database"
	"github.com/NachoxMacho/supermarkethelper/internal/traces"
	_ "net/http/pprof"
	pyroscope "github.com/grafana/pyroscope-go"
	pyroscope_pprof "github.com/grafana/pyroscope-go/http/pprof"
)

//go:embed public
var FS embed.FS

type httpFunc func(http.ResponseWriter, *http.Request) error

func ErrorHandler(handler httpFunc, logger slog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			logger.Error("dum dum error", "err", err.Error(), "trace", string(debug.Stack()))
			_, _ = fmt.Fprintf(w, "Internal Server Error %s", err)
		}
	}
}

func TraceInjector(handler httpFunc, traceProvider trace.Tracer) httpFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Inject the span into the request context
		ctx := r.Context()
		ctx, span := traceProvider.Start(ctx, r.URL.Path)
		ctx = traces.AddTraceProviderToContext(ctx, traceProvider)
		defer span.End()
		request := r.Clone(ctx)

		span.SetAttributes(
			attribute.String("method", r.Method),
			attribute.String("host", r.Host),
		)

		return handler(w, request)
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		slog.Warn("Error loading environment file(s).", slog.String("err", err.Error()))
	}

	pyroscopeAddr := os.Getenv("PYROSCOPE_ENDPOINT")

	if pyroscopeAddr != "" {
		pyroscope.Start(pyroscope.Config{
			ApplicationName: "supermarkethelper",
			ServerAddress: pyroscopeAddr,
			Logger: pyroscope.StandardLogger,
		})
	}

	traceProvider, shutdownFunc, err := traces.SetupTracer()
	if err != nil {
		slog.Error("error setting up traces", slog.String("error", err.Error()))
	}

	defer shutdownFunc(context.Background())

	dbPath := os.Getenv("DB_PATH")
	dbDriver := os.Getenv("DB_DRIVER")

	err = database.Initialize(dbDriver, dbPath)
	if err != nil {
		log.Fatal("Database failed to initialize: " + err.Error())
	}

	db := database.ConnectDB()
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))

	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	mux.HandleFunc("/debug/pprof/profile", pyroscope_pprof.Profile)
	mux.Handle("GET /public/", http.StripPrefix("/", http.FileServerFS(FS)))
	mux.HandleFunc("GET /session/{id}", ErrorHandler(TraceInjector(Homepage, traceProvider), *logger))
	mux.HandleFunc("GET /{$}", ErrorHandler(TraceInjector(Homepage, traceProvider), *logger))
	mux.HandleFunc("PUT /session/{id}/category/{category}", ErrorHandler(TraceInjector(SessionCategory(db, *logger), traceProvider), *logger))
	mux.HandleFunc("PUT /session/{id}/product/{product}/set", ErrorHandler(TraceInjector(SetProductSpecific, traceProvider), *logger))
	err = http.ListenAndServe(":42069", mux)
	if err != nil {
		slog.Error("Error starting server", slog.String("err", err.Error()))
	}
}
