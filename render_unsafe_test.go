package gowebi

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dop251/goja"
)

func BenchmarkUnsafeRender(b *testing.B) {
	comp, err := os.ReadFile("testdata/simplecomp.js")
	if err != nil {
		b.Fatal(err)
	}

	program, err := goja.Compile("", string(comp), false)
	if err != nil {
		b.Fatal(err)
	}

	bm := map[string]*Bundle{
		"test": {Program: program},
	}

	pool, err := newPool(1, bm)
	if err != nil {
		b.Fatal(err)
	}

	r := NewPooledRenderer(
		pool,
		bm,
		template.Must(template.New("").Parse(testTmpl)),
		true,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()

		err := r.Render(
			b.Context(),
			rr,
			http.StatusOK,
			RenderOptions{
				Name:      "test",
				NoHydrate: true,
			},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
