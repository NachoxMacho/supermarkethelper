package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/NachoxMacho/supermarkethelper/database"
	"github.com/NachoxMacho/supermarkethelper/internal/traces"
	"github.com/NachoxMacho/supermarkethelper/types"
	"github.com/NachoxMacho/supermarkethelper/views/home"
)

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

func MergeProducts(base types.ProductItem, override types.ProductItem) types.ProductItem {
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Category != "" {
		base.Category = override.Category
	}
	if override.BoxPrice != 0 {
		base.BoxPrice = override.BoxPrice
	}
	if override.ItemsPerBox != 0 {
		base.ItemsPerBox = override.ItemsPerBox
	}
	if override.ItemsPerShelf != 0 {
		base.ItemsPerShelf = override.ItemsPerShelf
	}
	if override.ShelvesInStore != 0 {
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

func GetCategoryNames(ctx context.Context) ([]types.CategoryToggleOutput, error) {

	ctx, span := traces.SetupSpan(ctx)
	defer span.End()

	categories, err := database.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]types.CategoryToggleOutput, 0, len(categories))
	for _, c := range categories {
		list = append(list, types.CategoryToggleOutput{Name: c.Name, ID: c.ID})
	}
	return list, nil
}

func Homepage(w http.ResponseWriter, r *http.Request) error {

	ctx := r.Context()
	ctx, span := traces.SetupSpan(ctx)
	defer span.End()

	id := r.PathValue("id")
	if id == "" {
		session, err := database.AddSession()
		if err != nil {
			return err
		}
		http.Redirect(w, r, fmt.Sprintf("/%s", session.ID), http.StatusTemporaryRedirect)
		return nil
	}

	// Should fetch session id from database, but technically not needed at the moment

	products, err := database.GetProducts(ctx)
	if err != nil {
		return err
	}

	categories, err := database.GetCategories(ctx)
	if err != nil {
		return err
	}

	categoryMap := map[int]database.Category{}
	for _, c := range categories {
		categoryMap[c.ID] = c
	}

	productSpecifics, err := database.GetSessionProductSpecifics(ctx)
	if err != nil {
		return err
	}

	sessionCategories, err := database.GetSessionCategories(ctx)
	if err != nil {
		return err
	}
	enabledCategories := make([]database.Category, 0, len(categories))

	for _, c := range sessionCategories {
		if c.SessionID == id {
			enabledCategories = append(enabledCategories, categoryMap[c.CategoryID])
		}
	}

	formattedProducts := make([]types.ProductItemOutput, 0, len(products))

	for _, p := range products {
		if p.ID == 0 {
			continue
		}

		if !slices.ContainsFunc(enabledCategories, func(c database.Category) bool { return c.ID == p.CategoryID }) {
			continue
		}

		category := ""
		specifics := database.SessionProductSpecific{}
		for _, c := range categories {
			if c.ID == p.CategoryID {
				category = c.Name
				break
			}
		}

		for _, s := range productSpecifics {
			if s.ProductID == p.ID && s.SessionID == id {
				specifics = s
				break
			}
		}

		newProduct := types.ProductItem{
			ID:             p.ID,
			Category:       category,
			Name:           p.Name,
			ItemsPerBox:    p.ItemsPerBox,
			ItemsPerShelf:  p.ItemsPerShelf,
			BoxPrice:       specifics.BoxPrice,
			ShelvesInStore: specifics.ShelvesInStore,
		}

		if newProduct.BoxPrice == 0 {
			newProduct.BoxPrice = p.DefaultBoxPrice
		}
		if newProduct.ShelvesInStore == 0 {
			newProduct.ShelvesInStore = p.DefaultShelvesInStore
		}
		formattedProducts = append(formattedProducts, FormatProduct(newProduct))
	}

	list, err := GetCategoryNames(ctx)
	if err != nil {
		return err
	}

	for i, c := range list {
		if slices.ContainsFunc(enabledCategories, func(cat database.Category) bool { return cat.ID == c.ID }) {
			list[i].Selected = true
		}
	}
	return home.Index(id, formattedProducts, list, false, "id").Render(context.TODO(), w)
}

func SessionCategory(db *sql.DB, logger slog.Logger) func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {

		ctx := r.Context()
		ctx, span := traces.SetupSpan(ctx)
		defer span.End()

		id := r.PathValue("id")
		if id == "" {
			return fmt.Errorf("missing id: %s", r.URL.Path)
		}

		category := r.PathValue("category")
		if category == "" {
			return fmt.Errorf("missing category: %s", r.URL.Path)
		}
		category, err := url.QueryUnescape(category)
		if err != nil {
			return fmt.Errorf("failed to unescape category: %w", err)
		}

		// logger.Info("Adding category", category, "to", id)

		err = database.ToggleSessionCategory(id, category, ctx)
		if err != nil {
			return err
		}

		return nil
	}
}

func SetProductSpecific(w http.ResponseWriter, r *http.Request) error {

	ctx := r.Context()
	ctx, span := traces.SetupSpan(ctx)
	defer span.End()

	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("missing id: %s", r.URL.Path)
	}

	product := r.PathValue("product")
	if product == "" {
		return fmt.Errorf("missing product: %s", r.URL.Path)
	}

	productID, err := strconv.Atoi(product)
	if err != nil {
		return err
	}

	err = r.ParseForm()
	if err != nil {
		return err
	}

	boxPrice := 0.0
	shelvesInStore := 0

	if boxPriceInput := r.PostForm.Get("box_price"); boxPriceInput != "" {
		boxPrice, err = strconv.ParseFloat(boxPriceInput, 64)
		if err != nil {
			return err
		}
	}

	if shelvesInStoreInput := r.PostForm.Get("box_price"); shelvesInStoreInput != "" {
		shelvesInStore, err = strconv.Atoi(shelvesInStoreInput)
		if err != nil {
			return err
		}
	}

	if shelvesInStore == 0 && boxPrice == 0 {
		return nil
	}

	err = database.SetProductSpecific(id, productID, fmt.Sprintf("%.2f", boxPrice), fmt.Sprintf("%d", shelvesInStore), ctx)
	if err != nil {
		return err
	}

	return nil
}
