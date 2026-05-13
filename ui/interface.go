package ui

import "time"

type GameUI interface {
	Init() error
	Run(loop func()) error
	Close() error
	Render(view TableView)
	WaitAction() UIAction
	WaitActionOrTimeout(d time.Duration) (UIAction, bool)
	ShowMessage(msg string, buttons []ButtonSpec)
	ClearMessage()
	SetPhase(phase UIPhase)
	SleepForRedraw(d time.Duration)
}
