package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	var (
		cacheDir         string
		port             string
		externalHostname string
	)

	pflag.StringVar(&cacheDir, "cache-dir", "/cache", "Directory to store cache")
	pflag.StringVar(&port, "port", "8080", "Port to artifact service will listen on")
	pflag.StringVar(&externalHostname, "external-hostname", "", "External hostname to use for download URLs")

	bindEnv(pflag.Lookup("cache-dir"), "CACHE_DIR")
	bindEnv(pflag.Lookup("port"), "PORT")
	bindEnv(pflag.Lookup("external-hostname"), "EXTERNAL_HOSTNAME")

	pflag.Parse()

	srv, err := NewLocalService(cacheDir, externalHostname, port)
	if err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}

	if err := Serve(port, srv); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
