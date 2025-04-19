package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/hasansino/goapp/internal/config"
)

var (
	// These variables are passed as arguments to compiler.
	buildDate   string
	buildCommit string
)

var cfg *config.Config

func init() {
	if len(buildDate) == 0 {
		buildDate = "dev"
	}
	if len(buildCommit) == 0 {
		buildCommit = "dev"
	}

	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)

	var err error
	cfg, err = config.New()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v\n", err)
	}
}

func main() {
	log.Printf("%v\n", cfg)

	// listen for exit signals
	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGINT, syscall.SIGTERM)
	shutdown(<-sys)
}

// shutdown implements all graceful shutdown logic.
func shutdown(_ os.Signal, closers ...io.Closer) {
	log.Println("Shutting down...")
	for _, c := range closers {
		if err := c.Close(); err != nil {
			log.Printf(
				"Error closing %s: %v",
				reflect.TypeOf(c).String(), err,
			)
		}
	}
	os.Exit(0)
}
