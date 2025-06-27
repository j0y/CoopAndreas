package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
	// Configure global logger
	zerolog.TimeFieldFormat = "15:04:05"
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	
	printBanner()
	
	// Create server instance
	srv := server.New()
	
	// Create network server
	netServer := network.NewServer(Port, srv)
	
	// Connect server and network server
	srv.SetNetworkServer(netServer)
	
	// Start server in goroutine
	go func() {
		if err := netServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	zlog.Info().Int("port", Port).Msg("Server started")
	
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
}
