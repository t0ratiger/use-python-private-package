package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv" // PoC のため妥協
)

type Config struct {
	CLIPath    string
	CLITimeout time.Duration
}

func loadConfig(envFile string) (Config, error) {
	// .envは開発用の補助。すでに設定された環境変数は上書きしない。
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("%sの読み込み: %w", envFile, err)
	}
	config := Config{
		CLIPath:    os.Getenv("PYTHON_CLI_PATH"),
		CLITimeout: 30 * time.Second,
	}
	if config.CLIPath == "" {
		config.CLIPath = ".venv/bin/python-private-package"
	}
	if value := os.Getenv("PYTHON_CLI_TIMEOUT"); value != "" {
		timeout, err := time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			return Config{}, fmt.Errorf("PYTHON_CLI_TIMEOUT は30sや1mなどの正の時間で指定してください")
		}
		config.CLITimeout = timeout
	}
	return config, nil
}
