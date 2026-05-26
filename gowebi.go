package gowebi

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var cfg *Config

type Config struct {
	BundleDir string `json:"bundleDir"`
	// Needs to be set in production.
	//
	// example: //go:embed dist/*
	BundleFS fs.FS
	// Enables vm pooling for lower allocations and better performance.
	// Default mode creates a new vm per request.
	//
	// WARN: global JS state may persist between requests.
	UnsafePoolMode bool
	UnsafePoolSize int
	// Suppresses server-side console logs from JS.
	SuppressConsoleLogs bool
	AutoBrowserReload   bool
	IsDev               bool `json:"isDev"`
}

type GoWebi struct {
	Renderer Renderer

	cfg       *Config
	bundleMap map[string]*Bundle
}

func (gw *GoWebi) BundleDir() string {
	return gw.cfg.BundleDir
}

func New(c *Config) (*GoWebi, error) {
	// set global cfg
	cfg = c

	if err := initBundleFS(cfg); err != nil {
		return nil, err
	}

	sr, err := cfg.BundleFS.Open("metafile.server.json")
	if err != nil {
		return nil, err
	}
	defer sr.Close()

	cr, err := cfg.BundleFS.Open("metafile.client.json")
	if err != nil {
		return nil, err
	}
	defer cr.Close()

	bundles, err := bundlesFromMetafile(sr, cr, cfg)
	if err != nil {
		return nil, err
	}

	renderer, err := initRenderer(cfg, bundles)
	if err != nil {
		return nil, err
	}

	if cfg.IsDev && cfg.AutoBrowserReload {
		ws := newWSHandler(filepath.Join(cfg.BundleDir, "index.html"))
		go func() {
			if err := ws.start(); err != nil {
				log.Printf("failed to start reload server: %v", err)
			}
		}()
	}

	return &GoWebi{
		cfg:       cfg,
		Renderer:  renderer,
		bundleMap: bundles,
	}, nil
}

func initRenderer(cfg *Config, bundles map[string]*Bundle) (Renderer, error) {
	tmpl, err := template.ParseFS(cfg.BundleFS, "index.html")
	if err != nil {
		return nil, err
	}
	if cfg.IsDev {
		return newDevRenderer(cfg, tmpl), nil
	}

	// production
	//
	// pool mode
	if cfg.UnsafePoolMode {
		if cfg.UnsafePoolSize == 0 {
			cfg.UnsafePoolSize = runtime.GOMAXPROCS(0)
		}

		pool, err := newPool(cfg.UnsafePoolSize, bundles)
		if err != nil {
			return nil, err
		}

		return NewPooledRenderer(pool, bundles, tmpl, cfg.IsDev), nil
	}

	// normal mode
	return NewRenderer(bundles, tmpl, cfg.IsDev), nil
}

func initBundleFS(cfg *Config) error {
	if cfg.IsDev {
		cfg.BundleFS = os.DirFS(cfg.BundleDir)
		return nil
	}

	// production
	if cfg.BundleFS == nil {
		return fmt.Errorf("BundleFS is required in production")
	}

	var err error
	cfg.BundleFS, err = fs.Sub(cfg.BundleFS, filepath.Clean(cfg.BundleDir))

	return err
}

func MustHandleInternalCmd(cfg *Config) {
	fgs := flag.NewFlagSet("gowebi", flag.ContinueOnError)

	esbuildCfg := fgs.Bool("gowebi-esbuild-config", false, "")

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	_ = fgs.Parse(args)
	if *esbuildCfg {
		if err := json.NewEncoder(os.Stdout).Encode(GetESBuildConfig(cfg)); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}
}
