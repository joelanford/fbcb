package main

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-registry/pkg/image/containerdregistry"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/joelanford/fbcb/internal/fbcb"
)

func main() {
	var (
		configFile      string
		packagesBaseDir string
		cacheDir        string
	)
	cmd := &cobra.Command{
		Use: "fbcb",
		Run: func(cmd *cobra.Command, args []string) {
			bc, err := fbcb.LoadBuildConfigFile(configFile)
			if err != nil {
				log.Fatalf("load build config file %q: %v", configFile, err)
			}

			configFileDir := filepath.Dir(configFile)
			pcs, err := fbcb.LoadPackageConfigs(filepath.Join(configFileDir, bc.PackagesBaseDir))
			if err != nil {
				log.Fatalf("load package configs from base directory %q: %v", packagesBaseDir, err)
			}

			destroyCache := cacheDir == ""
			if destroyCache {
				cacheDir, err = os.MkdirTemp("", "fbcb-image-cache-")
				if err != nil {
					log.Fatalf("create temporary image cache directory: %v", err)
				}
			}

			regOpts := []containerdregistry.RegistryOption{
				containerdregistry.WithLog(nullLogger()),
				containerdregistry.WithCacheDir(cacheDir),
			}

			reg, err := containerdregistry.NewRegistry(regOpts...)
			if err != nil {
				log.Fatalf("initialize image registry: %v", err)
			}

			if destroyCache {
				defer reg.Destroy()
			}

			builders, err := fbcb.CreateBuilders(*bc, pcs, reg)
			if err != nil {
				log.Fatalf("create builders for each catalog and package: %v", err)
			}

			for _, b := range builders {
				if err := b.Build(cmd.Context()); err != nil {
					log.Fatal(err)
				}
			}

			fmt.Println("\n\nSUMMARY:")
			for _, b := range builders {
				fmt.Printf("  Built catalog image %q with packages:\n", b.BuildConfig.Destination.OutputImage)
				for _, pkg := range b.Packages() {
					fmt.Printf("    %s\n", pkg)
				}
				fmt.Printf("\n")
			}
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "c", "fbcb.yaml", "Path to fbcb config file.")
	cmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Path of persistent image cache directory.")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}

func nullLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logrus.NewEntry(logger)
}
