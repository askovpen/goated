package ui

import (
	"github.com/rivo/tview"
)

type App struct {
	App         *tview.Application
	Layout      *tview.Flex
	Pages       *tview.Pages
	sb          *StatusBar
	al          *tview.Table
	im          IM
	showKludges bool
}

func NewApp() *App {
	a := &App{}
	a.App = tview.NewApplication()

	a.Pages = tview.NewPages()
	a.Pages.AddPage(a.AreaList())
	a.Pages.AddPage(a.AreaListQuit())

	a.sb = NewStatusBar(a)
	a.sb.Run()
	a.Layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.Pages, 0, 1, true).
		AddItem(a.sb.SB, 1, 1, false)
	return a
}
func (a *App) Run() error {
	return a.App.SetRoot(a.Layout, true).Run()
}
