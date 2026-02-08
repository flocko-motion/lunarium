package main

// rectsOverlap checks if two axis-aligned rectangles overlap.
// Each rect is defined as (x, y, w, h) where (x, y) is the top-left corner.
func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	if x1+w1 <= x2 || x2+w2 <= x1 {
		return false
	}
	if y1+h1 <= y2 || y2+h2 <= y1 {
		return false
	}
	return true
}
