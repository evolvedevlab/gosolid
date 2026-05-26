package gowebi

import (
	"maps"
	"os"
	"strconv"
)

type ESBuildConfig struct {
	// shared cfg
	// should be same for both client & server builds
	EntryPoints []string          `json:"entryPoints,omitempty"`
	Bundle      bool              `json:"bundle,omitempty"`
	Metafile    bool              `json:"metafile,omitempty"`
	Sourcemap   bool              `json:"sourcemap,omitempty"`
	Minify      bool              `json:"minify,omitempty"`
	TreeShaking bool              `json:"treeShaking,omitempty"`
	Conditions  []string          `json:"conditions,omitempty"`
	LogLevel    string            `json:"logLevel,omitempty"`
	Define      map[string]string `json:"define,omitempty"`

	// specific to build target
	Outdir     string   `json:"outdir,omitempty"`
	Format     string   `json:"format,omitempty"`
	Platform   string   `json:"platform,omitempty"`
	Target     []string `json:"target,omitempty"`
	Splitting  bool     `json:"splitting,omitempty"`
	EntryNames string   `json:"entryNames,omitempty"`
	ChunkNames string   `json:"chunkNames,omitempty"`
	AssetNames string   `json:"assetNames,omitempty"`
}

type ESBuildConfigData struct {
	Config Config        `json:"goConfig"`
	Server ESBuildConfig `json:"server"`
	Client ESBuildConfig `json:"client"`
}

func GetESBuildConfig(cfg *Config) ESBuildConfigData {
	isDev := cfg.IsDev
	environment := os.Getenv("ENVIRONMENT")

	shared := ESBuildConfig{
		EntryPoints: []string{"web/pages/*"},
		Bundle:      true,
		Metafile:    true,
		Sourcemap:   true,
		Minify:      !isDev,
		TreeShaking: true,
		Conditions:  []string{environment},
		LogLevel:    "info",
		Define: map[string]string{
			"process.env.ENVIRONMENT": strconv.Quote(environment),
		},
	}

	// server cfg
	server := shared

	server.Outdir = "dist/server"
	server.Format = "iife"
	server.Platform = "node"
	server.Target = []string{"node20"}
	server.EntryNames = "[name]"
	server.ChunkNames = "chunks/[name]"
	server.AssetNames = "assets/[name]"

	server.Define = maps.Clone(shared.Define)
	server.Define["__SERVER__"] = "true"

	// client cfg
	client := shared

	client.Outdir = "dist/client"
	client.Format = "esm"
	client.Platform = "browser"
	client.Target = []string{
		"chrome120",
		"firefox120",
		"safari17",
	}
	client.Splitting = true
	client.Sourcemap = isDev

	if isDev {
		client.EntryNames = "[name]"
		client.AssetNames = "assets/[name]"
	} else {
		client.EntryNames = "[name]-[hash]"
		client.AssetNames = "assets/[name]-[hash]"
	}

	client.ChunkNames = "chunks/[name]-[hash]"

	client.Define = maps.Clone(shared.Define)
	client.Define["__SERVER__"] = "false"

	return ESBuildConfigData{
		Config: *cfg,
		Server: server,
		Client: client,
	}
}
