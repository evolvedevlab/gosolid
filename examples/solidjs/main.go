package main

import (
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

// run: export ENVIRONMENT=development && node esbuild.config.js && go run .
func main() {
	app, err := gowebi.New(gowebi.Config{
		BundleDir: "./dist",
		IsDev:     os.Getenv("ENVIRONMENT") != "production",
	})
	if err != nil {
		log.Fatal(err)
	}

	// expose only the client bundle dir
	http.Handle("/client/", gowebi.ServeBundle(app.BundleDir()))
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
			Props: Data{Msg: "gello"},
		})
		log.Println("render error:", err, time.Since(start))
	})

	fmt.Println("listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
