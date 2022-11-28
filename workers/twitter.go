package workers

import (
	"context"
	"fmt"
	"os"

	//lint:ignore SA1019 deprecated and I don't care
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/icco/cacophony/models"
	"github.com/icco/cron/shared"
	"github.com/icco/cron/tweets"
	"github.com/icco/gutil/logging"
)

func Twitter(ctx context.Context, consumerKey, consumerSecret, accessToken, accessSecret string) error {
	log := logging.FromContext(ctx)
	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		return fmt.Errorf("consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(ctx, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return err
	}
	log.Debugw("found user", "user", user.ScreenName)

	// Home Timeline
	homeTimelineParams := &twitter.HomeTimelineParams{
		Count:           200,
		IncludeEntities: twitter.Bool(true),
	}
	homeTweets, resp, err := client.Timelines.HomeTimeline(homeTimelineParams)
	if resp.Header.Get("X-Rate-Limit-Remaining") == "0" {
		return fmt.Errorf("out of Rate Limit")
	}

	if err != nil {
		return err
	}

	c := tweets.Twitter{
		Config: shared.Config{
			Log: log,
		},
		GraphQLToken: os.Getenv("GQL_TOKEN"),
	}

	for _, t := range homeTweets {
		// Save tweet to DB via graphql
		if err := c.UploadTweet(ctx, t); err != nil {
			return fmt.Errorf("problem uploading tweet: %w", err)
		}

		for _, u := range t.Entities.Urls {
			if err := models.SaveTwitterURL(ctx, u.ExpandedURL, t.IDStr); err != nil {
				return fmt.Errorf("error saving url: %w", err)
			}
		}
	}

	return nil
}
