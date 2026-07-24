//go:build zyratemplate

// Package content holds this template's blog post data.
//
// Zyra does not have an MDX pipeline wired into its esbuild bundler yet
// (see README.md), so rather than inventing a fake MDX renderer, this
// keeps things honest: posts are a small in-Go slice of structured data
// (title, slug, excerpt, date, body), and pages/blog/*.tsx render that data
// as plain HTML/text. Swapping this for real .mdx file parsing is a
// natural drop-in extension once the framework's MDX pipeline ships —
// only this file (and the small loop in main.go that registers one static
// route per slug) would need to change.
package content

import "github.com/zyra-framework/zyra/pkg/zyra"

// Post is a single blog post.
type Post struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Date    string `json:"date"` // ISO 8601 (YYYY-MM-DD)
	Body    string `json:"body"`
}

// Posts is the full list of blog posts, ordered newest-first. Slugs are
// derived from each title with zyra.Slug.Make so they never drift out of
// sync with the human-readable title.
var Posts = buildPosts()

func buildPosts() []Post {
	posts := []Post{
		{
			Title: "Why we built Zyra as a single Go binary",
			Date:  "2026-01-12",
			Excerpt: "No Node.js, no CGO, no runtime dependency hell — just `./app` and you're live. " +
				"Here's the reasoning behind Zyra's zero-runtime-dependency bet.",
			Body: "Most fullstack frameworks quietly ask you to ship a Node.js runtime, native database " +
				"drivers, and a pile of transitive dependencies alongside your actual application code. " +
				"Zyra takes a different bet: compile everything — the HTTP server, the bundler, even a " +
				"pure-Go JavaScript engine for server-side rendering — into one static binary. Deploying " +
				"becomes `scp` and `systemctl restart`, not a multi-stage container build chasing native " +
				"module ABI mismatches.",
		},
		{
			Title: "Shipping a single binary to production",
			Date:  "2026-02-03",
			Excerpt: "A practical walkthrough of `zyra build` and what actually ends up embedded in the " +
				"resulting executable.",
			Body: "`zyra build` compiles your Go Actions, bundles your React pages with esbuild, compiles " +
				"Tailwind CSS with a standalone binary, and embeds all of it — client bundle, styles, SQL " +
				"migrations, even default email templates — into a single `//go:embed`-backed executable. " +
				"The result is a binary you can drop onto any Linux, macOS, or Windows host with " +
				"CGO_ENABLED=0 and nothing else installed, and it just runs.",
		},
		{
			Title: "Chasing a perfect Lighthouse score from day one",
			Date:  "2026-02-21",
			Excerpt: "This very landing page template is Zyra's own dogfooding example: SSG pages, " +
				"minimal JS, and automatic sitemap + meta tags out of the box.",
			Body: "Every page in this template renders with `renderMode = \"ssg\"`: HTML is computed once " +
				"(and cached) instead of on every request, so there is no database round-trip or template " +
				"render blocking the response. Combine that with automatic sitemap generation " +
				"(`seo.generateSitemap` in zyra.config.json) and per-page `meta()` exports for SEO tags, " +
				"and a fresh `zyra create --template landing-page` project starts with a strong Lighthouse " +
				"score instead of clawing one back after the fact.",
		},
	}

	for i := range posts {
		posts[i].Slug = zyra.Slug.Make(posts[i].Title)
	}
	return posts
}

// FindBySlug looks up a post by its slug.
func FindBySlug(slug string) (Post, bool) {
	for _, p := range Posts {
		if p.Slug == slug {
			return p, true
		}
	}
	return Post{}, false
}
