package database

import (
	"fmt"
	"slices"
	"strings"
)

func GetSessionCategories() ([]SessionCategory, error) {
	db := ConnectDB()

	rows, err := db.Query("select session_id, category_id from session_categories;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var sessionCategories []SessionCategory
	for rows.Next() {
		var c SessionCategory
		err = rows.Scan(&c.SessionID, &c.CategoryID)
		if err != nil {
			return nil, err
		}
		if c.SessionID == "" || c.CategoryID == 0 {
			continue
		}
		sessionCategories = append(sessionCategories, c)
	}

	defer rows.Close()

	return sessionCategories, nil
}

func ToggleSessionCategory(sessionID string, category string) error {
	categories, err := GetCategories()
	if err != nil {
		return err
	}
	matchingCategoryID := 0
	for _, c := range categories {
		if strings.EqualFold(c.Name, category) {
			matchingCategoryID = c.ID
		}
	}

	sessionCategories, err := GetSessionCategories()
	if err != nil {
		return err
	}

	match := slices.ContainsFunc(sessionCategories, func(s SessionCategory) bool {
		return s.SessionID == sessionID && s.CategoryID == matchingCategoryID
	})

	db := ConnectDB()
	if match {
		fmt.Println("Removing category", matchingCategoryID, "to", sessionID)
		stmt, err := db.Prepare(`delete from session_categories where session_id = ? and category_id = ? COLLATE NOCASE;`)
		if err != nil {
			return err
		}
		r, err := stmt.Exec(sessionID, fmt.Sprintf("%d", matchingCategoryID))
		if err != nil {
			return err
		}
		ra, err := r.RowsAffected()
		if err != nil {
			return err
		}
		fmt.Println("affected", ra, "rows")
		return err
	}

	fmt.Println("Adding category", category, "to", sessionID)
	_, err = db.Exec("insert into session_categories values (?, ?);", sessionID, fmt.Sprintf("%d", matchingCategoryID))
	return err
}
