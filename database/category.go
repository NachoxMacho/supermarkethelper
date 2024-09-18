package database

func GetCategories() ([]Category, error) {
	db := ConnectDB()

	rows, err := db.Query("select id, name, precursor from categories;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var categories []Category
	for rows.Next() {
		var c Category
		err = rows.Scan(&c.ID, &c.Name, &c.Precursor)
		if err != nil {
			return nil, err
		}
		if c.ID == 0 { continue }
		categories = append(categories, c)
	}

	defer rows.Close()

	return categories, nil
}

