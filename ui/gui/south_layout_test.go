package gui

import (
	"testing"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

func TestSouthSlotsAdaptiveStayWithinTableWithLargeHand(t *testing.T) {
	g := &GUI{
		sc: BuildScaleCtx(1280, 960, 1.0),
	}
	suits := []string{"鏂瑰潡", "姊呰姳", "绾㈠績", "榛戞"}
	cards := make([]baseui.CardView, 0, 54)
	for i := 0; i < 54; i++ {
		suit := suits[i%len(suits)]
		cards = append(cards, baseui.CardView{
			Suit:          "鈾?",
			EffectiveSuit: suit,
			Rank:          "A",
			FaceUp:        true,
			Trump:         suit == "榛戞",
		})
	}

	slots := g.southSlotsAdaptive(cards, map[int]bool{}, map[int]bool{}, "榛戞")
	if len(slots) != len(cards) {
		t.Fatalf("expected %d slots, got %d", len(cards), len(slots))
	}

	cardW, _ := g.sc.CardPhysSize()
	tableX := g.sc.PX(RefTableX)
	tableRight := tableX + g.sc.PXAbsolute(RefTableW)
	minX := slots[0].x
	maxRight := slots[0].x + cardW
	for _, slot := range slots[1:] {
		if slot.x < minX {
			minX = slot.x
		}
		if right := slot.x + cardW; right > maxRight {
			maxRight = right
		}
	}

	if minX < tableX {
		t.Fatalf("hand starts outside table: minX=%d tableX=%d", minX, tableX)
	}
	if maxRight > tableRight {
		t.Fatalf("hand exceeds table width: maxRight=%d tableRight=%d", maxRight, tableRight)
	}
}
