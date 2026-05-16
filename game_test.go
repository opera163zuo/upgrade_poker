package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 216 {
		t.Errorf("Expected 216 cards, got %d", len(deck))
	}
}

func TestDeal(t *testing.T) {
	g := NewGame(nil)
	g.Deal()

	totalCards := 0
	for _, p := range g.Players {
		if len(p.Hand) != 52 {
			t.Errorf("Expected 52 cards per player, got %d", len(p.Hand))
		}
		totalCards += len(p.Hand)
	}
	totalCards += len(g.BottomCards)
	if totalCards != 216 {
		t.Errorf("Expected 216 total cards, got %d", totalCards)
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

		for g.TrickCount < 52 {
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


// --- 牌型压制修复 tests ---

func TestPairNotBeatenByDiffSuitSameRankTrump(t *testing.T) {
	// 主牌是方块，用户出一对方块7
	// 对家出红桃2 + 方块2（其中2是级牌时，不同花色的两张级牌）
	// 系统应判用户更大，因为红桃2+方块2不算对子
	level := Rank2
	trumpSuit := SuitDiamond

	// 用户出一对方块7
	lead := []Card{
		{Suit: SuitDiamond, Rank: Rank7, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank7, Copy: 1},
	}

	// 对家出红桃2 + 方块2 — 不同花色同Rank = 不算对子
	response := []Card{
		{Suit: SuitHeart, Rank: Rank2, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank2, Copy: 1},
	}

	// Verify: analyzeSameSuit should NOT detect response as a pair
	responseGroups := AnalyzePlay(response, trumpSuit, level)
	isPair := false
	for _, g := range responseGroups {
		if g.IsPair {
			isPair = true
			break
		}
	}
	if isPair {
		t.Error("红桃2+方块2 should NOT be detected as a pair (different actual suits)")
	}

	// Verify: comparePlays should say lead wins
	leadSuit := EffectiveSuit(lead[0], trumpSuit, level)
	cmp := comparePlays(lead, response, trumpSuit, level, leadSuit)
	if cmp <= 0 {
		t.Errorf("comparePlays(方块7对, 红桃2+方块2) = %d, expected >0 (lead should win)", cmp)
	}
}

func TestKillNonTrumpPairNeedsTrumpPair(t *testing.T) {
	// 首出是一对非主牌对子（如梅花7）
	// 毙掉它需要一对主牌，不能用两张不同的主牌
	level := Rank2
	trumpSuit := SuitDiamond

	// 首出：梅花7对（非主牌）
	lead := []Card{
		{Suit: SuitClub, Rank: Rank7, Copy: 0},
		{Suit: SuitClub, Rank: Rank7, Copy: 1},
	}

	// 毙牌尝试：红桃2 + 方块7（都是主牌但不同花色不同Rank — 不是对子）
	badKill := []Card{
		{Suit: SuitHeart, Rank: Rank2, Copy: 0}, // off-level, trump
		{Suit: SuitDiamond, Rank: Rank7, Copy: 0}, // trump suit
	}

	// Compare: lead should win because badKill is not a pair
	leadSuit := EffectiveSuit(lead[0], trumpSuit, level)
	cmp := comparePlays(badKill, lead, trumpSuit, level, leadSuit)
	if cmp >= 0 {
		t.Error("Two different trump cards (non-pair) should NOT beat a non-trump pair")
	}

	// 正确的毙牌：方块7对（一对主牌）
	goodKill := []Card{
		{Suit: SuitDiamond, Rank: Rank7, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank7, Copy: 1},
	}
	cmp = comparePlays(goodKill, lead, trumpSuit, level, leadSuit)
	if cmp <= 0 {
		t.Error("A genuine trump pair should beat a non-trump pair")
	}
}

func TestPairRequiresSameActualSuit(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// Same rank but different suits — should NOT be a pair
	cards := []Card{
		{Suit: SuitSpade, Rank: Rank10, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank10, Copy: 1},
	}
	groups := AnalyzePlay(cards, trumpSuit, level)
	for _, g := range groups {
		if g.IsPair {
			t.Errorf("Spade10 + Diamond10 should NOT be a pair (different suits, both off-level rank=%d)", level)
		}
	}

	// Same rank AND same suit — SHOULD be a pair
	cards2 := []Card{
		{Suit: SuitDiamond, Rank: Rank10, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank10, Copy: 1},
	}
	groups2 := AnalyzePlay(cards2, trumpSuit, level)
	isPair := false
	for _, g := range groups2 {
		if g.IsPair {
			isPair = true
			break
		}
	}
	if !isPair {
		t.Error("Diamond10 + Diamond10 should be detected as a pair (same suit)")
	}
}

func TestPlayTypeRequiresSameActualSuit(t *testing.T) {
	level := Rank2
	trumpSuit := SuitDiamond

	// Two cards of same rank, different suits — playType should NOT be pair (2)
	cards := []Card{
		{Suit: SuitHeart, Rank: Rank2, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank2, Copy: 1},
	}
	pt := playType(cards, trumpSuit, level)
	if pt == 2 {
		t.Error("playType(红桃2+方块2) should NOT return 2 (pair) — different suits")
	}
}

func TestFindPairsInCardsRequiresSameActualSuit(t *testing.T) {
	level := Rank2
	trumpSuit := SuitDiamond

	cards := []Card{
		{Suit: SuitHeart, Rank: Rank2, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank2, Copy: 1},
	}
	pairs := findPairsInCards(cards, trumpSuit, level)
	if len(pairs) > 0 {
		t.Error("findPairsInCards should NOT find pair for 红桃2+方块2 (different actual suits)")
	}

	// Same actual suit should still work
	cards2 := []Card{
		{Suit: SuitDiamond, Rank: Rank2, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank2, Copy: 1},
	}
	pairs2 := findPairsInCards(cards2, trumpSuit, level)
	if len(pairs2) == 0 {
		t.Error("findPairsInCards should find pair for 方块2+方块2 (same actual suit)")
	}
}

func TestComparePlaysPairVsTwoLevelRankCards(t *testing.T) {
	// Scenario from bug report: trump=Diamond, lead=一对7(Diamond),
	// response=红桃2+方块2 (two different-suit level-rank cards when level=2)
	level := Rank2
	trumpSuit := SuitDiamond

	lead := []Card{
		{Suit: SuitDiamond, Rank: Rank7, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank7, Copy: 1},
	}
	response := []Card{
		{Suit: SuitHeart, Rank: Rank2, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank2, Copy: 1},
	}

	leadSuit := EffectiveSuit(lead[0], trumpSuit, level)
	cmp := comparePlays(lead, response, trumpSuit, level, leadSuit)
	if cmp <= 0 {
		t.Errorf("Lead pair should beat two level-rank cards of different suits: comparePlays=%d", cmp)
	}

	// Also test that the winner is correctly determined
	plays := make([][]Card, 4)
	plays[0] = lead
	plays[1] = response
	// Two passes for the other players
	plays[2] = nil
	plays[3] = nil
	winner := DetermineTrickWinner(plays, trumpSuit, level)
	if winner != 0 {
		t.Errorf("Player 0 (lead pair) should win, got winner=%d", winner)
	}
}

func TestFollowingNotForceMax(t *testing.T) {
	// validateFollowing should accept ANY valid following play,
	// not force the player to play the maximum card
	level := Rank10
	trumpSuit := SuitHeart

	// Lead: single spade K
	lead := []Card{{Suit: SuitSpade, Rank: RankK, Copy: 0}}

	// Hand has spade 7 and spade A
	hand := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitDiamond, Rank: Rank5, Copy: 0},
	}

	// Play spade 7 (minimum, not max) — should be valid
	playedMin := []Card{{Suit: SuitSpade, Rank: Rank7, Copy: 0}}
	if !ValidatePlay(playedMin, lead, hand, nil, trumpSuit, level) {
		t.Error("Playing spade 7 (not max) to follow spade K should be valid")
	}

	// Play spade A (max) — also valid
	playedMax := []Card{{Suit: SuitSpade, Rank: RankA, Copy: 0}}
	if !ValidatePlay(playedMax, lead, hand, nil, trumpSuit, level) {
		t.Error("Playing spade A (max) to follow spade K should also be valid")
	}
}

func TestValidateFollowingPairNotForceMax(t *testing.T) {
	// When following a pair, should be able to play any valid pair, not necessarily max
	level := Rank10
	trumpSuit := SuitHeart

	// Lead: spade K pair
	lead := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
	}

	// Hand has spade 7 pair and spade A pair
	hand := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
	}

	// Play spade 7 pair (not max) — should be valid
	playedMin := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
	}
	if !ValidatePlay(playedMin, lead, hand, nil, trumpSuit, level) {
		t.Error("Playing spade 7 pair (not max) to follow spade K pair should be valid")
	}
}

// --- 4-Deck Multi-Pair Tests ---

func TestMultiPairDetectionSameRankSuit(t *testing.T) {
	// 4 copies of the same card (same rank + same suit) should be detected as 2 pairs
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 2},
		{Suit: SuitSpade, Rank: RankA, Copy: 3},
	}

	groups := AnalyzePlay(cards, trumpSuit, level)

	pairCount := 0
	singleCount := 0
	for _, g := range groups {
		if g.IsPair {
			pairCount++
		}
		if g.IsSingle {
			singleCount++
		}
	}
	if pairCount != 2 {
		t.Errorf("Expected 2 pair groups from 4 copies of same card, got %d", pairCount)
	}
	if singleCount != 0 {
		t.Errorf("Expected 0 singles from 4 copies of same card, got %d", singleCount)
	}
	totalGroupCards := 0
	for _, g := range groups {
		totalGroupCards += len(g.Cards)
	}
	if totalGroupCards != 4 {
		t.Errorf("Expected total 4 cards in groups, got %d", totalGroupCards)
	}
}

func TestMultiPairLeadValid(t *testing.T) {
	// Leading with 2 pairs (4 cards of same rank+suit) should be valid
	level := Rank10
	trumpSuit := SuitHeart

	// 4 copies of SpadeA = 2 pairs
	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 2},
		{Suit: SuitSpade, Rank: RankA, Copy: 3},
	}

	hand := make([]Card, len(cards))
	copy(hand, cards)
	otherHands := [][]Card{
		{{Suit: SuitClub, Rank: RankK, Copy: 0}},
		{{Suit: SuitDiamond, Rank: RankK, Copy: 0}},
		{},
	}

	if !ValidatePlay(cards, nil, hand, otherHands, trumpSuit, level) {
		t.Error("Leading with 2 pairs (4 same cards) should be valid")
	}

	// playType should detect 2 pairs
	pt := playType(cards, trumpSuit, level)
	if pt != 2 {
		t.Errorf("playType(4 copies) = %d, expected 2 (pair level, 2 pairs)", pt)
	}
}

func TestMultiPairNotTractor(t *testing.T) {
	// 4 copies of same card (same rank+suit) should NOT be a tractor
	// (need consecutive ranks for tractor)
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 2},
		{Suit: SuitSpade, Rank: RankA, Copy: 3},
	}

	groups := AnalyzePlay(cards, trumpSuit, level)
	for _, g := range groups {
		if g.IsTractor {
			t.Error("4 copies of same rank+suit should NOT be detected as a tractor")
			return
		}
	}
}

func TestMultiPairFollowStructure(t *testing.T) {
	// Leading 2 pairs (4 cards), follower with 4 copies of same rank should match
	level := Rank10
	trumpSuit := SuitHeart

	// Lead: 2 pairs of SpadeK
	lead := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 2},
		{Suit: SuitSpade, Rank: RankK, Copy: 3},
	}

	// Follower has 4 Spade7 = 2 pairs of Spade7
	hand := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
		{Suit: SuitSpade, Rank: Rank7, Copy: 2},
		{Suit: SuitSpade, Rank: Rank7, Copy: 3},
	}

	// Play all 4 Spade7 — should be valid following play (same count, same suit)
	played := []Card{
		{Suit: SuitSpade, Rank: Rank7, Copy: 0},
		{Suit: SuitSpade, Rank: Rank7, Copy: 1},
		{Suit: SuitSpade, Rank: Rank7, Copy: 2},
		{Suit: SuitSpade, Rank: Rank7, Copy: 3},
	}

	if !ValidatePlay(played, lead, hand, nil, trumpSuit, level) {
		t.Error("Following 2-pair lead with 2 pairs should be valid")
	}
}

func TestMultiPairSeparateGroups(t *testing.T) {
	// 2 pairs of SpadeA + normal pair of SpadeK:
	// A and K are consecutive ranks, so this forms 1 tractor (KK AA) + 1 remaining pair (AA)
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 2},
		{Suit: SuitSpade, Rank: RankA, Copy: 3},
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
	}

	groups := AnalyzePlay(cards, trumpSuit, level)

	totalCards := 0
	tractorCount := 0
	pairCount := 0
	for _, g := range groups {
		totalCards += len(g.Cards)
		if g.IsTractor {
			tractorCount++
		}
		if g.IsPair {
			pairCount++
		}
	}
	// A and K are consecutive: 1 tractor (KK AA) + 1 remaining pair (AA) = 2 groups, 6 cards
	if tractorCount != 1 {
		t.Errorf("Expected 1 tractor from consecutive A+K pairs, got %d", tractorCount)
	}
	if pairCount != 1 {
		t.Errorf("Expected 1 remaining pair of SpadeA, got %d", pairCount)
	}
	if totalCards != 6 {
		t.Errorf("Expected 6 total cards in groups, got %d", totalCards)
	}
}


func TestFindPairsMultiCard(t *testing.T) {
	// Test that findPairsInCards returns 2 pairs for 4 copies
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 2},
		{Suit: SuitSpade, Rank: RankA, Copy: 3},
	}

	pairs := findPairsInCards(cards, trumpSuit, level)
	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs from 4 copies of same card, got %d", len(pairs))
	}
	for i, p := range pairs {
		if len(p) != 2 {
			t.Errorf("Pair %d has %d cards, expected 2", i, len(p))
		}
	}

	// Verify all 4 cards are covered and no duplicates
	seen := make(map[Card]bool)
	for _, p := range pairs {
		for _, c := range p {
			if seen[c] {
				t.Errorf("Card %v appears in multiple pairs", c)
			}
			seen[c] = true
		}
	}
	if len(seen) != 4 {
		t.Errorf("Expected 4 unique cards across pairs, got %d", len(seen))
	}
}
