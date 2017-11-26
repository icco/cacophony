package models

import (
	"time"
)

type SavedUrl struct {
	Id         int64
	Link       string
	TweetIds   []string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func AllSavedUrls() ([]*SavedUrl, error) {
	rows, err := db.Query("SELECT * FROM saved_urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*SavedUrl, 0)
	for rows.Next() {
		url := new(SavedUrl)
		err := rows.Scan()
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
