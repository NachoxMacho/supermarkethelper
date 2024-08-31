package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

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

	orderPathValue := r.URL.Query().Get("order")
	descendingOrder := orderPathValue == "desc"
	sortType := r.URL.Query().Get("sort")
	if sortType == "" {
		sortType = "id"
	}


	fmt.Println("descendingOrder", descendingOrder, "sortType", sortType)

	switch sortType {
	case "id":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].ID > products[j].ID != descendingOrder
		})
	case "name":
		sort.SliceStable(products, func(i, j int) bool {
			return strings.ToLower(products[i].Name) > strings.ToLower(products[j].Name) != descendingOrder
		})
	case "category":
		sort.SliceStable(products, func(i, j int) bool {
			return strings.ToLower(products[i].Category) > strings.ToLower(products[j].Category) != descendingOrder
		})
	case "box_price":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].BoxPrice > products[j].BoxPrice != descendingOrder
		})
	case "price_per_item":
		sort.SliceStable(products, func(i, j int) bool {
			return GetMarketPrice(products[i]) > GetMarketPrice(products[j]) != descendingOrder
		})
	case "boxes_per_shelf":
		sort.SliceStable(products, func(i, j int) bool {
			return GetBoxesPerShelf(products[i]) > GetBoxesPerShelf(products[j]) != descendingOrder
		})
	case "items_per_shelf":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].ItemsPerShelf > products[j].ItemsPerShelf != descendingOrder
		})
	case "shelves_in_store":
		sort.SliceStable(products, func(i, j int) bool {
			return products[i].ShelvesInStore > products[j].ShelvesInStore != descendingOrder
		})
	case "stocked_amount":
		sort.SliceStable(products, func(i, j int) bool {
			return GetTotalInventory(products[i]) > GetTotalInventory(products[j]) != descendingOrder
		})
	case "sale_price":
		sort.SliceStable(products, func(i, j int) bool {
			return GetSalePrice(products[i]) > GetSalePrice(products[j]) != descendingOrder
		})
	}

	formattedProducts := make([]types.ProductItemOutput, len(products))

	for i, p := range products {
		if p.ID == 0 {
			continue
		}
		formattedProducts[i] = FormatProduct(p)
	}

	return home.Index(formattedProducts, GetCategories(products), descendingOrder, sortType).Render(context.TODO(), w)
}

func GetLargestID() int {
	products, err := database.GetAllProducts()
	if err != nil {
		return -1
	}
	largest := 0
	for _, p := range products {
		if p.ID > largest {
			largest = p.ID
		}
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
	return math.Ceil(f/precision) * precision
}

func GetSalePrice(p types.ProductItem) float64 {
	return RoundToPrecision(GetMarketPrice(p)*1.45, 0.25)
}

func GetCategories(products []types.ProductItem) []string {
	categories := make([]string, 0, len(products))

	for _, p := range products {
		if p.Category == "" {
			continue
		}
		if !slices.Contains(categories, p.Category) {
			categories = append(categories, p.Category)
		}
	}

	return categories
}

func AddProduct(w http.ResponseWriter, r *http.Request) error {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	slog.Debug("recieved product to add", slog.String("body", string(bodyBytes)))

	newProduct := types.ProductItem{}

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		u, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			return err
		}

		if u.Get("box_price") != "" {
			boxPrice, err := strconv.ParseFloat(u.Get("box_price"), 64)
			if err != nil {
				return err
			}
			newProduct.BoxPrice = boxPrice
		}

		if u.Get("items_per_shelf") != "" {
			itemsPerShelf, err := strconv.Atoi(u.Get("items_per_shelf"))
			if err != nil {
				return err
			}
			newProduct.ItemsPerShelf = itemsPerShelf
		}

		if u.Get("items_per_box") != "" {
			itemsPerBox, err := strconv.Atoi(u.Get("items_per_box"))
			if err != nil {
				return err
			}
			newProduct.ItemsPerBox = itemsPerBox
		}

		if u.Get("shelves_in_store") != "" {
			shelvesInStore, err := strconv.Atoi(u.Get("shelves_in_store"))
			if err != nil {
				return err
			}
			newProduct.ShelvesInStore = shelvesInStore
		}
		newProduct.Name = u.Get("name")
		newProduct.Category = u.Get("category")
	case "application/json":
		err = json.Unmarshal(bodyBytes, &newProduct)
		if err != nil {
			return err
		}
	}

	if newProduct.ID == 0 {
		newProduct.ID = GetLargestID() + 1
	}

	products, err := database.GetAllProducts()
	if err != nil {
		return err
	}

	for _, p := range products {
		fmt.Println("comparing", p.ID, "==", newProduct.ID)
		if p.ID == newProduct.ID {
			w.WriteHeader(400)
			return fmt.Errorf("item already exists with ID: %d", p.ID)
		}
	}

	database.AddProduct(newProduct)

	formattedProducts := make([]types.ProductItemOutput, len(products))

	for i, p := range products {
		if p.ID == 0 {
			continue
		}
		formattedProducts[i] = FormatProduct(p)
	}

	return home.Index(formattedProducts, GetCategories(products), false, "id").Render(context.TODO(), w)
}

func ModifyProduct(w http.ResponseWriter, r *http.Request) error {

	pathID := r.PathValue("id")
	if pathID == "" {
		return errors.New("id missing in path")
	}
	id, err := strconv.Atoi(pathID)
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	slog.Debug("recieved product to modify", slog.String("body", string(bodyBytes)))

	incomingProduct := types.ProductItem{}

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		u, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			return err
		}

		if u.Get("box_price") != "" {
			boxPrice, err := strconv.ParseFloat(u.Get("box_price"), 64)
			if err != nil {
				return err
			}
			incomingProduct.BoxPrice = boxPrice
		}

		if u.Get("items_per_shelf") != "" {
			itemsPerShelf, err := strconv.Atoi(u.Get("items_per_shelf"))
			if err != nil {
				return err
			}
			incomingProduct.ItemsPerShelf = itemsPerShelf
		}

		if u.Get("items_per_box") != "" {
			itemsPerBox, err := strconv.Atoi(u.Get("items_per_box"))
			if err != nil {
				return err
			}
			incomingProduct.ItemsPerBox = itemsPerBox
		}

		if u.Get("shelves_in_store") != "" {
			shelvesInStore, err := strconv.Atoi(u.Get("shelves_in_store"))
			if err != nil {
				return err
			}
			incomingProduct.ShelvesInStore = shelvesInStore
		}
		incomingProduct.Name = u.Get("name")
		incomingProduct.Category = u.Get("category")
	case "application/json":
		err = json.Unmarshal(bodyBytes, &incomingProduct)
		if err != nil {
			return err
		}
	}

	incomingProduct.ID = id

	fmt.Println("Merging Objects")
	products, err := database.GetAllProducts()
	if err != nil {
		return err
	}

	for _, p := range products {
		fmt.Println("comparing", p.ID, "==", incomingProduct.ID)
		if p.ID == incomingProduct.ID {
			incomingProduct = MergeProducts(p, incomingProduct)
			break
		}
	}

	fmt.Println("Modifying Product to Database")

	err = database.ModifyProduct(incomingProduct)
	if err != nil {
		w.WriteHeader(400)
		return err
	}
	fmt.Println("Formatting")

	fp := FormatProduct(incomingProduct)

	return home.Row(fp, GetCategories(products)).Render(context.TODO(), w)
}

// This is the HTMX endpoint for the form
func AddOrModifyProduct(w http.ResponseWriter, r *http.Request) error {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	fmt.Println(string(bodyBytes))

	incomingProduct := types.ProductItem{}

	switch r.Header.Get("Content-Type") {
	case "application/x-www-form-urlencoded":
		u, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			return err
		}

		if u.Get("id") != "" {
			id, err := strconv.Atoi(u.Get("id"))
			if err != nil {
				return err
			}
			incomingProduct.ID = id
		}
		if u.Get("box_price") != "" {
			boxPrice, err := strconv.ParseFloat(u.Get("box_price"), 64)
			if err != nil {
				return err
			}
			incomingProduct.BoxPrice = boxPrice
		}

		if u.Get("items_per_shelf") != "" {
			itemsPerShelf, err := strconv.Atoi(u.Get("items_per_shelf"))
			if err != nil {
				return err
			}
			incomingProduct.ItemsPerShelf = itemsPerShelf
		}

		if u.Get("items_per_box") != "" {
			itemsPerBox, err := strconv.Atoi(u.Get("items_per_box"))
			if err != nil {
				return err
			}
			incomingProduct.ItemsPerBox = itemsPerBox
		}

		if u.Get("shelves_in_store") != "" {
			shelvesInStore, err := strconv.Atoi(u.Get("shelves_in_store"))
			if err != nil {
				return err
			}
			incomingProduct.ShelvesInStore = shelvesInStore
		}
		incomingProduct.Name = u.Get("name")
		incomingProduct.Category = u.Get("category")
	}

	fmt.Printf("Created Object %v\n", incomingProduct)

	products, err := database.GetAllProducts()
	if err != nil {
		return err
	}

	found := false

	for i, p := range products {
		if p.ID == incomingProduct.ID {
			incomingProduct = MergeProducts(p, incomingProduct)
			fmt.Printf("Modifying Object %v\n", incomingProduct)
			database.ModifyProduct(incomingProduct)
			products[i] = incomingProduct
			found = true
			break
		}
	}

	if !found && incomingProduct.ID != 0 {
		database.AddProduct(incomingProduct)
		products = append(products, incomingProduct)
	}

	formattedProducts := make([]types.ProductItemOutput, len(products))

	for i, p := range products {
		if p.ID == 0 {
			continue
		}
		formattedProducts[i] = FormatProduct(p)
	}

	return home.Index(formattedProducts, GetCategories(products), false, "id").Render(context.TODO(), w)
}

func MergeProducts(base types.ProductItem, override types.ProductItem) types.ProductItem {
	if override.Name != "" {
		fmt.Println("overriding name with", override.Name)
		base.Name = override.Name
	}
	if override.Category != "" {
		fmt.Println("overriding category with", override.Category)
		base.Category = override.Category
	}
	if override.BoxPrice != 0 {
		fmt.Println("overriding box_price with", override.BoxPrice)
		base.BoxPrice = override.BoxPrice
	}
	if override.ItemsPerBox != 0 {
		fmt.Println("overriding items_per_box with", override.ItemsPerBox)
		base.ItemsPerBox = override.ItemsPerBox
	}
	if override.ItemsPerShelf != 0 {
		fmt.Println("overriding items_per_shelf with", override.ItemsPerShelf)
		base.ItemsPerShelf = override.ItemsPerShelf
	}
	fmt.Println(override.ShelvesInStore)
	if override.ShelvesInStore != 0 {
		fmt.Println("overriding shelves_in_store with", override.ShelvesInStore)
		base.ShelvesInStore = override.ShelvesInStore
	}
	return base
}

func FormatProduct(p types.ProductItem) types.ProductItemOutput {
	fp := types.ProductItemOutput{}
	fp.ID = fmt.Sprintf("%d", p.ID)
	fp.BoxPrice = fmt.Sprintf("%.2f", p.BoxPrice)
	fp.Name = p.Name
	fp.ItemsPerShelf = fmt.Sprintf("%d", p.ItemsPerShelf)
	fp.PricePerItem = fmt.Sprintf("%.2f", GetMarketPrice(p))
	fp.Category = p.Category
	fp.ItemsPerBox = fmt.Sprintf("%d", p.ItemsPerBox)
	fp.BoxesPerShelf = fmt.Sprintf("%.2f", GetBoxesPerShelf(p))
	fp.ShelvesInStore = fmt.Sprintf("%d", p.ShelvesInStore)
	fp.StockedAmount = fmt.Sprintf("%d", GetTotalInventory(p))
	fp.SalePrice = fmt.Sprintf("%.2f", GetSalePrice(p))

	return fp
}

type httpFunc func(http.ResponseWriter, *http.Request) error

func ErrorHandler(handler httpFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			slog.Error("dum dum error", "err", err.Error())
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
		if p.ID == 0 {
			continue
		}
		formattedProducts[i] = FormatProduct(p)
	}

	fmt.Println("Rendering")

	return home.Index(formattedProducts, GetCategories(products), false, "id").Render(context.TODO(), w)
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
	// mux.HandleFunc("POST /products/:id", ErrorHandler(AddProduct))
	mux.HandleFunc("PUT /products/{id}/", ErrorHandler(ModifyProduct))
	mux.HandleFunc("POST /products", ErrorHandler(AddProduct))
	http.ListenAndServe(":42069", mux)
}
