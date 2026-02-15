package ui

// Braille oscilloscope renderer
//
// Each braille character (U+2800-U+28FF) is a 2x4 dot grid:
//   col0 col1
//   [0]  [3]   row 0 (top)
//   [1]  [4]   row 1
//   [2]  [5]   row 2
//   [6]  [7]   row 3 (bottom)
//
// Dot bit positions in the Unicode offset:
//   bit 0 = (0,0)  bit 3 = (1,0)
//   bit 1 = (0,1)  bit 4 = (1,1)
//   bit 2 = (0,2)  bit 5 = (1,2)
//   bit 6 = (0,3)  bit 7 = (1,3)

// brailleDot maps (col, row) to the bit offset in the braille codepoint
var brailleDot = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // col 0
	{0x08, 0x10, 0x20, 0x80}, // col 1
}

// RenderBraille renders float64 samples as a braille waveform.
// width = number of braille characters wide.
// height = number of braille characters tall (each is 4 dots high).
// Samples are expected in -1..1 range.
func RenderBraille(samples []float64, width, height int) []string {
	totalRows := height * 4 // total dot rows
	totalCols := width * 2  // total dot columns

	// Grid of braille characters, initialized to empty braille (U+2800)
	grid := make([][]rune, height)
	for r := 0; r < height; r++ {
		grid[r] = make([]rune, width)
	}

	if len(samples) == 0 || width == 0 || height == 0 {
		return gridToStrings(grid)
	}

	// Map samples to dot columns, drawing a connected line
	prevDotRow := -1
	for col := 0; col < totalCols; col++ {
		// Sample index for this dot column
		si := col * len(samples) / totalCols
		if si >= len(samples) {
			si = len(samples) - 1
		}

		// Map sample value (-1..1) to dot row (0=top, totalRows-1=bottom)
		s := samples[si]
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		// Invert: +1 -> row 0 (top), -1 -> row totalRows-1 (bottom)
		dotRow := int((1.0 - s) * 0.5 * float64(totalRows-1))
		if dotRow < 0 {
			dotRow = 0
		}
		if dotRow >= totalRows {
			dotRow = totalRows - 1
		}

		// Draw connected line from prevDotRow to dotRow
		if prevDotRow < 0 {
			setDot(grid, col, dotRow)
		} else {
			drawLine(grid, col-1, prevDotRow, col, dotRow)
		}
		prevDotRow = dotRow
	}

	return gridToStrings(grid)
}

// setDot sets a single dot in the braille grid
func setDot(grid [][]rune, dotCol, dotRow int) {
	charCol := dotCol / 2
	charRow := dotRow / 4
	if charRow < 0 || charRow >= len(grid) || charCol < 0 || charCol >= len(grid[0]) {
		return
	}
	localCol := dotCol % 2
	localRow := dotRow % 4
	grid[charRow][charCol] |= brailleDot[localCol][localRow]
}

// drawLine draws dots between two points using Bresenham's
func drawLine(grid [][]rune, x0, y0, x1, y1 int) {
	dx := x1 - x0
	dy := y1 - y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	for {
		setDot(grid, x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// gridToStrings converts the braille grid to string lines
func gridToStrings(grid [][]rune) []string {
	lines := make([]string, len(grid))
	for r, row := range grid {
		chars := make([]rune, len(row))
		for c, bits := range row {
			chars[c] = 0x2800 + bits
		}
		lines[r] = string(chars)
	}
	return lines
}
