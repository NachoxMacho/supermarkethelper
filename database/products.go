package database

import (
	"fmt"

	"github.com/NachoxMacho/supermarkethelper/types"
)

func GetAllProducts() ([]types.ProductItem, error) {
	db := ConnectDB()

	rows, err := db.Query("select * from products;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var products []types.ProductItem
	for rows.Next() {
		var p types.ProductItem
		err = rows.Scan(&p.ID, &p.Name, &p.Category, &p.BoxPrice, &p.ItemsPerBox, &p.ShelvesInStore, &p.ItemsPerShelf)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	defer rows.Close()

	return products, nil
}

func AddProduct(p types.ProductItem) error {
	db := ConnectDB()

	result, err := db.Exec("insert into products values (?, ?, ?, ?, ?, ?, ?);",
		fmt.Sprintf("%d", p.ID),
		p.Name,
		p.Category,
		fmt.Sprintf("%.2f", p.BoxPrice),
		fmt.Sprintf("%d", p.ItemsPerBox),
		fmt.Sprintf("%d", p.ShelvesInStore),
		fmt.Sprintf("%d", p.ItemsPerShelf),
	)

	if err != nil {
		return err
	}
	ra, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if ra != 1 {
		return fmt.Errorf("unknown change %d rows changed", ra)
	}

	return nil
}

func ModifyProduct(p types.ProductItem) error {
	db := ConnectDB()

	result, err := db.Exec("update products set name = ?, category = ?, box_price = ?, items_per_box = ?, shelves_in_store = ?, items_per_shelf = ? where id = ? limit 1;",
		p.Name,
		p.Category,
		fmt.Sprintf("%.2f", p.BoxPrice),
		fmt.Sprintf("%d", p.ItemsPerBox),
		fmt.Sprintf("%d", p.ShelvesInStore),
		fmt.Sprintf("%d", p.ItemsPerShelf),
		fmt.Sprintf("%d", p.ID),
	)

	if err != nil {
		return err
	}
	ra, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if ra != 1 {
		return fmt.Errorf("unknown change %d rows changed", ra)
	}

	return nil

}
