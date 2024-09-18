package database

import "github.com/google/uuid"

func GetSessions() ([]Session, error) {
	db := ConnectDB()

	rows, err := db.Query("select id from sessions;")
	if err != nil {
		return nil, err
	}

	// We have to loop through each row returned and construct the objects
	var sessions []Session
	for rows.Next() {
		var s Session
		err = rows.Scan(&s.ID)
		if err != nil {
			return nil, err
		}
		if s.ID == "" {
			continue
		}
		sessions = append(sessions, s)
	}

	defer rows.Close()

	return sessions, nil
}

func AddSession() (Session, error) {
	currentSessions, err := GetSessions()
	if err != nil {
		return Session{}, err
	}
	contains := true
	newUUID := uuid.New()

	for contains {
		contains = false
		for _, s := range currentSessions {
			if s.ID == newUUID.String() {
				contains = true
				newUUID = uuid.New()
				break
			}
		}
	}

	db := ConnectDB()
	_, err = db.Exec("insert into sessions values (?);", newUUID.String())
	if err != nil {
		return Session{}, err
	}

	return Session{ID: newUUID.String()}, nil
}
