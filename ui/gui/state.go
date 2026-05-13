package gui

import (
	"sync"
	"time"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

type state struct {
	mu          sync.RWMutex
	view        baseui.TableView
	actionCh    chan baseui.UIAction
	selected    map[int]bool
	cardRects   []rect
	buttonRects []buttonRect
	lastClick   time.Time
	lastCardIdx int
}

type buttonRect struct {
	rect
	action  baseui.UIAction
	enabled bool
}
