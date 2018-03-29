package runtime

import (
	"context"
	"testing"
)

func TestNilAsBool(t *testing.T) {
	ctx := context.Background()
	res := Nil.Bool(ctx)
	if res != false {
		t.Errorf("Nil as bool : expected %v, got %v", false, res)
	}
}

func TestNilAsString(t *testing.T) {
	ctx := context.Background()
	res := Nil.String(ctx)
	if res != NilString {
		t.Errorf("Nil as string : expected %s, got %s", NilString, res)
	}
}
