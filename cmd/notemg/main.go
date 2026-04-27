package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/notemg/notemg/internal/config"
	"github.com/notemg/notemg/internal/server"
)

//go:embed all:frontend/dist
var frontendEmbed embed.FS

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	dataDir := flag.String("data", "", "Data directory (overrides config)")
	port := flag.Int("port", 0, "Server port (overrides config)")
	dev := flag.Bool("dev", false, "Development mode")
	flag.Parse()

	command := flag.Arg(0)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *dataDir != "" {
		cfg.Data.Dir = *dataDir
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *dev {
		cfg.Server.Port = 5173
	}

	switch command {
	case "init":
		if err := runInit(cfg); err != nil {
			log.Fatalf("Init failed: %v", err)
		}
	case "serve", "":
		if err := runServe(cfg); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: notemg [init|serve] [flags]")
		os.Exit(1)
	}
}

func getFrontendFS() embed.FS {
	sub, err := fs.Sub(frontendEmbed, "frontend/dist")
	if err != nil {
		log.Println("Warning: frontend dist not found, serving API only")
	}
	_ = sub
	return frontendEmbed
}

func runInit(cfg *config.Config) error {
	if err := cfg.EnsureDataDirs(); err != nil {
		return err
	}
	fmt.Println("NoteMG initialized. Data directory:", cfg.Data.Dir)
	fmt.Println("Run 'notemg serve' to start the server, then set your password via the web UI.")
	return nil
}

func runServe(cfg *config.Config) error {
	if err := cfg.EnsureDataDirs(); err != nil {
		return err
	}

	srv := server.New(cfg, frontendEmbed)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
	}()

	return srv.Start()
}

var _ = http.ListenAndServe
