//go:build zyratemplate

package actions

import (
	"context"
	"testing"
)

func TestGreet_DefaultsToWorld(t *testing.T) {
	out, err := Greet(context.Background(), GreetInput{})
	if err != nil {
		t.Fatalf("Greet returned unexpected error: %v", err)
	}
	if out.Message != "Hello, world! This response came from a Go Action." {
		t.Errorf("unexpected message: %q", out.Message)
	}
}

func TestGreet_UsesProvidedName(t *testing.T) {
	out, err := Greet(context.Background(), GreetInput{Name: "  Zyra  "})
	if err != nil {
		t.Fatalf("Greet returned unexpected error: %v", err)
	}
	if out.Message != "Hello, Zyra! This response came from a Go Action." {
		t.Errorf("unexpected message: %q", out.Message)
	}
}
