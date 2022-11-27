package workers

import (
	"context"
	"fmt"

	"github.com/icco/gutil/logging"
	"github.com/mattn/go-mastodon"
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

	timeline, err := c.GetTimelinePublic(ctx, false, nil)
	if err != nil {
		return err
	}

	for i := len(timeline) - 1; i >= 0; i-- {
		log.Debugw("found toot", "toot", timeline[i])
	}

	return nil
}
