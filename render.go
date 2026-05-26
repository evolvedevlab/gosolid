package gowebi

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type Render struct {
	bundleMap map[string]*Bundle
	tmpl      *template.Template
	isDev     bool
}

func NewRenderer(bMap map[string]*Bundle, tmpl *template.Template, isDev bool) *Render {
	return &Render{
		bundleMap: bMap,
		tmpl:      tmpl,
		isDev:     isDev,
	}
}

func (r *Render) Render(ctx context.Context, w http.ResponseWriter, status int, opts RenderOptions) error {
	b, ok := r.bundleMap[opts.Name]
	if !ok {
		return fmt.Errorf("page bundle not found")
	}

	runtime, err := runVM(b.Program)
	if err != nil {
		if r.isDev {
			return writeDebugError(w, err)
		}
		return err
	}

	appHtml, err := getStaticHTML(runtime, opts.Props)
	if err != nil {
		if r.isDev {
			return writeDebugError(w, err)
		}
		return err
	}

	metadata := opts.Meta
	if metadata == nil {
		metadata, err = getMetadata(runtime, opts.Props)
		if err != nil {
			if r.isDev {
				return writeDebugError(w, err)
			}
			return err
		}
	}

	raw, err := json.Marshal(opts.Props)
	if err != nil {
		return err
	}

	script := getHydrationScript(raw, r.isDev)

	w.Header().Add("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(status)
	return r.tmpl.Execute(w, map[string]any{
		"Meta":               metadata,
		"Script":             template.HTML(script),
		"App":                template.HTML(appHtml),
		"HydrationScriptSrc": b.ClientPath,
		"NoHydrate":          opts.NoHydrate,
		"IsDev":              r.isDev,
		"Props":              opts.Props,
	})
}
