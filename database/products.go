package database

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/NachoxMacho/supermarkethelper/internal/traces"
	"github.com/NachoxMacho/supermarkethelper/types"
)

func GetProducts(ctx context.Context) ([]Product, error) {

	_, span := traces.SetupSpan(ctx)
	defer span.End()
	db := ConnectDB()

	rows, err := db.Query("select id, name, category_id, items_per_box, items_per_shelf, default_box_price, default_shelves_in_store from products;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.ItemsPerBox, &p.ItemsPerShelf, &p.DefaultBoxPrice, &p.DefaultShelvesInStore)
		if err != nil {
			return nil, err
		}
		if p.ID == 0 {
			continue
		}
		products = append(products, p)
	}

	defer rows.Close()

	return products, nil
}

func GetProduct(id int) (Product, error) {
	db := ConnectDB()

	rows, err := db.Query("select id, name, category_id, items_per_box, items_per_shelf, default_box_price, default_shelves_in_store from products where id = ?;", fmt.Sprintf("%d", id))
	if err != nil {
		return Product{}, err
	}

	// We have to loop through each row returned and construct the objects
	var p Product
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.ItemsPerBox, &p.ItemsPerShelf, &p.DefaultBoxPrice, &p.DefaultShelvesInStore)
		if err != nil {
			return Product{}, err
		}
	}

	defer rows.Close()

	return p, nil
}

func GetSessionProduct(sessionID string, productID int) (types.ProductItem, error) {

	db := ConnectDB()

	rows, err := db.Query(`
		select
			id,
			name,
			c.name,
			items_per_box,
			items_per_shelf,
			coalesce(sp.box_price, p.default_box_price),
			coalesce(sp.shelves_in_store, p.default_shelves_in_store),
		from
			products p
			inner join categories c on c.id = p.category_id
			left join product_specifics sp on p.id = sp.product_id and sp.session_id = ?
		where
			id = ?;
		`, sessionID, fmt.Sprintf("%d", productID))
	if err != nil {
		return types.ProductItem{}, err
	}

	// We have to loop through each row returned and construct the objects
	var p types.ProductItem
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Category, &p.ItemsPerBox, &p.ItemsPerShelf, &p.BoxPrice, &p.ShelvesInStore)
		if err != nil {
			return types.ProductItem{}, err
		}
	}

	defer rows.Close()

	return p, nil
}

func GetSessionProducts(sessionID string) ([]types.ProductItem, error) {

	db := ConnectDB()

	rows, err := db.Query(`
		select
			id,
			name,
			c.name,
			items_per_box,
			items_per_shelf,
			coalesce(sp.box_price, p.default_box_price),
			coalesce(sp.shelves_in_store, p.default_shelves_in_store),
		from
			products p
			inner join categories c on c.id = p.category_id
			left join product_specifics sp on p.id = sp.product_id and sp.session_id = ?
		;
		`, sessionID)
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var products []types.ProductItem
	for rows.Next() {
		var p types.ProductItem
		err = rows.Scan(&p.ID, &p.Name, &p.Category, &p.ItemsPerBox, &p.ItemsPerShelf, &p.BoxPrice, &p.ShelvesInStore)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	defer rows.Close()

	return products, nil
}
