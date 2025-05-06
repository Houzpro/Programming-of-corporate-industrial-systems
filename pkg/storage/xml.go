package storage

import (
	"context"
	"encoding/xml"
	"os"
	"time"

	"github.com/user/mastermind/pkg/logging"
)

type GameResult struct {
	XMLName     xml.Name     `xml:"GameResult"`
	StartTime   string       `xml:"StartTime"`
	EndTime     string       `xml:"EndTime"`
	SecretCode  string       `xml:"SecretCode"`
	PlayerStats []PlayerStat `xml:"PlayerStat"`
	Winner      string       `xml:"Winner"`
}

type PlayerStat struct {
	Name     string `xml:"Name"`
	Attempts int    `xml:"Attempts"`
}

func SaveGameResult(ctx context.Context, result GameResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		logging.Logger.ErrorContext(ctx, "Failed to create result file",
			"error", err, "filename", filename)
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(result); err != nil {
		logging.Logger.ErrorContext(ctx, "Failed to encode game result",
			"error", err, "filename", filename)
		return err
	}

	logging.Logger.InfoContext(ctx, "Game result saved successfully", "filename", filename)
	return nil
}

func NewGameResult(startTime time.Time, secretCode string) GameResult {
	return GameResult{
		StartTime:   startTime.Format(time.RFC3339),
		SecretCode:  secretCode,
		PlayerStats: []PlayerStat{},
	}
}

func (g *GameResult) Finalize(winner string) {
	g.EndTime = time.Now().Format(time.RFC3339)
	g.Winner = winner
}

func (g *GameResult) AddPlayerStat(name string, attempts int) {
	g.PlayerStats = append(g.PlayerStats, PlayerStat{
		Name:     name,
		Attempts: attempts,
	})
}
