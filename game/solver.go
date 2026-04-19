package game

// HintAllowed reports whether another hint can be used.
func HintAllowed(p *Puzzle) bool {
	return p.MaxHints <= 0 || p.HintsUsed < p.MaxHints
}

// Hint reveals one incorrect or empty cell from the solution.
// Returns (row, col, value) of the hinted cell, or (-1,-1,0) if no hint available.
func Hint(p *Puzzle) (int, int, int) {
	if !HintAllowed(p) {
		return -1, -1, 0
	}
	// Collect all unfilled or wrong active cells in random order.
	type candidate struct{ r, c int }
	var candidates []candidate
	for r := 0; r < p.GridSize; r++ {
		for c := 0; c < p.GridSize; c++ {
			if !p.Active[r][c] {
				continue
			}
			cell := p.Cells[r][c]
			if cell.State == StateGiven {
				continue
			}
			if cell.Value != p.Solution[r][c] {
				candidates = append(candidates, candidate{r, c})
			}
		}
	}
	if len(candidates) == 0 {
		return -1, -1, 0
	}
	// Pick the first candidate (we could randomise but deterministic is fine).
	pick := candidates[0]
	p.HintsUsed++
	// Apply as a given so the player can't erase it.
	p.Cells[pick.r][pick.c] = Cell{
		Value:  p.Solution[pick.r][pick.c],
		State:  StateGiven,
		CageID: p.Cells[pick.r][pick.c].CageID,
	}
	return pick.r, pick.c, p.Solution[pick.r][pick.c]
}

// Validate marks each filled cell as StateFilled or StateInvalid.
func Validate(p *Puzzle) {
	for r := 0; r < p.GridSize; r++ {
		for c := 0; c < p.GridSize; c++ {
			cell := &p.Cells[r][c]
			if !p.Active[r][c] || cell.State == StateGiven {
				continue
			}
			if cell.Value == 0 {
				cell.State = StateEmpty
			} else if cell.Value == p.Solution[r][c] {
				cell.State = StateFilled
			} else {
				cell.State = StateInvalid
			}
		}
	}
}

// SolveAll fills every remaining cell from the solution (used for "give up").
func SolveAll(p *Puzzle) {
	for r := 0; r < p.GridSize; r++ {
		for c := 0; c < p.GridSize; c++ {
			if !p.Active[r][c] {
				continue
			}
			if p.Cells[r][c].State != StateGiven {
				p.Cells[r][c] = Cell{
					Value:  p.Solution[r][c],
					State:  StateGiven,
					CageID: p.Cells[r][c].CageID,
				}
			}
		}
	}
}
