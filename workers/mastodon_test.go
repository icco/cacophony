package main

import (
	"context"
	"testing"
)

func TestMastodon(t *testing.T) {
	if err := mastodonCronWorker(context.Background()); err != nil {
		t.Errorf("mastdonCronWorker() was not nil: %+v", err)
	}
}
