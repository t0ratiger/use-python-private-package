package main

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGenerateCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.csv")
	if err := generateCSV(path, 123); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 124 || strings.Join(records[0], ",") != "ID,price,quantity" {
		t.Fatalf("unexpected CSV header or row count: %d", len(records))
	}
	for i, row := range records[1:] {
		price, priceErr := strconv.ParseFloat(row[1], 64)
		quantity, quantityErr := strconv.Atoi(row[2])
		if priceErr != nil || price < 1 || price > 1000 || quantityErr != nil || quantity < 1 || quantity > 10 {
			t.Fatalf("invalid row %d: %v", i+1, row)
		}
	}
	if records[123][0] != "A123" {
		t.Fatalf("unexpected final ID: %s", records[123][0])
	}
}

func TestRunRejectsInvalidRows(t *testing.T) {
	for _, rows := range []int{0, -1} {
		if err := run(context.Background(), rows, "missing-cli", io.Discard); err == nil {
			t.Fatalf("rows=%d should fail", rows)
		}
	}
}

func TestPythonCLI(t *testing.T) {
	cli, err := filepath.Abs(".venv/bin/python-private-package")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("Python CLIをインストールしてから実行してください: %v", err)
	}
	for _, tc := range []struct {
		name, csv, want, wantError string
	}{
		{"total", "ID,price,quantity\nA001,100,2\nA002,150.5,3\n", "651.5\n", ""},
		{"decimal", "ID,price,quantity\nA001,0.10,3\nA002,0.20,1\n", "0.50\n", ""},
		{"invalid", "ID,price,quantity\nA001,invalid,2\n", "", "price"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "input.csv")
			if err := os.WriteFile(input, []byte(tc.csv), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			got, err := calculateTotal(ctx, cli, input, filepath.Join(dir, "total.txt"))
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected CLI diagnostic containing %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
