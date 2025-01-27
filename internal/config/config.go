package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl       string `json:"db_url"`
	CurrentUser string `json:"current_user"`
}

func getUserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func Read() (Config, error) {
	homeDir, err := getUserHomeDir()
	if err != nil {
		return Config{}, err
	}
	fileName := filepath.Join(homeDir, configFileName)
	configFile, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		panic(err)
	}
	var cfg Config
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUser = user
	homeDir, err := getUserHomeDir()
	if err != nil {
		return err
	}
	fileName := filepath.Join(homeDir, configFileName)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
