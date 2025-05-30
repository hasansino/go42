package main

import (
	"log"

	"github.com/hasansino/cfg2env"

	"github.com/hasansino/go42/internal/config"
	"github.com/hasansino/go42/internal/tools"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	e := cfg2env.New(
		cfg2env.WithEnvironmentTagName(config.TagNameEnvVarName),
		cfg2env.WithDefaultValueTagName(config.TagNameDefaultValue),
		cfg2env.WithExportedFileName(".env.example"),
		cfg2env.WithExtraEntry("COMPOSE_PROJECT_NAME", cfg.Core.ServiceName),
		cfg2env.WithExtraTagExtraction(tools.ValidateRulesTagName),
	)
	err = e.ToFile(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
