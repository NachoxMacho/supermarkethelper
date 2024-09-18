package database

func GetSessionProductSpecifics() ([]SessionProductSpecific, error) {
	db := ConnectDB()

	rows, err := db.Query("select product_id, session_id, box_price, shelves_in_store from session_product_specifics;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var sessionCategories []SessionProductSpecific
	for rows.Next() {
		var c SessionProductSpecific
		err = rows.Scan(&c.ProductID, &c.SessionID, &c.BoxPrice, &c.ShelvesInStore)
		if err != nil {
			return nil, err
		}
		if c.SessionID == "" || c.ProductID == 0 { continue }
		sessionCategories = append(sessionCategories, c)
	}

	defer rows.Close()

	return sessionCategories, nil
}

