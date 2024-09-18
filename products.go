package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"slices"

	"github.com/NachoxMacho/supermarkethelper/database"
	"github.com/NachoxMacho/supermarkethelper/types"
	"github.com/NachoxMacho/supermarkethelper/views/home"
)

// func GetProducts(w http.ResponseWriter, r *http.Request) error {
//
//		products, err := database.GetProducts()
//		if err != nil {
//			return err
//		}
//
//		orderPathValue := r.URL.Query().Get("order")
//		descendingOrder := orderPathValue == "desc"
//		sortType := r.URL.Query().Get("sort")
//		if sortType == "" {
//			sortType = "id"
//		}
//
//		fmt.Println("descendingOrder", descendingOrder, "sortType", sortType)
//
//		switch sortType {
//		case "id":
//			sort.SliceStable(products, func(i, j int) bool {
//				return products[i].ID > products[j].ID != descendingOrder
//			})
//		case "name":
//			sort.SliceStable(products, func(i, j int) bool {
//				return strings.ToLower(products[i].Name) > strings.ToLower(products[j].Name) != descendingOrder
//			})
//		case "category":
//			sort.SliceStable(products, func(i, j int) bool {
//				return strings.ToLower(products[i].Category) > strings.ToLower(products[j].Category) != descendingOrder
//			})
//		case "box_price":
//			sort.SliceStable(products, func(i, j int) bool {
//				return products[i].BoxPrice > products[j].BoxPrice != descendingOrder
//			})
//		case "price_per_item":
//			sort.SliceStable(products, func(i, j int) bool {
//				return GetMarketPrice(products[i]) > GetMarketPrice(products[j]) != descendingOrder
//			})
//		case "items_per_box":
//			sort.SliceStable(products, func(i, j int) bool {
//				return products[i].ItemsPerBox > products[j].ItemsPerBox != descendingOrder
//			})
//		case "boxes_per_shelf":
//			sort.SliceStable(products, func(i, j int) bool {
//				return GetBoxesPerShelf(products[i]) > GetBoxesPerShelf(products[j]) != descendingOrder
//			})
//		case "items_per_shelf":
//			sort.SliceStable(products, func(i, j int) bool {
//				return products[i].ItemsPerShelf > products[j].ItemsPerShelf != descendingOrder
//			})
//		case "shelves_in_store":
//			sort.SliceStable(products, func(i, j int) bool {
//				return products[i].ShelvesInStore > products[j].ShelvesInStore != descendingOrder
//			})
//		case "stocked_amount":
//			sort.SliceStable(products, func(i, j int) bool {
//				return GetTotalInventory(products[i]) > GetTotalInventory(products[j]) != descendingOrder
//			})
//		case "sale_price":
//			sort.SliceStable(products, func(i, j int) bool {
//				return GetSalePrice(products[i]) > GetSalePrice(products[j]) != descendingOrder
//			})
//		}
//
//		formattedProducts := make([]types.ProductItemOutput, len(products))
//
//		for i, p := range products {
//			if p.ID == 0 {
//				continue
//			}
//			formattedProducts[i] = FormatProduct(p)
//		}
//
//		return home.Index(formattedProducts, GetCategories(products), descendingOrder, sortType).Render(context.TODO(), w)
//	}
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

func GetCategoryNames() ([]string, error) {
	categories, err := database.GetCategories()
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(categories))
	for _, c := range categories {
		list = append(list, c.Name)
	}
	return list, nil
}

func Homepage(w http.ResponseWriter, r *http.Request) error {

	// TODO: make this work
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

	products, err := database.GetProducts()
	if err != nil {
		return err
	}

	categories, err := database.GetCategories()
	if err != nil {
		return err
	}

	productSpecifics, err := database.GetSessionProductSpecifics()
	if err != nil {
		return err
	}

	sessionCategories, err := database.GetSessionCategories()
	if err != nil {
		return err
	}
	enabledCategoryIDs := make([]int, 0, len(categories))
	for _, c := range sessionCategories {
		if c.SessionID == id {
			enabledCategoryIDs = append(enabledCategoryIDs, c.CategoryID)
		}
	}

	formattedProducts := make([]types.ProductItemOutput, 0, len(products))

	for _, p := range products {
		if p.ID == 0 {
			continue
		}

		if slices.Contains(enabledCategoryIDs, p.CategoryID) {

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
			formattedProducts = append(formattedProducts, FormatProduct(newProduct))
		}
		// formattedProducts[i] = FormatProduct(p)
	}

	fmt.Println("Rendering")
	list, err := GetCategoryNames()
	if err != nil {
		return err
	}


	return home.Index(formattedProducts, list, false, "id").Render(context.TODO(), w)
}

func SessionCategory(w http.ResponseWriter, r *http.Request) error {

	id := r.PathValue("id")
	if id == "" {
		return fmt.Errorf("missing id: %s", r.URL.Path)
	}

	category := r.PathValue("category")
	if category == "" {
		return fmt.Errorf("missing category: %s", r.URL.Path)
	}

	fmt.Println("Adding category", category, "to", id)

	err := database.AddSessionCategory(id, category)
	if err != nil {
		return err
	}

	return nil
}
