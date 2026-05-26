package main

import (
	"embed"
	"fmt"
	"gowebi"
	"log"
	"net/http"
	"os"
	"time"
)

type Data struct {
	Msg string `json:"msg"`
}

//go:embed dist/*
var dist embed.FS

// run: export ENVIRONMENT=development && node esbuild.config.js && go run .
func main() {
	cfg := &gowebi.Config{
		BundleDir:         "./dist",
		BundleFS:          dist,
		AutoBrowserReload: true,
		IsDev:             os.Getenv("ENVIRONMENT") != "production",
	}

	// should be registered before
	gowebi.MustHandleInternalCmd(cfg)
	app, err := gowebi.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// expose only the client bundle dir
	http.Handle("/client/", gowebi.ServeBundle())
	http.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		err := app.Renderer.Render(r.Context(), w, http.StatusOK, gowebi.RenderOptions{
			Name: "web/pages/Dash.jsx",
		})
		log.Println("render error:", err, time.Since(start))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		start := time.Now()
		err := app.Renderer.Render(r.Context(), w, http.StatusOK, gowebi.RenderOptions{
			Name:  "web/pages/Home.jsx",
			Props: Data{Msg: "Hello"},
		})
		log.Println("render error:", err, time.Since(start))
	})

	fmt.Println("listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
