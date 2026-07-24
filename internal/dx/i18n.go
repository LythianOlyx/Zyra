package dx

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type localeCtxKey struct{}

// I18nManager handles multi-language translations and locale detection.
type I18nManager struct {
	mu           sync.RWMutex
	defaultLang  string
	translations map[string]map[string]string // locale -> key -> msg
}

func NewI18nManager() *I18nManager {
	return &I18nManager{
		defaultLang:  "en",
		translations: make(map[string]map[string]string),
	}
}

func (i *I18nManager) SetDefaultLocale(locale string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.defaultLang = locale
}

func (i *I18nManager) LoadTranslations(locale string, dict map[string]string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	loc := strings.ToLower(locale)
	if _, exists := i.translations[loc]; !exists {
		i.translations[loc] = make(map[string]string)
	}

	for k, v := range dict {
		i.translations[loc][k] = v
	}
}

func (i *I18nManager) DetectLocale(r *http.Request) string {
	if r == nil {
		return i.defaultLang
	}

	// 1. URL path prefix check e.g. /id/about or /en/contact
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 && len(parts[0]) == 2 {
		loc := strings.ToLower(parts[0])
		i.mu.RLock()
		_, exists := i.translations[loc]
		i.mu.RUnlock()
		if exists {
			return loc
		}
	}

	// 2. Cookie check
	if cookie, err := r.Cookie("_zyra_locale"); err == nil && cookie.Value != "" {
		loc := strings.ToLower(cookie.Value)
		i.mu.RLock()
		_, exists := i.translations[loc]
		i.mu.RUnlock()
		if exists {
			return loc
		}
	}

	// 3. Accept-Language header check
	accept := r.Header.Get("Accept-Language")
	if accept != "" {
		langs := strings.Split(accept, ",")
		for _, l := range langs {
			code := strings.TrimSpace(strings.Split(l, ";")[0])
			code = strings.ToLower(strings.Split(code, "-")[0])
			i.mu.RLock()
			_, exists := i.translations[code]
			i.mu.RUnlock()
			if exists {
				return code
			}
		}
	}

	return i.defaultLang
}

func (i *I18nManager) WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeCtxKey{}, locale)
}

func (i *I18nManager) LocaleFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(localeCtxKey{}).(string); ok && v != "" {
		return v
	}
	return i.defaultLang
}

func (i *I18nManager) Translate(ctx context.Context, locale, key string, args ...any) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	loc := locale
	if loc == "" {
		loc = i.LocaleFromContext(ctx)
	}

	dict, exists := i.translations[loc]
	if !exists {
		dict = i.translations[i.defaultLang]
	}

	tmpl, found := dict[key]
	if !found {
		// Fallback to default lang dict
		if defaultDict, hasDefault := i.translations[i.defaultLang]; hasDefault {
			tmpl, found = defaultDict[key]
		}
	}

	if !found {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(tmpl, args...)
	}
	return tmpl
}
