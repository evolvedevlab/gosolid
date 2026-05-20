package gowebi

import (
	"html/template"
	"os"
	"path/filepath"
)

type Config struct {
	BundleDir string
	IsDev     bool
}

type GoWebi struct {
	Renderer Renderer

	cfg       Config
	bundleMap map[string]*Bundle
}

func (gw *GoWebi) BundleDir() string {
	return gw.cfg.BundleDir
}

func New(cfg Config) (*GoWebi, error) {
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

	return &GoWebi{
		cfg:       cfg,
		Renderer:  NewRenderer(bundles, tmpl, cfg.IsDev),
		bundleMap: bundles,
	}, nil
}
