package database

func GetProducts() ([]Product, error) {
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
		if p.ID == 0 { continue }
		products = append(products, p)
	}

	defer rows.Close()

	return products, nil
}

