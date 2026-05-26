package gowebi

import (
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/dop251/goja"
)

type Bundle struct {
	Program    *goja.Program
	ServerPath string
	ClientPath string
}

type Metafile struct {
	Outputs map[string]*Output `json:"outputs"`
}

type Output struct {
	EntryPoint string `json:"entryPoint"`
}

func bundlesFromMetafile(sr, cr io.Reader, cfg *Config) (map[string]*Bundle, error) {
	sf, cf, err := loadMetafiles(sr, cr)
	if err != nil {
		return nil, err
	}

	var (
		bundles   = make(map[string]*Bundle, 1)
		clientOut = buildClientOutputMap(cf)
	)
	// As goja cannot use fs.FS to load sourcemaps
	// In prod, we write the server sourcemaps to filesystem from fs.FS
	gojaJSPath, err := prepareGojaSourcemaps(cfg.IsDev)
	if err != nil {
		return nil, err
	}

	for outPath, o := range sf.Outputs {
		if filepath.Ext(outPath) != ".js" {
			continue
		}
		outPathRel, err := filepath.Rel(cfg.BundleDir, outPath)
		if err != nil {
			return nil, err
		}
		// In dev, goja should resolve sourcemaps by itself from bundleDir
		if cfg.IsDev {
			gojaJSPath = outPath
		}

		// read server bundle js code
		code, err := fs.ReadFile(cfg.BundleFS, filepath.ToSlash(outPathRel))
		if err != nil {
			return nil, err
		}

		program, err := goja.Compile(gojaJSPath, string(code), true)
		if err != nil {
			return nil, err
		}

		clientPath, ok := clientOut[o.EntryPoint]
		if !ok {
			log.Printf("%s: hydration script not found", o.EntryPoint)
		}

		clientPath, err = filepath.Rel(cfg.BundleDir, clientPath)
		if err != nil {
			return nil, err
		}

		bundles[o.EntryPoint] = &Bundle{
			ServerPath: outPath,
			ClientPath: "/" + filepath.ToSlash(clientPath),
			Program:    program,
		}
	}

	return bundles, nil
}

// if given entrypoint is not found, it will return nil, nil
func devBundleFromEntrypoint(fsys fs.FS, entrypoint string) (*Bundle, error) {
	sr, err := fsys.Open("metafile.server.json")
	if err != nil {
		return nil, err
	}
	defer sr.Close()

	cr, err := fsys.Open("metafile.client.json")
	if err != nil {
		return nil, err
	}
	defer cr.Close()

	sf, cf, err := loadMetafiles(sr, cr)
	if err != nil {
		return nil, err
	}

	clientOut := buildClientOutputMap(cf)

	for outPath, o := range sf.Outputs {
		if o.EntryPoint != entrypoint {
			continue
		}
		if filepath.Ext(outPath) != ".js" {
			continue
		}

		outPathRel, err := filepath.Rel(cfg.BundleDir, outPath)
		if err != nil {
			return nil, err
		}

		// read server bundle js code
		code, err := fs.ReadFile(cfg.BundleFS, filepath.ToSlash(outPathRel))
		if err != nil {
			return nil, err
		}

		program, err := goja.Compile(outPath, string(code), true)
		if err != nil {
			return nil, err
		}

		clientPath, ok := clientOut[o.EntryPoint]
		if !ok {
			log.Printf("%s: hydration script not found", o.EntryPoint)
		}

		clientPath, err = filepath.Rel(cfg.BundleDir, clientPath)
		if err != nil {
			return nil, err
		}

		return &Bundle{
			ServerPath: outPath,
			ClientPath: "/" + filepath.ToSlash(clientPath),
			Program:    program,
		}, nil
	}

	return nil, nil
}

func extractAndWriteSourcemaps(fsys fs.FS, outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".map" {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		target := filepath.Join(outDir, filepath.Base(path))
		return os.WriteFile(target, data, 0644)
	})
}

func prepareGojaSourcemaps(isDev bool) (string, error) {
	if isDev {
		return "", nil
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(
		cacheDir,
		"gowebi",
		getBuildHash(),
		"sourcemaps",
	)

	dir += string(os.PathSeparator)
	if err := extractAndWriteSourcemaps(cfg.BundleFS, dir); err != nil {
		return "", err
	}

	return dir, nil
}

func loadMetafiles(sr, cr io.Reader) (sf Metafile, cf Metafile, err error) {
	if err := json.NewDecoder(sr).Decode(&sf); err != nil {
		return sf, cf, err
	}
	if err := json.NewDecoder(cr).Decode(&cf); err != nil {
		return sf, cf, err
	}

	return sf, cf, nil
}

// hydration script
// entrypoint to outPath
// takes client metafile
func buildClientOutputMap(cf Metafile) map[string]string {
	out := make(map[string]string)
	for outPath, o := range cf.Outputs {
		// exclude chunks
		if !strings.Contains(outPath, "/chunks/") && len(o.EntryPoint) > 0 {
			out[o.EntryPoint] = outPath
		}
	}

	return out
}

func getBuildHash() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			if len(s.Value) > 12 {
				return s.Value[:12]
			}
			return s.Value
		}
	}

	return "dev"
}
