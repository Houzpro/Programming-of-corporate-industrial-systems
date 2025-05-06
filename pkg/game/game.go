package game

import (
	"math/rand"
	"sync"
	"time"

	"github.com/user/mastermind/pkg/config"
)

type Game struct {
	SecretCode string
	mu         sync.Mutex
	StartTime  time.Time
	Config     *config.GameConfig
}

func NewGame(cfg *config.GameConfig) *Game {
	return &Game{
		SecretCode: GenerateCode(cfg.CodeLength),
		StartTime:  time.Now(),
		Config:     cfg,
	}
}

func GenerateCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = chars[r.Intn(len(chars))]
	}
	return string(code)
}

func (g *Game) CalculateMarkers(guess string) (black, white int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	guessRunes := []rune(guess)
	codeRunes := []rune(g.SecretCode)

	compareLength := len(guessRunes)
	if compareLength > len(codeRunes) {
		compareLength = len(codeRunes)
	}

	used := make([]bool, len(codeRunes))

	for i := 0; i < compareLength; i++ {
		if guessRunes[i] == codeRunes[i] {
			black++
			used[i] = true
		}
	}

	for i := 0; i < compareLength; i++ {
		if guessRunes[i] != codeRunes[i] {
			for j := range codeRunes {
				if !used[j] && guessRunes[i] == codeRunes[j] {
					white++
					used[j] = true
					break
				}
			}
		}
	}

	return
}

func (g *Game) CheckWin(guess string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return guess == g.SecretCode
}
