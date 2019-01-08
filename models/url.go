package models

import (
	"context"
	"time"

	"github.com/lib/pq"
)

// SavedURL stores a single url seen in a tweet.
type SavedURL struct {
	Link       string
	TweetIDs   []string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// SomeSavedURLs returns a subset of most recently seen urls.
func SomeSavedURLs(ctx context.Context, limit int) ([]*SavedURL, error) {
	rows, err := db.QueryContext(ctx, "SELECT link, created_at, modified_at, tweet_ids FROM saved_urls ORDER BY modified_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*SavedURL, 0)
	for rows.Next() {
		url := new(SavedURL)
		err := rows.Scan(&url.Link, &url.CreatedAt, &url.ModifiedAt, pq.Array(&url.TweetIDs))
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

// AllSavedURLs returns all of the urls ever seen.
func AllSavedURLs(ctx context.Context) ([]*SavedURL, error) {
	rows, err := db.QueryContext(ctx, "SELECT link, created_at, modified_at, tweet_ids FROM saved_urls ORDER BY modified_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]*SavedURL, 0)
	for rows.Next() {
		url := new(SavedURL)
		err := rows.Scan(&url.Link, &url.CreatedAt, &url.ModifiedAt, pq.Array(&url.TweetIDs))
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

// SaveURL does an upsert on a URL.
func SaveURL(ctx context.Context, link string, tweetID string) error {
	query := `
  INSERT into saved_urls (link, tweet_ids, created_at, modified_at)
  VALUES ($1, ARRAY [$2], transaction_timestamp(), transaction_timestamp())
  ON CONFLICT(link) DO UPDATE
  SET tweet_ids = saved_urls.tweet_ids || $2, modified_at = transaction_timestamp()
  WHERE NOT EXCLUDED.tweet_ids <@ saved_urls.tweet_ids;
  `
	_, err := db.ExecContext(ctx, query, link, tweetID)
	if err != nil {
		log.Printf("Query errored: %+v, $1: %+v, $2: %+v", query, link, tweetID)
	}

	return err
}
