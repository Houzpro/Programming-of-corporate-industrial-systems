package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/user/mastermind/pkg/logging"
)

func main() {
	logging.Init(logging.LogLevelInfo, nil)

	serverHost := flag.String("host", "localhost", "Server host")
	serverPort := flag.String("port", "8080", "Server port")
	flag.Parse()

	serverAddr := net.JoinHostPort(*serverHost, *serverPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logging.Logger.Info("Received signal, disconnecting", "signal", sig)
		cancel()
	}()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		logging.Logger.Error("Error connecting to server", "error", err, "serverAddr", serverAddr)
		os.Exit(1)
	}
	defer conn.Close()

	logging.Logger.Info("Connected to Mastermind server", "serverAddr", serverAddr)
	fmt.Println("Connected to Mastermind server")
	fmt.Println("Enter your guesses (4 characters A-Z or 0-9)")
	fmt.Println("Type 'quit' to exit")

	go readResponses(ctx, conn)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			guess := scanner.Text()
			if strings.ToLower(guess) == "quit" {
				logging.Logger.Info("User requested disconnect")
				fmt.Println("Disconnecting from server")
				return
			}

			guess = strings.ToUpper(guess)
			logging.Logger.Debug("Sending guess to server", "guess", guess)

			_, err := conn.Write([]byte(guess + "\n"))
			if err != nil {
				logging.Logger.Error("Error sending guess", "error", err)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Logger.Error("Error reading input", "error", err)
	}
}

func readResponses(ctx context.Context, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if scanner.Scan() {
				response := scanner.Text()
				fmt.Println("Server:", response)

				if strings.Contains(response, "win") || strings.Contains(response, "lose") {
					fmt.Println("Enter a new guess when the next round starts")
				}
			} else {
				if err := scanner.Err(); err != nil {
					logging.Logger.Error("Error reading server response", "error", err)
				} else {
					logging.Logger.Info("Server connection closed")
				}
				return
			}
		}
	}
}
