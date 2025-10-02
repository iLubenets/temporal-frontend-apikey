package main

import (
	"log"
	"os"
	"path"

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	logpkg "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/temporal"
)

func main() {
	logger := logpkg.NewZapLogger(logpkg.BuildZapLogger(logpkg.Config{Level: "info"}))

	env := os.Getenv(config.EnvKeyEnvironment)
	if env == "" {
		env = "development"
	}
	root := os.Getenv(config.EnvKeyRoot)
	if root == "" {
		root = "."
	}
	configDir := os.Getenv(config.EnvKeyConfigDir)
	if configDir == "" {
		configDir = "config"
	}
	configDirPath := path.Join(root, configDir)
	zone := os.Getenv(config.EnvKeyAvailabilityZone)
	cfg, err := config.LoadConfig(env, configDirPath, zone)
	if err != nil {
		log.Fatal(err)
	}

	// custom API-KEY config
	apiKeys := os.Getenv("TEMPORAL_API_KEYS")
	if apiKeys == "" {
		log.Fatal(err)
	}
	mapper, err := NewAPIKeyClaimMapper(apiKeys, logger)
	if err != nil {
		log.Fatal(err)
	}

	s, err := temporal.NewServer(
		temporal.ForServices([]string{
			string(primitives.FrontendService),
		}),
		temporal.WithConfig(cfg),
		temporal.InterruptOn(temporal.InterruptCh()),
		temporal.WithAuthorizer(authorization.NewDefaultAuthorizer()),
		// customer claim manager
		temporal.WithClaimMapper(func(*config.Config) authorization.ClaimMapper { return mapper }),
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Starting Temporal with API key protection")
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
