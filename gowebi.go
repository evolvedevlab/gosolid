package gowebi

import (
	"html/template"
	"os"
	"path/filepath"
)

var cfg Config

type Config struct {
	BundleDir string
	// Enables vm pooling for lower allocations and better performance.
	// Default mode creates a new vm per request.
	//
	// WARN: global JS state may persist between requests.
	UnsafePoolMode bool
	UnsafePoolSize int
	// Suppresses server-side console logs from JS.
	SuppressConsoleLogs bool
	IsDev               bool
}

type GoWebi struct {
	Renderer Renderer

	cfg       Config
	bundleMap map[string]*Bundle
}

func (gw *GoWebi) BundleDir() string {
	return gw.cfg.BundleDir
}

func New(c Config) (*GoWebi, error) {
	// set global cfg
	cfg = c

	tmpl, err := template.ParseFiles(filepath.Join(cfg.BundleDir, "index.html"))
	if err != nil {
		return nil, err
	}

	metafile := filepath.Join(cfg.BundleDir, "metafile.json")

	f, err := os.Open(metafile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bundles, err := bundleFromMetafile(f, cfg.BundleDir)
	if err != nil {
		return nil, err
	}

	var renderer Renderer
	if cfg.UnsafePoolMode {
		if cfg.UnsafePoolSize == 0 {
			cfg.UnsafePoolSize = 10
		}

		pool, err := newPool(cfg.UnsafePoolSize, bundles)
		if err != nil {
			return nil, err
		}

		renderer = NewPooledRenderer(pool, bundles, tmpl, cfg.IsDev)
	} else {
		renderer = NewRenderer(bundles, tmpl, cfg.IsDev)
	}

	return &GoWebi{
		cfg:       cfg,
		Renderer:  renderer,
		bundleMap: bundles,
	}, nil
}
