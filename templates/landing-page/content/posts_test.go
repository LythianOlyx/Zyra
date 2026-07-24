//go:build zyratemplate

package content

import "testing"

func TestPosts_AreNonEmptyWithUniqueSlugs(t *testing.T) {
	if len(Posts) == 0 {
		t.Fatal("expected at least one blog post")
	}

	seen := map[string]bool{}
	for _, p := range Posts {
		if p.Slug == "" {
			t.Errorf("post %q has an empty slug", p.Title)
		}
		if seen[p.Slug] {
			t.Errorf("duplicate slug %q", p.Slug)
		}
		seen[p.Slug] = true

		if p.Title == "" || p.Excerpt == "" || p.Date == "" || p.Body == "" {
			t.Errorf("post %q is missing a required field: %+v", p.Slug, p)
		}
	}
}

func TestFindBySlug_ReturnsKnownPost(t *testing.T) {
	want := Posts[0]
	got, ok := FindBySlug(want.Slug)
	if !ok {
		t.Fatalf("expected to find post with slug %q", want.Slug)
	}
	if got.Title != want.Title {
		t.Errorf("expected title %q, got %q", want.Title, got.Title)
	}
}

func TestFindBySlug_ReturnsFalseForUnknownSlug(t *testing.T) {
	if _, ok := FindBySlug("does-not-exist"); ok {
		t.Fatal("expected FindBySlug to return false for an unknown slug")
	}
}
