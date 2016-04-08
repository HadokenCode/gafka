package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

var (
	w, h        int
	pageSize    int
	selectedRow = 0
	detailView  = false
)

const (
	coldef = termbox.ColorDefault
)

func refreshSize() {
	w, h = termbox.Size()
	pageSize = h - 1 // footer=1
}

func redrawAll() {
	termbox.Clear(coldef, coldef)
	refreshSize()

	x, y := 1, 0
	lock.Lock()
	for i := len(events) - 1; i >= 0; i-- {
		evt := events[i]
		drawEvent(x, y, evt)
		y++
	}

	drawFooter()
	lock.Unlock()

	termbox.Flush()
}

func drawEvent(x, y int, evt interface{}) {
	fg_col, bg_col := coldef, coldef
	if y == selectedRow {
		fg_col = termbox.ColorBlack
		bg_col = termbox.ColorYellow
	}

	switch hook := evt.(type) {
	case *Webhook:
		var row string
		if len(hook.Commits) == 0 {
			row = fmt.Sprintf("%10s %20s %25s %d",
				" ",
				hook.User_name,
				hook.Repository.Name,
				hook.Total_commits_count)
		} else {
			row = fmt.Sprintf("%10s %20s %25s %s",
				since(hook.Commits[0].Timestamp),
				hook.User_name,
				hook.Repository.Name,
				hook.Commits[0].Message)
		}

		for i, c := range row {
			termbox.SetCell(1+i, y, c, fg_col, bg_col)
		}
		return

		for _, c := range hook.Commits {
			row := fmt.Sprintf("%10s %15s %10s %s",
				since(c.Timestamp),
				hook.User_name,
				hook.Repository.Name,
				c.Message)
			if len(row) > w {
				//row = row[:w]
			}

			for i, c := range row {
				termbox.SetCell(1+i, y, c, fg_col, bg_col)
			}
			y++
		}
	}
}

func drawSplash() {
	refreshSize()
	row := "loading gitlab events..."
	x, y := w/2-len(row)/2, h/2+1
	for i, c := range row {
		termbox.SetCell(x+i, y, c, termbox.ColorGreen, coldef)
	}
	termbox.Flush()
}

func drawFooter() {
	footerText := fmt.Sprintf(" Esc:Back   Enter:Detail projects: events:%d row:%d",
		len(events), selectedRow)
	for i := 0; i < w; i++ {
		termbox.SetCell(i, h-1, ' ', coldef, termbox.ColorBlue)
	}
	for i, c := range footerText {
		termbox.SetCell(i, h-1, c, coldef, termbox.ColorBlue)
	}
}
