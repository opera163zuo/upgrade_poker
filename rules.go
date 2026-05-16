package main

// Game rules: trump determination, card comparison, tractor detection, play validation

// IsTrump checks if a card is a trump card
// Trump cards: Jokers, level-rank cards of any suit, cards of trump suit
func IsTrump(card Card, trumpSuit Suit, level Rank) bool {
	// Jokers are always trump
	if card.IsJoker() {
		return true
	}
	// Level rank cards are always trump (e.g., all 10s when playing level 10)
	if card.Rank == level {
		return true
	}
	// Cards of the trump suit are trump
	if card.Suit == trumpSuit {
		return true
	}
	return false
}

// TrumpOrder returns the sort order of a trump card (higher = stronger)
// Order: BigJoker(100) > SmallJoker(90) > MainLevel(80) > OffLevel(70+) > TrumpA(14) > ... > Trump2(2)
func TrumpOrder(card Card, trumpSuit Suit, level Rank) int {
	if !IsTrump(card, trumpSuit, level) {
		return 0
	}

	// Jokers
	if card.IsJoker() {
		if card.Rank == RankBigJoker {
			return 100
		}
		return 90
	}

	// Level rank cards
	if card.Rank == level {
		if card.Suit == trumpSuit {
			return 80 // Main level rank
		}
		// Off-suit level rank: ordered by suit
		return 70 + int(card.Suit)
	}

	// Trump suit cards (not level rank)
	if card.Suit == trumpSuit {
		return int(card.Rank) // 3=3, 4=4, ..., A=14
	}

	return 0 // shouldn't reach here
}

// CompareCards compares two cards in the context of a trick
// Returns: -1 if a < b, 0 if equal, 1 if a > b
// This considers trump status and the lead suit
func CompareCards(a, b Card, trumpSuit Suit, level Rank, leadSuit Suit) int {
	aIsTrump := IsTrump(a, trumpSuit, level)
	bIsTrump := IsTrump(b, trumpSuit, level)

	// Trump always beats non-trump
	if aIsTrump && !bIsTrump {
		return 1
	}
	if !aIsTrump && bIsTrump {
		return -1
	}

	// Both trump: compare by trump order
	if aIsTrump && bIsTrump {
		ao := TrumpOrder(a, trumpSuit, level)
		bo := TrumpOrder(b, trumpSuit, level)
		if ao > bo {
			return 1
		}
		if ao < bo {
			return -1
		}
		// Same order (e.g., two off-suit level ranks of different suits) - same strength
		return 0
	}

	// Both non-trump
	// Must be same suit to compare
	aEffectiveSuit := EffectiveSuit(a, trumpSuit, level)
	bEffectiveSuit := EffectiveSuit(b, trumpSuit, level)

	// Cards of non-lead suit can't beat lead suit
	if aEffectiveSuit != bEffectiveSuit {
		if aEffectiveSuit == leadSuit {
			return 1
		}
		if bEffectiveSuit == leadSuit {
			return -1
		}
		// Both off-suit: can't compare, equal (neither wins)
		return 0
	}

	// Same suit, compare rank (higher rank wins)
	if a.Rank > b.Rank {
		return 1
	}
	if a.Rank < b.Rank {
		return -1
	}
	return 0
}

// EffectiveSuit returns the effective suit of a card for trick purposes
// Trump cards effectively belong to the "trump" suit
func EffectiveSuit(card Card, trumpSuit Suit, level Rank) Suit {
	if IsTrump(card, trumpSuit, level) {
		return trumpSuit
	}
	return card.Suit
}

// CardGroup represents a group of cards (single, pair, or tractor)
type CardGroup struct {
	Cards     []Card
	IsTractor bool
	IsPair    bool
	IsTriple  bool
	IsQuad    bool
	IsSingle  bool
	Suit      Suit // effective suit
}

// AnalyzePlay breaks down a player's played cards into groups
// (singles, pairs, tractors) for comparison with the lead
func AnalyzePlay(cards []Card, trumpSuit Suit, level Rank) []CardGroup {
	if len(cards) == 0 {
		return nil
	}

	// Group by effective suit
	suitGroups := make(map[Suit][]Card)
	for _, c := range cards {
		s := EffectiveSuit(c, trumpSuit, level)
		suitGroups[s] = append(suitGroups[s], c)
	}

	var result []CardGroup
	for suit, group := range suitGroups {
		groups := analyzeSameSuit(group, trumpSuit, level)
		for _, g := range groups {
			g.Suit = suit
			result = append(result, g)
		}
	}

	return result
}

// analyzeSameSuit analyzes cards of the same effective suit into groups
// Supports 4-deck: with 4 copies of the same card (count >= 4), produces
// groups of triples, quads, pairs, or singles depending on card counts.
// Priority: quad > triple > pair > single, with tractors detected at each level.
func analyzeSameSuit(cards []Card, trumpSuit Suit, level Rank) []CardGroup {
	if len(cards) == 0 {
		return nil
	}

	// Count cards by (rank, actual suit)
	type rsKey struct {
		rank Rank
		suit Suit
	}
	rsCount := make(map[rsKey]int)
	for _, c := range cards {
		k := rsKey{c.Rank, c.Suit}
		rsCount[k]++
	}

	// Build slot info for each (rank,suit) with counts
	type slot struct {
		rank  Rank
		suit  Suit
		cnt   int // total cards in this {rank,suit}
	}
	var slots []slot
	for k, cnt := range rsCount {
		if cnt >= 1 {
			slots = append(slots, slot{k.rank, k.suit, cnt})
		}
	}

	// Helper to get unit counts per {rank, suit}
	type baseGroupType int
	const (
		unitSingle baseGroupType = iota
		unitPair
		unitTriple
		unitQuad
	)
	getUnitCount := func(unit baseGroupType, rank Rank, suit Suit) int {
		for _, s := range slots {
			if s.rank == rank && s.suit == suit {
				switch unit {
				case unitQuad: return s.cnt / 4
				case unitTriple: return s.cnt / 3
				case unitPair: return s.cnt / 2
				case unitSingle: return s.cnt
				}
			}
		}
		return 0
	}

	// Collect ranks with each unit type
	collectRanks := func(unit baseGroupType) []Rank {
		seen := make(map[Rank]bool)
		var ranks []Rank
		for _, s := range slots {
			if seen[s.rank] { continue }
			switch unit {
			case unitQuad:
				if s.cnt >= 4 { seen[s.rank] = true; ranks = append(ranks, s.rank) }
			case unitTriple:
				if s.cnt >= 3 { seen[s.rank] = true; ranks = append(ranks, s.rank) }
			case unitPair:
				if s.cnt >= 2 { seen[s.rank] = true; ranks = append(ranks, s.rank) }
			case unitSingle:
				if s.cnt >= 1 { seen[s.rank] = true; ranks = append(ranks, s.rank) }
			}
		}
		return ranks
	}

	var result []CardGroup
	consumedCards := make(map[Card]bool)

	consumeRankCards := func(rank Rank, count int) []Card {
		var out []Card
		for _, c := range cards {
			if c.Rank == rank && !consumedCards[c] && len(out) < count {
				consumedCards[c] = true
				out = append(out, c)
			}
		}
		return out
	}

	// Detect quad-tractors (consecutive quads)
	quadRanks := collectRanks(unitQuad)
	if len(quadRanks) >= 2 {
		sortRanks(quadRanks, trumpSuit, level)
		quadTractors := findTractors(quadRanks, trumpSuit, level)
		for _, tRanks := range quadTractors {
			var tc []Card
			for _, r := range tRanks {
				// Consume 4 cards per quad
				for _, s := range slots {
					if s.rank == r && getUnitCount(unitQuad, r, s.suit) >= 1 {
						tc = append(tc, consumeRankCards(r, 4)...)
						break
					}
				}
			}
			if len(tc) >= 8 {
				result = append(result, CardGroup{Cards: tc, IsTractor: true, IsQuad: true})
			}
		}
	}

	// Detect triple-tractors (consecutive triples)
	tripleRanks := collectRanks(unitTriple)
	if len(tripleRanks) >= 2 {
		sortRanks(tripleRanks, trumpSuit, level)
		tripleTractors := findTractors(tripleRanks, trumpSuit, level)
		for _, tRanks := range tripleTractors {
			var tc []Card
			for _, r := range tRanks {
				// Consume 3 cards per triple
				for _, s := range slots {
					if s.rank == r && getUnitCount(unitTriple, r, s.suit) >= 1 {
						tc = append(tc, consumeRankCards(r, 3)...)
						break
					}
				}
			}
			if len(tc) >= 6 {
				result = append(result, CardGroup{Cards: tc, IsTractor: true, IsTriple: true})
			}
		}
	}

	// Detect pair-tractors (consecutive pairs)
	pairRanks := collectRanks(unitPair)
	if len(pairRanks) >= 2 {
		sortRanks(pairRanks, trumpSuit, level)
		pairTractors := findTractors(pairRanks, trumpSuit, level)
		for _, tRanks := range pairTractors {
			var tc []Card
			for _, r := range tRanks {
				// Consume 2 cards per pair
				for _, s := range slots {
					if s.rank == r && getUnitCount(unitPair, r, s.suit) >= 1 {
						tc = append(tc, consumeRankCards(r, 2)...)
						break
					}
				}
			}
			if len(tc) >= 4 {
				result = append(result, CardGroup{Cards: tc, IsTractor: true})
			}
		}
	}

	// Add remaining quad groups (not consumed by tractors)
	for _, s := range slots {
		numQuads := getUnitCount(unitQuad, s.rank, s.suit)
		for i := 0; i < numQuads; i++ {
			// Check how many unconsumed cards exist first
			unconsumed := 0
			for _, c := range cards {
				if c.Rank == s.rank && !consumedCards[c] && c.Suit == s.suit {
					unconsumed++
				}
			}
			if unconsumed >= 4 {
				cardsOut := consumeRankCards(s.rank, 4)
				if len(cardsOut) >= 4 {
					result = append(result, CardGroup{Cards: cardsOut, IsQuad: true})
				}
			}
		}
	}

	// Add remaining triple groups (not consumed by tractors)
	for _, s := range slots {
		numTriples := getUnitCount(unitTriple, s.rank, s.suit)
		for i := 0; i < numTriples; i++ {
			unconsumed := 0
			for _, c := range cards {
				if c.Rank == s.rank && !consumedCards[c] && c.Suit == s.suit {
					unconsumed++
				}
			}
			if unconsumed >= 3 {
				cardsOut := consumeRankCards(s.rank, 3)
				if len(cardsOut) >= 3 {
					result = append(result, CardGroup{Cards: cardsOut, IsTriple: true})
				}
			}
		}
	}

	// Add remaining pairs (not consumed by tractors/triples/quads)
	for _, s := range slots {
		numPairs := getUnitCount(unitPair, s.rank, s.suit)
		for i := 0; i < numPairs; i++ {
			unconsumed := 0
			for _, c := range cards {
				if c.Rank == s.rank && !consumedCards[c] && c.Suit == s.suit {
					unconsumed++
				}
			}
			if unconsumed >= 2 {
				cardsOut := consumeRankCards(s.rank, 2)
				if len(cardsOut) >= 2 {
					result = append(result, CardGroup{Cards: cardsOut, IsPair: true})
				}
			}
		}
	}

	// Add singles
	for _, c := range cards {
		if !consumedCards[c] {
			result = append(result, CardGroup{Cards: []Card{c}, IsSingle: true})
		}
	}

	return result
}

// findTractors finds consecutive pairs that form tractors
func findTractors(pairRanks []Rank, trumpSuit Suit, level Rank) [][]Rank {
	if len(pairRanks) < 2 {
		return nil
	}

	// Sort by effective rank order
	sortRanks(pairRanks, trumpSuit, level)

	var tractors [][]Rank
	current := []Rank{pairRanks[0]}

	for i := 1; i < len(pairRanks); i++ {
		if areConsecutiveRanks(current[len(current)-1], pairRanks[i], trumpSuit, level) {
			current = append(current, pairRanks[i])
		} else {
			if len(current) >= 2 {
				tractors = append(tractors, append([]Rank{}, current...))
			}
			current = []Rank{pairRanks[i]}
		}
	}
	if len(current) >= 2 {
		tractors = append(tractors, append([]Rank{}, current...))
	}

	return tractors
}

// areConsecutiveRanks checks if two ranks are consecutive for tractor detection.
// The level rank is skipped in the sequence because level-rank cards don't participate
// as normal suit cards. E.g. when level=5, ranks 4 and 6 ARE consecutive
// (5 is the level and not available as a normal suit rank).
func areConsecutiveRanks(a, b Rank, trumpSuit Suit, level Rank) bool {
	if a > b {
		a, b = b, a
	}
	if a == b {
		return false
	}
	// Neither rank can be the level rank itself
	if a == level || b == level {
		return false
	}
	// Advance one step, skipping the level rank if it's in between
	next := a + 1
	if next == level {
		next = level + 1
	}
	return next == b
}

// trumpRankOrder gives a linear order for trump rank adjacency
func trumpRankOrder(r Rank, trumpSuit Suit, level Rank) int {
	if r == level {
		return 20 // Level rank (highest in trump after jokers)
	}
	return int(r)
}

// sortRanks sorts ranks by their trump order (descending)
func sortRanks(ranks []Rank, trumpSuit Suit, level Rank) {
	sortSlice(ranks, func(a, b Rank) bool {
		return trumpRankOrder(a, trumpSuit, level) > trumpRankOrder(b, trumpSuit, level)
	})
}

func sortSlice(ranks []Rank, less func(a, b Rank) bool) {
	for i := 0; i < len(ranks); i++ {
		for j := i + 1; j < len(ranks); j++ {
			if less(ranks[j], ranks[i]) {
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}
}

// getCardsOfRank returns up to n cards of the given rank from the slice
func getCardsOfRank(cards []Card, rank Rank, n int) []Card {
	var result []Card
	for _, c := range cards {
		if c.Rank == rank && len(result) < n {
			result = append(result, c)
		}
	}
	return result
}

// ValidatePlay checks if the played cards are a legal play given the lead and the player's hand
func ValidatePlay(played []Card, lead []Card, hand []Card, allHands [][]Card, trumpSuit Suit, level Rank) bool {
	if len(played) == 0 {
		return false
	}

	// If leading (no lead cards), any valid combination is OK
	if len(lead) == 0 {
		return validateLeading(played, allHands, trumpSuit, level)
	}

	// Following: must follow suit and match the structure of the lead
	return validateFollowing(played, lead, hand, trumpSuit, level)
}

// validateLeading checks if a leading play is valid (甩牌 rules)
func validateLeading(cards []Card, allHands [][]Card, trumpSuit Suit, level Rank) bool {
	// Leading play: all cards must be of the same effective suit
	// (or all trump), and form valid groups (singles, pairs, tractors)

	if len(cards) == 0 {
		return false
	}

	// All cards must be of the same effective suit
	firstSuit := EffectiveSuit(cards[0], trumpSuit, level)
	for _, c := range cards[1:] {
		if EffectiveSuit(c, trumpSuit, level) != firstSuit {
			return false
		}
	}

	// The play must decompose cleanly into singles, pairs, and tractors
	groups := AnalyzePlay(cards, trumpSuit, level)
	totalCards := 0
	for _, g := range groups {
		totalCards += len(g.Cards)
	}
	if totalCards != len(cards) {
		return false
	}

	// Single group (single card, one pair, or one tractor) is always valid
	if len(groups) <= 1 {
		return true
	}

	// Multiple groups: 甩牌 only when every component is currently最大的。
	for _, g := range groups {
		if !isMaxGroup(g, allHands, trumpSuit, level) {
			return false
		}
	}
	return true
}

// validateFollowing checks if a following play is valid
func validateFollowing(played []Card, lead []Card, hand []Card, trumpSuit Suit, level Rank) bool {
	leadSuit := EffectiveSuit(lead[0], trumpSuit, level)

	// Must play the same number of cards as the lead
	if len(played) != len(lead) {
		return false
	}

	// Count lead-suit cards in hand
	leadSuitInHand := 0
	for _, c := range hand {
		if EffectiveSuit(c, trumpSuit, level) == leadSuit {
			leadSuitInHand++
		}
	}

	// Count lead-suit cards in the played set
	leadSuitPlayed := 0
	for _, c := range played {
		if EffectiveSuit(c, trumpSuit, level) == leadSuit {
			leadSuitPlayed++
		}
	}

	// Must use all lead-suit cards from hand if we have fewer than needed
	if leadSuitInHand > 0 {
		if leadSuitPlayed < min(leadSuitInHand, len(lead)) {
			return false
		}
	}

	leadGroups := AnalyzePlay(lead, trumpSuit, level)
	playedGroups := AnalyzePlay(played, trumpSuit, level)

	if len(leadGroups) == 1 && len(playedGroups) == 1 {
		leadGroup := leadGroups[0]
		playedGroup := playedGroups[0]
		if leadGroup.IsSingle && playedGroup.IsSingle {
			return true
		}
		if leadGroup.IsPair && playedGroup.IsPair {
			return true
		}
		if leadGroup.IsTractor && playedGroup.IsTractor && len(leadGroup.Cards) == len(playedGroup.Cards) {
			return true
		}
	}

	// Count tractors and pairs in lead
	leadTractorCount := 0
	leadPairCount := 0
	for _, g := range leadGroups {
		if g.IsTractor {
			leadTractorCount++
			leadPairCount += len(g.Cards) / 2
		} else if g.IsPair {
			leadPairCount++
		}
	}

	playedTractorCount := 0
	playedPairCount := 0
	for _, g := range playedGroups {
		if g.IsTractor {
			playedTractorCount++
			playedPairCount += len(g.Cards) / 2
		} else if g.IsPair {
			playedPairCount++
		}
	}

	if leadSuitInHand > 0 {
		// Must match structure: if lead has pairs, must play pairs if available
		if leadPairCount > 0 {
			availablePairs := countAvailablePairs(hand, leadSuit, trumpSuit, level)
			if availablePairs > 0 && playedPairCount < min(availablePairs, leadPairCount) {
				return false
			}
		}

		// Must match tractor count if possible
		if leadTractorCount > 0 {
			availableTractors := countAvailableTractors(hand, leadSuit, trumpSuit, level)
			if availableTractors > 0 && playedTractorCount < min(availableTractors, leadTractorCount) {
				return false
			}
		}

		return true
	}

	// No lead suit cards: can play any cards
	// But if playing trump (毙牌), must meet minimum tractor and pair counts
	allTrump := true
	for _, c := range played {
		if !IsTrump(c, trumpSuit, level) {
			allTrump = false
			break
		}
	}

	if allTrump {
		// 毙牌: tractor and pair counts must meet or exceed lead's
		if playedTractorCount < leadTractorCount {
			return false
		}
		if playedPairCount < leadPairCount {
			return false
		}
	}

	return true
}

// countAvailablePairs counts pairs available in hand for a given suit
func countAvailablePairs(hand []Card, suit Suit, trumpSuit Suit, level Rank) int {
	rankCount := make(map[Rank]int)
	for _, c := range hand {
		if EffectiveSuit(c, trumpSuit, level) == suit {
			rankCount[c.Rank]++
		}
	}
	count := 0
	for _, c := range rankCount {
		if c >= 2 {
			count++
		}
	}
	return count
}

// countAvailableTractors counts tractors available in hand for a given suit
func countAvailableTractors(hand []Card, suit Suit, trumpSuit Suit, level Rank) int {
	rankCount := make(map[Rank]int)
	for _, c := range hand {
		if EffectiveSuit(c, trumpSuit, level) == suit {
			rankCount[c.Rank]++
		}
	}

	var pairRanks []Rank
	for r, c := range rankCount {
		if c >= 2 {
			pairRanks = append(pairRanks, r)
		}
	}

	if len(pairRanks) < 2 {
		return 0
	}

	tractors := findTractors(pairRanks, trumpSuit, level)
	return len(tractors)
}

// IsKillPlay determines if the played cards constitute a "毙" (kill with trump)
func IsKillPlay(played []Card, lead []Card, trumpSuit Suit, level Rank) bool {
	// All played cards must be trump
	for _, c := range played {
		if !IsTrump(c, trumpSuit, level) {
			return false
		}
	}

	// Lead must be non-trump (if lead is also trump, it's not a "kill")
	for _, c := range lead {
		if !IsTrump(c, trumpSuit, level) {
			return true // At least one non-trump card in lead, and all played are trump
		}
	}

	return false
}

// DetermineTrickWinner determines which player wins the trick
// Returns the index of the winning player
func DetermineTrickWinner(plays [][]Card, trumpSuit Suit, level Rank) int {
	if len(plays) == 0 {
		return 0
	}

	// Find first non-empty play to determine lead suit
	leadIdx := -1
	for i, p := range plays {
		if len(p) > 0 {
			leadIdx = i
			break
		}
	}

	if leadIdx == -1 {
		return 0 // All plays empty, first player wins by default
	}

	leadSuit := EffectiveSuit(plays[leadIdx][0], trumpSuit, level)
	winner := leadIdx

	for i := 0; i < len(plays); i++ {
		if i == leadIdx {
			continue
		}
		if len(plays[i]) == 0 {
			continue
		}
		cmp := comparePlays(plays[i], plays[winner], trumpSuit, level, leadSuit)
		if cmp > 0 {
			winner = i
		}
		// cmp == 0: first player wins (先出者大), keep current winner
	}

	return winner
}

// comparePlays compares two plays to determine which is stronger
// Returns: 1 if a > b, -1 if a < b, 0 if equal
// Rules:
//  1. Trump beats non-trump (毙牌)
//  2. Among same trump/non-trump: higher play type wins (tractor > pair > singles)
//  3. Same type: compare best card
func comparePlays(a, b []Card, trumpSuit Suit, level Rank, leadSuit Suit) int {
	// If either is empty, the non-empty one wins
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	if len(a) == 0 {
		return -1
	}
	if len(b) == 0 {
		return 1
	}

	aIsTrump := isAllTrump(a, trumpSuit, level)
	bIsTrump := isAllTrump(b, trumpSuit, level)

	aType := playType(a, trumpSuit, level)
	bType := playType(b, trumpSuit, level)

	// If both are trump or both non-trump: compare purely by type then best card
	if aIsTrump == bIsTrump {
		if aType != bType {
			if aType > bType {
				return 1
			}
			return -1
		}
		// Same play type: compare best card
		aBest := bestCardInPlay(a, trumpSuit, level, leadSuit)
		bBest := bestCardInPlay(b, trumpSuit, level, leadSuit)
		return CompareCards(aBest, bBest, trumpSuit, level, leadSuit)
	}

	// 毙牌 scenario: one side is trump, the other is non-trump
	// The 毙牌 side must have AT LEAST the same structure type to win.
	// E.g. a single trump card cannot 毙 a non-trump pair.
	if aIsTrump && !bIsTrump {
		// a (trump) is trying to 毙 b (non-trump)
		if aType >= bType {
			return 1
		}
		return -1
	}
	// !aIsTrump && bIsTrump
	if bType >= aType {
		return -1
	}
	return 1
}

// isAllTrump checks if all cards in a play are trump
func isAllTrump(cards []Card, trumpSuit Suit, level Rank) bool {
	for _, c := range cards {
		if !IsTrump(c, trumpSuit, level) {
			return false
		}
	}
	return true
}

// playType returns the type strength of a play:
// 1 = singles (散牌)
// 2 = pair (对子)
// 3 = triple (三张同点)
// 4 = quad (四张同点)
// 5 = pair-tractor (对子拖拉机, consecutive pairs)
// 6 = triple-tractor (三张拖拉机, consecutive triples)
// 7 = quad-tractor (四张拖拉机, consecutive quads)
// For 4-deck: each {suit,rank} with count>=n contributes base units of size n.
func playType(cards []Card, trumpSuit Suit, level Rank) int {
	if len(cards) < 2 {
		return 1
	}

	// Count cards by (rank, actual suit)
	type rsKey struct {
		rank Rank
		suit Suit
	}
	rsCount := make(map[rsKey]int)
	for _, c := range cards {
		k := rsKey{c.Rank, c.Suit}
		rsCount[k]++
	}

	// Determine max consecutive rank sets by triple, quadruple threshold
	// Collect ranks with count>=2, >=3, >=4
	ranksWithPair := make(map[Rank]bool)
	ranksWithTriple := make(map[Rank]bool)
	ranksWithQuad := make(map[Rank]bool)
	for k, cnt := range rsCount {
		if cnt >= 2 {
			ranksWithPair[k.rank] = true
		}
		if cnt >= 3 {
			ranksWithTriple[k.rank] = true
		}
		if cnt >= 4 {
			ranksWithQuad[k.rank] = true
		}
	}
	
	// Helper: check if a set of ranks forms consecutive sequence
	hasConsecutive := func(ranks map[Rank]bool, minLength int) bool {
		var rlist []Rank
		for r := range ranks {
			if r != level { // level rank cards don't form normal sequences
				rlist = append(rlist, r)
			}
		}
		sortRanks(rlist, trumpSuit, level)
		streak := 1
		for i := 1; i < len(rlist); i++ {
			if areConsecutiveRanks(rlist[i-1], rlist[i], trumpSuit, level) {
				streak++
				if streak >= minLength {
					return true
				}
			} else {
				streak = 1
			}
		}
		return false
	}

	// Check quad-tractor (consecutive quads) - highest
	if hasConsecutive(ranksWithQuad, 2) && len(cards) >= 8 {
		return 7
	}

	// Check triple-tractor (consecutive triples)
	if hasConsecutive(ranksWithTriple, 2) && len(cards) >= 6 {
		return 6
	}

	// Check pair-tractor (consecutive pairs)
	if hasConsecutive(ranksWithPair, 2) && len(cards) >= 4 {
		return 5
	}

	// Count how many cards of the most common rank (by actual suit)
	maxSameRank := 0
	for _, cnt := range rsCount {
		if cnt > maxSameRank {
			maxSameRank = cnt
		}
	}

	// Single rank groups
	if maxSameRank >= 4 && len(cards)%4 == 0 && len(cards) <= 4 {
		return 4 // quad (四张)
	}
	if maxSameRank >= 3 && len(cards)%3 == 0 && len(cards) <= 3 {
		return 3 // triple (三张)
	}
	if maxSameRank >= 2 && len(cards) == 2 {
		return 2 // pair
	}

	// Remaining cases: singles or mixed groups
	if maxSameRank >= 2 {
		return 2 // at least pairs
	}
	return 1
}

// bestCardInPlay returns the strongest card in a play
func bestCardInPlay(cards []Card, trumpSuit Suit, level Rank, leadSuit Suit) Card {
	if len(cards) == 0 {
		return Card{}
	}

	best := cards[0]
	for _, c := range cards[1:] {
		if CompareCards(c, best, trumpSuit, level, leadSuit) > 0 {
			best = c
		}
	}
	return best
}

// CalculateTrickPoints calculates the total points in a trick
func CalculateTrickPoints(plays [][]Card) int {
	total := 0
	for _, play := range plays {
		for _, card := range play {
			total += card.Points()
		}
	}
	return total
}

// CalculateBottomMultiplier calculates the bottom card score multiplier
// Based on the last trick's winning play structure.
// Rule: 底牌倍率 = 最后一墩所出牌被分解的组数（每组翻一倍）
// Multiplier = 2 ^ (total number of groups in the winning play).
// Each group — single, pair, triple, quad, or tractor — counts as 1 multiplier unit.
func CalculateBottomMultiplier(winningPlay []Card, trumpSuit Suit, level Rank) int {
	if len(winningPlay) == 0 {
		return 2 // default 2x
	}

	groups := AnalyzePlay(winningPlay, trumpSuit, level)
	numGroups := len(groups)
	if numGroups == 0 {
		return 2
	}

	// Multiplier = 2 ^ numGroups
	result := 1
	for i := 0; i < numGroups; i++ {
		result *= 2
	}
	if result < 2 {
		return 2
	}
	return result
}

// getCardsOfRankSameSuit returns up to `limit` cards of the given rank from the SAME actual suit.
// For 4-deck: when count >= 4, this can return 4+ cards to support multi-pair detection.
// Level-rank cards of different actual suits should not form a pair, so this helper
// ensures pairs are only detected from the same suit.
func getCardsOfRankSameSuit(cards []Card, rank Rank, limit int) []Card {
	if limit <= 0 {
		limit = 2
	}
	// Group by actual suit
	suitCards := make(map[Suit][]Card)
	for _, c := range cards {
		if c.Rank == rank {
			suitCards[c.Suit] = append(suitCards[c.Suit], c)
		}
	}
	// Find a suit with at least 2 cards; return up to limit
	for _, group := range suitCards {
		if len(group) >= 2 {
			n := limit
			if len(group) < n {
				n = len(group)
			}
			return group[:n]
		}
	}
	return nil
}


// isMaxGroup checks if a card group is unbeatable among all hands
func isMaxGroup(g CardGroup, allHands [][]Card, trumpSuit Suit, level Rank) bool {
	// Collect all cards of the same effective suit from all hands
	var otherCards []Card
	for _, hand := range allHands {
		for _, c := range hand {
			if EffectiveSuit(c, trumpSuit, level) == g.Suit {
				otherCards = append(otherCards, c)
			}
		}
	}

	switch {
	case g.IsSingle:
		return isMaxSingleCard(g.Cards[0], otherCards, trumpSuit, level)
	case g.IsPair:
		return isMaxPairCards(g.Cards, otherCards, trumpSuit, level)
	case g.IsTractor:
		return isMaxTractorCards(g.Cards, otherCards, trumpSuit, level)
	case g.IsTriple:
		return isMaxTripleCards(g.Cards, otherCards, trumpSuit, level)
	case g.IsQuad:
		return isMaxQuadCards(g.Cards, otherCards, trumpSuit, level)
	}
	return true
}

// cardRankOrder returns a comparable order for a card within its effective suit
func cardRankOrder(card Card, trumpSuit Suit, level Rank) int {
	if IsTrump(card, trumpSuit, level) {
		return TrumpOrder(card, trumpSuit, level)
	}
	return int(card.Rank)
}

// isMaxSingleCard checks if a single card is the highest of its effective suit
func isMaxSingleCard(card Card, otherCards []Card, trumpSuit Suit, level Rank) bool {
	order := cardRankOrder(card, trumpSuit, level)
	for _, c := range otherCards {
		if cardRankOrder(c, trumpSuit, level) > order {
			return false
		}
	}
	return true
}

// isMaxPairCards checks if a pair is the highest pair of its effective suit
func isMaxPairCards(pairCards []Card, otherCards []Card, trumpSuit Suit, level Rank) bool {
	pairOrder := cardRankOrder(pairCards[0], trumpSuit, level)
	for _, c := range pairCards[1:] {
		if o := cardRankOrder(c, trumpSuit, level); o > pairOrder {
			pairOrder = o
		}
	}

	pairs := findPairsInCards(otherCards, trumpSuit, level)
	for _, p := range pairs {
		otherOrder := cardRankOrder(p[0], trumpSuit, level)
		for _, c := range p[1:] {
			if o := cardRankOrder(c, trumpSuit, level); o > otherOrder {
				otherOrder = o
			}
		}
		if otherOrder > pairOrder {
			return false
		}
	}
	return true
}

// isMaxTripleCards checks if a triple (3 same rank) is the highest triple of its effective suit
func isMaxTripleCards(tripleCards []Card, otherCards []Card, trumpSuit Suit, level Rank) bool {
	ourOrder := cardRankOrder(tripleCards[0], trumpSuit, level)
	for _, c := range tripleCards[1:] {
		if o := cardRankOrder(c, trumpSuit, level); o > ourOrder {
			ourOrder = o
		}
	}

	otherTriples := findTriplesInCards(otherCards, trumpSuit, level)
	for _, t := range otherTriples {
		tOrder := cardRankOrder(t[0], trumpSuit, level)
		for _, c := range t[1:] {
			if o := cardRankOrder(c, trumpSuit, level); o > tOrder {
				tOrder = o
			}
		}
		if tOrder > ourOrder {
			return false
		}
	}
	return true
}

// isMaxQuadCards checks if a quad (4 same rank+suit) is the highest quad of its effective suit
func isMaxQuadCards(quadCards []Card, otherCards []Card, trumpSuit Suit, level Rank) bool {
	ourOrder := cardRankOrder(quadCards[0], trumpSuit, level)
	for _, c := range quadCards[1:] {
		if o := cardRankOrder(c, trumpSuit, level); o > ourOrder {
			ourOrder = o
		}
	}

	otherQuads := findQuadsInCards(otherCards, trumpSuit, level)
	for _, q := range otherQuads {
		qOrder := cardRankOrder(q[0], trumpSuit, level)
		for _, c := range q[1:] {
			if o := cardRankOrder(c, trumpSuit, level); o > qOrder {
				qOrder = o
			}
		}
		if qOrder > ourOrder {
			return false
		}
	}
	return true
}

// isMaxTractorCards checks if a tractor is the highest tractor of its effective suit
func isMaxTractorCards(tractorCards []Card, otherCards []Card, trumpSuit Suit, level Rank) bool {
	ourHighest := cardRankOrder(tractorCards[0], trumpSuit, level)
	for _, c := range tractorCards {
		if o := cardRankOrder(c, trumpSuit, level); o > ourHighest {
			ourHighest = o
		}
	}

	otherTractors := findTractorsInCards(otherCards, trumpSuit, level)
	for _, t := range otherTractors {
		tHighest := cardRankOrder(t[0], trumpSuit, level)
		for _, c := range t {
			if o := cardRankOrder(c, trumpSuit, level); o > tHighest {
				tHighest = o
			}
		}
		if tHighest > ourHighest {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ResolveThrow checks if a leading multi-card play is a valid throw (all groups are max).
// If all groups are unbeatable, returns the cards as-is.
// Otherwise, retains only the safe (unbeatable) components and removes the rest.
// Single groups (single pair, single tractor) are always played as-is.
// Only 甩牌 with multiple groups is checked and potentially reduced.
func ResolveThrow(cards []Card, allHands [][]Card, trumpSuit Suit, level Rank) []Card {
	if len(cards) <= 1 || len(allHands) == 0 {
		return cards
	}

	groups := AnalyzePlay(cards, trumpSuit, level)

	// Single group (one pair, one tractor, one single) — always valid, play as-is
	if len(groups) <= 1 {
		return cards
	}

	// Keep only the groups that are safe (unbeatable)
	var safeCards []Card
	for _, g := range groups {
		if isMaxGroup(g, allHands, trumpSuit, level) {
			safeCards = append(safeCards, g.Cards...)
		}
	}

	// If no safe groups, fall back to the smallest single card
	if len(safeCards) == 0 {
		smallest := cards[0]
		smallestOrder := cardRankOrder(smallest, trumpSuit, level)
		for _, c := range cards[1:] {
			if o := cardRankOrder(c, trumpSuit, level); o < smallestOrder {
				smallest = c
				smallestOrder = o
			}
		}
		return []Card{smallest}
	}

	return safeCards
}
