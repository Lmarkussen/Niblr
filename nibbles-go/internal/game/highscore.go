package game

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type highScoreFile struct {
	HighScore int `json:"high_score"`
}

func LoadHighScore() (int, error) {
	path, err := highScorePath()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var file highScoreFile
	if err := json.Unmarshal(data, &file); err != nil {
		return 0, err
	}
	return file.HighScore, nil
}

func SaveHighScore(score int) error {
	path, err := highScorePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(highScoreFile{HighScore: score}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func highScorePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nibbles-go", "highscore.json"), nil
}
