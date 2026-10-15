// Package swaggerui serves a vendored copy of Swagger UI.
//
// Swagger UI (https://github.com/swagger-api/swagger-ui) is distributed under the
// Apache License 2.0, see dist/LICENSE and dist/NOTICE.
//
// The package depends on the standard library only, so importing it from generated
// code does not add any third-party dependencies to the resulting binary.
package swaggerui

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist
var dist embed.FS

//go:embed index.html.tmpl
var indexSource string

var (
	indexTemplate = template.Must(template.New("index").Parse(indexSource))
	assets        = func() map[string]struct{} {
		entries, err := fs.ReadDir(dist, "dist")
		if err != nil {
			panic(err)
		}
		m := make(map[string]struct{}, len(entries))
		for _, e := range entries {
			if !e.IsDir() {
				m[e.Name()] = struct{}{}
			}
		}
		return m
	}()
)

// Assets returns the embedded swagger-ui-dist files.
func Assets() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// Config configures Handler.
type Config struct {
	// SpecURL is the URL of the OpenAPI document to render.
	//
	// A relative URL is resolved by the browser against the page URL,
	// so "openapi.json" works wherever the handler is mounted.
	SpecURL string
	// Title is the HTML page title. Defaults to "Swagger UI".
	Title string
	// AssetsURL is the base URL (with a trailing slash) to load Swagger UI
	// static files from, for example a CDN like "https://unpkg.com/swagger-ui-dist@5/".
	//
	// If empty, the files embedded into this package are served by Handler.
	AssetsURL string
}

// Handler serves Swagger UI: the index page at "/" (or "") and, unless Config.AssetsURL
// is set, static files at "/<name>".
//
// Paths are relative to the mount point: use http.StripPrefix to mount the handler
// under a sub-path.
type Handler struct {
	cfg   Config
	index []byte
}

// New creates a new Handler.
func New(cfg Config) *Handler {
	if cfg.Title == "" {
		cfg.Title = "Swagger UI"
	}
	if cfg.AssetsURL != "" && !strings.HasSuffix(cfg.AssetsURL, "/") {
		cfg.AssetsURL += "/"
	}

	oauth2Redirect := "oauth2-redirect.html"
	if cfg.AssetsURL != "" {
		oauth2Redirect = cfg.AssetsURL + oauth2Redirect
	}

	var buf bytes.Buffer
	if err := indexTemplate.Execute(&buf, struct {
		Config
		OAuth2RedirectURL string
	}{
		Config:            cfg,
		OAuth2RedirectURL: oauth2Redirect,
	}); err != nil {
		// Template is embedded and data is plain strings, so this is a programmer error.
		panic(err)
	}

	return &Handler{
		cfg:   cfg,
		index: buf.Bytes(),
	}
}

// Match reports whether ServeHTTP would serve given path.
//
// The path must be relative to the mount point, i.e. "/" for the index page.
func (h *Handler) Match(p string) bool {
	switch p {
	case "", "/":
		return true
	}
	if h.cfg.AssetsURL != "" {
		return false
	}
	name, ok := strings.CutPrefix(p, "/")
	if !ok {
		return false
	}
	_, ok = assets[name]
	return ok
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	p := r.URL.Path
	switch p {
	case "", "/":
		h.serveIndex(w, r)
		return
	}
	if !h.Match(p) {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimPrefix(p, "/")
	hdr := w.Header()
	hdr.Set("Cache-Control", "public, max-age=86400")
	hdr.Set("ETag", `"`+Version+"-"+name+`"`)
	// Name is validated against the list of embedded files by Match.
	//#nosec G703
	http.ServeFileFS(w, r, dist, "dist/"+name)
}

func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()
	hdr.Set("Content-Type", "text/html; charset=utf-8")
	hdr.Set("Cache-Control", "no-cache")
	hdr.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(h.index)
}
