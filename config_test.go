package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		env           map[string]string
		wantPath      string
		wantTimeout   time.Duration
		wantError     bool
	}{
		{name: "missing file", wantPath: ".venv/bin/python-private-package", wantTimeout: 30 * time.Second},
		{name: "dotenv", content: "PYTHON_CLI_PATH='/path with spaces/cli'\nPYTHON_CLI_TIMEOUT=1m\n", wantPath: "/path with spaces/cli", wantTimeout: time.Minute},
		{name: "environment overrides dotenv", content: "PYTHON_CLI_PATH=/file/cli\nPYTHON_CLI_TIMEOUT=1m\n", env: map[string]string{"PYTHON_CLI_PATH": "/env/cli", "PYTHON_CLI_TIMEOUT": "2m"}, wantPath: "/env/cli", wantTimeout: 2 * time.Minute},
		{name: "empty environment uses defaults", content: "PYTHON_CLI_PATH=/file/cli\nPYTHON_CLI_TIMEOUT=1m\n", env: map[string]string{"PYTHON_CLI_PATH": "", "PYTHON_CLI_TIMEOUT": ""}, wantPath: ".venv/bin/python-private-package", wantTimeout: 30 * time.Second},
		{name: "invalid timeout", env: map[string]string{"PYTHON_CLI_TIMEOUT": "abc"}, wantError: true},
		{name: "zero timeout", env: map[string]string{"PYTHON_CLI_TIMEOUT": "0s"}, wantError: true},
		{name: "negative timeout", env: map[string]string{"PYTHON_CLI_TIMEOUT": "-1s"}, wantError: true},
		{name: "malformed dotenv", content: "PYTHON_CLI_PATH=\"unterminated\n", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// godotenv.Loadが変更する環境変数もテスト終了時に復元する。
			for _, key := range []string{"PYTHON_CLI_PATH", "PYTHON_CLI_TIMEOUT"} {
				t.Setenv(key, "")
				if err := os.Unsetenv(key); err != nil {
					t.Fatal(err)
				}
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			path := filepath.Join(t.TempDir(), ".env")
			if tc.content != "" {
				if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := loadConfig(path)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected configuration error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.CLIPath != tc.wantPath || got.CLITimeout != tc.wantTimeout {
				t.Fatalf("got %+v; want path=%q, timeout=%v", got, tc.wantPath, tc.wantTimeout)
			}
		})
	}
}
