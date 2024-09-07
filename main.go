package main

import (
	"embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/NachoxMacho/supermarkethelper/database"
)

//go:embed public
var FS embed.FS

type httpFunc func(http.ResponseWriter, *http.Request) error

func ErrorHandler(handler httpFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.Error("dum dum error", "err", err.Error())
			fmt.Fprintf(w, "Internal Server Error %s", err)
		}
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		slog.Warn("Error loading environment file(s).", slog.String("err", err.Error()))
	}

	dbPath := os.Getenv("DB_PATH")
	dbDriver := os.Getenv("DB_DRIVER")

	err = database.Initialize(dbDriver, dbPath)
	if err != nil {
		log.Fatal("Database failed to initialize: " + err.Error())
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	mux := http.NewServeMux()
	mux.Handle("GET /public/*", http.StripPrefix("/", http.FileServerFS(FS)))
	mux.HandleFunc("GET /", ErrorHandler(Homepage))
	mux.HandleFunc("GET /products", ErrorHandler(GetProducts))
	// mux.HandleFunc("POST /products/:id", ErrorHandler(AddProduct))
	mux.HandleFunc("PUT /products/{id}/", ErrorHandler(ModifyProduct))
	mux.HandleFunc("POST /products", ErrorHandler(AddProduct))
	http.ListenAndServe(":42069", mux)
}
