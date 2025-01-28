package progressbar

import (
	"fmt"
	"io"
	"os"
)

func NewProgressBar(step int) ProgressBar {
	return ProgressBar{step: step}
}

type ProgressBar struct {
	step int
	prev int
}

func (b *ProgressBar) Draw(percent int) {
	if percent < b.prev+b.step {
		return
	}
	defer func() {
		for b.prev < percent {
			b.prev += b.step
		}
	}()

	var bar string
	for i := 0; i < 20; i++ {
		if i < percent/5 {
			bar += "#"
		} else {
			bar += "."
		}
	}

	_, err := io.WriteString(os.Stderr, fmt.Sprintf("[%s] %d%%\n", bar, percent))
	if err != nil {
		panic(err)
	}
}
