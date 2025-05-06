package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/user/mastermind/pkg/config"
	"github.com/user/mastermind/pkg/game"
	"github.com/user/mastermind/pkg/logging"
	"github.com/user/mastermind/pkg/storage"
)

type GameServer struct {
	currentGame   *game.Game
	players       map[string]net.Conn
	mu            sync.Mutex
	playerResults map[string]int
	gameResult    storage.GameResult
	waitGroup     sync.WaitGroup
	minPlayers    int
	gameActive    bool
	config        *config.Config
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewGameServer(minPlayers int, cfg *config.Config) *GameServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameServer{
		players:       make(map[string]net.Conn),
		playerResults: make(map[string]int),
		minPlayers:    minPlayers,
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *GameServer) StartServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logging.Logger.ErrorContext(s.ctx, "Failed to start server", "error", err, "port", port)
		return err
	}
	defer listener.Close()

	logging.Logger.InfoContext(s.ctx, "Server started", "port", port)
	s.startNewRound()

	for {
		select {
		case <-s.ctx.Done():
			logging.Logger.InfoContext(s.ctx, "Server shutting down")
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				logging.Logger.ErrorContext(s.ctx, "Error accepting connection", "error", err)
				continue
			}

			s.mu.Lock()
			if len(s.players) >= s.config.Game.MaxPlayers {
				conn.Write([]byte("Server is full. Try again later.\n"))
				conn.Close()
				s.mu.Unlock()
				continue
			}

			playerName := fmt.Sprintf("Player%d", len(s.players)+1)
			s.players[playerName] = conn
			s.playerResults[playerName] = 0
			logging.Logger.InfoContext(s.ctx, "Player joined", "player", playerName)

			if len(s.players) >= s.minPlayers && !s.gameActive {
				s.gameActive = true
				logging.Logger.InfoContext(s.ctx, "Game starting", "players", len(s.players))
				for _, playerConn := range s.players {
					playerConn.Write([]byte("Game is starting! Try to guess the secret code.\n"))
				}
			}

			s.mu.Unlock()
			s.waitGroup.Add(1)
			playerCtx, cancel := context.WithCancel(s.ctx)
			defer cancel()
			go s.handlePlayer(playerCtx, conn, playerName)
		}
	}
}

func (s *GameServer) Stop() {
	s.cancel()
	s.waitGroup.Wait()
	logging.Logger.InfoContext(context.Background(), "Server stopped")
}

func (s *GameServer) startNewRound() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentGame = game.NewGame(&s.config.Game)
	s.gameResult = storage.NewGameResult(s.currentGame.StartTime, s.currentGame.SecretCode)
	s.gameActive = len(s.players) >= s.minPlayers
	s.playerResults = make(map[string]int)

	logging.Logger.InfoContext(s.ctx, "New round started", "secretCode", s.currentGame.SecretCode)
}

func (s *GameServer) handlePlayer(ctx context.Context, conn net.Conn, name string) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.players, name)
		s.mu.Unlock()
		s.waitGroup.Done()
		logging.Logger.InfoContext(ctx, "Player left", "player", name)
	}()

waitLoop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.mu.Lock()
			active := s.gameActive
			s.mu.Unlock()

			if active {
				break waitLoop
			}

			conn.Write([]byte("Waiting for more players to join...\n"))
			time.Sleep(2 * time.Second)
		}
	}

	attempts := 0
	for attempts < s.config.Game.MaxAttempts {
		select {
		case <-ctx.Done():
			return
		default:
			var timeoutCtx context.Context
			var cancel context.CancelFunc

			if s.config.Server.EnableTimeout {
				timeoutCtx, cancel = context.WithTimeout(ctx, s.config.Server.InputTimeout)
				defer cancel()
				go s.sendTimeoutWarnings(timeoutCtx, conn)
			} else {
				timeoutCtx = ctx
				cancel = func() {} // No-op cancel function
			}

			readCh := make(chan string)
			errCh := make(chan error)

			go func() {
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					errCh <- err
					return
				}
				readCh <- string(buffer[:n])
			}()

			select {
			case <-ctx.Done():
				cancel()
				return
			case <-timeoutCtx.Done():
				if s.config.Server.EnableTimeout && ctx.Err() == nil {
					conn.Write([]byte("Time's up! You took too long to make a guess.\n"))
					continue
				}
			case err := <-errCh:
				logging.Logger.ErrorContext(ctx, "Error reading from player", "player", name, "error", err)
				cancel()
				return
			case input := <-readCh:
				cancel()
				guess := strings.TrimSpace(input)
				guess = strings.ToUpper(guess)

				if len(guess) != s.config.Game.CodeLength {
					conn.Write([]byte(fmt.Sprintf("Invalid guess! The code must be exactly %d characters long.\n", s.config.Game.CodeLength)))
					continue
				}

				attempts++

				s.mu.Lock()
				s.playerResults[name] = attempts
				black, white := s.currentGame.CalculateMarkers(guess)

				s.checkAllPlayersFinished()

				s.mu.Unlock()

				logging.Logger.InfoContext(ctx, "Player made guess",
					"player", name,
					"guess", guess,
					"black", black,
					"white", white,
					"attempts", attempts)

				response := fmt.Sprintf("Black: %d, White: %d\n", black, white)
				conn.Write([]byte(response))

				if s.currentGame.CheckWin(guess) {
					s.mu.Lock()
					s.endRound(ctx, name)
					s.mu.Unlock()
					conn.Write([]byte("You win! The code was " + s.currentGame.SecretCode + "\n"))
					return
				}

				if attempts >= s.config.Game.MaxAttempts {
					conn.Write([]byte("You have reached maximum attempts. Waiting for other players to finish...\n"))

					s.mu.Lock()
					allFinished := s.checkAllPlayersFinished()
					if allFinished {
						s.endRoundNoWinner(ctx)
						s.mu.Unlock()
						return
					}
					s.mu.Unlock()
				}
			}
		}
	}
}

func (s *GameServer) sendTimeoutWarnings(ctx context.Context, conn net.Conn) {
	initialWarning := time.After(s.config.Server.InputTimeout - s.config.Server.WarningTime)
	finalWarning := time.After(s.config.Server.InputTimeout - 3*time.Second)

	select {
	case <-ctx.Done():
		return
	case <-initialWarning:
		conn.Write([]byte(fmt.Sprintf("Warning: You have %d seconds left to make your guess!\n", int(s.config.Server.WarningTime.Seconds()))))

		// Send intermediate warnings (only one per interval)
		remainingTime := s.config.Server.WarningTime - 3*time.Second
		if remainingTime > 0 {
			intermediateWarning := time.After(remainingTime / 2)
			select {
			case <-ctx.Done():
				return
			case <-intermediateWarning:
				secondsLeft := int(remainingTime.Seconds() / 2)
				if secondsLeft > 0 {
					conn.Write([]byte(fmt.Sprintf("Warning: %d seconds remaining!\n", secondsLeft)))
				}
			}
		}

		// Final warning
		select {
		case <-ctx.Done():
			return
		case <-finalWarning:
			conn.Write([]byte("Final warning: 3 seconds left!\n"))
		}
	}
}

func (s *GameServer) checkAllPlayersFinished() bool {
	if len(s.players) == 0 {
		return true
	}

	for name := range s.players {
		if attempts, exists := s.playerResults[name]; !exists || attempts < s.config.Game.MaxAttempts {
			return false
		}
	}
	return true
}

func (s *GameServer) broadcastMessage(message string) {
	for name, conn := range s.players {
		_, err := conn.Write([]byte(message))
		if err != nil {
			logging.Logger.ErrorContext(s.ctx, "Error sending message to player", "player", name, "error", err)
		}
	}
}

func (s *GameServer) endRoundNoWinner(ctx context.Context) {
	for name, attempts := range s.playerResults {
		s.gameResult.AddPlayerStat(name, attempts)
	}

	s.gameResult.Finalize("No Winner")

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("results_%s.xml", timestamp)
	err := storage.SaveGameResult(ctx, s.gameResult, filename)
	if err != nil {
		logging.Logger.ErrorContext(ctx, "Error saving game results", "error", err)
	}

	message := fmt.Sprintf("Game over! No one guessed the code: %s\n", s.currentGame.SecretCode)
	s.broadcastMessage(message + "Starting new round shortly...\n")

	logging.Logger.InfoContext(ctx, "Round ended with no winner", "secretCode", s.currentGame.SecretCode)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			s.startNewRound()
			s.broadcastMessage("New round started! Try to guess the secret code.\n")
		}
	}()
}

func (s *GameServer) endRound(ctx context.Context, winner string) {
	for name, attempts := range s.playerResults {
		s.gameResult.AddPlayerStat(name, attempts)
	}

	s.gameResult.Finalize(winner)

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("results_%s.xml", timestamp)
	err := storage.SaveGameResult(ctx, s.gameResult, filename)
	if err != nil {
		logging.Logger.ErrorContext(ctx, "Error saving game results", "error", err)
	}

	message := fmt.Sprintf("Game over! %s wins with the code %s\n", winner, s.currentGame.SecretCode)
	s.broadcastMessage(message + "Starting new round shortly...\n")

	logging.Logger.InfoContext(ctx, "Round ended with winner", "winner", winner, "secretCode", s.currentGame.SecretCode)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			s.startNewRound()
			s.broadcastMessage("New round started! Try to guess the secret code.\n")
		}
	}()
}
