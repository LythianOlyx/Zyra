//go:build zyratemplate

package content

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Post represents a blog article.
type Post struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Excerpt     string    `json:"excerpt"`
	BodyHTML    string    `json:"bodyHtml"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"publishedAt"`
}

var SeedPosts = []Post{
	{
		Slug:        "getting-started-with-zyra",
		Title:       "Getting Started with Zyra Framework",
		Excerpt:     "Learn how Zyra eliminates Node.js runtime dependencies with single Go binary compilation.",
		Author:      "Zyra Team",
		Tags:        []string{"Go", "React", "Fullstack"},
		PublishedAt: time.Now().Add(-168 * time.Hour),
		BodyHTML: `
			<p>Zyra is a zero-runtime-dependency fullstack Go 1.23+ and React web framework.</p>
			<h2>Code Example</h2>
			<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Hello, Zyra!")
}</code></pre>
		`,
	},
	{
		Slug:        "zero-cgo-embedded-ssr",
		Title:       "Zero CGO Embedded JS Server-Side Rendering",
		Excerpt:     "How Zyra embeds the Goja JS engine into a thread-safe pool for ultra-fast SSG and SSR.",
		Author:      "Systems Lead",
		Tags:        []string{"Architecture", "SSR", "Goja"},
		PublishedAt: time.Now().Add(-72 * time.Hour),
		BodyHTML: `
			<p>By embedding the dopamine goja engine, Zyra renders React components server-side without needing Node.js or Bun installed on the server.</p>
		`,
	},
}

// RSS Feed structures
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// GenerateRSS XML marshals the posts into an RSS 2.0 feed string.
func GenerateRSS(baseURL string) (string, error) {
	var items []RSSItem
	for _, p := range SeedPosts {
		items = append(items, RSSItem{
			Title:       p.Title,
			Link:        fmt.Sprintf("%s/blog/%s", baseURL, p.Slug),
			Description: p.Excerpt,
			PubDate:     p.PublishedAt.Format(time.RFC1123Z),
			GUID:        fmt.Sprintf("%s/blog/%s", baseURL, p.Slug),
		})
	}

	feed := RSSFeed{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       "Zyra Blog CMS",
			Link:        baseURL,
			Description: "Latest news and articles from Zyra Blog CMS",
			Items:       items,
		},
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}
