package gowebi

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dop251/goja"
)

type Renderer interface {
	Render(ctx context.Context, w http.ResponseWriter, status int, options RenderOptions) error
}

type RenderOptions struct {
	Name string
	// Whether to hydrate the dom with SSR'd data.
	// Default is false (hydration is on).
	NoHydrate bool
	Meta      map[string]string
	Props     any
}

type pool struct {
	ch chan *poolItem
}

type poolItem struct {
	vm    *goja.Runtime
	pages map[string]*Runtime
}

func newPool(size int, bundles map[string]*Bundle) (*pool, error) {
	pool := &pool{
		ch: make(chan *poolItem, size),
	}

	for i := 0; i < size; i++ {
		vm := newVM(cfg.SuppressConsoleLogs)
		pages := make(map[string]*Runtime, len(bundles))

		for outPath, b := range bundles {
			runtime, err := runWithVM(vm, b.Program)
			if err != nil {
				return nil, err
			}

			pages[outPath] = runtime
			// after pre-warm we dont use program in runtime
			if i >= size {
				b.Program = nil
			}
		}

		pool.ch <- &poolItem{
			vm:    vm,
			pages: pages,
		}
	}

	return pool, nil
}

func (p *pool) get(ctx context.Context) (*poolItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case item := <-p.ch:
		return item, nil
	}
}

// put will never block
//
// WARN: pool has fixed no. of items, do not put new pool item in runtime.
func (p *pool) put(item *poolItem) {
	p.ch <- item
}

func getStaticHTML(runtime *Runtime, data any) (string, error) {
	rr, err := runtime.Render(
		goja.Undefined(),
		runtime.VM.ToValue(data),
	)
	if err != nil {
		return "", err
	}

	return rr.String(), nil
}

func getMetadata(runtime *Runtime, data any) (map[string]string, error) {
	// meta fn was not provided
	if runtime.Meta == nil {
		return nil, nil
	}

	mr, err := runtime.Meta(
		goja.Undefined(),
		runtime.VM.ToValue(data),
	)
	if err != nil {
		return nil, err
	}

	var metadata map[string]string
	if err := runtime.VM.ExportTo(mr, &metadata); err != nil {
		return nil, fmt.Errorf("metadata export failed: %w", err)
	}

	return metadata, nil
}

func getHydrationScript(raw []byte, isDev bool) string {
	script := `<script>
window._$HY = {};
window.__DATA__ = ` + string(raw) + `
</script>`

	if isDev {
		script += "\n" + browserReloadScript
	}
	return script
}

type devRender struct {
	cfg  *Config
	tmpl *template.Template
}

func newDevRenderer(cfg *Config, tmpl *template.Template) *devRender {
	return &devRender{
		cfg:  cfg,
		tmpl: tmpl,
	}
}

func (dr *devRender) Render(ctx context.Context, w http.ResponseWriter, status int, opts RenderOptions) error {
	b, err := devBundleFromEntrypoint(dr.cfg.BundleFS, opts.Name)
	if err != nil {
		return err
	}

	bundles := map[string]*Bundle{
		opts.Name: b,
	}

	r := NewRenderer(bundles, dr.tmpl, dr.cfg.IsDev)
	return r.Render(ctx, w, status, opts)
}

func writeDebugError(w http.ResponseWriter, e error) error {
	w.Header().Add("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(e.Error()))
	return err
}
