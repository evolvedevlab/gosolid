package gowebi

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type PoolRender struct {
	pool      *pool
	bundleMap map[string]*Bundle
	tmpl      *template.Template
	debug     bool
}

func NewPooledRenderer(pool *pool, bMap map[string]*Bundle, tmpl *template.Template, debug bool) *PoolRender {
	return &PoolRender{
		pool:      pool,
		bundleMap: bMap,
		tmpl:      tmpl,
		debug:     debug,
	}
}

func (r *PoolRender) Render(ctx context.Context, w http.ResponseWriter, status int, opts RenderOptions) error {
	pi, err := r.pool.get(ctx)
	if err != nil {
		return fmt.Errorf("request cancelled: %w", err)
	}
	// defer r.pool.put(pl) // better to do it before network Write

	b, ok := r.bundleMap[opts.Name]
	if !ok {
		return fmt.Errorf("page bundle not found")
	}

	runtime := pi.pages[opts.Name]

	appHtml, err := getStaticHTML(runtime, opts.Props)
	if err != nil {
		if r.debug {
			return writeDebugError(w, err)
		}
		return err
	}

	metadata := opts.Meta
	if metadata == nil {
		metadata, err = getMetadata(runtime, opts.Props)
		if err != nil {
			if r.debug {
				return writeDebugError(w, err)
			}
			return err
		}
	}

	// return back to the pool
	r.pool.put(pi)

	raw, err := json.Marshal(opts.Props)
	if err != nil {
		return err
	}

	script := `<script>
window._$HY = {};
window.__DATA__ = ` + string(raw) + `
</script>`

	w.Header().Add("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(status)
	return r.tmpl.Execute(w, map[string]any{
		"Meta":               metadata,
		"Script":             template.HTML(script),
		"App":                template.HTML(appHtml),
		"HydrationScriptSrc": b.ClientPath,
		"NoHydrate":          opts.NoHydrate,
		"Props":              opts.Props,
	})
}
