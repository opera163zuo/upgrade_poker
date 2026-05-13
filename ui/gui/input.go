package gui

import (
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

func (g *GUI) updateInput() {
	g.handleKeyboard()

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
	waitingForHuman := g.st.view.WaitingForHuman
	hintIdx := append([]int(nil), g.st.view.HintCardIdx...)
	g.st.mu.RUnlock()

	for _, b := range buttonRects {
		if !b.enabled || !b.contains(x, y) {
			continue
		}
		action := b.action
		switch action.Type {
		case baseui.ActionSelectBid:
			g.chooseBid(action)
			return
		case baseui.ActionCancel:
			g.clearSelected()
			return
		case baseui.ActionHint:
			g.applyHintSelection(hintIdx)
			return

		case baseui.ActionConfirm:
			g.submitSelectionOrConfirm()
			return
		case baseui.ActionPlay:
			action.CardIdx = g.selectedIndices()
			if len(action.CardIdx) == 0 {
				return
			}
		}
		g.st.actionCh <- action
		return
	}
	if phase != baseui.PhasePlaying && phase != baseui.PhaseDiscard {
		return
	}
	if phase == baseui.PhasePlaying && !waitingForHuman {
		return
	}
	for idx := len(cardRects) - 1; idx >= 0; idx-- {
		r := cardRects[idx]
		if !r.contains(x, y) {
			continue
		}
		selected, double := g.toggleCardSelection(idx)
		if double {
			if phase == baseui.PhasePlaying {
				g.st.actionCh <- baseui.UIAction{Type: baseui.ActionPlay, CardIdx: selected}
			} else if phase == baseui.PhaseDiscard && len(selected) == g.discardCount() {
				g.st.actionCh <- baseui.UIAction{Type: baseui.ActionPlay, CardIdx: selected}
			}
		}
		return
	}
}

func (g *GUI) handleKeyboard() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.clearSelected()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.submitSelectionOrConfirm()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.quickBid()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.submitFixedAction(baseui.ActionPass)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.applyCurrentHint()
	}
}

func (g *GUI) quickBid() {
	g.st.mu.RLock()
	phase := g.st.view.Phase
	buttons := append([]buttonRect(nil), g.st.buttonRects...)
	g.st.mu.RUnlock()
	if phase != baseui.PhaseBidding {
		return
	}
	if action, ok := g.currentSelectedBidAction(buttons); ok {
		g.st.actionCh <- action
		return
	}
	for _, b := range buttons {
		if b.enabled && b.action.Type == baseui.ActionSelectBid {
			g.chooseBid(b.action)
			if action, ok := g.currentSelectedBidAction(buttons); ok {
				g.st.actionCh <- action
			}
			return
		}
	}
}

func (g *GUI) submitFixedAction(actionType baseui.ActionType) {
	g.st.mu.RLock()
	buttons := append([]buttonRect(nil), g.st.buttonRects...)
	g.st.mu.RUnlock()
	for _, b := range buttons {
		if b.enabled && b.action.Type == actionType {
			g.st.actionCh <- b.action
			return
		}
	}
}

func (g *GUI) submitSelectionOrConfirm() {
	g.st.mu.RLock()
	phase := g.st.view.Phase
	waitingForHuman := g.st.view.WaitingForHuman
	buttons := append([]buttonRect(nil), g.st.buttonRects...)
	g.st.mu.RUnlock()
	if phase == baseui.PhaseBidding {
		if action, ok := g.currentSelectedBidAction(buttons); ok {
			g.st.actionCh <- action
		}
		return
	}
	indices := g.selectedIndices()
	if len(indices) > 0 {
		switch phase {
		case baseui.PhasePlaying:
			if waitingForHuman {
				g.st.actionCh <- baseui.UIAction{Type: baseui.ActionPlay, CardIdx: indices}
			}
		case baseui.PhaseDiscard:
			if len(indices) == g.discardCount() {
				g.st.actionCh <- baseui.UIAction{Type: baseui.ActionPlay, CardIdx: indices}
			}
		}
		return
	}
	for _, b := range buttons {
		if b.enabled && b.action.Type == baseui.ActionConfirm {
			g.st.actionCh <- b.action
			return
		}
	}
}

func (g *GUI) selectedIndices() []int {
	g.st.mu.RLock()
	defer g.st.mu.RUnlock()
	return selectedIndicesLocked(g.st.selected)
}

func selectedIndicesLocked(selected map[int]bool) []int {
	indices := make([]int, 0, len(selected))
	for idx := range selected {
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}

func (g *GUI) toggleCardSelection(idx int) ([]int, bool) {
	now := time.Now()
	g.st.mu.Lock()
	defer g.st.mu.Unlock()
	if g.st.selected == nil {
		g.st.selected = map[int]bool{}
	}
	if g.st.selected[idx] {
		delete(g.st.selected, idx)
	} else {
		g.st.selected[idx] = true
	}
	selected := selectedIndicesLocked(g.st.selected)
	double := idx == g.st.lastCardIdx && now.Sub(g.st.lastClick) < 350*time.Millisecond
	g.st.lastClick = now
	g.st.lastCardIdx = idx
	return selected, double
}

func (g *GUI) discardCount() int {
	g.st.mu.RLock()
	defer g.st.mu.RUnlock()
	return g.st.view.DiscardCount
}

func (g *GUI) clearSelected() {
	g.st.mu.Lock()
	g.st.selected = map[int]bool{}
	g.st.mu.Unlock()
}

func (g *GUI) applyCurrentHint() {
	g.st.mu.RLock()
	hintIdx := append([]int(nil), g.st.view.HintCardIdx...)
	g.st.mu.RUnlock()
	g.applyHintSelection(hintIdx)
}

func (g *GUI) applyHintSelection(indices []int) {
	if len(indices) == 0 {
		return
	}
	g.st.mu.Lock()
	g.st.selected = map[int]bool{}
	for _, idx := range indices {
		g.st.selected[idx] = true
	}
	g.st.mu.Unlock()
}


func (g *GUI) chooseBid(action baseui.UIAction) {
	g.st.mu.Lock()
	g.st.selectedBidType = action.BidType
	g.st.selectedBidSuit = action.BidSuit
	g.st.selectedBidChoice = action.BidType + "|" + action.BidSuit
	g.st.mu.Unlock()
}

func (g *GUI) currentSelectedBidAction(buttons []buttonRect) (baseui.UIAction, bool) {
	g.st.mu.RLock()
	selectedKey := g.st.selectedBidChoice
	selectedSuit := g.st.selectedBidSuit
	g.st.mu.RUnlock()
	for _, b := range buttons {
		if !b.enabled || b.action.Type != baseui.ActionSelectBid {
			continue
		}
		key := b.action.BidType + "|" + b.action.BidSuit
		if key == selectedKey {
			return baseui.UIAction{Type: baseui.ActionBid, BidType: b.action.BidType, BidSuit: b.action.BidSuit}, true
		}
	}
	for _, b := range buttons {
		if !b.enabled || b.action.Type != baseui.ActionSelectBid {
			continue
		}
		if b.action.BidSuit == selectedSuit {
			return baseui.UIAction{Type: baseui.ActionBid, BidType: b.action.BidType, BidSuit: b.action.BidSuit}, true
		}
	}
	return baseui.UIAction{}, false
}
