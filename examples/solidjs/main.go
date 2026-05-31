package main

import (
	"embed"
	"fmt"
	"gowebi"
	"log"
	"net/http"
)

type Data struct {
	Msg string `json:"msg"`
}

//go:embed dist/*
var dist embed.FS

// run: make watch
// prod: make run
func main() {
	app, err := gowebi.New(gowebi.WithBundleFS(dist))
	if err != nil {
		log.Fatal(err)
	}

	const listenAddr = ":3000"

	http.Handle("/client/", gowebi.ServeBundle())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		app.Renderer.Render(r.Context(), w, http.StatusOK, gowebi.RenderOptions{
			Name:  "web/pages/Home.jsx",
			Props: Data{Msg: "Hello"},
		})
	})

	fmt.Printf("listening at %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
