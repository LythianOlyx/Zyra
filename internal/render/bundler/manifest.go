package bundler

// Manifest resolves a page route to the script(s) the HTML shell must
// include. Because Splitting uses real ES module `import` statements
// between chunks, a browser only needs a single `<script type="module">`
// tag for the entry file itself — it will fetch any statically-imported
// shared chunks automatically. So in the common case ScriptsFor returns a
// single-element slice, but the signature stays a slice for flexibility.
//
// The zero value is not useful; a *Manifest is obtained from a
// successful Build call.
type Manifest struct {
	// scripts maps a route to its ordered, PublicPath-prefixed script
	// URLs.
	scripts map[string][]string
	// entryFiles maps a route to the absolute on-disk path of its
	// primary (entry) output file.
	entryFiles map[string]string
}

// newManifest returns an empty, ready-to-populate Manifest.
func newManifest() *Manifest {
	return &Manifest{
		scripts:    make(map[string][]string),
		entryFiles: make(map[string]string),
	}
}

// set records that route resolves to publicURL (its script URL, already
// PublicPath-prefixed) and absPath (its on-disk output file). It is only
// called while assembling a Manifest inside Build.
func (m *Manifest) set(route, publicURL, absPath string) {
	m.scripts[route] = []string{publicURL}
	m.entryFiles[route] = absPath
}

// ScriptsFor returns the ordered list of public script URLs (already
// prefixed with Config.PublicPath) to include as
// `<script type="module" src="...">` tags for the given route. It
// returns nil if route is unknown.
func (m *Manifest) ScriptsFor(route string) []string {
	if m == nil {
		return nil
	}
	scripts := m.scripts[route]
	if len(scripts) == 0 {
		return nil
	}
	out := make([]string, len(scripts))
	copy(out, scripts)
	return out
}

// EntryFile returns the on-disk path (under Config.OutDir) of the
// primary output file for route, and whether it was found.
func (m *Manifest) EntryFile(route string) (string, bool) {
	if m == nil {
		return "", false
	}
	path, ok := m.entryFiles[route]
	return path, ok
}

// Routes returns a defensive copy of the full route-to-scripts mapping, so
// callers (notably the `zyra` CLI's build/dev pipeline) can persist the
// manifest to disk — e.g. as "dist/client/manifest.json" — for a
// standalone generated application's own process to load at runtime
// without ever importing this internal package directly.
func (m *Manifest) Routes() map[string][]string {
	if m == nil {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(m.scripts))
	for route, scripts := range m.scripts {
		copied := make([]string, len(scripts))
		copy(copied, scripts)
		out[route] = copied
	}
	return out
}
