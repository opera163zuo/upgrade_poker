package gui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

func (g *GUI) updateInput() {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.mouseDown = false
		return
	}
	if g.mouseDown {
		return
	}
	g.mouseDown = true
	x, y := ebiten.CursorPosition()

	g.st.mu.RLock()
	buttonRects := append([]buttonRect(nil), g.st.buttonRects...)
	cardRects := append([]rect(nil), g.st.cardRects...)
	phase := g.st.view.Phase
	g.st.mu.RUnlock()

	for _, b := range buttonRects {
		if b.contains(x, y) {
			g.st.actionCh <- baseui.UIAction{Type: b.action}
			return
		}
	}
	if phase != baseui.PhasePlaying && phase != baseui.PhaseDiscard {
		return
	}
	for idx, r := range cardRects {
		if !r.contains(x, y) {
			continue
		}
		now := time.Now()
		g.st.mu.Lock()
		if g.st.selected == nil {
			g.st.selected = map[int]bool{}
		}
		if g.st.selected[idx] {
			delete(g.st.selected, idx)
		} else {
			g.st.selected[idx] = true
		}
		selected := make([]int, 0, len(g.st.selected))
		for i := range g.st.selected {
			selected = append(selected, i)
		}
		double := now.Sub(g.st.lastClick) < 350*time.Millisecond
		g.st.lastClick = now
		g.st.mu.Unlock()
		if double && phase == baseui.PhasePlaying {
			g.st.actionCh <- baseui.UIAction{Type: baseui.ActionPlay, CardIdx: selected}
		}
		return
	}
}
