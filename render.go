package gosolid

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dop251/goja"
)

type Renderer interface {
	Render(w http.ResponseWriter, status int, options RenderOptions) error
}

type RenderOptions struct {
	Name string
	// whether to hydrate the dom with SSR'd data (JS won't work).
	// default is false.
	NoHydrate bool
	Props     any
}

type Render struct {
	bundleMap map[string]*Bundle
	tmpl      *template.Template
	debug     bool
}

func NewRenderer(bMap map[string]*Bundle, tmpl *template.Template, debug bool) *Render {
	return &Render{
		bundleMap: bMap,
		tmpl:      tmpl,
		debug:     debug,
	}
}

func (r *Render) Render(w http.ResponseWriter, status int, opts RenderOptions) error {
	b, ok := r.bundleMap[opts.Name]
	if !ok {
		return fmt.Errorf("page bundle not found")
	}

	runtime, err := runVM(b.Program)
	if err != nil {
		if r.debug {
			return writeDebugError(w, err)
		}
		return err
	}

	appHtml, err := getStaticHTML(runtime, opts.Props)
	if err != nil {
		if r.debug {
			return writeDebugError(w, err)
		}
		return err
	}

	metadata, err := getMetadata(runtime, opts.Props)
	if err != nil {
		if r.debug {
			return writeDebugError(w, err)
		}
		return err
	}

	raw, err := json.Marshal(opts.Props)
	if err != nil {
		return err
	}

	head := fmt.Sprintf(`
<script>
    globalThis._$HY = {};
    window.__DATA__ = %s
</script>
`, string(raw))

	w.Header().Add("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(status)
	return r.tmpl.Execute(w, map[string]any{
		"Meta":               metadata,
		"Head":               template.HTML(head),
		"App":                template.HTML(appHtml),
		"HydrationScriptSrc": b.ClientPath,
		"Hydrate":            !opts.NoHydrate,
		"Props":              opts.Props,
	})
}

func getStaticHTML(runtime *Runtime, data any) (string, error) {
	rr, err := runtime.Render(
		goja.Undefined(),
		runtime.VM.ToValue(data),
	)
	if err != nil {
		return "", err
	}

	html, ok := rr.Export().(string)
	if !ok {
		return "", fmt.Errorf("invalid render export output")
	}

	return html, nil
}

func getMetadata(runtime *Runtime, data any) (map[string]any, error) {
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

	metadata, ok := mr.Export().(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid meta export output")
	}

	return metadata, nil
}

func writeDebugError(w http.ResponseWriter, e error) error {
	w.Header().Add("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(e.Error()))
	return err
}
