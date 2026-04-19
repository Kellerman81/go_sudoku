//go:build windows

package ui

func pdfSavePath() string { return "sudoku_book.pdf" }
func pdfSupported() bool  { return true }
