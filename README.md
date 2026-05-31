# GoWebi

GoWebi is a server side rendering library/plugin for Go that lets you use frontend frameworks while keeping Go as the only server runtime.

It renders components to HTML on the server and hydrates them on the client. No Node.js server is required in production.

## Features

- Framework agnostic (works with any JS framework that does not rely heavily on Node.js APIs)
- Single binary deployment
- Auto browser reload in development
- New VM per request or pooled VM rendering
- Built in developer utilities

## Supported Frameworks

- [SolidJS](https://www.solidjs.com)

## Requirements

- Go
- Node.js
- Preferred package manager

Node.js is only required for bundling (development).

## Getting Started

GoWebi requires a small amount of frontend setup and build configuration.

For now, the easiest way to get started is from the SolidJS example:

```bash
git clone https://github.com/evolvedevlab/gowebi && cd gowebi && rm -rf .git && find . -mindepth 1 -maxdepth 1 ! -name examples -exec rm -rf {} \; && mv examples/solidjs/* . && rm -rf examples
```

This will leave you with a ready-to-use SolidJS project. **Check Makefile for commands.***

## Installation

```bash
go get -u github.com/evolvedevlab/gowebi@latest
```

## Quick Start

```go
app, err := gowebi.New()
if err != nil {
	log.Fatal(err)
}
```

**To embed generated bundle:**

```go
//go:embed dist/*
var dist embed.FS

app, err := gowebi.New(gowebi.WithBundleFS(dist))
```

**For production:**

```sh
go build -o ./bin/main && ENVIRONMENT=production ./bin/main
```

`ENVIRONMENT=production` has to be set. The dev/prod mode can be customized with `gowebi.WithIsDev(false)`

## Rendering

```go
type Data struct {
	Msg string `json:"msg"`
}

err := renderer.Render(r.Context(), w, http.StatusOK, gowebi.RenderOptions{
	Name: "web/pages/Home.jsx",
	Props: map[string]any{
		"title": "Hello",
		"data": Data{Msg: "Hi!"},
	},
})
```

**Props can be a map or struct. When using structs, JSON tags are used for field names inside the component.**

## VM Pooling

By default, NewRenderer creates a new JS VM for every render.

VM pooling keeps a set of pre-warmed VMs alive and reuses them between requests. This significantly reduces allocations and improves rendering performance.

Enable it with:

```go
app, err := gowebi.New(
	gowebi.WithUnsafePoolMode(true),
)
```

You can also configure the pool size:

```go
app, err := gowebi.New(
	gowebi.WithUnsafePoolMode(true),
	gowebi.WithUnsafePoolSize(4),
)
```

WARN: **Do not store request specific data globally when using pooled VMs, as they're are reused between requests.**
