// Package scaffold implements `zyra create`'s project generation: listing
// the 10 official starter templates (embedded into the binary via
// github.com/zyra-framework/zyra/templates), rendering their placeholder
// tokens, copying Zyra's frontend runtime alongside them, and writing a
// standalone go.mod so the generated project is an independent Go module
// that only ever imports the public pkg/zyra and pkg/zyra/app API — never
// any internal/ package — per 02-ARCHITECTURE.md's package boundary rules.
package scaffold

// Template describes one official starter template shown by `zyra create`.
type Template struct {
	ID          string
	Title       string
	Description string
}

// Templates returns the 10 official starter templates, in the same order
// as zyraStrategy/10-CLI-AND-PROJECT-TEMPLATES.md.
func Templates() []Template {
	return []Template{
		{ID: "blank", Title: "blank", Description: "Empty & minimal — learn Zyra's core concepts from scratch"},
		{ID: "saas-starter", Title: "saas-starter", Description: "Auth + billing + dashboard (the most complete starter)"},
		{ID: "dashboard-admin", Title: "dashboard-admin", Description: "Admin panel: data tables, CRUD forms, granular RBAC"},
		{ID: "landing-page", Title: "landing-page", Description: "Marketing site: SEO-first, pricing, blog, all SSG"},
		{ID: "ecommerce", Title: "ecommerce", Description: "Storefront + cart + Stripe checkout + admin catalog"},
		{ID: "ai-chat", Title: "ai-chat", Description: "Streaming LLM chat app, ChatGPT-style, SSE-based"},
		{ID: "blog-cms", Title: "blog-cms", Description: "Blog/content site with MDX, RSS, syntax highlighting"},
		{ID: "realtime-collab", Title: "realtime-collab", Description: "Real-time kanban board: presence + optimistic updates"},
		{ID: "api-only", Title: "api-only", Description: "Headless backend: Go Actions + OpenAPI, no React pages"},
		{ID: "portfolio", Title: "portfolio", Description: "Simple personal website with a contact form + email"},
	}
}

// Find returns the Template with the given id, if it is one of the 10
// official starter templates.
func Find(id string) (Template, bool) {
	for _, t := range Templates() {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}
