package models

import (
	"log"
	"time"

	"github.com/lib/pq"
)

type SavedUrl struct {
	Link       string
	TweetIds   []string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func AllSavedUrls() ([]*SavedUrl, error) {
	rows, err := db.Query("SELECT link, created_at, modified_at, tweet_ids FROM saved_urls ORDER BY modified_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*SavedUrl, 0)
	for rows.Next() {
		url := new(SavedUrl)
		err := rows.Scan(&url.Link, &url.CreatedAt, &url.ModifiedAt, pq.Array(&url.TweetIds))
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

func SaveUrl(link string, tweetId string) error {
	query := `
  INSERT into saved_urls (link, tweet_ids, created_at, modified_at)
  VALUES ($1, ARRAY [$2], transaction_timestamp(), transaction_timestamp())
  ON CONFLICT(link) DO UPDATE
  SET tweet_ids = saved_urls.tweet_ids || $2, modified_at = transaction_timestamp()
  WHERE NOT EXCLUDED.tweet_ids <@ saved_urls.tweet_ids;
  `
	_, err := db.Exec(query, link, tweetId)
	if err != nil {
		log.Printf("Query errored: %+v, $1: %+v, $2: %+v", query, link, tweetId)
	}

	return err
}
