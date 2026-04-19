package ui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/Kellerman81/go_sudoku/game"
	"github.com/Kellerman81/go_sudoku/locale"
	"github.com/Kellerman81/go_sudoku/pdf"
	hist "github.com/Kellerman81/go_sudoku/storage"
)

// Screen constants.
type screen int

const (
	screenMenu screen = iota
	screenGame
	screenHistory
	screenPDF
	screenAbout
)

// UI colours.
var (
	colPrimary   = nrgba(25, 90, 175, 255)
	colDanger    = nrgba(180, 30, 30, 255)
	colSuccess   = nrgba(30, 145, 60, 255)
	colPage      = nrgba(245, 247, 250, 255)
	colCard      = nrgba(255, 255, 255, 255)
	colBorder    = nrgba(210, 215, 225, 255)
	colTextMuted = nrgba(120, 125, 135, 255)
	colText      = nrgba(20, 25, 35, 255)
)

// App is the root application state.
type App struct {
	win     *app.Window
	th      *material.Theme
	history *hist.History
	lang    locale.Lang

	cur      screen
	puzzle   *game.Puzzle
	board    BoardState
	timer    string
	tickStop chan struct{}
	paused     bool
	pauseStart time.Time

	// menu widgets
	gameTypeSel   int // 0=Normal 1=Killer 2=Samurai
	difficultySel int // 0=Easy 1=Medium 2=Hard 3=Expert
	maxHintsSel   int // index into hintOptions
	maxErrorsSel  int // index into errorOptions
	btnGameType   [3]widget.Clickable
	btnDifficulty [4]widget.Clickable
	btnHintLimit  [5]widget.Clickable
	btnErrLimit   [5]widget.Clickable
	btnLang       [2]widget.Clickable
	btnNewGame    widget.Clickable
	btnHistory    widget.Clickable
	btnExportPDF  widget.Clickable
	btnAbout      widget.Clickable
	generating    bool

	// about screen
	btnAboutBack widget.Clickable
	btnGitHub    widget.Clickable

	// game state
	gameOver      bool // true once won or lost (stops input)
	gameWon       bool // true = win, false = loss
	finalElapsed  int  // seconds at game end (frozen)
	finalPoints   int  // score at game end (frozen)
	btnPlayAgain widget.Clickable
	btnOverMenu  widget.Clickable

	// game widgets
	btnHint     widget.Clickable
	btnUndo     widget.Clickable
	btnCheck    widget.Clickable
	btnGiveUp   widget.Clickable
	btnNoteMode widget.Clickable
	btnMenu     widget.Clickable
	btnNum      [9]widget.Clickable
	btnDelete   widget.Clickable
	confirm     *confirmState // non-nil when a confirm dialog is showing

	// history widgets
	btnBackHist  widget.Clickable
	btnClearHist widget.Clickable
	histList     widget.List

	// PDF export screen
	pdfNormalEd   widget.Editor
	pdfKillerEd   widget.Editor
	pdfSamuraiEd  widget.Editor
	pdfDiffSel    int // difficulty for generated puzzles
	pdfPerPageSel int // index into pdfPerPageOptions
	pdfSolutions  bool
	pdfHistory    bool
	btnPDFBack    widget.Clickable
	btnPDFExport  widget.Clickable
	btnPDFSol     widget.Clickable
	btnPDFHist    widget.Clickable
	btnPDFDiff    [4]widget.Clickable
	btnPDFPerPage [5]widget.Clickable
	pdfExporting  bool
	pdfStatusMsg  string
}

type confirmState struct {
	msg    string
	btnOK  widget.Clickable
	btnCan widget.Clickable
	onOK   func()
}

// hintOptions and errorOptions map selector index → limit value (0 = unlimited).
var hintOptions = []string{"∞", "1", "2", "3", "5"}
var hintValues = []int{0, 1, 2, 3, 5}
var errorOptions = []string{"∞", "1", "3", "5", "10"}
var errorValues = []int{0, 1, 3, 5, 10}

var pdfPerPageOptions = []string{"1", "2", "4", "6", "9"}
var pdfPerPageValues = []int{1, 2, 4, 6, 9}

func NewApp(w *app.Window, th *material.Theme) *App {
	s := hist.LoadSettings()
	return &App{
		win:           w,
		th:            th,
		history:       hist.Load(),
		timer:         "00:00",
		board:         NewBoardState(),
		gameTypeSel:   s.GameType,
		difficultySel: s.Difficulty,
		maxHintsSel:   s.MaxHintsSel,
		maxErrorsSel:  s.MaxErrorsSel,
		lang:          locale.Lang(s.Lang),
	}
}

// tr returns the current string table.
func (a *App) tr() *locale.Strings {
	return locale.Get(a.lang)
}

// Run is the main event loop.
func (a *App) Run() error {
	var ops op.Ops
	for {
		e := a.win.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.ConfigEvent:
			if e.Config.Focused {
				a.resumeTimer()
			} else {
				a.pauseTimer()
			}
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.update(gtx)
			a.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// update processes widget interactions before drawing.
func (a *App) update(gtx layout.Context) {
	switch a.cur {
	case screenMenu:
		a.updateMenu(gtx)
	case screenGame:
		a.updateGame(gtx)
	case screenHistory:
		a.updateHistory(gtx)
	case screenPDF:
		a.updatePDF(gtx)
	case screenAbout:
		// handled inline
	}
}

func (a *App) layout(gtx layout.Context) {
	// Page background.
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, colPage)

	switch a.cur {
	case screenMenu:
		a.layoutMenu(gtx)
	case screenGame:
		a.layoutGame(gtx)
	case screenHistory:
		a.layoutHistory(gtx)
	case screenPDF:
		a.layoutPDF(gtx)
	case screenAbout:
		a.layoutAbout(gtx)
	}
}

// ---- Menu ---------------------------------------------------------------

func (a *App) updateMenu(gtx layout.Context) {
	if a.generating {
		return
	}
	if a.btnNewGame.Clicked(gtx) {
		a.startGame()
	}
}

func (a *App) layoutMenu(gtx layout.Context) {
	tr := a.tr()
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(380)
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.th, unit.Sp(32), tr.AppTitle)
				lbl.Color = colPrimary
				lbl.Alignment = text.Middle
				lbl.Font.Weight = 700
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.Language)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							langSel := int(a.lang)
							dims := a.layoutToggleGroup(gtx, locale.Names, a.btnLang[:], &langSel)
							if locale.Lang(langSel) != a.lang {
								a.lang = locale.Lang(langSel)
								hist.Settings{
									GameType:     a.gameTypeSel,
									Difficulty:   a.difficultySel,
									MaxHintsSel:  a.maxHintsSel,
									MaxErrorsSel: a.maxErrorsSel,
									Lang:         int(a.lang),
								}.Save()
							}
							return dims
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.GameType)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutToggleGroup(gtx, tr.GameTypes(), a.btnGameType[:], &a.gameTypeSel)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.Difficulty)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutToggleGroup(gtx, tr.Difficulties(), a.btnDifficulty[:], &a.difficultySel)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return a.layoutLabel(gtx, tr.MaxHints)
										}),
										layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return a.layoutToggleGroup(gtx, hintOptions, a.btnHintLimit[:], &a.maxHintsSel)
										}),
									)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return a.layoutLabel(gtx, tr.MaxErrors)
										}),
										layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return a.layoutToggleGroup(gtx, errorOptions, a.btnErrLimit[:], &a.maxErrorsSel)
										}),
									)
								}),
							)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.generating {
					lbl := material.Label(a.th, unit.Sp(16), tr.Generating)
					lbl.Color = colTextMuted
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}
				return a.layoutPrimaryBtn(gtx, &a.btnNewGame, tr.NewGame)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.btnHistory.Clicked(gtx) {
					a.cur = screenHistory
				}
				return a.layoutSecondaryBtn(gtx, &a.btnHistory, tr.History)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !pdfSupported() {
					return layout.Dimensions{}
				}
				if a.btnExportPDF.Clicked(gtx) {
					a.cur = screenPDF
				}
				return a.layoutSecondaryBtn(gtx, &a.btnExportPDF, tr.ExportPDF)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.btnAbout.Clicked(gtx) {
					a.cur = screenAbout
				}
				return a.layoutSecondaryBtn(gtx, &a.btnAbout, tr.About)
			}),
		)
	})
}

// ---- About screen ----------------------------------------------------------

func (a *App) layoutAbout(gtx layout.Context) {
	tr := a.tr()
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(380)
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.btnAboutBack.Clicked(gtx) {
								a.cur = screenMenu
							}
							return a.layoutSmallBtn(gtx, &a.btnAboutBack, tr.Back, colPrimary)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(20), tr.AboutTitle)
							lbl.Font.Weight = 600
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(26), "Go Sudoku")
							lbl.Color = colPrimary
							lbl.Font.Weight = 700
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(13), tr.AboutDesc)
							lbl.Color = colTextMuted
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.AboutFeatures)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(13), tr.AboutFeatureList)
							lbl.Color = colText
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.AboutSource)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.btnGitHub.Clicked(gtx) {
								openURL("https://github.com/Kellerman81/go_sudoku")
							}
							return material.Clickable(gtx, &a.btnGitHub, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.th, unit.Sp(13), "https://github.com/Kellerman81/go_sudoku")
								lbl.Color = colPrimary
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.AboutTech)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(13), tr.AboutTechList)
							lbl.Color = colTextMuted
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}

// ---- Game ---------------------------------------------------------------

func (a *App) updateGame(gtx layout.Context) {
	// Game-over overlay buttons.
	if a.gameOver {
		if a.btnPlayAgain.Clicked(gtx) {
			a.gameOver = false
			a.startGame()
		}
		if a.btnOverMenu.Clicked(gtx) {
			a.gameOver = false
			a.cur = screenMenu
		}
		return
	}

	if a.confirm != nil {
		if a.confirm.btnOK.Clicked(gtx) {
			fn := a.confirm.onOK
			a.confirm = nil
			fn()
			return
		}
		if a.confirm.btnCan.Clicked(gtx) {
			a.confirm = nil
		}
		return
	}

	// Helper: apply a number to the selected cell and check win/loss.
	applyNum := func(n int) {
		if a.board.SelRow < 0 || a.board.SelCol < 0 {
			return
		}
		if a.board.NoteMode {
			if n > 0 {
				a.puzzle.ToggleNote(a.board.SelRow, a.board.SelCol, n)
			}
		} else {
			isErr := a.puzzle.SetCell(a.board.SelRow, a.board.SelCol, n)
			if isErr && a.puzzle.MaxErrors > 0 && a.puzzle.ErrorsMade >= a.puzzle.MaxErrors {
				a.onLose()
				return
			}
			if a.puzzle.IsComplete() {
				a.onWin()
			}
		}
		a.win.Invalidate()
	}

	// Board keyboard/click input.
	if val, ok := a.board.Update(gtx, a.puzzle); ok {
		applyNum(val)
	}

	// Number pad buttons.
	for n := 1; n <= 9; n++ {
		if a.btnNum[n-1].Clicked(gtx) {
			applyNum(n)
		}
	}
	if a.btnDelete.Clicked(gtx) {
		if a.board.SelRow >= 0 && a.board.SelCol >= 0 {
			a.puzzle.SetCell(a.board.SelRow, a.board.SelCol, 0)
		}
	}

	if a.btnNoteMode.Clicked(gtx) {
		a.board.NoteMode = !a.board.NoteMode
	}
	if a.btnHint.Clicked(gtx) {
		r, c, _ := game.Hint(a.puzzle)
		if r >= 0 {
			a.board.HintRow, a.board.HintCol = r, c
			a.board.HintExpiry = time.Now().Add(2 * time.Second)
			a.board.SelRow, a.board.SelCol = r, c
			go func() {
				time.Sleep(2 * time.Second)
				a.win.Invalidate()
			}()
		}
	}
	if a.btnUndo.Clicked(gtx) {
		a.puzzle.Undo()
	}
	if a.btnCheck.Clicked(gtx) {
		game.Validate(a.puzzle)
	}
	if a.btnGiveUp.Clicked(gtx) {
		a.confirm = &confirmState{
			msg: a.tr().GiveUpMsg,
			onOK: func() {
				a.stopTicker()
				game.SolveAll(a.puzzle)
				a.recordResult(false, true, a.puzzle.ElapsedSeconds())
			},
		}
	}
	if a.btnMenu.Clicked(gtx) {
		a.confirm = &confirmState{
			msg: a.tr().AbandonMsg,
			onOK: func() {
				a.stopTicker()
				a.cur = screenMenu
			},
		}
	}
}

func (a *App) layoutGame(gtx layout.Context) {
	if a.puzzle == nil {
		return
	}
	if a.confirm != nil {
		a.layoutConfirm(gtx)
		return
	}
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.board.Layout(gtx, a.th, a.puzzle)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Dp(160)
					gtx.Constraints.Max.X = gtx.Dp(160)
					return a.layoutSidePanel(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, a.layoutNumPad)
		}),
	)
	// Game-over overlay drawn on top.
	if a.gameOver {
		a.layoutGameOverlay(gtx)
	}
}

func (a *App) layoutSidePanel(gtx layout.Context) layout.Dimensions {
	p := a.puzzle
	tr := a.tr()

	noteLabel := tr.NotesOff
	if a.board.NoteMode {
		noteLabel = tr.NotesOn
	}

	// Build hint label with remaining count.
	hintLabel := tr.Hint
	hintCol := colPrimary
	if !game.HintAllowed(p) {
		hintLabel = tr.HintNone
		hintCol = colTextMuted
	} else if p.MaxHints > 0 {
		hintLabel = fmt.Sprintf(tr.HintLeft, p.MaxHints-p.HintsUsed)
	}

	// Build error status.
	errText := fmt.Sprintf(tr.ErrorsCount, p.ErrorsMade)
	errCol := colTextMuted
	if p.MaxErrors > 0 {
		errText = fmt.Sprintf(tr.ErrorsLimit, p.ErrorsMade, p.MaxErrors)
		if p.ErrorsMade >= p.MaxErrors {
			errCol = colDanger
		} else if p.ErrorsMade >= p.MaxErrors-1 {
			errCol = nrgba(200, 120, 0, 255)
		}
	}

	return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.th, unit.Sp(13),
					fmt.Sprintf("%s · %s", p.Type, p.Difficulty))
				lbl.Color = colTextMuted
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.th, unit.Sp(28), a.timer)
				lbl.Font.Weight = 600
				lbl.Color = colText
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				hintText := fmt.Sprintf(tr.HintsCount, p.HintsUsed)
				if p.MaxHints > 0 {
					hintText = fmt.Sprintf(tr.HintsLimit, p.HintsUsed, p.MaxHints)
				}
				lbl := material.Label(a.th, unit.Sp(13), hintText)
				lbl.Color = colTextMuted
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(a.th, unit.Sp(13), errText)
				lbl.Color = errCol
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnNoteMode, noteLabel, colPrimary)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnHint, hintLabel, hintCol)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnUndo, tr.Undo, colPrimary)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnCheck, tr.Check, colSuccess)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnGiveUp, tr.GiveUp, colDanger)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutSmallBtn(gtx, &a.btnMenu, tr.Menu, colTextMuted)
			}),
		)
	})
}

func (a *App) layoutNumPad(gtx layout.Context) layout.Dimensions {
	btnW := gtx.Dp(52)
	total := 9*btnW + 8*gtx.Dp(4)
	gtx.Constraints.Max.X = total
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 19)
		for i := range 9 {
			i := i
			n := i + 1
			done := a.numberComplete(n)
			if i > 0 {
				children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout))
			}
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(52)
				gtx.Constraints.Max.X = gtx.Dp(52)
				gtx.Constraints.Min.Y = gtx.Dp(44)
				btn := material.Button(a.th, &a.btnNum[i], fmt.Sprintf("%d", n))
				if done {
					btn.Background = nrgba(190, 195, 205, 255)
					btn.Color = nrgba(140, 145, 155, 255)
				}
				return btn.Layout(gtx)
			}))
		}
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

// ---- History ---------------------------------------------------------------

func (a *App) updateHistory(_ layout.Context) {}

func (a *App) layoutHistory(gtx layout.Context) {
	tr := a.tr()
	stats := a.history.ComputeStats()
	layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.btnBackHist.Clicked(gtx) {
							a.cur = screenMenu
						}
						return a.layoutSmallBtn(gtx, &a.btnBackHist, tr.Back, colPrimary)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.th, unit.Sp(20), tr.HistoryTitle)
						lbl.Font.Weight = 600
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if a.btnClearHist.Clicked(gtx) {
							a.history = &hist.History{}
							_ = a.history.Save()
						}
						return a.layoutSmallBtn(gtx, &a.btnClearHist, tr.Clear, colDanger)
					}),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
				rows := []struct{ k, v string }{
					{tr.GamesPlayed, fmt.Sprintf("%d", stats.GamesPlayed)},
					{tr.Won, fmt.Sprintf("%d", stats.GamesWon)},
					{tr.GaveUp, fmt.Sprintf("%d", stats.GaveUp)},
					{tr.TotalHints, fmt.Sprintf("%d", stats.TotalHints)},
					{tr.BestTime, hist.FormatDuration(stats.MinTimeSec)},
					{tr.AvgWinTime, hist.FormatDuration(stats.AvgTimeSec)},
					{tr.AvgPoints, formatPts(stats.AvgPoints)},
					{tr.MaxPoints, formatPts(stats.MaxPoints)},
				}
				children := make([]layout.FlexChild, 0, len(rows))
				for _, row := range rows {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.th, unit.Sp(13), row.k)
								lbl.Color = colTextMuted
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.th, unit.Sp(13), row.v)
								lbl.Font.Weight = 600
								return lbl.Layout(gtx)
							}),
						)
					}))
					children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(a.th, &a.histList).Layout(gtx, len(a.history.Results),
				func(gtx layout.Context, i int) layout.Dimensions {
					r := a.history.Results[len(a.history.Results)-1-i]
					result := tr.Abandoned
					if r.Won {
						result = tr.Won
					} else if r.GaveUp {
						result = tr.GaveUp
					}
					return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.th, unit.Sp(13),
									fmt.Sprintf("%s · %s · %s", r.GameType, r.Difficulty, r.StartTime.Format("2006-01-02")))
								lbl.Color = colTextMuted
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								col := colText
								if r.Won {
									col = colSuccess
								} else if r.GaveUp {
									col = colDanger
								}
								pts := r.Points
								if pts == 0 && r.Won {
									pts = hist.ComputePoints(true, r.DurationS, r.HintsUsed, r.ErrorsMade)
								}
								scorePart := ""
								if r.Won {
									scorePart = fmt.Sprintf("  ★%s", formatPts(pts))
								}
								lbl := material.Label(a.th, unit.Sp(13),
									fmt.Sprintf("%s  %s  %dh %de%s", result, hist.FormatDuration(r.DurationS), r.HintsUsed, r.ErrorsMade, scorePart))
								lbl.Color = col
								return lbl.Layout(gtx)
							}),
						)
					})
				})
		}),
	)
}

// ---- Confirm dialog --------------------------------------------------------

func (a *App) layoutConfirm(gtx layout.Context) {
	c := a.confirm
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(300)
		return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.th, unit.Sp(16), c.msg)
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return a.layoutSmallBtn(gtx, &c.btnCan, a.tr().Cancel, colTextMuted)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return a.layoutSmallBtn(gtx, &c.btnOK, a.tr().OK, colPrimary)
						}),
					)
				}),
			)
		})
	})
}

// ---- Layout helpers --------------------------------------------------------

func (a *App) layoutCard(gtx layout.Context, content layout.Widget) layout.Dimensions {
	return widget.Border{
		Color:        colBorder,
		Width:        unit.Dp(1),
		CornerRadius: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, colCard)
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, content)
	})
}

func (a *App) layoutLabel(gtx layout.Context, s string) layout.Dimensions {
	lbl := material.Label(a.th, unit.Sp(12), s)
	lbl.Color = colTextMuted
	lbl.Font.Weight = 600
	return lbl.Layout(gtx)
}

func (a *App) layoutToggleGroup(gtx layout.Context, options []string, btns []widget.Clickable, sel *int) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(options)*2)
	for i, opt := range options {
		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// Check click on the persistent button state.
			if btns[i].Clicked(gtx) {
				*sel = i
			}
			selected := *sel == i
			bg := colPrimary
			if !selected {
				bg = nrgba(230, 232, 240, 255)
			}
			h := gtx.Dp(36)
			r := gtx.Dp(6)
			defer clip.RRect{
				Rect: image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, h)},
				SE:   r, SW: r, NE: r, NW: r,
			}.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, bg)
			return material.Clickable(gtx, &btns[i], func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, h))
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.th, unit.Sp(13), opt)
					if selected {
						lbl.Color = nrgba(255, 255, 255, 255)
						lbl.Font.Weight = 600
					} else {
						lbl.Color = colText
					}
					return lbl.Layout(gtx)
				})
			})
		}))
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
}

func (a *App) layoutPrimaryBtn(gtx layout.Context, btn *widget.Clickable, label string) layout.Dimensions {
	b := material.Button(a.th, btn, label)
	b.Background = colPrimary
	b.Color = nrgba(255, 255, 255, 255)
	return b.Layout(gtx)
}

func (a *App) layoutSecondaryBtn(gtx layout.Context, btn *widget.Clickable, label string) layout.Dimensions {
	b := material.Button(a.th, btn, label)
	b.Background = nrgba(230, 232, 240, 255)
	b.Color = colText
	return b.Layout(gtx)
}

func (a *App) layoutSmallBtn(gtx layout.Context, btn *widget.Clickable, label string, col color.NRGBA) layout.Dimensions {
	b := material.Button(a.th, btn, label)
	b.Background = col
	b.Color = nrgba(255, 255, 255, 255)
	b.Inset = layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(6), Left: unit.Dp(10), Right: unit.Dp(10)}
	return b.Layout(gtx)
}

// layoutCountEditor renders a numeric text input for PDF puzzle counts.
func (a *App) layoutCountEditor(gtx layout.Context, ed *widget.Editor) layout.Dimensions {
	ed.SingleLine = true
	ed.MaxLen = 4
	e := material.Editor(a.th, ed, "0")
	e.TextSize = unit.Sp(14)
	dims := widget.Border{
		Color:        colBorder,
		Width:        unit.Dp(1),
		CornerRadius: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(80)
		gtx.Constraints.Min.X = gtx.Dp(80)
		return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
	})
	// Drain events after layout (editor populates its queue during Layout).
	for {
		_, ok := ed.Update(gtx)
		if !ok {
			break
		}
		txt := ed.Text()
		filtered := ""
		for _, ch := range txt {
			if ch >= '0' && ch <= '9' {
				filtered += string(ch)
			}
		}
		if filtered != txt {
			ed.SetText(filtered)
		}
	}
	return dims
}

// editorInt parses an editor's text as a non-negative integer (0 on error).
// formatPts formats a score with thousands separators, e.g. 8333 → "8,333".
func formatPts(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func editorInt(ed *widget.Editor) int {
	n, err := strconv.Atoi(ed.Text())
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// ---- Game logic helpers -----------------------------------------------------

func (a *App) startGame() {
	a.generating = true
	a.gameOver = false
	a.win.Invalidate()
	hist.Settings{
		GameType:     a.gameTypeSel,
		Difficulty:   a.difficultySel,
		MaxHintsSel:  a.maxHintsSel,
		MaxErrorsSel: a.maxErrorsSel,
		Lang:         int(a.lang),
	}.Save()
	maxH := hintValues[a.maxHintsSel]
	maxE := errorValues[a.maxErrorsSel]
	go func() {
		gt := game.GameType(a.gameTypeSel)
		diff := game.Difficulty(a.difficultySel)
		var p *game.Puzzle
		switch gt {
		case game.Normal:
			p = game.GenerateNormal(diff)
		case game.Killer:
			p = game.GenerateKiller(diff)
			p.AssignCageIDs()
		case game.Samurai:
			p = game.GenerateSamurai(diff)
		}
		p.MaxHints = maxH
		p.MaxErrors = maxE
		a.puzzle = p
		a.board = NewBoardState()
		a.timer = "00:00"
		a.generating = false
		a.cur = screenGame
		a.startTicker()
		a.win.Invalidate()
	}()
}

func (a *App) onWin() {
	a.stopTicker()
	elapsed := a.puzzle.ElapsedSeconds()
	a.finalElapsed = elapsed
	a.finalPoints = hist.ComputePoints(true, elapsed, a.puzzle.HintsUsed, a.puzzle.ErrorsMade)
	a.recordResult(true, false, elapsed)
	a.gameOver = true
	a.gameWon = true
	a.win.Invalidate()
}

func (a *App) onLose() {
	a.stopTicker()
	elapsed := a.puzzle.ElapsedSeconds()
	a.finalElapsed = elapsed
	a.finalPoints = 0
	game.SolveAll(a.puzzle)
	a.recordResult(false, false, elapsed)
	a.gameOver = true
	a.gameWon = false
	a.win.Invalidate()
}

func (a *App) recordResult(won, gaveUp bool, elapsed int) {
	if a.puzzle == nil {
		return
	}
	p := a.puzzle
	a.history.Append(hist.GameResult{
		GameType:   p.Type.String(),
		Difficulty: p.Difficulty.String(),
		StartTime:  p.StartTime,
		DurationS:  elapsed,
		Won:        won,
		GaveUp:     gaveUp,
		HintsUsed:  p.HintsUsed,
		ErrorsMade: p.ErrorsMade,
		Points:     hist.ComputePoints(won, elapsed, p.HintsUsed, p.ErrorsMade),
	})
}

func (a *App) startTicker() {
	a.stopTicker()
	a.tickStop = make(chan struct{})
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if a.puzzle != nil {
					a.timer = hist.FormatDuration(a.puzzle.ElapsedSeconds())
				}
				a.win.Invalidate()
			case <-a.tickStop:
				return
			}
		}
	}()
}

func (a *App) stopTicker() {
	if a.tickStop != nil {
		close(a.tickStop)
		a.tickStop = nil
	}
}

func (a *App) pauseTimer() {
	if a.cur != screenGame || a.gameOver || a.paused || a.puzzle == nil {
		return
	}
	a.paused = true
	a.pauseStart = time.Now()
	a.stopTicker()
}

func (a *App) resumeTimer() {
	if !a.paused || a.puzzle == nil {
		return
	}
	// Shift StartTime forward by the paused duration so ElapsedSeconds stays accurate.
	a.puzzle.StartTime = a.puzzle.StartTime.Add(time.Since(a.pauseStart))
	a.paused = false
	a.startTicker()
}

// ---- PDF screen ------------------------------------------------------------

func (a *App) updatePDF(_ layout.Context) {}

func (a *App) layoutPDF(gtx layout.Context) {
	tr := a.tr()
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(420)
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.btnPDFBack.Clicked(gtx) {
								a.cur = screenMenu
							}
							return a.layoutSmallBtn(gtx, &a.btnPDFBack, tr.Back, colPrimary)
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(a.th, unit.Sp(20), tr.PDFTitle)
							lbl.Font.Weight = 600
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.NormalPuzzles)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutCountEditor(gtx, &a.pdfNormalEd)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.KillerPuzzles)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutCountEditor(gtx, &a.pdfKillerEd)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.SamuraiPuzzles)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutCountEditor(gtx, &a.pdfSamuraiEd)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.Difficulty)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutToggleGroup(gtx, tr.Difficulties(), a.btnPDFDiff[:], &a.pdfDiffSel)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutLabel(gtx, tr.PuzzlesPerPage)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.layoutToggleGroup(gtx, pdfPerPageOptions, a.btnPDFPerPage[:], &a.pdfPerPageSel)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							if a.btnPDFSol.Clicked(gtx) {
								a.pdfSolutions = !a.pdfSolutions
							}
							label := tr.SolutionsOff
							bg := nrgba(230, 232, 240, 255)
							fg := colText
							if a.pdfSolutions {
								label = tr.SolutionsOn
								bg = colSuccess
								fg = nrgba(255, 255, 255, 255)
							}
							b := material.Button(a.th, &a.btnPDFSol, label)
							b.Background = bg
							b.Color = fg
							return b.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							if a.btnPDFHist.Clicked(gtx) {
								a.pdfHistory = !a.pdfHistory
							}
							label := tr.HistoryOff
							bg := nrgba(230, 232, 240, 255)
							fg := colText
							if a.pdfHistory {
								label = tr.HistoryOn
								bg = colSuccess
								fg = nrgba(255, 255, 255, 255)
							}
							b := material.Button(a.th, &a.btnPDFHist, label)
							b.Background = bg
							b.Color = fg
							return b.Layout(gtx)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.btnPDFExport.Clicked(gtx) && !a.pdfExporting {
					a.pdfExporting = true
					a.pdfStatusMsg = tr.GeneratingPDF
					a.win.Invalidate()
					opts := pdf.BookOptions{
						NormalCount:    editorInt(&a.pdfNormalEd),
						KillerCount:    editorInt(&a.pdfKillerEd),
						SamuraiCount:   editorInt(&a.pdfSamuraiEd),
						Difficulty:     game.Difficulty(a.pdfDiffSel),
						PerPage:        pdfPerPageValues[a.pdfPerPageSel],
						Solutions:      a.pdfSolutions,
						IncludeHistory: a.pdfHistory,
					}
					history := a.history
					go func() {
						// Save next to the executable on Windows; home dir on Linux.
						savePath := pdfSavePath()
						f, err := os.Create(savePath)
						if err != nil {
							a.pdfStatusMsg = "Error: " + err.Error()
							a.pdfExporting = false
							a.win.Invalidate()
							return
						}
						err = pdf.ExportBook(f, opts, history)
						f.Close()
						if err != nil {
							a.pdfStatusMsg = "Error: " + err.Error()
						} else {
							a.pdfStatusMsg = fmt.Sprintf(a.tr().PDFSaved, savePath)
						}
						a.pdfExporting = false
						a.win.Invalidate()
					}()
				}
				if a.pdfExporting {
					lbl := material.Label(a.th, unit.Sp(14), tr.GeneratingPDF)
					lbl.Color = colTextMuted
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}
				return a.layoutPrimaryBtn(gtx, &a.btnPDFExport, tr.GeneratePDF)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.pdfStatusMsg == "" {
					return layout.Dimensions{}
				}
				col := colSuccess
				if len(a.pdfStatusMsg) >= 5 && a.pdfStatusMsg[:5] == "Error" {
					col = colDanger
				}
				lbl := material.Label(a.th, unit.Sp(13), a.pdfStatusMsg)
				lbl.Color = col
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
		)
	})
}

// layoutGameOverlay draws a semi-transparent win/lose panel over the game.
func (a *App) layoutGameOverlay(gtx layout.Context) {
	tr := a.tr()
	defer clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, color.NRGBA{A: 160})

	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(300)
		return a.layoutCard(gtx, func(gtx layout.Context) layout.Dimensions {
			title := fmt.Sprintf("%s  ★%s", tr.YouWon, formatPts(a.finalPoints))
			titleCol := colSuccess
			sub := fmt.Sprintf(tr.WonSub, hist.FormatDuration(a.finalElapsed), a.puzzle.HintsUsed, a.puzzle.ErrorsMade)
			if !a.gameWon {
				title = tr.GameOver
				titleCol = colDanger
				sub = fmt.Sprintf(tr.LostSub, a.puzzle.ErrorsMade)
			}
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.th, unit.Sp(28), title)
					lbl.Color = titleCol
					lbl.Font.Weight = 700
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(a.th, unit.Sp(14), sub)
					lbl.Alignment = text.Middle
					lbl.Color = colTextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return a.layoutSmallBtn(gtx, &a.btnOverMenu, tr.BackToMenu, colTextMuted)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return a.layoutSmallBtn(gtx, &a.btnPlayAgain, tr.PlayAgain, colPrimary)
						}),
					)
				}),
			)
		})
	})
}

// numberComplete returns true when digit n is fully placed on the board.
func (a *App) numberComplete(n int) bool {
	if a.puzzle == nil {
		return false
	}
	return a.puzzle.ValueComplete(n)
}
