package workers

import (
	"context"
	"fmt"

	"github.com/icco/cacophony/models"
	"github.com/icco/gutil/logging"
	"github.com/mattn/go-mastodon"
	"mvdan.cc/xurls/v2"
)

func Mastodon(ctx context.Context, server, clientID, clientSecret, accessToken string) error {
	log := logging.FromContext(ctx)
	if server == "" || clientID == "" || clientSecret == "" {
		return fmt.Errorf("server, client id and client secret required")
	}

	if accessToken == "" {
		return fmt.Errorf("user password and email cannot be empty string")
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       server,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
	})

	if err := c.AuthenticateApp(ctx); err != nil {
		return err
	}

	var statuses []*mastodon.Status
	var pg mastodon.Pagination
	limit := 1000
	for len(statuses) < limit && pg.MaxID != "" {
		timeline, err := c.GetTimelinePublic(ctx, false, &pg)
		if err != nil {
			return err
		}
		statuses = append(statuses, timeline...)
		log.Debugw("got statuses", "count", len(statuses), "pagination", pg)
	}

	for k, v := range statuses {
		log.Debugw("found toot", "count", k, "toot", v)

		var urls []string
		if v.Card != nil && v.Card.URL != "" {
			urls = append(urls, v.Card.URL)
		}

		rxStrict := xurls.Strict()
		contentUrls := rxStrict.FindAllString(v.Content, -1)
		urls = append(urls, contentUrls...)

		for _, url := range urls {
			if err := models.SaveMastodonURL(ctx, url, v.URL); err != nil {
				return fmt.Errorf("error saving url: %w", err)
			}
		}
	}

	return nil
}
