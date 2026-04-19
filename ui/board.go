package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/Kellerman81/go_sudoku/game"
)

// Cell drawing sizes in Dp units.
const (
	cellDpNormal  unit.Dp = 52
	cellDpSamurai unit.Dp = 26 // smaller so 21-cell grid fits on screen
	boxGapDp      unit.Dp = 3
	boardMarginDp unit.Dp = 10
)

// Colours for the board.
var (
	bColBg        = nrgba(250, 250, 250, 255)
	bColGiven     = nrgba(25, 25, 25, 255)
	bColFilled    = nrgba(30, 80, 180, 255)
	bColInvalid   = nrgba(200, 30, 30, 255)
	bColNote      = nrgba(110, 110, 110, 255)
	bColSelected  = nrgba(170, 210, 255, 255)
	bColHighlight = nrgba(210, 230, 255, 255)
	bColSameVal   = nrgba(195, 215, 255, 255)
	bColInactive  = nrgba(190, 190, 190, 255)
	bColGridThick = nrgba(50, 50, 50, 255)
	bColHint      = nrgba(100, 220, 100, 255)
)

// BoardState holds the interactive state of the Sudoku board.
type BoardState struct {
	SelRow, SelCol int // -1 = nothing selected
	NoteMode       bool
	HintRow        int
	HintCol        int
	HintExpiry     time.Time
}

func NewBoardState() BoardState {
	return BoardState{SelRow: -1, SelCol: -1, HintRow: -1, HintCol: -1}
}

// Update reads input events and updates state.
// Returns (number, true) when the user enters a digit that should be applied.
func (b *BoardState) Update(gtx layout.Context, p *game.Puzzle) (inputVal int, inputOK bool) {
	cellPx := gtx.Dp(cellDpFor(p))
	gapPx := gtx.Dp(boxGapDp)
	marginPx := gtx.Dp(boardMarginDp)

	// Pointer events
	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: b,
			Kinds:  pointer.Press,
		})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok || pe.Kind != pointer.Press {
			continue
		}
		gtx.Execute(key.FocusCmd{Tag: b}) // request focus immediately on click
		col := pixelToCell(int(pe.Position.X), cellPx, gapPx, marginPx, p.BoxSize, p.GridSize)
		row := pixelToCell(int(pe.Position.Y), cellPx, gapPx, marginPx, p.BoxSize, p.GridSize)
		if row >= 0 && col >= 0 && p.Active[row][col] {
			b.SelRow, b.SelCol = row, col
		} else if row < 0 || col < 0 {
			b.SelRow, b.SelCol = -1, -1
		}
	}

	// Keyboard events — handles key.Event (arrows, delete) and key.EditEvent
	// (WM_CHAR on Windows: regular digits, numpad digits, letters like N).
	for {
		ev, ok := gtx.Event(
			key.FocusFilter{Target: b},
			key.Filter{Focus: b, Name: key.NameDeleteBackward},
			key.Filter{Focus: b, Name: key.NameUpArrow},
			key.Filter{Focus: b, Name: key.NameDownArrow},
			key.Filter{Focus: b, Name: key.NameLeftArrow},
			key.Filter{Focus: b, Name: key.NameRightArrow},
		)
		if !ok {
			break
		}
		switch ev := ev.(type) {
		case key.EditEvent:
			if len(ev.Text) != 1 {
				continue
			}
			ch := ev.Text[0]
			if ch == 'n' || ch == 'N' {
				b.NoteMode = !b.NoteMode
			} else if ch >= '1' && ch <= '9' && b.SelRow >= 0 && b.SelCol >= 0 {
				return int(ch - '0'), true
			} else if ch == '0' && b.SelRow >= 0 && b.SelCol >= 0 {
				return 0, true
			}
		case key.Event:
			if ev.State != key.Press {
				continue
			}
			switch ev.Name {
			case key.NameDeleteBackward:
				if b.SelRow >= 0 && b.SelCol >= 0 {
					return 0, true
				}
			case key.NameUpArrow:
				if b.SelRow > 0 {
					b.SelRow--
					for b.SelRow > 0 && p.Active != nil && !p.Active[b.SelRow][b.SelCol] {
						b.SelRow--
					}
				}
			case key.NameDownArrow:
				if b.SelRow < p.GridSize-1 {
					b.SelRow++
					for b.SelRow < p.GridSize-1 && p.Active != nil && !p.Active[b.SelRow][b.SelCol] {
						b.SelRow++
					}
				}
			case key.NameLeftArrow:
				if b.SelCol > 0 {
					b.SelCol--
					for b.SelCol > 0 && p.Active != nil && !p.Active[b.SelRow][b.SelCol] {
						b.SelCol--
					}
				}
			case key.NameRightArrow:
				if b.SelCol < p.GridSize-1 {
					b.SelCol++
					for b.SelCol < p.GridSize-1 && p.Active != nil && !p.Active[b.SelRow][b.SelCol] {
						b.SelCol++
					}
				}
			}
		}
	}
	return 0, false
}

func cellDpFor(p *game.Puzzle) unit.Dp {
	if p.Type == game.Samurai {
		return cellDpSamurai
	}
	return cellDpNormal
}

// Layout draws the sudoku board.
func (b *BoardState) Layout(gtx layout.Context, th *material.Theme, p *game.Puzzle) layout.Dimensions {
	cellPx := gtx.Dp(cellDpFor(p))
	gapPx := gtx.Dp(boxGapDp)
	marginPx := gtx.Dp(boardMarginDp)
	sz := p.GridSize
	boxes := sz / p.BoxSize

	totalSz := marginPx*2 + sz*cellPx + (boxes-1)*gapPx

	// Register pointer area for the whole board.
	boardStack := clip.Rect(image.Rect(0, 0, totalSz, totalSz)).Push(gtx.Ops)
	event.Op(gtx.Ops, b)
	boardStack.Pop()

	// Board background.
	fillRect(gtx.Ops, image.Rect(0, 0, totalSz, totalSz), bColBg)

	hintActive := b.HintRow >= 0 && time.Now().Before(b.HintExpiry)

	// Draw cells.
	for row := range sz {
		for col := range sz {
			x, y := cellOrigin(col, row, cellPx, gapPx, marginPx, p.BoxSize)
			cr := image.Rect(x, y, x+cellPx, y+cellPx)

			if !p.Active[row][col] {
				fillRect(gtx.Ops, cr, bColInactive)
				continue
			}

			cell := p.Cells[row][col]
			bg := b.cellBg(row, col, cell, p, hintActive)
			fillRect(gtx.Ops, cr, bg)

			// Killer mode: draw a vivid orange border ring inside the selected cell
			// instead of overriding the cage background colour. Orange contrasts
			// clearly against both the dark cage borders and the pastel cage fills.
			if p.Type == game.Killer && row == b.SelRow && col == b.SelCol {
				const sw = 4
				sc := nrgba(230, 100, 0, 255)
				fillRect(gtx.Ops, image.Rect(cr.Min.X+2, cr.Min.Y+2, cr.Max.X-2, cr.Min.Y+2+sw), sc)
				fillRect(gtx.Ops, image.Rect(cr.Min.X+2, cr.Max.Y-2-sw, cr.Max.X-2, cr.Max.Y-2), sc)
				fillRect(gtx.Ops, image.Rect(cr.Min.X+2, cr.Min.Y+2, cr.Min.X+2+sw, cr.Max.Y-2), sc)
				fillRect(gtx.Ops, image.Rect(cr.Max.X-2-sw, cr.Min.Y+2, cr.Max.X-2, cr.Max.Y-2), sc)
			}

			if cell.Value != 0 {
				textCol := cellTextColor(cell.State)
				bold := cell.State == game.StateGiven
				drawCenteredText(gtx, th, fmt.Sprintf("%d", cell.Value), textCol, bold, unit.Sp(22), x, y, cellPx, cellPx)
			} else {
				// Pencil marks (3×3 grid inside the cell).
				for n := 1; n <= 9; n++ {
					if cell.Notes[n] {
						nr, nc := (n-1)/3, (n-1)%3
						nx := x + nc*(cellPx/3)
						ny := y + nr*(cellPx/3)
						drawCenteredText(gtx, th, fmt.Sprintf("%d", n), bColNote, false, unit.Sp(10), nx, ny, cellPx/3, cellPx/3)
					}
				}
			}

			// Killer cage sum label in top-left corner of the cage's first cell.
			if p.Type == game.Killer && cell.CageID >= 0 {
				cage := &p.Cages[cell.CageID]
				if len(cage.Cells) > 0 && cage.Cells[0][0] == row && cage.Cells[0][1] == col {
					drawSmallCornerText(gtx, th, fmt.Sprintf("%d", cage.Sum),
						nrgba(60, 60, 60, 255), x+5, y+4, cellPx/2, cellPx/4)
				}
			}
		}
	}

	// Draw grid lines.
	for i := 0; i <= sz; i++ {
		isThick := i%p.BoxSize == 0
		lw := 1
		if isThick {
			lw = 3
		}
		lc := nrgba(170, 170, 170, 255)
		if isThick {
			lc = bColGridThick
		}

		hy := marginPx + i*cellPx + (i/p.BoxSize)*gapPx
		if i == sz {
			hy = totalSz - marginPx
		}
		fillRect(gtx.Ops, image.Rect(marginPx, hy-lw/2, totalSz-marginPx, hy-lw/2+lw), lc)

		vx := marginPx + i*cellPx + (i/p.BoxSize)*gapPx
		if i == sz {
			vx = totalSz - marginPx
		}
		fillRect(gtx.Ops, image.Rect(vx-lw/2, marginPx, vx-lw/2+lw, totalSz-marginPx), lc)
	}

	// Killer: draw cage borders inside cells on top of grid lines.
	if p.Type == game.Killer {
		drawKillerCageBorders(gtx, p, cellPx, gapPx, marginPx)
	}

	return layout.Dimensions{Size: image.Pt(totalSz, totalSz)}
}

// drawKillerCageBorders draws 2-px inset borders along cage boundaries so cage
// regions remain visible regardless of selection state.
func drawKillerCageBorders(gtx layout.Context, p *game.Puzzle, cellPx, gapPx, marginPx int) {
	const inset = 1
	const bw = 2
	lc := nrgba(40, 40, 40, 230)
	sz := p.GridSize

	cageID := func(r, c int) int {
		if r < 0 || r >= sz || c < 0 || c >= sz || !p.Active[r][c] {
			return -2 // distinct sentinel
		}
		return p.Cells[r][c].CageID
	}

	for row := 0; row < sz; row++ {
		for col := 0; col < sz; col++ {
			if !p.Active[row][col] {
				continue
			}
			x, y := cellOrigin(col, row, cellPx, gapPx, marginPx, p.BoxSize)
			cid := cageID(row, col)
			if cageID(row-1, col) != cid {
				fillRect(gtx.Ops, image.Rect(x+inset, y+inset, x+cellPx-inset, y+inset+bw), lc)
			}
			if cageID(row+1, col) != cid {
				fillRect(gtx.Ops, image.Rect(x+inset, y+cellPx-inset-bw, x+cellPx-inset, y+cellPx-inset), lc)
			}
			if cageID(row, col-1) != cid {
				fillRect(gtx.Ops, image.Rect(x+inset, y+inset, x+inset+bw, y+cellPx-inset), lc)
			}
			if cageID(row, col+1) != cid {
				fillRect(gtx.Ops, image.Rect(x+cellPx-inset-bw, y+inset, x+cellPx-inset, y+cellPx-inset), lc)
			}
		}
	}
}

func (b *BoardState) cellBg(row, col int, cell game.Cell, p *game.Puzzle, hintActive bool) color.NRGBA {
	if hintActive && row == b.HintRow && col == b.HintCol {
		return bColHint
	}
	isSelected := row == b.SelRow && col == b.SelCol

	if p.Type == game.Killer && cell.CageID >= 0 {
		// In killer mode always show the cage colour unchanged.
		// Selection is indicated by a border ring drawn on top (see Layout).
		return nrgbaFrom(p.Cages[cell.CageID].Color)
	}

	if isSelected {
		return bColSelected
	}
	if b.SelRow >= 0 && !isSelected &&
		(row == b.SelRow || col == b.SelCol || sameBox(row, col, b.SelRow, b.SelCol, p.BoxSize)) {
		return bColHighlight
	}
	if b.SelRow >= 0 {
		sv := p.Cells[b.SelRow][b.SelCol].Value
		if sv != 0 && cell.Value == sv {
			return bColSameVal
		}
	}
	return nrgba(255, 255, 255, 255)
}

// ---- helpers ---------------------------------------------------------------

func cellOrigin(col, row, cellPx, gapPx, marginPx, boxSize int) (int, int) {
	x := marginPx + col*cellPx + (col/boxSize)*gapPx
	y := marginPx + row*cellPx + (row/boxSize)*gapPx
	return x, y
}

func pixelToCell(px, cellPx, gapPx, marginPx, boxSize, gridSize int) int {
	px -= marginPx
	if px < 0 {
		return -1
	}
	// Account for the gap added before each box (col/boxSize * gapPx).
	for cell := 0; cell < gridSize; cell++ {
		start := cell*cellPx + (cell/boxSize)*gapPx
		if px >= start && px < start+cellPx {
			return cell
		}
	}
	return -1
}

func fillRect(ops *op.Ops, r image.Rectangle, col color.NRGBA) {
	defer clip.Rect(r).Push(ops).Pop()
	paint.Fill(ops, col)
}

func drawCenteredText(gtx layout.Context, th *material.Theme, text string, col color.NRGBA, bold bool, sz unit.Sp, x, y, w, h int) {
	defer op.Offset(image.Pt(x, y)).Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rect(0, 0, w, h)).Push(gtx.Ops).Pop()
	childGtx := gtx
	childGtx.Constraints = layout.Exact(image.Pt(w, h))
	layout.Center.Layout(childGtx, func(gtx layout.Context) layout.Dimensions {
		l := material.Label(th, sz, text)
		l.Color = col
		if bold {
			l.Font.Weight = font.Bold
		}
		return l.Layout(gtx)
	})
}

func drawSmallCornerText(gtx layout.Context, th *material.Theme, text string, col color.NRGBA, x, y, w, h int) {
	defer op.Offset(image.Pt(x, y)).Push(gtx.Ops).Pop()
	defer clip.Rect(image.Rect(0, 0, w, h)).Push(gtx.Ops).Pop()
	childGtx := gtx
	childGtx.Constraints = layout.Exact(image.Pt(w, h))
	l := material.Label(th, unit.Sp(9), text)
	l.Color = col
	l.Layout(childGtx)
}

func cellTextColor(state game.CellState) color.NRGBA {
	switch state {
	case game.StateGiven:
		return bColGiven
	case game.StateFilled:
		return bColFilled
	case game.StateInvalid:
		return bColInvalid
	default:
		return bColFilled
	}
}

func sameBox(r1, c1, r2, c2, boxSize int) bool {
	return r1/boxSize == r2/boxSize && c1/boxSize == c2/boxSize
}

func nrgba(r, g, b, a uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func nrgbaFrom(c color.RGBA) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func blendNRGBA(a, b color.NRGBA) color.NRGBA {
	return color.NRGBA{
		R: (a.R + b.R) / 2,
		G: (a.G + b.G) / 2,
		B: (a.B + b.B) / 2,
		A: 255,
	}
}
