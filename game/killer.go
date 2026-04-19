package game

import (
	"math/rand"
)

// GenerateKiller creates a killer sudoku puzzle.
// Killer sudoku has no given numbers; instead, cells are grouped into
// "cages" and each cage shows the sum of its cells.
func GenerateKiller(diff Difficulty) *Puzzle {
	p := NewPuzzle(Killer, diff)
	solution := generateSolvedGrid(9, 3)

	for r := range 9 {
		for c := range 9 {
			p.Solution[r][c] = solution[r][c]
			p.Active[r][c] = true
		}
	}

	p.Cages = buildCages(solution, 9)

	// Killer typically shows no givens; remove all, rely only on cages.
	// For easier difficulties we reveal some cells.
	switch diff {
	case Easy:
		// Reveal some cells as givens.
		revealKillerCells(p, solution, 16)
	case Medium:
		revealKillerCells(p, solution, 8)
	default:
		// Hard/Expert: no givens
	}
	return p
}

// buildCages partitions the 9×9 grid into killer cages of size 2-5.
// Each cage is guaranteed to contain unique digits (killer sudoku rule).
func buildCages(solution [][]int, size int) []Cage {
	assigned := make([][]int, size)
	for i := range assigned {
		assigned[i] = make([]int, size)
		for j := range assigned[i] {
			assigned[i][j] = -1
		}
	}

	var cages []Cage
	cageID := 0
	colorIdx := 0

	positions := shuffledPositions(size)
	for _, pos := range positions {
		r, c := pos[0], pos[1]
		if assigned[r][c] != -1 {
			continue
		}
		maxSize := 2 + rand.Intn(4) // 2-5
		cage := growCage(r, c, maxSize, size, assigned, solution)
		if len(cage) == 0 {
			cage = [][2]int{{r, c}}
		}
		sum := 0
		for _, cell := range cage {
			sum += solution[cell[0]][cell[1]]
			assigned[cell[0]][cell[1]] = cageID
		}
		cages = append(cages, Cage{
			ID:    cageID,
			Sum:   sum,
			Cells: cage,
			Color: CageColors[colorIdx%len(CageColors)],
		})
		cageID++
		colorIdx++
	}
	return cages
}

// growCage does a BFS-like expansion from (startR, startC).
// Only neighbours whose solution value is not already in the cage are eligible,
// enforcing the killer sudoku rule that every cage contains unique digits.
func growCage(startR, startC, maxSize, size int, assigned [][]int, solution [][]int) [][2]int {
	cage := [][2]int{{startR, startC}}
	assigned[startR][startC] = -2 // temp mark
	used := [10]bool{}
	used[solution[startR][startC]] = true

	dirs := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	seen := map[[2]int]bool{}
	for len(cage) < maxSize {
		var candidates [][2]int
		for _, cell := range cage {
			for _, d := range dirs {
				nr, nc := cell[0]+d[0], cell[1]+d[1]
				key := [2]int{nr, nc}
				if nr < 0 || nr >= size || nc < 0 || nc >= size {
					continue
				}
				if assigned[nr][nc] != -1 || seen[key] {
					continue
				}
				// Reject if the candidate's value already appears in this cage.
				if used[solution[nr][nc]] {
					continue
				}
				seen[key] = true
				candidates = append(candidates, key)
			}
		}
		if len(candidates) == 0 {
			break
		}
		pick := candidates[rand.Intn(len(candidates))]
		cage = append(cage, pick)
		assigned[pick[0]][pick[1]] = -2
		used[solution[pick[0]][pick[1]]] = true
	}
	// Revert temp marks
	for _, cell := range cage {
		assigned[cell[0]][cell[1]] = -1
	}
	return cage
}

func shuffledPositions(size int) [][2]int {
	positions := make([][2]int, 0, size*size)
	for r := range size {
		for c := range size {
			positions = append(positions, [2]int{r, c})
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	return positions
}

func revealKillerCells(p *Puzzle, solution [][]int, count int) {
	positions := shuffledPositions(9)
	revealed := 0
	for _, pos := range positions {
		if revealed >= count {
			break
		}
		r, c := pos[0], pos[1]
		p.Cells[r][c] = Cell{
			Value:  solution[r][c],
			State:  StateGiven,
			CageID: p.Cells[r][c].CageID,
		}
		revealed++
	}
}

// AssignCageIDs copies cage IDs back onto puzzle cells.
func (p *Puzzle) AssignCageIDs() {
	for id, cage := range p.Cages {
		for _, cell := range cage.Cells {
			p.Cells[cell[0]][cell[1]].CageID = id
		}
	}
}

// CageOf returns the cage that contains cell (r,c), or nil.
func (p *Puzzle) CageOf(r, c int) *Cage {
	id := p.Cells[r][c].CageID
	if id < 0 || id >= len(p.Cages) {
		return nil
	}
	return &p.Cages[id]
}

// KillerCageError returns true if the cage containing (r,c) has a sum error.
// A cage is "in error" when all cells are filled but the sum doesn't match.
func (p *Puzzle) KillerCageError(cageID int) bool {
	if cageID < 0 || cageID >= len(p.Cages) {
		return false
	}
	cage := p.Cages[cageID]
	sum := 0
	for _, cell := range cage.Cells {
		v := p.Cells[cell[0]][cell[1]].Value
		if v == 0 {
			return false // not fully filled yet
		}
		sum += v
	}
	return sum != cage.Sum
}
