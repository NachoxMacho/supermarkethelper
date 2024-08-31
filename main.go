package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/NachoxMacho/supermarkethelper/database"
	"github.com/NachoxMacho/supermarkethelper/types"
	"github.com/NachoxMacho/supermarkethelper/views/home"
)

//go:embed public
var FS embed.FS

func GetProducts(w http.ResponseWriter, r *http.Request) error {

	products, err := database.GetAllProducts()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(products)
	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(jsonData)
	return nil
}

func GetLargestID() int {
	products, err := database.GetAllProducts()
	if err != nil {
		return -1
	}
	largest := 0
	for _, p := range products {
		if p.ID > largest { largest = p.ID }
	}
	return largest
}

func GetBoxesPerShelf(p types.ProductItem) float64 {
	return float64(p.ItemsPerShelf) / float64(p.ItemsPerBox)
}

func GetMarketPrice(p types.ProductItem) float64 {
	return p.BoxPrice / float64(p.ItemsPerBox)
}

// n is the number of shelves
func GetTotalInventory(p types.ProductItem) int {
	return p.ItemsPerShelf * p.ShelvesInStore
}

func RoundToPrecision(f float64, precision float64) float64 {
	return math.Ceil(f / precision) * precision
}

func GetSalePrice(marketPrice float64, mult float64) float64 {
	return RoundToPrecision(marketPrice*mult, 0.25)
}

func AddProduct(w http.ResponseWriter, r *http.Request) error {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	slog.Debug("recieved product to add", slog.String("body", string(body)))

	newProduct := types.ProductItem{}

	err = json.Unmarshal(body, &newProduct)
	if err != nil {
		return err
	}

	if newProduct.ID == 0 {
		newProduct.ID = GetLargestID() + 1
	}

	database.AddProduct(newProduct)


	w.WriteHeader(200)
	return nil
}


func ModifyProduct(w http.ResponseWriter, r *http.Request) error {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	slog.Debug("recieved product to modify", slog.String("body", string(body)))

	newProduct := types.ProductItem{}

	err = json.Unmarshal(body, &newProduct)
	if err != nil {
		return err
	}
	err = database.ModifyProduct(newProduct)
	if err != nil {
		w.WriteHeader(400)
		return err
	}
	w.WriteHeader(200)
	return nil
}

type httpFunc func (http.ResponseWriter, *http.Request) error

func ErrorHandler(handler httpFunc) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Internal Server Error %s", err)
		}
	}
}

func Homepage(w http.ResponseWriter, r *http.Request) error {

	products, err := database.GetAllProducts()
	if err != nil {
		return err
	}

	formattedProducts := make([]types.ProductItemOutput, len(products))

	for i, p := range products {
		formattedProducts[i].ID = fmt.Sprintf("%d", p.ID)
		formattedProducts[i].BoxPrice = fmt.Sprintf("%.2f", p.BoxPrice)
		formattedProducts[i].Name = p.Name
		formattedProducts[i].ItemsPerShelf = fmt.Sprintf("%d", p.ItemsPerShelf)
		formattedProducts[i].PricePerItem = fmt.Sprintf("%.2f", GetMarketPrice(p))
		formattedProducts[i].Category = p.Category
		formattedProducts[i].ItemsPerBox = fmt.Sprintf("%d",p.ItemsPerBox)
		formattedProducts[i].BoxesPerShelf = fmt.Sprintf("%.2f", GetBoxesPerShelf(p))
		formattedProducts[i].ShelvesInStore = fmt.Sprintf("%d", p.ShelvesInStore)
		formattedProducts[i].StockedAmount = fmt.Sprintf("%d", GetTotalInventory(p))
		formattedProducts[i].SalePrice = fmt.Sprintf("%.2f", GetSalePrice(GetMarketPrice(p), 1.45))
	}


	return home.Index(formattedProducts).Render(context.TODO(), w)
}

func main() {
	err := database.Initialize()
	if err != nil {
		log.Fatal("Database failed to initialize: " + err.Error())
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	mux := http.NewServeMux()
	mux.Handle("GET /public/*", http.StripPrefix("/", http.FileServerFS(FS)))
	mux.HandleFunc("GET /", ErrorHandler(Homepage))
	mux.HandleFunc("GET /products", ErrorHandler(GetProducts))
	mux.HandleFunc("POST /products", ErrorHandler(AddProduct))
	mux.HandleFunc("PUT /products", ErrorHandler(ModifyProduct))
	http.ListenAndServe(":42069", mux)
}
