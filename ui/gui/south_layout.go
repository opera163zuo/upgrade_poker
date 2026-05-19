package gui

import (
	"sort"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

func (g *GUI) southSlotsAdaptive(cards []baseui.CardView, selected map[int]bool,
	biddingRaise map[int]bool, trumpSuit string) []southSlot {

	order := g.suitRowOrder(cards, trumpSuit)
	suitIndices := map[string][]int{}
	for idx, c := range cards {
		key := effectiveSuit(c)
		suitIndices[key] = append(suitIndices[key], idx)
	}

	var ordered []int
	for _, suit := range order {
		ordered = append(ordered, suitIndices[suit]...)
	}
	if len(ordered) == 0 {
		return nil
	}

	physCardW, physCardH := g.sc.CardPhysSize()
	tableX := g.sc.PX(RefTableX)
	tableW := g.sc.PXAbsolute(RefTableW)
	withinCount, betweenCount := southHandGapCounts(cards, ordered)
	withinSuitGap, betweenSuitGap := g.fitSouthHandGaps(withinCount, betweenCount, physCardW, tableW)
	totalWidth := southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap)

	startX := tableX + (tableW-totalWidth)/2
	preferredStartX := g.sc.PX(RefSouthHandX)
	if startX < preferredStartX {
		startX = preferredStartX
	}
	maxStartX := tableX + tableW - totalWidth
	if startX > maxStartX {
		startX = maxStartX
	}

	y := g.sc.PX(RefSouthHandY)
	var slots []southSlot
	for pos, idx := range ordered {
		var slotX int
		if pos == 0 {
			slotX = startX
		} else {
			gap := withinSuitGap
			if effectiveSuit(cards[ordered[pos-1]]) != effectiveSuit(cards[idx]) {
				gap = betweenSuitGap
			}
			slotX = slots[pos-1].x + gap
		}
		slotY := y
		if selected[idx] || biddingRaise[idx] {
			slotY -= g.sc.PXAbsolute(18)
		}
		slots = append(slots, southSlot{
			idx: idx,
			x:   slotX,
			y:   slotY,
			w:   physCardW,
			h:   physCardH,
		})
	}

	sort.Slice(slots, func(i, j int) bool { return slots[i].idx < slots[j].idx })
	return slots
}

func southHandGapCounts(cards []baseui.CardView, ordered []int) (int, int) {
	withinCount := 0
	betweenCount := 0
	for i := 1; i < len(ordered); i++ {
		if effectiveSuit(cards[ordered[i-1]]) != effectiveSuit(cards[ordered[i]]) {
			betweenCount++
			continue
		}
		withinCount++
	}
	return withinCount, betweenCount
}

func (g *GUI) fitSouthHandGaps(withinCount, betweenCount, physCardW, maxWidth int) (int, int) {
	withinSuitGap := g.sc.PXAbsolute(RefSouthHandGap)
	betweenSuitGap := g.sc.PXAbsolute(22)
	if southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap) <= maxWidth {
		return withinSuitGap, betweenSuitGap
	}

	minWithinGap := g.sc.PXAbsolute(4)
	if minWithinGap < 1 {
		minWithinGap = 1
	}
	if withinCount > 0 {
		maxWithinGap := (maxWidth - physCardW - betweenCount*betweenSuitGap) / withinCount
		if maxWithinGap < withinSuitGap {
			withinSuitGap = maxWithinGap
		}
		if withinSuitGap < minWithinGap {
			withinSuitGap = minWithinGap
		}
	}
	if southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap) <= maxWidth {
		return withinSuitGap, betweenSuitGap
	}

	minBetweenGap := withinSuitGap + g.sc.PXAbsolute(3)
	if minBetweenGap < g.sc.PXAbsolute(6) {
		minBetweenGap = g.sc.PXAbsolute(6)
	}
	if betweenCount > 0 {
		maxBetweenGap := (maxWidth - physCardW - withinCount*withinSuitGap) / betweenCount
		if maxBetweenGap < betweenSuitGap {
			betweenSuitGap = maxBetweenGap
		}
		if betweenSuitGap < minBetweenGap {
			betweenSuitGap = minBetweenGap
		}
	}
	if southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap) <= maxWidth {
		return withinSuitGap, betweenSuitGap
	}

	if withinCount > 0 {
		maxWithinGap := (maxWidth - physCardW - betweenCount*betweenSuitGap) / withinCount
		if maxWithinGap < withinSuitGap {
			withinSuitGap = maxWithinGap
		}
		if withinSuitGap < 1 {
			withinSuitGap = 1
		}
	}
	if southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap) <= maxWidth {
		return withinSuitGap, betweenSuitGap
	}

	if betweenCount > 0 {
		maxBetweenGap := (maxWidth - physCardW - withinCount*withinSuitGap) / betweenCount
		if maxBetweenGap < betweenSuitGap {
			betweenSuitGap = maxBetweenGap
		}
		if betweenSuitGap < 1 {
			betweenSuitGap = 1
		}
	}
	return withinSuitGap, betweenSuitGap
}

func southHandTotalWidth(physCardW, withinCount, betweenCount, withinSuitGap, betweenSuitGap int) int {
	return physCardW + withinCount*withinSuitGap + betweenCount*betweenSuitGap
}
