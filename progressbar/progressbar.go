package progressbar

import (
	pb "github.com/schollz/progressbar/v2"
)

var (
	isQuiet     bool
	progressbar *pb.ProgressBar
)

// CreateProgressbar to indicate progress
func CreateProgressbar(total int, quiet bool) {
	isQuiet = quiet
	if !isQuiet {
		progressbar = pb.New(total)
	}
}

// AddToBar adds a specific amount to the progressbar
func AddToBar(amount int) {
	if !isQuiet {
		progressbar.Add(1)
	}
}
