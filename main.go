package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/text"
	"gioui.org/widget/material"

	"github.com/Kellerman81/go_sudoku/ui"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Go Sudoku"),
			app.Size(1024, 720),
		)
		th := material.NewTheme()
		th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
		a := ui.NewApp(w, th)
		if err := a.Run(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}
