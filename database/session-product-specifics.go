package database

import (
	"fmt"
	"slices"
)

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
		if c.SessionID == "" || c.ProductID == 0 {
			continue
		}
		sessionCategories = append(sessionCategories, c)
	}

	defer rows.Close()

	return sessionCategories, nil
}

func SetProductSpecific(sessionID string, productID int, boxPrice string, shelvesInStore string) error {

	sessionProductSpecifics, err := GetSessionProductSpecifics()
	if err != nil {
		return err
	}

	match := slices.ContainsFunc(sessionProductSpecifics, func(s SessionProductSpecific) bool {
		return s.SessionID == sessionID && s.ProductID == productID
	})

	db := ConnectDB()
	if !match {
		_, err = db.Exec("insert into session_product_specifics values (?, ?, ?, ?)", fmt.Sprintf("%d", productID), sessionID, boxPrice, shelvesInStore)
		return err
	}

	r, err := db.Exec("update session_product_specifics set box_price = ?, shelves_in_store = ? where session_id = ? and product_id = ?", boxPrice, shelvesInStore, sessionID, fmt.Sprintf("%d", productID))
	if err != nil {
		return err
	}

	ra, err := r.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Println("affected", ra, "rows")
	return nil
}
