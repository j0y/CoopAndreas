package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"coopandreas-server/internal/network"
	"coopandreas-server/internal/server"
	cooptypes "coopandreas-server/internal/types"
)

const (
	Version = cooptypes.ServerVersion
	Port    = 6767
)

func main() {
	// Parse command line flags
	var logLevel string
	flag.StringVar(&logLevel, "log-level", "info", "Set log level (trace, debug, info, warn, error, fatal, panic)")
	flag.Parse()

	// Configure global logger
	zerolog.TimeFieldFormat = "15:04:05"
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Set log level based on flag
	level, err := parseLogLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid log level '%s': %v", logLevel, err)
	}
	zerolog.SetGlobalLevel(level)

	printBanner()

	zlog.Info().Str("logLevel", logLevel).Msg("Server starting with log level")

	// Create server instance
	srv := server.New()

	// Create ENet-compatible network server
	netServer := network.NewServer(Port, srv)

	// Connect server and network server
	srv.SetNetworkServer(netServer)

	// Start server in goroutine
	go func() {
		if err := netServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	zlog.Info().Int("port", Port).Msg("ENet server started - compatible with C++ clients")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	zlog.Info().Msg("Server shutting down...")
	netServer.Stop()
	zlog.Info().Msg("Server shutdown complete")
}

func printBanner() {
	fmt.Println("[!] : Support:")
	fmt.Println("- https://github.com/Tornamic/CoopAndreas")
	fmt.Println("- https://discord.gg/TwQsR4qxVx")
	fmt.Println("- coopandreasmod@gmail.com")
	fmt.Println()
	fmt.Printf("[!] : CoopAndreas Server Go Prototype\n")
	fmt.Printf("[!] : Version : %s\n", Version)
	fmt.Printf("[!] : Platform : Go Runtime\n")
	fmt.Println()
	fmt.Println("[!] : Usage:")
	fmt.Println("  -log-level string")
	fmt.Println("        Set log level (trace, debug, info, warn, error, fatal, panic) (default \"info\")")
}

// parseLogLevel converts a string log level to zerolog.Level
func parseLogLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s (valid levels: trace, debug, info, warn, error, fatal, panic)", level)
	}
}
