package main

import "testing"

// ===================== E-1: 牌型/拖拉机/毙牌单元测试 =====================

func TestE1_ThreeCardGroupType(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// 3 same rank/suit = triple
	cards := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 2},
	}
	groups := AnalyzePlay(cards, trumpSuit, level)
	if len(groups) != 1 {
		t.Errorf("Expected 1 group for 3 same cards, got %d", len(groups))
	}
	if !groups[0].IsTriple {
		t.Errorf("3 same cards should be IsTriple=true, got IsTriple=%v", groups[0].IsTriple)
	} else {
		t.Log("OK: 3 same cards → IsTriple=true")
	}

	// 4 same rank/suit = quad
	cards4 := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 2},
		{Suit: SuitSpade, Rank: RankK, Copy: 3},
	}
	groups4 := AnalyzePlay(cards4, trumpSuit, level)
	if len(groups4) != 1 {
		t.Errorf("Expected 1 group for 4 same cards, got %d", len(groups4))
	}
	if !groups4[0].IsQuad {
		t.Errorf("4 same cards should be IsQuad=true, got IsQuad=%v", groups4[0].IsQuad)
	} else {
		t.Log("OK: 4 same cards → IsQuad=true")
	}
}

func TestE1_TractorDetection(t *testing.T) {
	level := Rank5
	trumpSuit := SuitHeart

	// 3 cards (1 pair + 1 single) should NOT be a tractor
	cards3 := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
	}
	groups3 := AnalyzePlay(cards3, trumpSuit, level)
	foundTractor := false
	for _, g := range groups3 {
		if g.IsTractor {
			foundTractor = true
		}
	}
	if foundTractor {
		t.Error("FAIL: 3 cards (pair+single) should NOT be a tractor")
	} else {
		t.Log("OK: 3 cards NOT a tractor")
	}

	// K+K+Q+Q (4 cards) should be a tractor when level=5 — K and Q consecutive skipping 5
	cards4 := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	groups4 := AnalyzePlay(cards4, trumpSuit, level)
	foundTractor = false
	for _, g := range groups4 {
		if g.IsTractor && len(g.Cards) == 4 {
			foundTractor = true
		}
	}
	if !foundTractor {
		t.Error("FAIL: K+K+Q+Q should be a tractor when level=5 (K=13, Q=12, consecutive skipping 5)")
	} else {
		t.Log("OK: K+K+Q+Q detected as tractor (K and Q consecutive when level=5)")
	}

	// 4+4+6+6 should be a tractor when level=5 (4 and 6 consecutive skipping 5)
	cards46 := []Card{
		{Suit: SuitSpade, Rank: Rank4, Copy: 0},
		{Suit: SuitSpade, Rank: Rank4, Copy: 1},
		{Suit: SuitSpade, Rank: Rank6, Copy: 0},
		{Suit: SuitSpade, Rank: Rank6, Copy: 1},
	}
	groups46 := AnalyzePlay(cards46, trumpSuit, level)
	foundTractor = false
	for _, g := range groups46 {
		if g.IsTractor && len(g.Cards) == 4 {
			foundTractor = true
		}
	}
	if !foundTractor {
		t.Error("FAIL: 4+4+6+6 should be a tractor when level=5")
	} else {
		t.Log("OK: 4+4+6+6 detected as tractor")
	}

	// A+A+J+J should NOT be a tractor (A=14, J=11, not consecutive)
	cardsAJ := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankJ, Copy: 0},
		{Suit: SuitSpade, Rank: RankJ, Copy: 1},
	}
	groupsAJ := AnalyzePlay(cardsAJ, trumpSuit, level)
	foundTractor = false
	for _, g := range groupsAJ {
		if g.IsTractor {
			foundTractor = true
		}
	}
	if foundTractor {
		t.Error("FAIL: K+K+J+J should NOT be a tractor (K=13, J=11, not consecutive)")
	} else {
		t.Log("OK: K+K+J+J correctly NOT a tractor")
	}

	// K+K+A+A should be a tractor when level=5 (K=13, A=14, consecutive)
	cardsKA := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankA, Copy: 0},
		{Suit: SuitSpade, Rank: RankA, Copy: 1},
	}
	groupsKA := AnalyzePlay(cardsKA, trumpSuit, level)
	foundTractor = false
	for _, g := range groupsKA {
		if g.IsTractor && len(g.Cards) == 4 {
			foundTractor = true
		}
	}
	if !foundTractor {
		t.Error("FAIL: K+K+A+A should be a tractor when level=5 (K=13, A=14 consecutive)")
	} else {
		t.Log("OK: K+K+A+A detected as tractor")
	}
}

func TestE1_QuadVsPairComparison(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// 4-of-a-kind should beat lower rank 4-of-a-kind
	cardsK := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 2},
		{Suit: SuitSpade, Rank: RankK, Copy: 3},
	}
	cardsQ := []Card{
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 2},
		{Suit: SuitSpade, Rank: RankQ, Copy: 3},
	}

	cmp := comparePlays(cardsK, cardsQ, trumpSuit, level, SuitSpade)
	if cmp <= 0 {
		t.Errorf("FAIL: K quad should beat Q quad, got cmp=%d", cmp)
	} else {
		t.Log("OK: K quad beats Q quad")
	}
}

func TestE1_KillPlayRules(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// Lead: pair tractor (KK+QQ in spades)
	lead := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0},
		{Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}

	// 4 trump cards (quad) should be able to 毙 (kill) a pair tractor
	hand := []Card{
		{Suit: SuitHeart, Rank: RankA, Copy: 0},
		{Suit: SuitHeart, Rank: RankA, Copy: 1},
		{Suit: SuitHeart, Rank: RankA, Copy: 2},
		{Suit: SuitHeart, Rank: RankA, Copy: 3},
	}
	otherHands := [][]Card{{}, {}}

	valid := ValidatePlay(hand, lead, hand, otherHands, trumpSuit, level)
	t.Logf("4 trump cards (quad) vs pair tractor lead: valid=%v", valid)

	// Lead: 3 cards (triple) — can 4 trump cards kill it?
	// The rule says: pair count and tractor count must match.
	// A 3-card lead has 0 pairs, 0 tractors. 4 trump cards has 0 pairs, 0 tractors.
	// So technically valid as allTrump path: playedTractorCount >= leadTractorCount ✓
	lead3 := []Card{
		{Suit: SuitSpade, Rank: RankK, Copy: 0},
		{Suit: SuitSpade, Rank: RankK, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 2},
	}
	valid3 := ValidatePlay(hand, lead3, hand, otherHands, trumpSuit, level)
	t.Logf("4 trump cards vs 3-card lead (triple): valid=%v", valid3)
}

// ===================== E-2: 甩牌处罚回归测试 =====================

func TestE2_SingleTractorNotResolved(t *testing.T) {
	level := Rank5
	trumpSuit := SuitHeart

	pair4 := []Card{{Suit: SuitHeart, Rank: Rank4, Copy: 0}, {Suit: SuitHeart, Rank: Rank4, Copy: 1}}
	pair6 := []Card{{Suit: SuitHeart, Rank: Rank6, Copy: 0}, {Suit: SuitHeart, Rank: Rank6, Copy: 1}}
	tractorCards := append(pair4, pair6...)

	allHands := [][]Card{{}, {}, {}}

	result := ResolveThrow(tractorCards, allHands, trumpSuit, level)
	if len(result) != 4 {
		t.Errorf("FAIL: Single tractor should not be resolved, got %d cards instead of 4", len(result))
	} else {
		t.Log("OK: Single tractor not resolved")
	}
}

func TestE2_SinglePairNotResolved(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	pair := []Card{{Suit: SuitSpade, Rank: RankQ, Copy: 0}, {Suit: SuitSpade, Rank: RankQ, Copy: 1}}
	allHands := [][]Card{{}, {}}

	result := ResolveThrow(pair, allHands, trumpSuit, level)
	if len(result) != 2 {
		t.Errorf("FAIL: Single pair should not be resolved, got %d cards instead of 2", len(result))
	} else {
		t.Log("OK: Single pair not resolved")
	}
}

func TestE2_MultiGroupThrowResolved(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0}, {Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	allHands := [][]Card{
		{{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1}},
		{}, {},
	}

	result := ResolveThrow(cards, allHands, trumpSuit, level)
	if len(result) == 4 {
		t.Error("FAIL: Multi-group throw with non-max components should be resolved")
	}
	if len(result) != 2 {
		t.Errorf("FAIL: Expected 2 cards (AA safe) after resolution, got %d", len(result))
	} else {
		t.Log("OK: Multi-group throw resolved to safe AA only")
	}
}

func TestE2_ThrowWithAllMaxKeepsAll(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1},
	}
	allHands := [][]Card{{}, {}, {}}

	result := ResolveThrow(cards, allHands, trumpSuit, level)
	if len(result) != 4 {
		t.Errorf("FAIL: All-max multi-group throw should keep all 4 cards, got %d", len(result))
	} else {
		t.Log("OK: All-max throw keeps all 4 cards")
	}
}

func TestE2_MultiGroupSomeMaxSomeNot(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	cards := []Card{
		{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
		{Suit: SuitSpade, Rank: RankQ, Copy: 0}, {Suit: SuitSpade, Rank: RankQ, Copy: 1},
	}
	allHands := [][]Card{
		{{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1}},
		{}, {},
	}

	result := ResolveThrow(cards, allHands, trumpSuit, level)
	if len(result) != 2 {
		t.Errorf("FAIL: Expected 2 cards (AA only), got %d", len(result))
	}
	for _, c := range result {
		if c.Rank != RankA {
			t.Errorf("FAIL: Expected RankA cards, got Rank%v", c.Rank)
		}
	}
	t.Log("OK: Only max groups retained")
}

// ===================== E-3: 底牌倍率验证 =====================

func TestE3_BottomMultiplier(t *testing.T) {
	level := Rank10
	trumpSuit := SuitHeart

	// NOTE: Multiplier = 2^(number of non-single CardGroups in winning play).
	// Each pair/triple/quad/tractor counts as 1 unit.
	// Tractor detection depends on level: with level=10, A(14) and K(13) ARE consecutive (skipping 10),
	// but 4→6 is NOT consecutive because level-10 ≠ 5, so rank 5 is between them.
	tests := []struct {
		name     string
		cards    []Card
		expected int
	}{
		{"Single card", []Card{{Suit: SuitSpade, Rank: RankA, Copy: 0}}, 2},
		{"One pair", []Card{{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1}}, 2},
		// AA+KK (A=14, K=13): consecutive when level=10 → 1 tractor group → 2x
		{"Two pairs merge into tractor", []Card{
			{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
			{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1},
		}, 2},
		// AA+KK+QQ (Q=12,K=13,A=14): all consecutive when level=10 → 1 tractor group → 2x
		{"Three pairs merge into tractor", []Card{
			{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
			{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1},
			{Suit: SuitSpade, Rank: RankQ, Copy: 0}, {Suit: SuitSpade, Rank: RankQ, Copy: 1},
		}, 2},
		{"Empty play", []Card{}, 2},
		// 4+4+6+6: with level=10, 4→6 NOT consecutive → 2 separate pair groups → 4x
		{"Two non-consecutive pairs", []Card{
			{Suit: SuitSpade, Rank: Rank4, Copy: 0}, {Suit: SuitSpade, Rank: Rank4, Copy: 1},
			{Suit: SuitSpade, Rank: Rank6, Copy: 0}, {Suit: SuitSpade, Rank: Rank6, Copy: 1},
		}, 4},
		// 4+4+6+6+8+8+9+9: pair(4)+pair(6)+tractor(8,9) = 3 groups → 8x
		{"Two pairs + one pair-tractor", []Card{
			{Suit: SuitSpade, Rank: Rank4, Copy: 0}, {Suit: SuitSpade, Rank: Rank4, Copy: 1},
			{Suit: SuitSpade, Rank: Rank6, Copy: 0}, {Suit: SuitSpade, Rank: Rank6, Copy: 1},
			{Suit: SuitSpade, Rank: Rank8, Copy: 0}, {Suit: SuitSpade, Rank: Rank8, Copy: 1},
			{Suit: SuitSpade, Rank: Rank9, Copy: 0}, {Suit: SuitSpade, Rank: Rank9, Copy: 1},
		}, 8},
		// A (single, ignored) + KK (pair) → 1 unit → 2x
		{"Single + pair", []Card{
			{Suit: SuitSpade, Rank: RankA, Copy: 0},
			{Suit: SuitSpade, Rank: RankK, Copy: 0}, {Suit: SuitSpade, Rank: RankK, Copy: 1},
		}, 2},
		{"One quad", []Card{
			{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
			{Suit: SuitSpade, Rank: RankA, Copy: 2}, {Suit: SuitSpade, Rank: RankA, Copy: 3},
		}, 2},
		{"One triple", []Card{
			{Suit: SuitSpade, Rank: RankA, Copy: 0}, {Suit: SuitSpade, Rank: RankA, Copy: 1},
			{Suit: SuitSpade, Rank: RankA, Copy: 2},
		}, 2},
	}

	for _, tt := range tests {
		m := CalculateBottomMultiplier(tt.cards, trumpSuit, level)
		if m != tt.expected {
			t.Errorf("FAIL: %s: expected %dx, got %dx", tt.name, tt.expected, m)
		} else {
			t.Logf("OK: %s → %dx", tt.name, m)
		}
	}
}
