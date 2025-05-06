package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/mastermind/pkg/config"
	"github.com/user/mastermind/pkg/logging"
	"github.com/user/mastermind/pkg/server"
)

func main() {
	logging.Init(logging.LogLevelInfo, nil)

	port := flag.String("port", "8080", "Server port")
	minPlayers := flag.Int("min-players", 2, "Minimum number of players to start the game")
	enableTimeout := flag.Bool("enable-timeout", false, "Enable time limit for player moves")
	flag.Parse()

	cfg, err := config.LoadConfig("")
	if err != nil {
		logging.Logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set the timeout flag based on command-line argument
	cfg.Server.EnableTimeout = *enableTimeout

	if *minPlayers < 1 || *minPlayers > cfg.Game.MaxPlayers {
		logging.Logger.Error("Invalid number of minimum players",
			"minPlayers", *minPlayers,
			"maxPlayers", cfg.Game.MaxPlayers)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logging.Logger.Info("Received signal, shutting down", "signal", sig)
		cancel()
	}()

	gameServer := server.NewGameServer(*minPlayers, cfg)
	logging.Logger.Info("Starting Mastermind server", "port", *port, "minPlayers", *minPlayers)

	go func() {
		if err := gameServer.StartServer(*port); err != nil {
			logging.Logger.Error("Error starting server", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	fmt.Println("Server shutting down...")
}
