package scaffold

import "strings"

// TemplateData is the value passed to every embedded template file when
// rendering "[[ ]]"-delimited placeholders (see render.go). Custom
// delimiters are used instead of Go's default "{{ }}" because React/JSX
// source commonly contains literal "{{" (e.g. `style={{ color: 'red' }}`),
// which would otherwise collide with text/template's own syntax.
type TemplateData struct {
	AppName             string
	ModulePath          string
	DatabaseDriver      string
	DatabaseURL         string
	AuthStrategy        string
	EnableObservability string // "true" | "false", for .env.example
	FrameworkVersion    string
}

func buildTemplateData(opts Options) TemplateData {
	driver, url := opts.Database.driverAndURL()

	authStrategy := ""
	if opts.EnableAuth {
		authStrategy = "session"
	}

	observability := "false"
	if opts.EnableObservability {
		observability = "true"
	}

	return TemplateData{
		AppName:             opts.AppName,
		ModulePath:          opts.ModulePath,
		DatabaseDriver:      driver,
		DatabaseURL:         url,
		AuthStrategy:        authStrategy,
		EnableObservability: observability,
		FrameworkVersion:    opts.FrameworkVersion,
	}
}

// sanitizeModulePath derives a reasonable default Go module path from a
// human-entered app name (e.g. "My Cool App" -> "my-cool-app").
func sanitizeModulePath(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var sb strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			sb.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == '.' || r == '/':
			sb.WriteRune(r)
			lastDash = false
		case r == ' ':
			if !lastDash {
				sb.WriteRune('-')
				lastDash = true
			}
		}
	}
	result := strings.Trim(sb.String(), "-")
	if result == "" {
		result = "zyra-app"
	}
	return result
}
