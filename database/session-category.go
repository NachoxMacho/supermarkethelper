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

func AddSessionCategory(session_id string, category string) error {
	categories, err := GetCategories()
	if err != nil {
		return err
	}
	matchingCategoryID := 0
	for _, c := range categories {
		if strings.ToLower(c.Name) == strings.ToLower(category) {
			matchingCategoryID = c.ID
		}
	}

	sessionCategories, err := GetSessionCategories()
	if err != nil {
		return err
	}

	match := slices.ContainsFunc(sessionCategories, func(s SessionCategory) bool {
		return s.SessionID == session_id && s.CategoryID == matchingCategoryID
	})

	if match {
		return nil
	}

	db := ConnectDB()
	_, err = db.Exec("insert into session_categories values (?, ?);", session_id, fmt.Sprintf("%d",matchingCategoryID))
	return err
}
