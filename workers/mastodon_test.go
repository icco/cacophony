package workers

import (
	"context"
	"testing"
)

func TestMastodon(t *testing.T) {
	if err := Mastodon(context.Background(), "", "", "", "", ""); err != nil {
		t.Errorf("Mastdon() was not nil: %+v", err)
	}
}
