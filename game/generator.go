package game

import (
	"math/rand"
)

// generateSolvedGrid fills a size×size grid with a valid sudoku solution.
func generateSolvedGrid(size, boxSize int) [][]int {
	grid := make([][]int, size)
	for i := range grid {
		grid[i] = make([]int, size)
	}
	// Seed diagonal boxes – they share no constraints with each other.
	for b := 0; b < size; b += boxSize {
		fillBox(grid, b, b, boxSize)
	}
	solveFull(grid, size, boxSize)
	return grid
}

func fillBox(grid [][]int, startRow, startCol, boxSize int) {
	nums := rand.Perm(boxSize * boxSize)
	idx := 0
	for r := startRow; r < startRow+boxSize; r++ {
		for c := startCol; c < startCol+boxSize; c++ {
			grid[r][c] = nums[idx] + 1
			idx++
		}
	}
}

// solveFull fills remaining empty cells via backtracking (randomised value order).
func solveFull(grid [][]int, size, boxSize int) bool {
	for r := range size {
		for c := range size {
			if grid[r][c] == 0 {
				vals := rand.Perm(size)
				for _, v := range vals {
					v++
					if isValid(grid, r, c, v, size, boxSize) {
						grid[r][c] = v
						if solveFull(grid, size, boxSize) {
							return true
						}
						grid[r][c] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

// SolveDeterministic fills a grid deterministically (for hint/check purposes).
func SolveDeterministic(grid [][]int, size, boxSize int) bool {
	for r := range size {
		for c := range size {
			if grid[r][c] == 0 {
				for v := 1; v <= size; v++ {
					if isValid(grid, r, c, v, size, boxSize) {
						grid[r][c] = v
						if SolveDeterministic(grid, size, boxSize) {
							return true
						}
						grid[r][c] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

func isValid(grid [][]int, row, col, val, size, boxSize int) bool {
	for c := range size {
		if grid[row][c] == val {
			return false
		}
	}
	for r := range size {
		if grid[r][col] == val {
			return false
		}
	}
	br := (row / boxSize) * boxSize
	bc := (col / boxSize) * boxSize
	for r := br; r < br+boxSize; r++ {
		for c := bc; c < bc+boxSize; c++ {
			if grid[r][c] == val {
				return false
			}
		}
	}
	return true
}

func copyGrid(src [][]int, size int) [][]int {
	dst := make([][]int, size)
	for i := range src {
		dst[i] = make([]int, size)
		copy(dst[i], src[i])
	}
	return dst
}

// countSolutions returns the number of solutions (capped at 2) for uniqueness checks.
func countSolutions(grid [][]int, size, boxSize int) int {
	g := copyGrid(grid, size)
	count := 0
	countHelper(g, size, boxSize, &count)
	return count
}

func countHelper(grid [][]int, size, boxSize int, count *int) {
	if *count > 1 {
		return
	}
	for r := range size {
		for c := range size {
			if grid[r][c] == 0 {
				for v := 1; v <= size; v++ {
					if isValid(grid, r, c, v, size, boxSize) {
						grid[r][c] = v
						countHelper(grid, size, boxSize, count)
						grid[r][c] = 0
					}
				}
				return
			}
		}
	}
	*count++
}

// removeCellsUnique removes cells while preserving uniqueness of the solution.
func removeCellsUnique(puzzle, solution [][]int, givens, size, boxSize int) {
	total := size * size
	target := total - givens

	positions := make([][2]int, 0, total)
	for r := range size {
		for c := range size {
			positions = append(positions, [2]int{r, c})
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	removed := 0
	for _, pos := range positions {
		if removed >= target {
			break
		}
		r, c := pos[0], pos[1]
		val := puzzle[r][c]
		puzzle[r][c] = 0
		if countSolutions(puzzle, size, boxSize) == 1 {
			removed++
		} else {
			puzzle[r][c] = val
		}
	}
}

// GenerateNormal creates a standard 9×9 sudoku puzzle graded to match diff.
// It retries generation until the puzzle's required solving techniques match
// the requested difficulty (up to maxAttempts tries, then uses the last result).
func GenerateNormal(diff Difficulty) *Puzzle {
	const maxAttempts = 60
	target := targetGrade[diff]
	givens := GivensCount[diff]

	var bestPuzzle [][]int
	var bestSolution [][]int

	for attempt := 0; attempt < maxAttempts; attempt++ {
		solution := generateSolvedGrid(9, 3)
		puzzle := copyGrid(solution, 9)
		removeCellsUnique(puzzle, solution, givens, 9, 3)

		grade := RateNormal(puzzle)
		if grade == target {
			bestPuzzle = puzzle
			bestSolution = solution
			break
		}
		// Keep the last attempt as fallback
		if bestPuzzle == nil {
			bestPuzzle = puzzle
			bestSolution = solution
		}
	}

	p := NewPuzzle(Normal, diff)
	for r := range 9 {
		for c := range 9 {
			p.Solution[r][c] = bestSolution[r][c]
			p.Active[r][c] = true
		}
	}
	for r := range 9 {
		for c := range 9 {
			if bestPuzzle[r][c] != 0 {
				p.Cells[r][c] = Cell{
					Value:  bestPuzzle[r][c],
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
