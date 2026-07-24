package scaffold

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// binaryExtensions lists file extensions copied byte-for-byte, never run
// through the template engine (images/fonts can't contain "[[ ]]"
// placeholders, and treating them as text would corrupt them).
var binaryExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
	".webp": true, ".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
	".otf": true,
}

// copyEmbeddedDir walks every file under root in fsys and writes it to the
// equivalent path under destRoot, rendering "[[ ]]" placeholders (see
// TemplateData) in every non-binary file along the way.
func copyEmbeddedDir(fsys fs.FS, root, destRoot string, data TemplateData) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		target := destRoot
		if rel != "." {
			target = filepath.Join(destRoot, rel)
		}

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			content = stripTemplateBuildTag(content)
		}

		if !isBinaryFile(path) {
			rendered, err := renderTemplate(path, content, data)
			if err != nil {
				return err
			}
			content = rendered
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}

func isBinaryFile(path string) bool {
	return binaryExtensions[strings.ToLower(filepath.Ext(path))]
}

// templateBuildTagLine marks every .go file in templates/** as excluded
// from the Zyra framework repository's own `go build ./...`/`go vet
// ./...`/`go test ./...`: template sources contain unrendered "[[ ]]"
// placeholders (e.g. in import paths) that are not valid Go on their own,
// and (unlike a nested go.mod) a build tag does not stop //go:embed from
// including the file. It must never reach a generated project, so
// stripTemplateBuildTag removes it (plus the blank line that must follow
// a build-constraint comment per Go's own syntax rules) while copying.
const templateBuildTagLine = "//go:build zyratemplate"

func stripTemplateBuildTag(content []byte) []byte {
	const prefix = templateBuildTagLine + "\n\n"
	if bytes.HasPrefix(content, []byte(prefix)) {
		return content[len(prefix):]
	}
	return content
}

func renderTemplate(name string, content []byte, data TemplateData) ([]byte, error) {
	tmpl, err := template.New(name).Delims("[[", "]]").Parse(string(content))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
