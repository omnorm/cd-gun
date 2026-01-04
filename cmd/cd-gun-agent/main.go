package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/omnorm/cd-gun/internal/app"
)

const version = "0.1.1"

func main() {
	var (
		configPath  = flag.String("config", "/etc/cd-gun/config.yaml", "Path to configuration file")
		logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		showVersion = flag.Bool("version", false, "Show version")
		help        = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("cd-gun version %s\n", version)
		os.Exit(0)
	}

	// Create and start the app
	app, err := app.NewApp(*configPath, *logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`CD-Gun - Universal CD/GitOps Agent v%s

Usage: cd-gun-agent [options]

Options:
  -config string
        Path to configuration file (default "/etc/cd-gun/config.yaml")
  -log-level string
        Log level: debug, info, warn, error (default "info")
  -version
        Show version and exit
  -help
        Show this help message and exit

Signals:
  SIGHUP  - Reload configuration
  SIGUSR1 - Force check all repositories
  SIGTERM - Graceful shutdown

Example:
  cd-gun-agent -config /etc/cd-gun/config.yaml -log-level info

Documentation:
  https://github.com/omnorm/cd-gun

`, version)
}
