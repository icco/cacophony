package workers

import (
	"context"
	"fmt"

	"github.com/mattn/go-mastodon"
)

func Mastodon(ctx context.Context, server, clientID, clientSecret, accessToken string) error {
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
	})

	if err := c.AuthenticateToken(ctx, accessToken, "urn:ietf:wg:oauth:2.0:oob"); err != nil {
		return err
	}

	timeline, err := c.GetTimelineHome(context.Background(), nil)
	if err != nil {
		return err
	}
	for i := len(timeline) - 1; i >= 0; i-- {
		fmt.Println(timeline[i])
	}

	return nil
}
