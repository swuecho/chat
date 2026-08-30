package requestctx

import (
	"context"
	"testing"
)

func TestPrincipalRoundTrip(t *testing.T) {
	want := Principal{UserID: 42, Role: "admin"}
	ctx := WithPrincipal(context.Background(), want)
	got, err := PrincipalFrom(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("PrincipalFrom() = %#v, want %#v", got, want)
	}
}

func TestPrincipalMissing(t *testing.T) {
	if _, err := PrincipalFrom(context.Background()); err == nil {
		t.Fatal("PrincipalFrom() accepted a missing principal")
	}
}
