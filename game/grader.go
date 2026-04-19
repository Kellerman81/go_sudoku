package game

import "math/bits"

// grade constants mirror Difficulty but for internal grading.
const (
	gradeEasy   = 0
	gradeMedium = 1
	gradeHard   = 2
	gradeExpert = 3
)

// cands[r][c] is a bitmask of possible values: bit v is set if v is still a candidate.
type cands [9][9]uint16

const fullMask = uint16(0b1111111110) // bits 1-9

func buildCands(grid [][]int) cands {
	var c cands
	for r := 0; r < 9; r++ {
		for col := 0; col < 9; col++ {
			if grid[r][col] == 0 {
				c[r][col] = fullMask
			}
		}
	}
	for r := 0; r < 9; r++ {
		for col := 0; col < 9; col++ {
			if v := grid[r][col]; v != 0 {
				c.eliminate(r, col, v)
			}
		}
	}
	return c
}

func (c *cands) eliminate(r, col, v int) {
	bit := uint16(1 << v)
	for i := 0; i < 9; i++ {
		c[r][i] &^= bit
		c[i][col] &^= bit
	}
	br, bc := (r/3)*3, (col/3)*3
	for rr := br; rr < br+3; rr++ {
		for cc := bc; cc < bc+3; cc++ {
			c[rr][cc] &^= bit
		}
	}
	c[r][col] = 0
}

func bitCount(x uint16) int { return bits.OnesCount16(x) }

func lowBit(x uint16) int {
	for v := 1; v <= 9; v++ {
		if x&(1<<v) != 0 {
			return v
		}
	}
	return 0
}

// applyNakedSingles places any cell with exactly one candidate. Returns count placed.
func applyNakedSingles(grid [][]int, c *cands) int {
	placed := 0
	for r := 0; r < 9; r++ {
		for col := 0; col < 9; col++ {
			if grid[r][col] == 0 && bitCount(c[r][col]) == 1 {
				v := lowBit(c[r][col])
				grid[r][col] = v
				c.eliminate(r, col, v)
				placed++
			}
		}
	}
	return placed
}

// unitCells returns the 9 (r,c) positions of a unit: type 0=row, 1=col, 2=box.
func unitCells(kind, idx int) [9][2]int {
	var cells [9][2]int
	switch kind {
	case 0: // row idx
		for i := 0; i < 9; i++ {
			cells[i] = [2]int{idx, i}
		}
	case 1: // col idx
		for i := 0; i < 9; i++ {
			cells[i] = [2]int{i, idx}
		}
	case 2: // box idx (0-8, row-major)
		br, bc := (idx/3)*3, (idx%3)*3
		i := 0
		for r := br; r < br+3; r++ {
			for col := bc; col < bc+3; col++ {
				cells[i] = [2]int{r, col}
				i++
			}
		}
	}
	return cells
}

// applyHiddenSingles places any digit that can only go in one cell per unit.
func applyHiddenSingles(grid [][]int, c *cands) int {
	placed := 0
	for kind := 0; kind < 3; kind++ {
		for idx := 0; idx < 9; idx++ {
			cells := unitCells(kind, idx)
			for v := 1; v <= 9; v++ {
				bit := uint16(1 << v)
				count, last := 0, [2]int{}
				for _, rc := range cells {
					if grid[rc[0]][rc[1]] == 0 && c[rc[0]][rc[1]]&bit != 0 {
						count++
						last = rc
					}
				}
				if count == 1 {
					r, col := last[0], last[1]
					grid[r][col] = v
					c.eliminate(r, col, v)
					placed++
				}
			}
		}
	}
	return placed
}

// applyNakedPairs eliminates candidates using naked pairs within a unit.
func applyNakedPairs(c *cands) int {
	elim := 0
	for kind := 0; kind < 3; kind++ {
		for idx := 0; idx < 9; idx++ {
			cells := unitCells(kind, idx)
			for i := 0; i < 9; i++ {
				ri, ci := cells[i][0], cells[i][1]
				if bitCount(c[ri][ci]) != 2 {
					continue
				}
				for j := i + 1; j < 9; j++ {
					rj, cj := cells[j][0], cells[j][1]
					if c[ri][ci] != c[rj][cj] {
						continue
					}
					pair := c[ri][ci]
					for k := 0; k < 9; k++ {
						rk, ck := cells[k][0], cells[k][1]
						if (rk == ri && ck == ci) || (rk == rj && ck == cj) {
							continue
						}
						before := c[rk][ck]
						c[rk][ck] &^= pair
						if c[rk][ck] != before {
							elim++
						}
					}
				}
			}
		}
	}
	return elim
}

// applyPointingPairs: if all candidates for a digit in a box are in one row/col,
// eliminate that digit from the rest of that row/col outside the box.
func applyPointingPairs(c *cands) int {
	elim := 0
	for box := 0; box < 9; box++ {
		br, bc := (box/3)*3, (box%3)*3
		for v := 1; v <= 9; v++ {
			bit := uint16(1 << v)
			rows, cols := [3]bool{}, [3]bool{}
			found := false
			for r := br; r < br+3; r++ {
				for col := bc; col < bc+3; col++ {
					if c[r][col]&bit != 0 {
						rows[r-br] = true
						cols[col-bc] = true
						found = true
					}
				}
			}
			if !found {
				continue
			}
			rowCount, colCount := 0, 0
			lockedRow, lockedCol := -1, -1
			for i := 0; i < 3; i++ {
				if rows[i] {
					rowCount++
					lockedRow = br + i
				}
				if cols[i] {
					colCount++
					lockedCol = bc + i
				}
			}
			if rowCount == 1 {
				for col := 0; col < 9; col++ {
					if col >= bc && col < bc+3 {
						continue
					}
					before := c[lockedRow][col]
					c[lockedRow][col] &^= bit
					if c[lockedRow][col] != before {
						elim++
					}
				}
			}
			if colCount == 1 {
				for r := 0; r < 9; r++ {
					if r >= br && r < br+3 {
						continue
					}
					before := c[r][lockedCol]
					c[r][lockedCol] &^= bit
					if c[r][lockedCol] != before {
						elim++
					}
				}
			}
		}
	}
	return elim
}

// applyNakedTriples eliminates via naked triples (three cells share ≤3 candidates).
func applyNakedTriples(c *cands) int {
	elim := 0
	for kind := 0; kind < 3; kind++ {
		for idx := 0; idx < 9; idx++ {
			cells := unitCells(kind, idx)
			for i := 0; i < 9; i++ {
				ri, ci := cells[i][0], cells[i][1]
				ni := bitCount(c[ri][ci])
				if ni < 2 || ni > 3 {
					continue
				}
				for j := i + 1; j < 9; j++ {
					rj, cj := cells[j][0], cells[j][1]
					nj := bitCount(c[rj][cj])
					if nj < 2 || nj > 3 {
						continue
					}
					union2 := c[ri][ci] | c[rj][cj]
					if bitCount(union2) > 3 {
						continue
					}
					for k := j + 1; k < 9; k++ {
						rk, ck := cells[k][0], cells[k][1]
						nk := bitCount(c[rk][ck])
						if nk < 2 || nk > 3 {
							continue
						}
						triple := union2 | c[rk][ck]
						if bitCount(triple) != 3 {
							continue
						}
						for m := 0; m < 9; m++ {
							rm, cm := cells[m][0], cells[m][1]
							if (rm == ri && cm == ci) || (rm == rj && cm == cj) || (rm == rk && cm == ck) {
								continue
							}
							before := c[rm][cm]
							c[rm][cm] &^= triple
							if c[rm][cm] != before {
								elim++
							}
						}
					}
				}
			}
		}
	}
	return elim
}

// applyXWing: if a candidate appears in exactly 2 rows in the same 2 columns,
// eliminate it from those columns in all other rows (and vice versa for cols).
func applyXWing(c *cands) int {
	elim := 0
	for v := 1; v <= 9; v++ {
		bit := uint16(1 << v)
		// rows that have exactly 2 candidates for v
		type rowInfo struct {
			idx  int
			cols [2]int
		}
		var twoRows []rowInfo
		for r := 0; r < 9; r++ {
			var found []int
			for col := 0; col < 9; col++ {
				if c[r][col]&bit != 0 {
					found = append(found, col)
				}
			}
			if len(found) == 2 {
				twoRows = append(twoRows, rowInfo{r, [2]int{found[0], found[1]}})
			}
		}
		for i := 0; i < len(twoRows); i++ {
			for j := i + 1; j < len(twoRows); j++ {
				if twoRows[i].cols != twoRows[j].cols {
					continue
				}
				c1, c2 := twoRows[i].cols[0], twoRows[i].cols[1]
				r1, r2 := twoRows[i].idx, twoRows[j].idx
				for r := 0; r < 9; r++ {
					if r == r1 || r == r2 {
						continue
					}
					for _, col := range []int{c1, c2} {
						before := c[r][col]
						c[r][col] &^= bit
						if c[r][col] != before {
							elim++
						}
					}
				}
			}
		}
		// same but for columns
		var twoCols []struct {
			idx  int
			rows [2]int
		}
		for col := 0; col < 9; col++ {
			var found []int
			for r := 0; r < 9; r++ {
				if c[r][col]&bit != 0 {
					found = append(found, r)
				}
			}
			if len(found) == 2 {
				twoCols = append(twoCols, struct {
					idx  int
					rows [2]int
				}{col, [2]int{found[0], found[1]}})
			}
		}
		for i := 0; i < len(twoCols); i++ {
			for j := i + 1; j < len(twoCols); j++ {
				if twoCols[i].rows != twoCols[j].rows {
					continue
				}
				r1, r2 := twoCols[i].rows[0], twoCols[i].rows[1]
				col1, col2 := twoCols[i].idx, twoCols[j].idx
				for col := 0; col < 9; col++ {
					if col == col1 || col == col2 {
						continue
					}
					for _, r := range []int{r1, r2} {
						before := c[r][col]
						c[r][col] &^= bit
						if c[r][col] != before {
							elim++
						}
					}
				}
			}
		}
	}
	return elim
}

func isSolved(grid [][]int) bool {
	for r := 0; r < 9; r++ {
		for col := 0; col < 9; col++ {
			if grid[r][col] == 0 {
				return false
			}
		}
	}
	return true
}

// RateNormal grades a 9×9 puzzle by the hardest technique needed to solve it.
// Returns gradeEasy/gradeMedium/gradeHard/gradeExpert.
func RateNormal(puzzle [][]int) int {
	// Work on a copy
	grid := copyGrid(puzzle, 9)
	c := buildCands(grid)
	grade := gradeEasy

	for !isSolved(grid) {
		progress := false

		// Easy: naked + hidden singles
		for {
			n := applyNakedSingles(grid, &c)
			n += applyHiddenSingles(grid, &c)
			if n == 0 {
				break
			}
			progress = true
		}
		if isSolved(grid) {
			break
		}

		// Medium: naked pairs + pointing pairs
		n := applyNakedPairs(&c)
		n += applyPointingPairs(&c)
		if n > 0 {
			if grade < gradeMedium {
				grade = gradeMedium
			}
			progress = true
			continue
		}

		// Hard: naked triples + X-wing
		n = applyNakedTriples(&c)
		n += applyXWing(&c)
		if n > 0 {
			if grade < gradeHard {
				grade = gradeHard
			}
			progress = true
			continue
		}

		if !progress {
			// Nothing worked — Expert level (requires guessing / advanced)
			return gradeExpert
		}
	}
	return grade
}
