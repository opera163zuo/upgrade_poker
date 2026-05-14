package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 108 {
		t.Errorf("Expected 108 cards, got %d", len(deck))
	}
}

func TestDeal(t *testing.T) {
	g := NewGame(nil)
	g.Deal()

	totalCards := 0
	for _, p := range g.Players {
		if len(p.Hand) != 25 {
			t.Errorf("Expected 25 cards per player, got %d", len(p.Hand))
		}
		totalCards += len(p.Hand)
	}
	totalCards += len(g.BottomCards)
	if totalCards != 108 {
		t.Errorf("Expected 108 total cards, got %d", totalCards)
	}
	if len(g.BottomCards) != 8 {
		t.Errorf("Expected 8 bottom cards, got %d", len(g.BottomCards))
	}
}

func TestTrumpDetection(t *testing.T) {
	// When playing level 10, heart is trump
	level := Rank10
	trumpSuit := SuitHeart

	// Big joker is always trump
	if !IsTrump(Card{Suit: SuitJoker, Rank: RankBigJoker}, trumpSuit, level) {
		t.Error("Big joker should be trump")
	}
	// Small joker is always trump
	if !IsTrump(Card{Suit: SuitJoker, Rank: RankSmallJoker}, trumpSuit, level) {
		t.Error("Small joker should be trump")
	}
	// Level rank of any suit is trump
	if !IsTrump(Card{Suit: SuitSpade, Rank: Rank10}, trumpSuit, level) {
		t.Error("Level rank of any suit should be trump")
	}
	// 2 is no longer a permanent trump unless it is the current level or trump suit
	if IsTrump(Card{Suit: SuitClub, Rank: Rank2}, trumpSuit, level) {
		t.Error("2 should not be a permanent trump")
	}
	// Trump suit cards are trump
	if !IsTrump(Card{Suit: SuitHeart, Rank: Rank7}, trumpSuit, level) {
		t.Error("Trump suit cards should be trump")
	}
	// Non-trump suit non-level card is not trump
	if IsTrump(Card{Suit: SuitSpade, Rank: Rank7}, trumpSuit, level) {
		t.Error("Non-trump suit non-level card should not be trump")
	}
}

func TestTrumpOrder(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	bigJoker := Card{Suit: SuitJoker, Rank: RankBigJoker}
	smallJoker := Card{Suit: SuitJoker, Rank: RankSmallJoker}
	mainLevel := Card{Suit: SuitHeart, Rank: Rank10}
	offLevel := Card{Suit: SuitSpade, Rank: Rank10}
	trump2 := Card{Suit: SuitHeart, Rank: Rank2}
	trumpA := Card{Suit: SuitHeart, Rank: RankA}

	cases := []struct {
		a, b     Card
		expected int // 1 if a > b
	}{
		{bigJoker, smallJoker, 1},
		{smallJoker, mainLevel, 1},
		{mainLevel, offLevel, 1},
		{offLevel, trumpA, 1},
		{trumpA, trump2, 1},
	}

	for _, c := range cases {
		result := CompareCards(c.a, c.b, trumpSuit, level, trumpSuit)
		if result != c.expected {
			t.Errorf("CompareCards(%v, %v) = %d, expected %d", c.a, c.b, result, c.expected)
		}
	}
}

func TestCompareNonTrump(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	spadeA := Card{Suit: SuitSpade, Rank: RankA}
	spadeK := Card{Suit: SuitSpade, Rank: RankK}
	clubA := Card{Suit: SuitClub, Rank: RankA}

	// Same suit: A > K
	if CompareCards(spadeA, spadeK, trumpSuit, level, SuitSpade) != 1 {
		t.Error("Spade A should beat Spade K")
	}
	// Different non-trump suits: lead suit wins
	if CompareCards(spadeA, clubA, trumpSuit, level, SuitSpade) != 1 {
		t.Error("Lead suit A should beat off-suit A")
	}
	// Trump beats non-trump
	trumpA := Card{Suit: SuitHeart, Rank: RankA}
	if CompareCards(trumpA, spadeA, trumpSuit, level, SuitSpade) != 1 {
		t.Error("Trump should beat non-trump")
	}
}

func TestCardPoints(t *testing.T) {
	c5 := Card{Rank: Rank5}
	if c5.Points() != 5 {
		t.Error("5 should be 5 points")
	}
	c10 := Card{Rank: Rank10}
	if c10.Points() != 10 {
		t.Error("10 should be 10 points")
	}
	ck := Card{Rank: RankK}
	if ck.Points() != 10 {
		t.Error("K should be 10 points")
	}
	ca := Card{Rank: RankA}
	if ca.Points() != 0 {
		t.Error("A should be 0 points")
	}
}

func TestValidateLeadingPlay(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// Single card is always valid
	single := []Card{{Suit: SuitSpade, Rank: RankA}}
	if !ValidatePlay(single, nil, single, nil, trumpSuit, level) {
		t.Error("Single card lead should be valid")
	}

	// Cards of different suits are invalid
	mixed := []Card{
		{Suit: SuitSpade, Rank: RankA},
		{Suit: SuitClub, Rank: RankK},
	}
	if ValidatePlay(mixed, nil, mixed, nil, trumpSuit, level) {
		t.Error("Mixed suit lead should be invalid")
	}
}

func TestAIPlayNoCrash(t *testing.T) {
	// Simulate a full game with 4 AI players to ensure no panics
	rng := rand.New(rand.NewSource(42))

	g := &Game{
		Level: [2]Rank{Rank3, Rank3},
		rng:   rng,
	}
	g.Players[PositionSouth] = NewPlayer(PositionSouth, false) // All AI
	g.Players[PositionWest] = NewPlayer(PositionWest, false)
	g.Players[PositionNorth] = NewPlayer(PositionNorth, false)
	g.Players[PositionEast] = NewPlayer(PositionEast, false)
	g.Dealer = PlayerPosition(rng.Intn(4))

	// Play a few hands
	for hand := 0; hand < 3; hand++ {
		g.Deal()
		g.RunBiddingPhase()
		if g.TrumpSuit == SuitJoker && g.CurrentBid == nil {
			g.TrumpSuit = SuitHeart // fallback for no-trump games in test
		}
		g.DiscardBottom()

		g.TrickCount = 0
		g.TeamScore = [2]int{0, 0}
		leadPlayer := g.Dealer

		for g.TrickCount < 25 {
			level := g.DealerLevel()
			trick := NewTrick(leadPlayer, g.TrumpSuit, level)
			g.CurrentTrick = trick

			currentPlayer := leadPlayer
			for i := 0; i < 4; i++ {
				player := g.Players[currentPlayer]
				var cards []Card
				if trick.PlayerCount() == 0 {
					cards = aiLead(player, g.TrumpSuit, level, g)
				} else {
					cards = aiFollow(player, trick, g.TrumpSuit, level, g)
				}

				// Validate
				var leadCards []Card
				if trick.PlayerCount() > 0 {
					leadCards = trick.LeadCards()
				}
				otherHands := make([][]Card, 0, 3)
				for j := range g.Players {
					if j != int(player.Position) {
						otherHands = append(otherHands, g.Players[j].Hand)
					}
				}
				if !ValidatePlay(cards, leadCards, player.Hand, otherHands, g.TrumpSuit, level) {
					cards = aiSafePlay(player, trick, g.TrumpSuit, level)
				}

				trick.AddPlay(currentPlayer, cards)
				player.RemoveCards(cards)
				currentPlayer = currentPlayer.Next()
			}

			winner := trick.Winner()
			points := trick.Points()
			winnerTeam := PlayerTeam(winner)
			g.TeamScore[winnerTeam] += points
			g.TrickCount++
			leadPlayer = winner
		}

		fmt.Printf("Hand %d: Team0=%d Team1=%d\n", hand, g.TeamScore[0], g.TeamScore[1])
	}
}

func TestThrowCards(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// Player has both copies of A and Q in spades (non-consecutive pairs)
	hand := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}

	// Other players have no spades at all
	otherHands := [][]Card{
		{{Suit: SuitClub, Rank: RankA, Copy: 0}},
		{{Suit: SuitDiamond, Rank: RankA, Copy: 0}},
		{},
	}

	// Throw AA pair + QQ pair - both max since no one else has higher spade pairs
	throw := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	if !ValidatePlay(throw, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Throw AA+QQ should be valid when no one else has higher spade pairs")
	}

	// Other player has both spade Ks - QQ pair is no longer max
	otherHands2 := [][]Card{
		{{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1}},
		{{Suit: SuitDiamond, Rank: RankA, Copy: 0}},
		{},
	}
	if ValidatePlay(throw, nil, hand, otherHands2, trumpSuit, level) {
		t.Error("Throw AA+QQ should be invalid when someone else has KK pair (beats QQ)")
	}

	// Single pair is always valid (no throw check for single group)
	pair := []Card{
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	if !ValidatePlay(pair, nil, hand, otherHands2, trumpSuit, level) {
		t.Error("Single pair lead should always be valid")
	}

	// Single card is always valid
	single := []Card{{Suit: SuitSpade, Rank: RankA, Copy: 0}}
	if !ValidatePlay(single, nil, hand, otherHands2, trumpSuit, level) {
		t.Error("Single card lead should always be valid")
	}

	// Throw singles: A + Q - both must be max
	throwSingles := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
	}
	// Valid when no one has spade A (player has both) and no one has spade > Q
	if !ValidatePlay(throwSingles, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Throw A+Q singles should be valid when no one has higher")
	}
	// Invalid when someone has spade K (> Q)
	if ValidatePlay(throwSingles, nil, hand, otherHands2, trumpSuit, level) {
		t.Error("Throw A+Q singles should be invalid when someone has K (> Q)")
	}
}

func TestValidateFollowingPair(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart
	lead := []Card{{Suit: SuitSpade, Rank: RankQ, Copy: 0}, {Suit: SuitSpade, Rank: RankQ, Copy: 1}}
	hand := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
		{Suit: SuitSpade, Rank: Rank9, Copy: 0},
		{Suit: SuitClub, Rank: RankA, Copy: 0},
	}
	played := []Card{{Suit: SuitSpade, Rank: Rank7, Copy: 0}, {Suit: SuitSpade, Rank: Rank7, Copy: 1}}
	if !ValidatePlay(played, lead, hand, nil, trumpSuit, level) {
		t.Fatal("expected following pair to be a valid play")
	}
}

func TestCanBidIncludesSingleLevel(t *testing.T) {
	player := NewPlayer(PositionWest, false)
	player.Hand = []Card{
		{Suit: SuitHeart, Rank: Rank5},
		{Suit: SuitSpade, Rank: Rank9},
	}

	bids := CanBid(player, Rank5)
	if len(bids) == 0 {
		t.Fatal("expected a single-level bid to be available")
	}
	found := false
	for _, bid := range bids {
		if bid.Type == BidSingleLevel && bid.Suit == SuitHeart {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heart single-level bid, got %#v", bids)
	}
}

func TestRunBiddingPhaseSetsDealerAndTrumpFromAIBid(t *testing.T) {
	g := &Game{Level: [2]Rank{Rank5, Rank5}}
	g.Players[PositionSouth] = NewPlayer(PositionSouth, false)
	g.Players[PositionWest] = NewPlayer(PositionWest, false)
	g.Players[PositionNorth] = NewPlayer(PositionNorth, false)
	g.Players[PositionEast] = NewPlayer(PositionEast, false)
	g.Dealer = PositionSouth
	g.TrumpSuit = SuitJoker

	g.Players[PositionSouth].Hand = []Card{{Suit: SuitClub, Rank: Rank9}}
	g.Players[PositionWest].Hand = []Card{{Suit: SuitHeart, Rank: Rank5}}
	g.Players[PositionNorth].Hand = []Card{{Suit: SuitSpade, Rank: Rank8}}
	g.Players[PositionEast].Hand = []Card{{Suit: SuitDiamond, Rank: Rank7}}

	if restarted := g.RunBiddingPhase(); restarted {
		t.Fatal("unexpected restart during bidding")
	}
	if g.CurrentBid == nil {
		t.Fatal("expected AI to make a bid with a single level card")
	}
	if g.CurrentBid.Type != BidSingleLevel {
		t.Fatalf("expected single-level bid, got %v", g.CurrentBid.Type)
	}
	if g.TrumpSuit != SuitHeart {
		t.Fatalf("expected trump suit heart, got %v", g.TrumpSuit)
	}
	if g.Dealer != PositionWest {
		t.Fatalf("expected dealer west, got %v", g.Dealer)
	}
}

func TestSortOrderTrumpLast(t *testing.T) {
	// Test that trump cards are sorted to the right (last) in CompareForSort
	// and non-trump suits follow 方块→梅花→红桃→黑桃 order

	// Scenario: level 5, Spade is trump
	level := Rank5
	trumpSuit := SuitSpade

	// Create a mixed hand
	cards := []Card{
		{Suit: SuitHeart, Rank: RankA},   // non-trump, should be middle
		{Suit: SuitClub, Rank: RankK},    // non-trump, should be left of Heart
		{Suit: SuitSpade, Rank: Rank7},   // trump suit, should be rightmost group
		{Suit: SuitDiamond, Rank: RankQ}, // non-trump, should be leftmost group
		{Suit: SuitJoker, Rank: RankBigJoker}, // trump, should be rightmost group
		{Suit: SuitClub, Rank: Rank5},    // level card of non-trump suit, is trump
	}

	SortCards(cards, trumpSuit, level)

	// Expected order by groups (left to right):
	// 方块Q | 梅花K | 红心A | 黑桃7 + 梅花5(级牌) + 大王
	// In each group: small to large

	// Find the first trump card index
	firstTrumpIdx := -1
	for i, c := range cards {
		if IsTrump(c, trumpSuit, level) {
			firstTrumpIdx = i
			break
		}
	}
	if firstTrumpIdx < 0 {
		t.Fatal("Expected at least one trump card in sorted order")
	}

	// All non-trump cards must come before first trump
	for i := 0; i < firstTrumpIdx; i++ {
		if IsTrump(cards[i], trumpSuit, level) {
			t.Errorf("Non-trump section at index %d contains trump card %v", i, cards[i])
		}
	}

	// All trump cards must be in the last segment
	for i := firstTrumpIdx; i < len(cards); i++ {
		if !IsTrump(cards[i], trumpSuit, level) {
			t.Errorf("Trump section at index %d contains non-trump card %v", i, cards[i])
		}
	}

	// Verify suit order among non-trump cards
	nonTrumpSuits := make([]Suit, 0)
	for i := 0; i < firstTrumpIdx; i++ {
		nonTrumpSuits = append(nonTrumpSuits, cards[i].Suit)
	}
	// Should appear in order: Diamond → Club → Heart
	expectedNonTrumpOrder := []Suit{SuitDiamond, SuitClub, SuitHeart}
	if len(nonTrumpSuits) != len(expectedNonTrumpOrder) {
		t.Errorf("Expected %d non-trump suits, got %d: %v", len(expectedNonTrumpOrder), len(nonTrumpSuits), nonTrumpSuits)
	} else {
		for i, s := range expectedNonTrumpOrder {
			if nonTrumpSuits[i] != s {
				t.Errorf("Non-trump suit at position %d = %d, expected %d", i, nonTrumpSuits[i], s)
			}
		}
	}
}

func TestNonTrumpDisplayOrder(t *testing.T) {
	// Test that NonTrumpDisplayOrder excludes the trump suit
	// and keeps Diamond→Club→Heart→Spade relative order

	order := NonTrumpDisplayOrder(SuitSpade)
	expected := []Suit{SuitDiamond, SuitClub, SuitHeart}
	if len(order) != len(expected) {
		t.Errorf("Expected %d suits, got %d: %v", len(expected), len(order), order)
	}
	for i, s := range expected {
		if order[i] != s {
			t.Errorf("At index %d: expected %d, got %d", i, s, order[i])
		}
	}

	// Test with heart as trump
	order2 := NonTrumpDisplayOrder(SuitHeart)
	expected2 := []Suit{SuitDiamond, SuitClub, SuitSpade}
	if len(order2) != len(expected2) {
		t.Errorf("Expected %d suits, got %d: %v", len(expected2), len(order2), order2)
	}
	for i, s := range expected2 {
		if order2[i] != s {
			t.Errorf("At index %d: expected %d, got %d", i, s, order2[i])
		}
	}

	// Test with diamond as trump
	order3 := NonTrumpDisplayOrder(SuitDiamond)
	expected3 := []Suit{SuitClub, SuitHeart, SuitSpade}
	if len(order3) != len(expected3) {
		t.Errorf("Expected %d suits, got %d: %v", len(expected3), len(order3), order3)
	}
	for i, s := range expected3 {
		if order3[i] != s {
			t.Errorf("At index %d: expected %d, got %d", i, s, order3[i])
		}
	}
}

// --- Consecutive pair (tractor) tests with level-rank skipping ---

func TestConsecutivePairsSkipLevelRank(t *testing.T) {
	// When level=5, ♥4+♥4 and ♥6+♥6 should form a tractor
	// because the level rank (5) is skipped in the non-trump suit
	level := Rank5
	trumpSuit := SuitSpade

	pair4 := []Card{{Suit: SuitHeart, Rank: Rank4, Copy: 0}, {Suit: SuitHeart, Rank: Rank4, Copy: 1}}
	pair6 := []Card{{Suit: SuitHeart, Rank: Rank6, Copy: 0}, {Suit: SuitHeart, Rank: Rank6, Copy: 1}}

	// 44 + 66 should be detected as a tractor when level=5
	tractorCards := append(pair4, pair6...)
	groups := AnalyzePlay(tractorCards, trumpSuit, level)

	foundTractor := false
	for _, g := range groups {
		if g.IsTractor && len(g.Cards) == 4 {
			foundTractor = true
			break
		}
	}
	if !foundTractor {
		t.Error("44+66 should be detected as a tractor when level=5 (skipping rank 5)")
	}

	// Verify the tractor is a valid leading play
	hand := make([]Card, len(tractorCards))
	copy(hand, tractorCards)
	otherHands := [][]Card{
		{{Suit: SuitClub, Rank: RankA, Copy: 0}},
		{{Suit: SuitDiamond, Rank: RankA, Copy: 0}},
		{},
	}
	if !ValidatePlay(tractorCards, nil, hand, otherHands, trumpSuit, level) {
		t.Error("44+66 tractor should be a valid leading play when level=5")
	}
}

func TestLevelRankItselfNotConsecutiveWithOthers(t *testing.T) {
	// Level-rank pairs should NOT be consecutive with other pairs
	level := Rank5
	trumpSuit := SuitSpade

	pair5 := []Card{{Suit: SuitHeart, Rank: Rank5, Copy: 0}, {Suit: SuitHeart, Rank: Rank5, Copy: 1}}
	pair6 := []Card{{Suit: SuitHeart, Rank: Rank6, Copy: 0}, {Suit: SuitHeart, Rank: Rank6, Copy: 1}}

	// 55 (level rank) + 66 should NOT be a tractor
	cards := append(pair5, pair6...)
	groups := AnalyzePlay(cards, trumpSuit, level)

	for _, g := range groups {
		if g.IsTractor {
			t.Error("55+66 should NOT be a tractor when level=5 (level-rank 5 is not consecutive with 6)")
			return
		}
	}
}

func TestNormalConsecutivePairsUnaffected(t *testing.T) {
	// Normal consecutive pairs (77+88) should still work
	level := Rank5
	trumpSuit := SuitSpade

	pair7 := []Card{{Suit: SuitHeart, Rank: Rank7, Copy: 0}, {Suit: SuitHeart, Rank: Rank7, Copy: 1}}
	pair8 := []Card{{Suit: SuitHeart, Rank: Rank8, Copy: 0}, {Suit: SuitHeart, Rank: Rank8, Copy: 1}}

	cards := append(pair7, pair8...)
	groups := AnalyzePlay(cards, trumpSuit, level)

	foundTractor := false
	for _, g := range groups {
		if g.IsTractor && len(g.Cards) == 4 {
			foundTractor = true
			break
		}
	}
	if !foundTractor {
		t.Error("77+88 should still be a tractor (level=5 does not affect them)")
	}
}

// --- ResolveThrow pair protection tests ---

func TestSinglePairNotResolvedByResolveThrow(t *testing.T) {
	// A single pair should NEVER be resolved (reduced to single card) by ResolveThrow
	level := Rank10
	trumpSuit := SuitHeart

	// Create a pair of ♠Q — not max because other hands have ♠A pairs
	pair := []Card{
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	allHands := [][]Card{
		{{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1}}, // higher pair
		{},
		{},
	}

	result := ResolveThrow(pair, allHands, trumpSuit, level)
	if len(result) != 2 {
		t.Errorf("Single pair should not be resolved, got %d cards instead of 2", len(result))
	}
}

func TestSingleTractorNotResolvedByResolveThrow(t *testing.T) {
	level := Rank5
	trumpSuit := SuitSpade

	// 44+66 tractor when level=5 (now detected as tractor)
	pair4 := []Card{{Suit: SuitHeart, Rank: Rank4, Copy: 0}, {Suit: SuitHeart, Rank: Rank4, Copy: 1}}
	pair6 := []Card{{Suit: SuitHeart, Rank: Rank6, Copy: 0}, {Suit: SuitHeart, Rank: Rank6, Copy: 1}}
	tractorCards := append(pair4, pair6...)

	// Other hands have higher pairs
	allHands := [][]Card{
		{{Suit: SuitHeart, Rank: RankA, Copy: 0}, {Suit: SuitHeart, Rank: RankA, Copy: 1}},
		{},
		{},
	}

	result := ResolveThrow(tractorCards, allHands, trumpSuit, level)
	if len(result) != 4 {
		t.Errorf("Single tractor should not be resolved, got %d cards instead of 4", len(result))
	}
}

func TestMultiGroupThrowStillResolved(t *testing.T) {
	// Multi-group 甩牌 should still be resolved (existing behavior for non-max groups)
	level := Rank10
	trumpSuit := SuitHeart

	// AA pair + QQ pair — QQ is not max if others have KK
	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	allHands := [][]Card{
		{{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1}}, // KK beats QQ
		{},
		{},
	}

	result := ResolveThrow(cards, allHands, trumpSuit, level)
	if len(result) == 4 {
		t.Error("Multi-group throw with non-max components should be resolved to single card")
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 card after resolution, got %d", len(result))
	}
}

// --- Pair validation tests ---

func TestValidatePairLeadingAlwaysValid(t *testing.T) {
	// A simple pair should always be valid when leading,
	// regardless of whether it's max
	level := Rank10
	trumpSuit := SuitHeart

	pair := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
	}
	hand := make([]Card, len(pair))
	copy(hand, pair)
	// Other players have higher pairs, but a single pair lead should still be valid
	otherHands := [][]Card{
		{{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1}},
		{},
		{},
	}
	if !ValidatePlay(pair, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Single pair should always be valid as leading play")
	}
}

func TestValidateTrumpPairLeadingAlwaysValid(t *testing.T) {
	// Trump pair should also be valid when leading
	level := Rank10
	trumpSuit := SuitHeart

	pair := []Card{
		{Suit: SuitHeart, Rank: RankA, Copy: 0},
		{Suit: SuitHeart, Rank: RankA, Copy: 1},
	}
	hand := make([]Card, len(pair))
	copy(hand, pair)
	otherHands := [][]Card{
		{{Suit: SuitHeart, Rank: RankK, Copy: 0}, {Suit: SuitHeart, Rank: RankK, Copy: 1}},
		{},
		{},
	}
	if !ValidatePlay(pair, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Trump pair should always be valid as leading play")
	}
}

func TestValidateLevelPairLeading(t *testing.T) {
	// Level-rank pair should also be valid
	level := Rank5
	trumpSuit := SuitHeart

	// Two level-rank cards from different suits form a trump pair
	pair := []Card{
		{Suit: SuitSpade, Rank: Rank5, Copy: 0},
		{Suit: SuitClub, Rank: Rank5, Copy: 1},
	}
	hand := make([]Card, len(pair))
	copy(hand, pair)
	otherHands := [][]Card{
		{{Suit: SuitHeart, Rank: RankK, Copy: 0}},
		{},
		{},
	}
	if !ValidatePlay(pair, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Level-rank pair should be valid as leading play")
	}
}

// --- 大小王对子 tests ---

func TestJokerPairs(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// 大王+大王 should be a valid pair
	bigJokerPair := []Card{
		{Suit: SuitJoker, Rank: RankBigJoker, Copy: 0},
		{Suit: SuitJoker, Rank: RankBigJoker, Copy: 1},
	}
	groups := AnalyzePlay(bigJokerPair, trumpSuit, level)
	foundPair := false
	for _, g := range groups {
		if g.IsPair && len(g.Cards) == 2 {
			foundPair = true
			break
		}
	}
	if !foundPair {
		t.Error("大王+大王 should be detected as a pair")
	}

	// 小王+小王 should be a valid pair
	smallJokerPair := []Card{
		{Suit: SuitJoker, Rank: RankSmallJoker, Copy: 0},
		{Suit: SuitJoker, Rank: RankSmallJoker, Copy: 1},
	}
	groups = AnalyzePlay(smallJokerPair, trumpSuit, level)
	foundPair = false
	for _, g := range groups {
		if g.IsPair && len(g.Cards) == 2 {
			foundPair = true
			break
		}
	}
	if !foundPair {
		t.Error("小王+小王 should be detected as a pair")
	}

	// 大王+小王 should NOT be a pair
	mixedJoker := []Card{
		{Suit: SuitJoker, Rank: RankBigJoker, Copy: 0},
		{Suit: SuitJoker, Rank: RankSmallJoker, Copy: 1},
	}
	groups = AnalyzePlay(mixedJoker, trumpSuit, level)
	for _, g := range groups {
		if g.IsPair {
			t.Error("大王+小王 should NOT be detected as a pair")
			return
		}
	}
}

func TestValidateBigJokerPairLeading(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	pair := []Card{
		{Suit: SuitJoker, Rank: RankBigJoker, Copy: 0},
		{Suit: SuitJoker, Rank: RankBigJoker, Copy: 1},
	}
	hand := make([]Card, len(pair))
	copy(hand, pair)
	otherHands := [][]Card{
		{},
		{},
		{},
	}
	if !ValidatePlay(pair, nil, hand, otherHands, trumpSuit, level) {
		t.Error("大王对 should be a valid leading play")
	}
}
