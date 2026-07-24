package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	clientruntime "github.com/LythianOlyx/Zyra/runtime/client"
	"github.com/LythianOlyx/Zyra/templates"
)

// DefaultFrameworkVersion is the `github.com/LythianOlyx/Zyra` module
// version requirement written into a generated project's go.mod when
// Options.FrameworkVersion is left empty.
const DefaultFrameworkVersion = "v1.0.0-alpha.1"

// Options configures a single `zyra create` project generation run.
type Options struct {
	// AppName is the human-readable project name (used in zyra.config.json,
	// README.md, page titles, etc).
	AppName string
	// Template is one of the 10 official starter template IDs (see
	// Templates()).
	Template string
	// Database selects the database engine wired into zyra.config.json /
	// .env.example.
	Database Database
	// EnableAuth toggles whether the generated zyra.config.json configures
	// an auth strategy ("session") or leaves auth unconfigured.
	EnableAuth bool
	// EnableObservability toggles the ZYRA_OBSERVABILITY env var written to
	// .env.example (OpenTelemetry tracing + Prometheus /metrics).
	EnableObservability bool
	// InitGit runs `git init` (+ an initial commit) in Dest after
	// generation, when true.
	InitGit bool

	// Dest is the destination directory the project is written to.
	// Defaults to "./<AppName>" (sanitized) when empty.
	Dest string
	// ModulePath is the Go module path written to go.mod. Defaults to a
	// sanitized version of AppName when empty.
	ModulePath string
	// FrameworkVersion is the `github.com/LythianOlyx/Zyra` version
	// requirement written to go.mod. Defaults to DefaultFrameworkVersion.
	FrameworkVersion string
	// FrameworkReplacePath, when non-empty, adds a
	// `replace github.com/LythianOlyx/Zyra => <path>` directive to the
	// generated go.mod. This exists purely for local framework development
	// and this repository's own CI template validation (see
	// scaffold_test.go) — end users creating real projects never set this.
	FrameworkReplacePath string
}

// Result describes the outcome of a successful Generate call.
type Result struct {
	Dest       string
	ModulePath string
	Template   Template
}

func (o Options) withDefaults() (Options, error) {
	if o.AppName == "" {
		return o, fmt.Errorf("scaffold: AppName must not be empty")
	}
	if o.Database == "" {
		o.Database = DatabaseSkip
	}
	if o.ModulePath == "" {
		o.ModulePath = sanitizeModulePath(o.AppName)
	}
	if o.Dest == "" {
		o.Dest = o.AppName
	}
	if o.FrameworkVersion == "" {
		o.FrameworkVersion = DefaultFrameworkVersion
	}
	return o, nil
}

// Generate scaffolds a new Zyra project from opts.Template into opts.Dest.
func Generate(opts Options) (*Result, error) {
	opts, err := opts.withDefaults()
	if err != nil {
		return nil, err
	}

	tmpl, ok := Find(opts.Template)
	if !ok {
		return nil, fmt.Errorf("scaffold: unknown template %q (run `zyra create` without --template to see the list)", opts.Template)
	}

	if entries, err := os.ReadDir(opts.Dest); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("scaffold: destination %q already exists and is not empty", opts.Dest)
	}

	if err := os.MkdirAll(opts.Dest, 0o755); err != nil {
		return nil, fmt.Errorf("scaffold: failed to create destination directory: %w", err)
	}

	data := buildTemplateData(opts)

	if err := copyEmbeddedDir(templates.FS, tmpl.ID, opts.Dest, data); err != nil {
		return nil, fmt.Errorf("scaffold: failed to copy template %q: %w", tmpl.ID, err)
	}

	runtimeDest := filepath.Join(opts.Dest, "runtime", "client")
	if err := copyEmbeddedDir(clientruntime.FS, ".", runtimeDest, data); err != nil {
		return nil, fmt.Errorf("scaffold: failed to copy runtime/client: %w", err)
	}

	if err := writeGoMod(opts); err != nil {
		return nil, err
	}

	if opts.InitGit {
		if err := gitInit(opts.Dest); err != nil {
			return nil, fmt.Errorf("scaffold: git init failed: %w", err)
		}
	}

	return &Result{Dest: opts.Dest, ModulePath: opts.ModulePath, Template: tmpl}, nil
}

func writeGoMod(opts Options) error {
	content := fmt.Sprintf("module %s\n\ngo 1.23\n\nrequire github.com/LythianOlyx/Zyra %s\n",
		opts.ModulePath, opts.FrameworkVersion)
	if opts.FrameworkReplacePath != "" {
		content += fmt.Sprintf("\nreplace github.com/LythianOlyx/Zyra => %s\n", opts.FrameworkReplacePath)
	}
	return os.WriteFile(filepath.Join(opts.Dest, "go.mod"), []byte(content), 0o644)
}

func gitInit(dest string) error {
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dest
	if err := cmd.Run(); err != nil {
		return err
	}
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = dest
	if err := addCmd.Run(); err != nil {
		return err
	}
	commitCmd := exec.Command("git", "-c", "user.email=zyra@localhost", "-c", "user.name=Zyra CLI",
		"commit", "-q", "-m", "chore: initial commit from `zyra create`")
	commitCmd.Dir = dest
	return commitCmd.Run()
}
