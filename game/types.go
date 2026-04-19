package game

import (
	"image/color"
	"time"
)

type GameType int

const (
	Normal  GameType = iota
	Killer
	Samurai
)

func (g GameType) String() string {
	switch g {
	case Normal:
		return "Normal"
	case Killer:
		return "Killer"
	case Samurai:
		return "Samurai"
	default:
		return "Unknown"
	}
}

type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
	Expert
)

func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	case Expert:
		return "Expert"
	default:
		return "Unknown"
	}
}

// GivensCount is the number of filled cells shown to the player per difficulty.
// These are tuned to work together with the technique-based grader: fewer givens
// generally forces harder solving paths.
var GivensCount = map[Difficulty]int{
	Easy:   38,
	Medium: 30,
	Hard:   25,
	Expert: 22,
}

// targetGrade maps each difficulty to the grader grade it must reach.
var targetGrade = map[Difficulty]int{
	Easy:   gradeEasy,
	Medium: gradeMedium,
	Hard:   gradeHard,
	Expert: gradeExpert,
}

type CellState int

const (
	StateEmpty   CellState = iota
	StateGiven             // fixed clue
	StateFilled            // player filled, correct
	StateInvalid           // player filled, wrong
)

type Cell struct {
	Value  int
	State  CellState
	Notes  [10]bool // pencil marks, index 1-9
	CageID int      // killer sudoku cage id (-1 = none)
}

type Cage struct {
	ID    int
	Sum   int
	Cells [][2]int // [row, col]
	Color color.RGBA
}

type Move struct {
	Row, Col int
	OldValue int
	NewValue int
	OldNotes [10]bool
	OldState CellState
}

// Puzzle is the full state of an active game.
type Puzzle struct {
	Type       GameType
	Difficulty Difficulty
	// Grid dimensions
	GridSize int // 9 for Normal/Killer, 21 for Samurai
	BoxSize  int // always 3

	Cells    [][]Cell
	Solution [][]int
	Active   [][]bool // which cells are part of the puzzle

	Cages    []Cage   // killer only
	Subgrids [][2]int // samurai: [rowOffset, colOffset] for each 9x9

	StartTime  time.Time
	HintsUsed  int
	ErrorsMade int // total wrong-value placements (not undone)
	MoveStack  []Move

	// Limits set by the player before the game starts (0 = unlimited).
	MaxHints  int
	MaxErrors int
}

var SamuraiSubgrids = [][2]int{
	{0, 0}, {0, 12}, {6, 6}, {12, 0}, {12, 12},
}

func NewPuzzle(gt GameType, diff Difficulty) *Puzzle {
	p := &Puzzle{
		Type:       gt,
		Difficulty: diff,
		BoxSize:    3,
		StartTime:  time.Now(),
	}
	switch gt {
	case Normal, Killer:
		p.GridSize = 9
	case Samurai:
		p.GridSize = 21
		p.Subgrids = SamuraiSubgrids
	}
	sz := p.GridSize
	p.Cells = make([][]Cell, sz)
	p.Solution = make([][]int, sz)
	p.Active = make([][]bool, sz)
	for i := range p.Cells {
		p.Cells[i] = make([]Cell, sz)
		p.Solution[i] = make([]int, sz)
		p.Active[i] = make([]bool, sz)
		for j := range p.Cells[i] {
			p.Cells[i][j].CageID = -1
		}
	}
	return p
}

// SetCell applies a player move (value 0 = clear).
// Returns true if this placement was a new error (wrong value in a non-error cell).
func (p *Puzzle) SetCell(row, col, val int) bool {
	c := &p.Cells[row][col]
	if c.State == StateGiven {
		return false
	}
	move := Move{
		Row:      row,
		Col:      col,
		OldValue: c.Value,
		NewValue: val,
		OldNotes: c.Notes,
		OldState: c.State,
	}
	p.MoveStack = append(p.MoveStack, move)
	c.Value = val
	c.Notes = [10]bool{}
	newError := false
	switch {
	case val == 0:
		c.State = StateEmpty
	case val == p.Solution[row][col]:
		c.State = StateFilled
	default:
		// Only count as a new error if this cell wasn't already wrong.
		if move.OldState != StateInvalid {
			p.ErrorsMade++
			newError = true
		}
		c.State = StateInvalid
	}
	// Remove this value from notes of all peers in the same box, row, and column.
	if val != 0 {
		boxR := (row / p.BoxSize) * p.BoxSize
		boxC := (col / p.BoxSize) * p.BoxSize
		for r := 0; r < p.GridSize; r++ {
			for cc := 0; cc < p.GridSize; cc++ {
				if r == row && cc == col {
					continue
				}
				if !p.Active[r][cc] {
					continue
				}
				if r == row || cc == col || (r >= boxR && r < boxR+p.BoxSize && cc >= boxC && cc < boxC+p.BoxSize) {
					p.Cells[r][cc].Notes[val] = false
				}
			}
		}
	}
	return newError
}

// ValueComplete returns true when digit n has been correctly placed in all cells.
func (p *Puzzle) ValueComplete(n int) bool {
	for r := 0; r < p.GridSize; r++ {
		for c := 0; c < p.GridSize; c++ {
			if !p.Active[r][c] {
				continue
			}
			if p.Solution[r][c] == n && p.Cells[r][c].Value != n {
				return false
			}
		}
	}
	return true
}

func (p *Puzzle) ToggleNote(row, col, note int) {
	c := &p.Cells[row][col]
	if c.State == StateGiven || c.Value != 0 {
		return
	}
	move := Move{
		Row:      row,
		Col:      col,
		OldValue: c.Value,
		NewValue: c.Value,
		OldNotes: c.Notes,
		OldState: c.State,
	}
	p.MoveStack = append(p.MoveStack, move)
	c.Notes[note] = !c.Notes[note]
}

func (p *Puzzle) Undo() bool {
	if len(p.MoveStack) == 0 {
		return false
	}
	m := p.MoveStack[len(p.MoveStack)-1]
	p.MoveStack = p.MoveStack[:len(p.MoveStack)-1]
	c := &p.Cells[m.Row][m.Col]
	c.Value = m.OldValue
	c.Notes = m.OldNotes
	c.State = m.OldState
	return true
}

func (p *Puzzle) IsComplete() bool {
	for r := 0; r < p.GridSize; r++ {
		for c := 0; c < p.GridSize; c++ {
			if !p.Active[r][c] {
				continue
			}
			if p.Cells[r][c].Value != p.Solution[r][c] {
				return false
			}
		}
	}
	return true
}

func (p *Puzzle) ElapsedSeconds() int {
	return int(time.Since(p.StartTime).Seconds())
}

// CageColors is the palette used for killer sudoku cages.
var CageColors = []color.RGBA{
	{R: 255, G: 220, B: 220, A: 255},
	{R: 220, G: 255, B: 220, A: 255},
	{R: 220, G: 220, B: 255, A: 255},
	{R: 255, G: 255, B: 200, A: 255},
	{R: 255, G: 225, B: 185, A: 255},
	{R: 200, G: 240, B: 255, A: 255},
	{R: 255, G: 215, B: 255, A: 255},
	{R: 210, G: 255, B: 240, A: 255},
	{R: 255, G: 240, B: 215, A: 255},
	{R: 225, G: 215, B: 255, A: 255},
}
