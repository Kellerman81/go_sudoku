// Package pdf provides Sudoku puzzle book export functionality.
package pdf

import (
	"fmt"
	"io"

	"github.com/jung-kurt/gofpdf"

	"github.com/Kellerman81/go_sudoku/game"
	hist "github.com/Kellerman81/go_sudoku/storage"
)

const pageMargin = 15.0

// BookOptions controls what gets generated into the PDF.
type BookOptions struct {
	NormalCount    int
	KillerCount    int
	SamuraiCount   int
	Difficulty     game.Difficulty
	PerPage        int // 1, 2, 4, 6, or 9 puzzles per page
	Solutions      bool
	IncludeHistory bool
}

// ExportBook generates fresh puzzles and writes a PDF book to w.
func ExportBook(w io.Writer, opts BookOptions, history *hist.History) error {
	// Generate all puzzles.
	var puzzles []*game.Puzzle
	for i := 0; i < opts.NormalCount; i++ {
		puzzles = append(puzzles, game.GenerateNormal(opts.Difficulty))
	}
	for i := 0; i < opts.KillerCount; i++ {
		p := game.GenerateKiller(opts.Difficulty)
		p.AssignCageIDs()
		puzzles = append(puzzles, p)
	}
	for i := 0; i < opts.SamuraiCount; i++ {
		puzzles = append(puzzles, game.GenerateSamurai(opts.Difficulty))
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Go Sudoku Puzzle Book", false)
	pdf.SetAuthor("Go Sudoku", false)
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(false, pageMargin)

	addTitlePage(pdf, opts, len(puzzles))
	addPuzzlePages(pdf, puzzles, opts.PerPage, false)
	if opts.Solutions {
		addSolutionsDivider(pdf)
		addPuzzlePages(pdf, puzzles, opts.PerPage, true)
	}
	if opts.IncludeHistory && history != nil && len(history.Results) > 0 {
		addHistoryPage(pdf, history)
	}
	return pdf.Output(w)
}

// ExportSinglePuzzle writes one puzzle (current game) to w — used for quick export.
func ExportSinglePuzzle(w io.Writer, p *game.Puzzle) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pageMargin, pageMargin, pageMargin)
	pdf.SetAutoPageBreak(false, pageMargin)
	pdf.AddPage()
	addPuzzlePage(pdf, p, 1, false, pageMargin, pageMargin, 210-2*pageMargin)
	pdf.AddPage()
	addPuzzlePage(pdf, p, 1, true, pageMargin, pageMargin, 210-2*pageMargin)
	return pdf.Output(w)
}

// ---- page builders ---------------------------------------------------------

func addTitlePage(pdf *gofpdf.Fpdf, opts BookOptions, total int) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 28)
	pdf.Ln(40)
	pdf.CellFormat(0, 14, "Go Sudoku", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 14)
	pdf.CellFormat(0, 10, "Puzzle Book", "", 1, "C", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 8, fmt.Sprintf("Difficulty: %s", opts.Difficulty), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Total puzzles: %d", total), "", 1, "C", false, 0, "")
	if opts.NormalCount > 0 {
		pdf.CellFormat(0, 7, fmt.Sprintf("  Normal: %d", opts.NormalCount), "", 1, "C", false, 0, "")
	}
	if opts.KillerCount > 0 {
		pdf.CellFormat(0, 7, fmt.Sprintf("  Killer: %d", opts.KillerCount), "", 1, "C", false, 0, "")
	}
	if opts.SamuraiCount > 0 {
		pdf.CellFormat(0, 7, fmt.Sprintf("  Samurai: %d", opts.SamuraiCount), "", 1, "C", false, 0, "")
	}
}

func addSolutionsDivider(pdf *gofpdf.Fpdf) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 22)
	pdf.Ln(60)
	pdf.CellFormat(0, 14, "Solutions", "", 1, "C", false, 0, "")
}

type puzzleSlot struct {
	puzzle *game.Puzzle
	num    int
}

// addPuzzlePages lays puzzles onto pages according to perPage setting.
// Samurai puzzles always get a full page regardless of perPage.
func addPuzzlePages(pdf *gofpdf.Fpdf, puzzles []*game.Puzzle, perPage int, solution bool) {
	// A4 printable area: 210 - 2*margin wide, 297 - 2*margin tall.
	const pageW = 210 - 2*pageMargin
	const pageH = 297 - 2*pageMargin

	var normal []puzzleSlot
	for i, p := range puzzles {
		if p.Type == game.Samurai {
			if len(normal) > 0 {
				layoutBatch(pdf, normal, perPage, solution, pageW, pageH)
				normal = nil
			}
			pdf.AddPage()
			addPuzzlePage(pdf, p, i+1, solution, pageMargin, pageMargin, pageW)
		} else {
			normal = append(normal, puzzleSlot{p, i + 1})
		}
	}
	if len(normal) > 0 {
		layoutBatch(pdf, normal, perPage, solution, pageW, pageH)
	}
}

// layoutBatch places non-Samurai puzzles using a grid layout.
func layoutBatch(pdf *gofpdf.Fpdf, slots []puzzleSlot, perPage int, solution bool, pageW, pageH float64) {
	if perPage <= 0 {
		perPage = 1
	}
	// Determine grid columns/rows for this perPage.
	cols, rows := gridDims(perPage)
	cellW := pageW / float64(cols)
	cellH := pageH / float64(rows)

	idx := 0
	for idx < len(slots) {
		pdf.AddPage()
		for slot := 0; slot < perPage && idx < len(slots); slot++ {
			col := slot % cols
			row := slot / cols
			x := pageMargin + float64(col)*cellW
			y := pageMargin + float64(row)*cellH

			// For 9×9 grids, use a square inscribed in the cell with some padding.
			pad := 4.0
			gridSize := min2(cellW-2*pad, cellH-2*pad-10) // 10 mm for label
			gx := x + (cellW-gridSize)/2
			gy := y + 10 // space for label

			addPuzzlePage(pdf, slots[idx].puzzle, slots[idx].num, solution, gx, gy, gridSize)
			idx++
		}
	}
}

func gridDims(perPage int) (cols, rows int) {
	switch perPage {
	case 1:
		return 1, 1
	case 2:
		return 1, 2
	case 4:
		return 2, 2
	case 6:
		return 2, 3
	case 9:
		return 3, 3
	default:
		return 1, 1
	}
}

// addPuzzlePage draws one puzzle (or solution) grid starting at (x,y) with given width.
func addPuzzlePage(pdf *gofpdf.Fpdf, p *game.Puzzle, num int, solution bool, x, y, width float64) {
	label := "Puzzle"
	if solution {
		label = "Solution"
	}
	heading := fmt.Sprintf("%s #%d – %s %s", label, num, p.Type, p.Difficulty)

	// Small heading above the grid.
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetXY(x, y)
	pdf.CellFormat(width, 6, heading, "", 1, "C", false, 0, "")
	y += 6

	switch p.Type {
	case game.Samurai:
		drawSamuraiGrid(pdf, p, x, y, width, solution)
	default:
		drawNineGrid(pdf, p, x, y, width, solution)
	}
}

// ---- grid drawers ----------------------------------------------------------

// drawNineGrid draws a 9×9 sudoku grid at (x, y) with given width in mm.
func drawNineGrid(pdf *gofpdf.Fpdf, p *game.Puzzle, x, y, width float64, solution bool) {
	cell := width / 9
	boxW := cell * 3

	// Killer cage backgrounds.
	if p.Type == game.Killer {
		for r := range 9 {
			for c := range 9 {
				cageID := p.Cells[r][c].CageID
				if cageID >= 0 && cageID < len(p.Cages) {
					col := p.Cages[cageID].Color
					pdf.SetFillColor(int(col.R), int(col.G), int(col.B))
					pdf.Rect(x+float64(c)*cell, y+float64(r)*cell, cell, cell, "F")
				}
			}
		}
	}

	// Thin cell lines.
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.2)
	for i := 0; i <= 9; i++ {
		pdf.Line(x+float64(i)*cell, y, x+float64(i)*cell, y+width)
		pdf.Line(x, y+float64(i)*cell, x+width, y+float64(i)*cell)
	}
	// Thick box lines.
	pdf.SetDrawColor(40, 40, 40)
	pdf.SetLineWidth(0.7)
	for i := 0; i <= 3; i++ {
		pdf.Line(x+float64(i)*boxW, y, x+float64(i)*boxW, y+width)
		pdf.Line(x, y+float64(i)*boxW, x+width, y+float64(i)*boxW)
	}

	// Numbers. cell is in mm; fontSize is in pt (1pt≈0.353mm), so multiply by ~2.8
	// to get text that fills roughly half the cell height.
	fontSize := cell * 1.8
	if fontSize < 8 {
		fontSize = 8
	}
	for r := range 9 {
		for c := range 9 {
			if !p.Active[r][c] {
				continue
			}
			cx := x + float64(c)*cell
			cy := y + float64(r)*cell

			var val int
			var isGiven bool
			if solution {
				val = p.Solution[r][c]
				isGiven = true
			} else {
				val = p.Cells[r][c].Value
				isGiven = p.Cells[r][c].State == game.StateGiven
			}

			if val != 0 {
				if isGiven {
					pdf.SetFont("Helvetica", "B", fontSize)
					pdf.SetTextColor(0, 0, 0)
				} else {
					pdf.SetFont("Helvetica", "", fontSize)
					pdf.SetTextColor(30, 80, 180)
				}
				pdf.SetXY(cx, cy+(cell-fontSize*0.35)/2)
				pdf.CellFormat(cell, fontSize*0.4, fmt.Sprintf("%d", val), "", 0, "C", false, 0, "")
			}

			// Killer cage sum.
			if p.Type == game.Killer && !solution {
				cageID := p.Cells[r][c].CageID
				if cageID >= 0 && cageID < len(p.Cages) {
					cage := p.Cages[cageID]
					if len(cage.Cells) > 0 && cage.Cells[0][0] == r && cage.Cells[0][1] == c {
						sumFontSize := fontSize * 0.4
						if sumFontSize < 5 {
							sumFontSize = 5
						}
						pdf.SetFont("Helvetica", "", sumFontSize)
						pdf.SetTextColor(80, 80, 80)
						pdf.SetXY(cx+0.5, cy+0.5)
						pdf.CellFormat(cell/2, sumFontSize*0.4, fmt.Sprintf("%d", cage.Sum), "", 0, "L", false, 0, "")
					}
				}
			}
		}
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)

	if p.Type == game.Killer && !solution {
		pdf.SetFont("Helvetica", "I", 6)
		pdf.SetXY(x, y+width+0.5)
		pdf.CellFormat(width, 4, "Corner numbers = cage sum", "", 0, "L", false, 0, "")
	}
}

// drawSamuraiGrid draws the 5-subgrid samurai layout scaled to fit width×width.
func drawSamuraiGrid(pdf *gofpdf.Fpdf, p *game.Puzzle, x, y, width float64, solution bool) {
	// Samurai is 21 cells wide; scale so full grid = width.
	cellMM := width / 21.0
	subW := 9 * cellMM

	for _, sg := range game.SamuraiSubgrids {
		ro, co := sg[0], sg[1]
		subX := x + float64(co)*cellMM
		subY := y + float64(ro)*cellMM

		// Thin grid.
		pdf.SetDrawColor(180, 180, 180)
		pdf.SetLineWidth(0.2)
		for i := 0; i <= 9; i++ {
			pdf.Line(subX+float64(i)*cellMM, subY, subX+float64(i)*cellMM, subY+subW)
			pdf.Line(subX, subY+float64(i)*cellMM, subX+subW, subY+float64(i)*cellMM)
		}
		// Thick box lines.
		pdf.SetDrawColor(40, 40, 40)
		pdf.SetLineWidth(0.5)
		for i := 0; i <= 3; i++ {
			bw := 3 * cellMM
			pdf.Line(subX+float64(i)*bw, subY, subX+float64(i)*bw, subY+subW)
			pdf.Line(subX, subY+float64(i)*bw, subX+subW, subY+float64(i)*bw)
		}
		// Numbers.
		fontSize := cellMM * 1.8
		if fontSize < 6 {
			fontSize = 6
		}
		for r := range 9 {
			for c := range 9 {
				gr, gc := ro+r, co+c
				if !p.Active[gr][gc] {
					continue
				}
				var val int
				var bold bool
				if solution {
					val = p.Solution[gr][gc]
					bold = true
				} else {
					val = p.Cells[gr][gc].Value
					bold = p.Cells[gr][gc].State == game.StateGiven
				}
				if val != 0 {
					if bold {
						pdf.SetFont("Helvetica", "B", fontSize)
					} else {
						pdf.SetFont("Helvetica", "", fontSize)
					}
					cx := subX + float64(c)*cellMM
					cy := subY + float64(r)*cellMM
					pdf.SetXY(cx, cy+(cellMM-fontSize*0.35)/2)
					pdf.CellFormat(cellMM, fontSize*0.4, fmt.Sprintf("%d", val), "", 0, "C", false, 0, "")
				}
			}
		}
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
}

func addHistoryPage(pdf *gofpdf.Fpdf, history *hist.History) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Ln(10)
	pdf.CellFormat(0, 10, "Game History", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	stats := history.ComputeStats()
	pdf.SetFont("Helvetica", "", 11)
	for _, l := range []string{
		fmt.Sprintf("Games Played: %d", stats.GamesPlayed),
		fmt.Sprintf("Won: %d  |  Gave Up: %d  |  Total Hints: %d", stats.GamesWon, stats.GaveUp, stats.TotalHints),
		fmt.Sprintf("Best Time: %s  |  Avg Win Time: %s", hist.FormatDuration(stats.MinTimeSec), hist.FormatDuration(stats.AvgTimeSec)),
	} {
		pdf.CellFormat(0, 7, l, "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 9)
	colW := []float64{28, 26, 24, 20, 16, 30, 16}
	headers := []string{"Type", "Difficulty", "Result", "Time", "Hints", "Date", "Errors"}
	for i, h := range headers {
		pdf.CellFormat(colW[i], 6, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, r := range history.Results {
		result := "Abandoned"
		if r.Won {
			result = "Won"
		} else if r.GaveUp {
			result = "Gave Up"
		}
		row := []string{
			r.GameType, r.Difficulty, result,
			hist.FormatDuration(r.DurationS),
			fmt.Sprintf("%d", r.HintsUsed),
			r.StartTime.Format("2006-01-02"),
			fmt.Sprintf("%d", r.ErrorsMade),
		}
		for i, cell := range row {
			pdf.CellFormat(colW[i], 5, cell, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
		if pdf.GetY() > 280 {
			pdf.AddPage()
			pdf.SetFont("Helvetica", "", 8)
		}
	}
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
