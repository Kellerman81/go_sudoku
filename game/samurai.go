package game

// GenerateSamurai creates a samurai sudoku: five overlapping 9×9 grids on a
// 21×21 canvas.  The overlapping 3×3 boxes are shared between grids.
//
// Subgrid positions (row-offset, col-offset):
//
//	Top-left  (0,0)   Top-right  (0,12)
//	          Center  (6,6)
//	Bot-left (12,0)   Bot-right (12,12)
//
// Overlap boxes live at global positions:
//
//	TL∩Center  rows 6-8,  cols 6-8
//	TR∩Center  rows 6-8,  cols 12-14
//	BL∩Center  rows 12-14, cols 6-8
//	BR∩Center  rows 12-14, cols 12-14
func GenerateSamurai(diff Difficulty) *Puzzle {
	p := NewPuzzle(Samurai, diff)

	// We use a 21×21 int grid to hold the full solution.
	fullSol := make([][]int, 21)
	for i := range fullSol {
		fullSol[i] = make([]int, 21)
	}

	// Generate the center grid first (offset 6,6).
	center := generateSolvedGrid(9, 3)
	placeSubgrid(fullSol, center, 6, 6)

	// Generate each corner subgrid, seeding the shared box from the center.
	corners := [][2]int{{0, 0}, {0, 12}, {12, 0}, {12, 12}}
	// Center sub-boxes that overlap with corners (local coords in center grid):
	// TL corner shares center's top-left box  (0,0)
	// TR corner shares center's top-right box (0,6)
	// BL corner shares center's bottom-left box (6,0)
	// BR corner shares center's bottom-right box (6,6)
	centerShared := [][2]int{{0, 0}, {0, 6}, {6, 0}, {6, 6}}
	// In the corner grid the shared box is always the inner corner:
	// TL → bottom-right (6,6), TR → bottom-left (6,0),
	// BL → top-right (0,6),   BR → top-left (0,0)
	cornerShared := [][2]int{{6, 6}, {6, 0}, {0, 6}, {0, 0}}

	for i, off := range corners {
		sub := generateSolvedGridWithSeed(center, 9, 3,
			centerShared[i], cornerShared[i])
		placeSubgrid(fullSol, sub, off[0], off[1])
	}

	// Mark active cells and copy solution.
	markActive(p, fullSol)

	// Build per-subgrid removals according to difficulty.
	puzzle := copyGrid(fullSol, 21)
	givens := GivensCount[diff]
	for _, sg := range SamuraiSubgrids {
		removeInSubgrid(puzzle, fullSol, sg[0], sg[1], givens, 9, 3)
	}
	// The four 3×3 overlap zones are shared between two subgrids each.
	// Always restore them as givens so that cross-grid uniqueness is guaranteed
	// (removeInSubgrid only checks uniqueness within its own 9×9 subgrid).
	overlapZones := [][4]int{
		{6, 6, 8, 8},     // TL corner ∩ center
		{6, 12, 8, 14},   // TR corner ∩ center
		{12, 6, 14, 8},   // BL corner ∩ center
		{12, 12, 14, 14}, // BR corner ∩ center
	}
	for _, z := range overlapZones {
		for r := z[0]; r <= z[2]; r++ {
			for c := z[1]; c <= z[3]; c++ {
				puzzle[r][c] = fullSol[r][c]
			}
		}
	}

	// Populate puzzle cells.
	for r := range 21 {
		for c := range 21 {
			p.Solution[r][c] = fullSol[r][c]
			if !p.Active[r][c] {
				continue
			}
			if puzzle[r][c] != 0 {
				p.Cells[r][c] = Cell{
					Value:  puzzle[r][c],
					State:  StateGiven,
					CageID: -1,
				}
			} else {
				p.Cells[r][c].CageID = -1
			}
		}
	}
	return p
}

// generateSolvedGridWithSeed creates a solved 9×9 grid while preserving one
// already-filled 3×3 box (the overlap with the center grid).
// seedSrc is the top-left corner of the seed box in src; seedDst is the
// top-left corner in the new grid.
// Retries until solveFull succeeds (pre-filled diagonal boxes can occasionally
// conflict with the seeded box's row/column constraints).
func generateSolvedGridWithSeed(src [][]int, size, boxSize int, seedSrc, seedDst [2]int) [][]int {
	seedBoxRow := seedDst[0] / boxSize
	seedBoxCol := seedDst[1] / boxSize
	for {
		grid := make([][]int, size)
		for i := range grid {
			grid[i] = make([]int, size)
		}
		// Copy the seeded box.
		for r := range boxSize {
			for c := range boxSize {
				grid[seedDst[0]+r][seedDst[1]+c] = src[seedSrc[0]+r][seedSrc[1]+c]
			}
		}
		// Pre-fill only diagonal boxes that share neither a row-band nor a
		// column-band with the seeded box, so their random values cannot
		// conflict with the seed's row/column constraints.
		for b := 0; b < size/boxSize; b++ {
			if b == seedBoxRow || b == seedBoxCol {
				continue
			}
			fillBox(grid, b*boxSize, b*boxSize, boxSize)
		}
		if solveFull(grid, size, boxSize) {
			return grid
		}
		// solveFull failed: the randomly filled diagonal boxes conflicted with
		// the seed. Retry with a fresh random fill.
	}
}

func placeSubgrid(full [][]int, sub [][]int, rowOff, colOff int) {
	for r := range 9 {
		for c := range 9 {
			full[rowOff+r][colOff+c] = sub[r][c]
		}
	}
}

func markActive(p *Puzzle, full [][]int) {
	for _, sg := range SamuraiSubgrids {
		ro, co := sg[0], sg[1]
		for r := ro; r < ro+9; r++ {
			for c := co; c < co+9; c++ {
				p.Active[r][c] = true
				p.Solution[r][c] = full[r][c]
			}
		}
	}
}

// removeInSubgrid removes cells inside one 9×9 subgrid of the full 21×21 grid.
func removeInSubgrid(puzzle, solution [][]int, rowOff, colOff, givens, size, boxSize int) {
	// Extract the subgrid.
	sub := make([][]int, size)
	subSol := make([][]int, size)
	for r := range size {
		sub[r] = make([]int, size)
		subSol[r] = make([]int, size)
		for c := range size {
			sub[r][c] = puzzle[rowOff+r][colOff+c]
			subSol[r][c] = solution[rowOff+r][colOff+c]
		}
	}
	removeCellsUnique(sub, subSol, givens, size, boxSize)
	// Write back (do not overwrite cells already removed by a neighbouring subgrid).
	for r := range size {
		for c := range size {
			if sub[r][c] != 0 {
				puzzle[rowOff+r][colOff+c] = sub[r][c]
			} else {
				puzzle[rowOff+r][colOff+c] = 0
			}
		}
	}
}
