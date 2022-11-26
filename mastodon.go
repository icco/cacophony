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
	clientID := flags.String("client-id", "", "Mastodon client ID")
	clientSecret := flags.String("client-secret", "", "Mastodon client Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "MASTODON")

	c := mastodon.NewClient(&mastodon.Config{
		Server:       *server,
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
	})

	if err := c.Authenticate(ctx, "your-email", "your-password"); err != nil {
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
