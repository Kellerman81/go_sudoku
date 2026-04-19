//go:build linux && !android

package ui

import (
	"os"
	"path/filepath"
)

func pdfSavePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "sudoku_book.pdf"
	}
	return filepath.Join(home, "sudoku_book.pdf")
}

func pdfSupported() bool { return true }
