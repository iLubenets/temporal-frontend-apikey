package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/ilubenets/temporal-apikey/src/authorizer"
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
		log.Fatalf("config [%s/%s.yaml] not found or corrupted: %v", configDirPath, env, err)
	}

	claimMappers := authorizer.NewMultiClaimMapper(logger)
	// Prefer API key processing first so JWT errors do not short-circuit
	if apiKeys := os.Getenv("TEMPORAL_API_KEYS"); apiKeys != "" {
		apiKeyClaimMapper, err := authorizer.NewAPIKeyClaimMapper(apiKeys, logger)
		if err != nil {
			log.Fatalf("ApiKeyClaimMapper: %v", err)
		}
		claimMappers.Add("apiKeyClaimMapper", apiKeyClaimMapper)
	}

	if strings.EqualFold(cfg.Global.Authorization.ClaimMapper, "default") {
		claimMappers.Add("defaultJWTClaimMapper", authorization.NewDefaultJWTClaimMapper(
			authorization.NewDefaultTokenKeyProvider(&cfg.Global.Authorization, logger), &cfg.Global.Authorization, logger,
		))
	}

	s, err := temporal.NewServer(
		temporal.ForServices([]string{
			string(primitives.FrontendService),
		}),
		temporal.WithConfig(cfg),
		temporal.InterruptOn(temporal.InterruptCh()),
		temporal.WithAuthorizer(authorization.NewDefaultAuthorizer()),
		// customer claim manager
		temporal.WithClaimMapper(func(*config.Config) authorization.ClaimMapper { return claimMappers }),
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Starting Temporal with API key protection")
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
