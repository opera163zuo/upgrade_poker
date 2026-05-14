package gui

import (
	"sync"
	"time"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

type state struct {
	mu                sync.RWMutex
	view              baseui.TableView
	actionCh          chan baseui.UIAction
	selected          map[int]bool
	cardRects         []Rect   // 物理像素坐标的牌矩形
	buttonRects       []buttonRect
	lastClick         time.Time
	lastCardIdx       int
	selectedBidType   string
	selectedBidSuit   string
	selectedBidChoice string
}

type buttonRect struct {
	Rect
	action  baseui.UIAction
	enabled bool
}
