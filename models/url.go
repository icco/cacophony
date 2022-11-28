package models

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

// SavedURL stores a single url seen in a tweet.
type SavedURL struct {
	Link         string
	TweetIDs     []string
	MastodonURLs []string
	CreatedAt    time.Time
	ModifiedAt   time.Time
}

// SomeSavedURLs returns a subset of most recently seen urls.
func SomeSavedURLs(ctx context.Context, limit int) ([]*SavedURL, error) {
	query := "SELECT link, created_at, modified_at, tweet_ids FROM saved_urls ORDER BY modified_at DESC LIMIT $1"
	rows, err := db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Errorw("query errored", "query", query, "limit", limit, zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var urls []*SavedURL
	for rows.Next() {
		url := new(SavedURL)
		if err := rows.Scan(&url.Link, &url.CreatedAt, &url.ModifiedAt, pq.Array(&url.TweetIDs)); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get some saved urls: %w", err)
	}
	return urls, nil
}

// AllSavedURLs returns all of the urls ever seen.
func AllSavedURLs(ctx context.Context) ([]*SavedURL, error) {
	query := "SELECT link, created_at, modified_at, tweet_ids FROM saved_urls ORDER BY modified_at DESC"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Errorw("query errored", "query", query, zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var urls []*SavedURL
	for rows.Next() {
		url := new(SavedURL)
		err := rows.Scan(&url.Link, &url.CreatedAt, &url.ModifiedAt, pq.Array(&url.TweetIDs))
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get all saved urls: %w", err)
	}
	return urls, nil
}

// SaveTwitterURL does an upsert on a URL.
func SaveTwitterURL(ctx context.Context, link, tweetID string) error {
	query := `
  INSERT into saved_urls (link, tweet_ids, created_at, modified_at)
  VALUES ($1, ARRAY [$2], transaction_timestamp(), transaction_timestamp())
  ON CONFLICT(link) DO UPDATE
  SET tweet_ids = saved_urls.tweet_ids || $2, modified_at = transaction_timestamp()
  WHERE NOT EXCLUDED.tweet_ids <@ saved_urls.tweet_ids;`
	if _, err := db.ExecContext(ctx, query, link, tweetID); err != nil {
		log.Errorw("twitter insert query errored", "query", query, "link", link, "tweet", tweetID, zap.Error(err))
		return err
	}

	return nil
}

// SaveMastodonURL does an upsert on a URL.
func SaveMastodonURL(ctx context.Context, link, mastoUrl string) error {
	query := `
  INSERT into saved_urls (link, mastodon_urls, created_at, modified_at)
  VALUES ($1, ARRAY [$2], transaction_timestamp(), transaction_timestamp())
  ON CONFLICT(link) DO UPDATE
  SET mastodon_urls = saved_urls.mastodon_urls || $2, modified_at = transaction_timestamp()
  WHERE NOT EXCLUDED.mastodon_urls <@ saved_urls.mastodon_urls;`
	if _, err := db.ExecContext(ctx, query, link, mastoUrl); err != nil {
		log.Errorw("mastodon insert query errored", "query", query, "link", link, "toot", mastoUrl, zap.Error(err))
		return err
	}

	return nil
}
