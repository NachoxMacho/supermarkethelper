package database

import (
	"context"
	"errors"

	"github.com/NachoxMacho/supermarkethelper/internal/traces"
)

func GetCategories(ctx context.Context) (categories []Category, err error) {

	_, span := traces.SetupSpan(ctx)
	defer span.End()

	db := ConnectDB()

	rows, err := db.Query("select id, name, precursor from categories;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	for rows.Next() {
		var c Category
		err = rows.Scan(&c.ID, &c.Name, &c.Precursor)
		if err != nil {
			return nil, err
		}
		if c.ID == 0 {
			continue
		}
		categories = append(categories, c)
	}

	defer func() {
		closingErr := rows.Close()
		if err != nil {
			errors.Join(err, closingErr)
		}
	}()

	return categories, nil
}
