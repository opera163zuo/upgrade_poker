package tui

import (
	"time"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

type legacyUI interface {
	Init() error
	Run(loop func()) error
	Close() error
	Render(view baseui.TableView)
	WaitForAction() baseui.UIAction
	WaitForActionOrTimeout(d time.Duration) (baseui.UIAction, bool)
	ShowMessage(msg string, buttons []baseui.ButtonSpec)
	ClearMessage()
	SetPhase(phase baseui.UIPhase)
	SleepForRedraw(d time.Duration)
}

type TUIAdapter struct {
	inner legacyUI
}

func NewAdapter(inner legacyUI) *TUIAdapter        { return &TUIAdapter{inner: inner} }
func (a *TUIAdapter) Init() error                  { return a.inner.Init() }
func (a *TUIAdapter) Run(loop func()) error        { return a.inner.Run(loop) }
func (a *TUIAdapter) Close() error                 { return a.inner.Close() }
func (a *TUIAdapter) Render(view baseui.TableView) { a.inner.Render(view) }
func (a *TUIAdapter) WaitAction() baseui.UIAction  { return a.inner.WaitForAction() }
func (a *TUIAdapter) WaitActionOrTimeout(d time.Duration) (baseui.UIAction, bool) {
	return a.inner.WaitForActionOrTimeout(d)
}
func (a *TUIAdapter) ShowMessage(msg string, buttons []baseui.ButtonSpec) {
	a.inner.ShowMessage(msg, buttons)
}
func (a *TUIAdapter) ClearMessage()                  { a.inner.ClearMessage() }
func (a *TUIAdapter) SetPhase(phase baseui.UIPhase)  { a.inner.SetPhase(phase) }
func (a *TUIAdapter) SleepForRedraw(d time.Duration) { a.inner.SleepForRedraw(d) }
