package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	config, err := loadConfig(".env")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	rows := flag.Int("rows", 10, "生成するCSVのデータ行数（1以上）")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), config.CLITimeout)
	defer cancel()
	if err := run(ctx, *rows, config.CLIPath, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, rows int, cli string, output io.Writer) error {
	if rows < 1 {
		return fmt.Errorf("-rows は1以上で指定してください")
	}
	dir, err := os.MkdirTemp("", "go-python-poc-*")
	if err != nil {
		return fmt.Errorf("一時ディレクトリの作成: %w", err)
	}
	defer os.RemoveAll(dir)

	inputPath := filepath.Join(dir, "input.csv")
	outputPath := filepath.Join(dir, "total.txt")
	if err := generateCSV(inputPath, rows); err != nil {
		return fmt.Errorf("CSVの生成: %w", err)
	}
	total, err := calculateTotal(ctx, cli, inputPath, outputPath)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(output, total); err != nil {
		return fmt.Errorf("標準出力への書き込み: %w", err)
	}
	return nil
}

func generateCSV(path string, rows int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(f)
	err = writer.Write([]string{"ID", "price", "quantity"})
	for i := 1; err == nil && i <= rows; i++ {
		// 単価は1.00〜1000.00。整数で扱い、小数の丸め誤差を避ける。
		cents := rand.IntN(99901) + 100
		err = writer.Write([]string{
			fmt.Sprintf("A%03d", i),
			fmt.Sprintf("%d.%02d", cents/100, cents%100),
			strconv.Itoa(rand.IntN(10) + 1),
		})
	}
	writer.Flush()
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if err := writer.Error(); err != nil {
		return err
	}
	return closeErr
}

func calculateTotal(ctx context.Context, cli, inputPath, outputPath string) (string, error) {
	cmd := exec.CommandContext(ctx, cli, inputPath, "--output", outputPath)
	diagnostics, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("Python CLIの実行: %w", ctx.Err())
		}
		return "", fmt.Errorf("Python CLIの実行: %w\n%s", err, strings.TrimSpace(string(diagnostics)))
	}
	total, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("計算結果の読み込み: %w", err)
	}
	return string(total), nil
}
