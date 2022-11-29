package workers

import (
	"context"
	"testing"
)

func TestTwitter(t *testing.T) {
	if err := Twitter(context.Background(), "", "", "", ""); err != nil {
		t.Errorf("Mastdon() was not nil: %+v", err)
	}
}
