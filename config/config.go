package config

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
)

type Config struct {
	SPOTIFY_ID     string       `json:"spotifyid"`
	SPOTIFY_SECRET string       `json:"spotifysecret"`
	Token          oauth2.Token `json:"token"`
	BackupPath     string       `json:"backuppath"`
}

func CreateEmpty() *Config {
	return &Config{
		SPOTIFY_ID:     "",
		SPOTIFY_SECRET: "",
		Token:          oauth2.Token{},
		BackupPath:     "",
	}
}

func LoadFromFile(path string) (*Config, error) {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	if config.SPOTIFY_ID == "" {
		config.SPOTIFY_ID = os.Getenv("SPOTIFY_ID")
	}
	if config.SPOTIFY_SECRET == "" {
		config.SPOTIFY_SECRET = os.Getenv("SPOTIFY_SECRET")
	}
	return &config, nil
}

func WriteToFile(path string, config Config) error {
	file, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, file, os.ModePerm)
}
