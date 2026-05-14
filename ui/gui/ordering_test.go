package gui

import (
	"testing"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

func TestSuitRowOrderTrumpLast(t *testing.T) {
	trumpCases := []struct {
		trumpSuit     string
		expectedOrder []string
	}{
		{trumpSuit: "黑桃", expectedOrder: []string{"方块", "梅花", "红心", "黑桃"}},
		{trumpSuit: "红心", expectedOrder: []string{"方块", "梅花", "黑桃", "红心"}},
		{trumpSuit: "方块", expectedOrder: []string{"梅花", "红心", "黑桃", "方块"}},
		{trumpSuit: "梅花", expectedOrder: []string{"方块", "红心", "黑桃", "梅花"}},
		{trumpSuit: "王", expectedOrder: []string{"方块", "梅花", "红心", "黑桃", "王"}},
	}

	for _, tc := range trumpCases {
		t.Run("trump_"+tc.trumpSuit, func(t *testing.T) {
			cards := []baseui.CardView{
				{Suit: "♠", EffectiveSuit: "黑桃", Rank: "A", FaceUp: true},
				{Suit: "♥", EffectiveSuit: "红心", Rank: "K", FaceUp: true},
				{Suit: "♦", EffectiveSuit: "方块", Rank: "Q", FaceUp: true},
				{Suit: "♣", EffectiveSuit: "梅花", Rank: "J", FaceUp: true},
				{Suit: "♠", EffectiveSuit: "黑桃", Rank: "10", FaceUp: true},
				{Suit: "♥", EffectiveSuit: "红心", Rank: "9", FaceUp: true},
				{Suit: "♦", EffectiveSuit: "方块", Rank: "8", FaceUp: true},
				{Suit: "♣", EffectiveSuit: "梅花", Rank: "7", FaceUp: true},
				{Suit: "♠", EffectiveSuit: "黑桃", Rank: "6", FaceUp: true},
			}
			for i := range cards {
				if cards[i].EffectiveSuit == tc.trumpSuit {
					cards[i].Trump = true
				}
			}

			g := &GUI{}
			order := g.suitRowOrder(cards, tc.trumpSuit)

			if len(order) < 1 {
				t.Fatal("empty order")
			}
			lastSuit := order[len(order)-1]
			if lastSuit != tc.trumpSuit {
				t.Errorf("trump suit %q should be last in order, got %q (order: %v)",
					tc.trumpSuit, lastSuit, order)
			}
			if len(order) != len(tc.expectedOrder) {
				t.Errorf("expected %d suits, got %d: %v", len(tc.expectedOrder), len(order), order)
				return
			}
			for i := range order {
				if order[i] != tc.expectedOrder[i] {
					t.Errorf("position %d: expected %q, got %q (full order: %v)",
						i, tc.expectedOrder[i], order[i], order)
					break
				}
			}
		})
	}
}

func TestSouthSlotsTrumpOnRight(t *testing.T) {
	g := &GUI{}
	cards := []baseui.CardView{
		{Suit: "♦", EffectiveSuit: "方块", Rank: "10", FaceUp: true},
		{Suit: "♣", EffectiveSuit: "梅花", Rank: "K", FaceUp: true},
		{Suit: "♥", EffectiveSuit: "红心", Rank: "A", FaceUp: true},
		{Suit: "♠", EffectiveSuit: "黑桃", Rank: "5", FaceUp: true, Trump: true},
		{Suit: "♠", EffectiveSuit: "黑桃", Rank: "A", FaceUp: true, Trump: true},
	}
	selected := map[int]bool{}
	biddingRaise := map[int]bool{}
	trumpSuit := "黑桃"

	slots := g.southSlots(cards, selected, biddingRaise, trumpSuit)

	if len(slots) != len(cards) {
		t.Fatalf("expected %d slots, got %d", len(cards), len(slots))
	}

	trumpMinX := 9999
	nonTrumpMaxX := 0
	for _, slot := range slots {
		c := cards[slot.idx]
		if c.Trump {
			if slot.x < trumpMinX {
				trumpMinX = slot.x
			}
		} else {
			if slot.x > nonTrumpMaxX {
				nonTrumpMaxX = slot.x
			}
		}
	}

	if trumpMinX <= nonTrumpMaxX {
		t.Errorf("trump cards (min x=%d) should be to the right of non-trump cards (max x=%d)",
			trumpMinX, nonTrumpMaxX)
	}
}

func TestEffectiveSuitSymmetry(t *testing.T) {
	cards := []baseui.CardView{
		{Suit: "♠", EffectiveSuit: "黑桃"},
		{Suit: "♥", EffectiveSuit: "红心"},
		{Suit: "♦", EffectiveSuit: "方块"},
		{Suit: "♣", EffectiveSuit: "梅花"},
	}
	for _, c := range cards {
		key := effectiveSuit(c)
		if key == "" {
			t.Errorf("effectiveSuit returned empty for suit=%q effective=%q", c.Suit, c.EffectiveSuit)
		}
	}
	if got := effectiveSuit(baseui.CardView{Suit: "♣", EffectiveSuit: "梅花"}); got != "梅花" {
		t.Errorf("expected 梅花, got %q", got)
	}
	if got := effectiveSuit(baseui.CardView{Suit: "♣", EffectiveSuit: ""}); got != "♣" {
		t.Errorf("expected ♣ fallback, got %q", got)
	}
}
