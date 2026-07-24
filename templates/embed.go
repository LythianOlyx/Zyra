// Package templates embeds Zyra's 10 official starter project templates
// (see zyraStrategy/10-CLI-AND-PROJECT-TEMPLATES.md) directly into the
// `zyra` binary, so `zyra create` never needs network access or a
// checked-out copy of the framework source to scaffold a new project.
package templates

import "embed"

//go:embed all:blank all:saas-starter all:dashboard-admin all:landing-page all:ecommerce all:ai-chat all:blog-cms all:realtime-collab all:api-only all:portfolio
var FS embed.FS
