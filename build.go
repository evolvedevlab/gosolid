package gowebi

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

type Bundle struct {
	Program    *goja.Program
	ServerPath string
	ClientPath string
}

type Metafile struct {
	Server Build `json:"server"`
	Client Build `json:"client"`
}

type Build struct {
	Outputs map[string]*Output `json:"outputs"`
}

type Output struct {
	EntryPoint string `json:"entryPoint"`
}

func bundleFromMetafile(r io.Reader, dir string) (map[string]*Bundle, error) {
	var mf Metafile
	if err := json.NewDecoder(r).Decode(&mf); err != nil {
		return nil, err
	}

	bundles := make(map[string]*Bundle)

	// hydration script
	// entrypoint to outPath
	clientOut := make(map[string]string, len(mf.Client.Outputs))
	for outPath, o := range mf.Client.Outputs {
		// exclude chunks
		if !strings.Contains(outPath, "/chunks/") && len(o.EntryPoint) > 0 {
			clientOut[o.EntryPoint] = outPath
		}
	}

	for outPath, o := range mf.Server.Outputs {
		if filepath.Ext(outPath) != ".js" {
			continue
		}

		code, err := os.ReadFile(outPath)
		if err != nil {
			return nil, err
		}

		program, err := goja.Compile(outPath, string(code), false)
		if err != nil {
			return nil, err
		}

		clientPath, ok := clientOut[o.EntryPoint]
		if !ok {
			log.Printf("%s: hydration script not found", o.EntryPoint)
		}

		clientPath, err = filepath.Rel(dir, clientPath)
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
