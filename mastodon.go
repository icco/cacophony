package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/coreos/pkg/flagutil"
	"github.com/mattn/go-mastodon"
)

func mastodonCronWorker(ctx context.Context) error {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	server := flags.String("server", "https://merveilles.town", "Mastodon server")
	clientID := flags.String("client-id", "", "Mastodon Client ID")
	clientSecret := flags.String("client-secret", "", "Mastodon Client Secret")
	userPassword := flags.String("user-password", "", "Mastodon User Password")
	userEmail := flags.String("user-email", "nat@natwelch.com", "Mastodon User Email")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "MASTODON")

	if *server == "" || *clientID == "" || *clientSecret == "" {
		return fmt.Errorf("server, client id and client secret required")
	}

	if *userPassword == "" || *userEmail == "" {
		return fmt.Errorf("user password and email cannot be empty string")
	}

	c := mastodon.NewClient(&mastodon.Config{
		Server:       *server,
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
	})

	if err := c.Authenticate(ctx, *userEmail, *userPassword); err != nil {
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
