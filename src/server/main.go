package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ilubenets/temporal-apikey/src/authorizer"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	logpkg "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/temporal"
)

func main() {
	logger := logpkg.NewZapLogger(logpkg.BuildZapLogger(logpkg.Config{Level: "debug"}))

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

	// DEBUG: print what we’ll pass to LoadConfig
	log.Printf("[BOOT] TEMPORAL_ENVIRONMENT=%q", env)
	log.Printf("[BOOT] TEMPORAL_ROOT=%q", root)
	log.Printf("[BOOT] TEMPORAL_CONFIG_DIR=%q", configDir)
	log.Printf("[BOOT] TEMPORAL_AVAILABILITY_ZONE=%q", zone)
	log.Printf("[BOOT] Effective configDirPath=%q", configDirPath)

	// DEBUG: list directories and show candidate filenames
	log.Printf("[BOOT] %s", dumpDir("/etc/temporal"))
	log.Printf("[BOOT] %s", dumpDir(configDirPath))

	// show which files we expect
	cand := []string{
		path.Join(configDirPath, env+".yaml"),
		path.Join(configDirPath, "base.yaml"),
		path.Join(configDirPath, "temporal.yaml"),
	}
	for _, f := range cand {
		_, err := os.Stat(f)
		log.Printf("[BOOT] check %s -> %v", f, err)
	}
	// --- end debug block ---

	cfg, err := config.LoadConfig(env, configDirPath, zone)
	if err != nil {
		log.Printf("[BOOT] LoadConfig error: %v", err)
		time.Sleep(time.Second * 60 * 15)
		log.Fatal(err)
	}

	// custom API-KEY auth
	mapper, err := authorizer.NewAPIKeyClaimMapper(logger)
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

func dumpDir(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Sprintf("ls %s -> ERROR: %v", dir, err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ls %s:\n", dir)
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Fprintf(&b, "  %s  %10d  %s\n",
			map[bool]string{true: "D", false: "-"}[e.IsDir()],
			info.Size(), e.Name())
		if e.IsDir() {
			// one level deeper
			sub := path.Join(dir, e.Name())
			subEntries, _ := os.ReadDir(sub)
			for _, se := range subEntries {
				si, _ := se.Info()
				fmt.Fprintf(&b, "    └─ %s %10d %s\n",
					map[bool]string{true: "D", false: "-"}[se.IsDir()],
					si.Size(), se.Name())
			}
		}
	}
	return b.String()
}
