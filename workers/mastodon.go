package main

import (
	"context"
	"fmt"

	"github.com/mattn/go-mastodon"
)

func Mastodon(ctx context.Context, server, clientID, clientSecret, userEmail, userPassword string) error {
	if server == "" || clientID == "" || clientSecret == "" {
		return fmt.Errorf("server, client id and client secret required")
	}

	if userPassword == "" || userEmail == "" {
		return fmt.Errorf("user password and email cannot be empty string")
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       server,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})

	if err := c.Authenticate(ctx, userEmail, userPassword); err != nil {
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
