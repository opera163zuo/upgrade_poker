package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

// GamePhase represents the current phase of the game
type GamePhase int

const (
	PhaseDealing GamePhase = iota
	PhaseBidding
	PhaseDiscarding
	PhasePlaying
	PhaseScoring
	PhaseGameOver
)

// Game represents the overall game state
type Game struct {
	Players     [4]*Player
	Deck        []Card
	BottomCards []Card
	TrumpSuit   Suit
	Level       [2]Rank
	Dealer      PlayerPosition
	CurrentBid  *Bid
	Phase       GamePhase

	CurrentTrick *Trick
	TrickCount   int
	TrickWinner  PlayerPosition

	TeamScore [2]int

	rng *rand.Rand

	ui        baseui.GameUI
	drawOrder []Card // current hand display order (maps index to Card)

	uiPhase           baseui.UIPhase
	uiMessage         string
	uiButtons         []baseui.ButtonSpec
	uiBidChoices      []baseui.BidChoice
	uiWaitingForHuman bool
	uiSelectedIdx     map[int]bool
	uiDiscardCount    int
	uiTrickWinner     PlayerPosition
	uiTrickPoints     int
	uiDealCounts      [4]int
	uiThinkingPos     PlayerPosition
	uiThinking        bool
}

func NewGame(gameUI baseui.GameUI) *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		Level:         [2]Rank{Rank2, Rank2},
		rng:           rng,
		ui:            gameUI,
		uiPhase:       baseui.PhaseWelcome,
		uiSelectedIdx: map[int]bool{},
	}

	g.Players[PositionSouth] = NewPlayer(PositionSouth, true)
	g.Players[PositionWest] = NewPlayer(PositionWest, false)
	g.Players[PositionNorth] = NewPlayer(PositionNorth, false)
	g.Players[PositionEast] = NewPlayer(PositionEast, false)

	g.Dealer = PlayerPosition(rng.Intn(4))

	return g
}

func (g *Game) DealerTeam() Team {
	return PlayerTeam(g.Dealer)
}

func (g *Game) DealerLevel() Rank {
	return g.Level[g.DealerTeam()]
}

func (g *Game) OpponentTeam() Team {
	if g.DealerTeam() == Team0 {
		return Team1
	}
	return Team0
}

func (g *Game) resetForNewMatch() {
	ng := NewGame(g.ui)
	*g = *ng
}

func (g *Game) waitAction() (baseui.UIAction, bool) {
	action := g.ui.WaitAction()
	if action.Type == baseui.ActionRestart {
		return action, true
	}
	return action, false
}

func (g *Game) waitActionOrTimeout(d time.Duration) (baseui.UIAction, bool, bool) {
	action, timedOut := g.ui.WaitActionOrTimeout(d)
	if action.Type == baseui.ActionRestart {
		return action, false, true
	}
	return action, timedOut, false
}

func (g *Game) Deal() {
	g.Deck = NewDeck()
	ShuffleDeck(g.Deck, g.rng)

	for i := range g.Players {
		g.Players[i].Hand = make([]Card, 0)
	}

	idx := 0
	for round := 0; round < 25; round++ {
		for player := 0; player < 4; player++ {
			g.Players[player].AddCard(g.Deck[idx])
			idx++
		}
	}

	g.BottomCards = make([]Card, 8)
	copy(g.BottomCards, g.Deck[idx:idx+8])
}

// DealAnimated deals cards one by one with animation
// DealAnimated deals cards one by one with animation
// During dealing, if human can bid (亮主), pause and ask
func (g *Game) DealAnimated() bool {
	g.Deck = NewDeck()
	ShuffleDeck(g.Deck, g.rng)

	for i := range g.Players {
		g.Players[i].Hand = make([]Card, 0)
	}
	g.uiDealCounts = [4]int{}
	g.render()
	g.setPhase(baseui.PhaseDealing)
	g.CurrentBid = nil

	level := g.DealerLevel()
	humanAsked := false

	idx := 0
	for round := 0; round < 25; round++ {
		for player := 0; player < 4; player++ {
			g.Players[player].AddCard(g.Deck[idx])
			g.uiDealCounts[player]++
			idx++
		}
		g.ui.SleepForRedraw(40 * time.Millisecond)

		// After each round, check if human (South) can bid
		if !humanAsked && round >= 2 {
			human := g.Players[PositionSouth]
			possibleBids := CanBid(human, level)
			var validBids []Bid
			for _, bid := range possibleBids {
				if g.CurrentBid == nil || CanOverrideBid(bid, *g.CurrentBid) {
					validBids = append(validBids, bid)
				}
			}
			if len(validBids) > 0 {
				g.setBidOptions(validBids)
				g.setPhase(baseui.PhaseBidding)
				g.showMessage("请选择亮主方式", nil)

				for {
					action, restarted := g.waitAction()
					if restarted {
						return true
					}
					if action.Type == baseui.ActionBid {
						if bid := g.matchBidAction(validBids, action); bid != nil {
							b := *bid
							b.Player = PositionSouth
							g.CurrentBid = &b
							humanAsked = true
							break
						}
					} else if action.Type == baseui.ActionPass {
						humanAsked = true
						break
					}
				}
				g.showMessage("", nil)
				g.setPhase(baseui.PhaseDealing)
			}
		}

		// AI players check for bid during dealing (only override if higher)
		for p := 0; p < 4; p++ {
			pos := PlayerPosition(p)
			if g.Players[pos].IsHuman {
				continue
			}
			bid := g.aiBidSimple(g.Players[pos], func() []Bid {
				possibleBids := CanBid(g.Players[pos], level)
				var vb []Bid
				for _, b := range possibleBids {
					if g.CurrentBid == nil || CanOverrideBid(b, *g.CurrentBid) {
						vb = append(vb, b)
					}
				}
				return vb
			}())
			if bid != nil {
				g.CurrentBid = bid
			}
		}
	}

	g.BottomCards = make([]Card, 8)
	copy(g.BottomCards, g.Deck[idx:idx+8])

	// Brief pause to show final state
	g.ui.SleepForRedraw(150 * time.Millisecond)
	return false
}

// RunBiddingPhase handles bidding with TUI
func (g *Game) RunBiddingPhase() bool {
	level := g.DealerLevel()

	for i := 0; i < 4; i++ {
		pos := PlayerPosition(i)
		player := g.Players[pos]

		possibleBids := CanBid(player, level)
		var validBids []Bid
		for _, bid := range possibleBids {
			if g.CurrentBid == nil || CanOverrideBid(bid, *g.CurrentBid) {
				validBids = append(validBids, bid)
			}
		}

		if len(validBids) == 0 {
			continue
		}

		if player.IsHuman {
			g.setBidOptions(validBids)
			g.setPhase(baseui.PhaseBidding)
			g.showMessage("请选择亮主方式", nil)

			for {
				action, restarted := g.waitAction()
				if restarted {
					return true
				}
				if action.Type == baseui.ActionBid {
					if bid := g.matchBidAction(validBids, action); bid != nil {
						b := *bid
						g.CurrentBid = &b
						break
					}
				} else if action.Type == baseui.ActionPass {
					break
				}
			}
			g.showMessage("", nil)
		} else {
			// AI bidding logic
			bid := g.aiBidSimple(player, validBids)
			if bid != nil {
				g.CurrentBid = bid
			}
		}
	}

	if g.CurrentBid != nil {
		g.Dealer = g.CurrentBid.Player
	}
	g.TrumpSuit = GetTrumpSuit(g.CurrentBid)

	for _, p := range g.Players {
		p.SortHand(g.TrumpSuit, level)
	}
	return false
}

func (g *Game) aiBidSimple(player *Player, validBids []Bid) *Bid {
	for _, bid := range validBids {
		if bid.Type == BidPairJoker {
			return &bid
		}
	}
	for _, bid := range validBids {
		if bid.Type == BidTripleLevel {
			return &bid
		}
	}
	for _, bid := range validBids {
		if bid.Type == BidPairLevel {
			suitCount := player.CountSuit(bid.Suit, bid.Suit, g.DealerLevel())
			if suitCount >= 4 {
				return &bid
			}
		}
	}
	return nil
}

// DiscardBottom handles the dealer picking up and discarding bottom cards
func (g *Game) DiscardBottom() bool {
	dealer := g.Players[g.Dealer]

	if dealer.IsHuman {
		// Show bottom cards
		g.showMessage(fmt.Sprintf("底牌：\n%s", cardsToString(g.BottomCards)),
			[]baseui.ButtonSpec{{ID: string(baseui.ActionConfirm), Label: "[Enter:确认]", Enabled: true}})
		if _, restarted := g.waitAction(); restarted {
			return true
		}
		g.showMessage("", nil)

		dealer.Hand = append(dealer.Hand, g.BottomCards...)
		dealer.SortHand(g.TrumpSuit, g.DealerLevel())

		// Let human choose 8 cards to discard via UI
		g.uiDiscardCount = 8
		g.setPhase(baseui.PhaseDiscard)

		for {
			action, restarted := g.waitAction()
			if restarted {
				return true
			}
			if action.Type == baseui.ActionPlay && len(action.CardIdx) == 8 {
				var discards []Card
				for _, idx := range action.CardIdx {
					if idx >= 0 && idx < len(g.drawOrder) {
						discards = append(discards, g.drawOrder[idx])
					}
				}
				if len(discards) == 8 {
					dealer.RemoveCards(discards)
					g.BottomCards = discards
					break
				}
			}
		}
		g.uiDiscardCount = 0
		g.setPhase(baseui.PhasePlaying)
	} else {
		dealer.Hand = append(dealer.Hand, g.BottomCards...)
		dealer.SortHand(g.TrumpSuit, g.DealerLevel())
		g.BottomCards = aiDiscard(dealer, g.TrumpSuit, g.DealerLevel())
	}

	dealer.SortHand(g.TrumpSuit, g.DealerLevel())
	return false
}

// PlayTrickFromLead handles one trick
func (g *Game) PlayTrickFromLead(leadPlayer PlayerPosition) (PlayerPosition, bool) {
	level := g.DealerLevel()
	trick := NewTrick(leadPlayer, g.TrumpSuit, level)
	g.CurrentTrick = trick

	currentPlayer := leadPlayer
	for i := 0; i < 4; i++ {
		player := g.Players[currentPlayer]

		var cards []Card
		var restarted bool
		if player.IsHuman {
			g.uiWaitingForHuman = true
			g.render()
			cards, restarted = g.humanPlayTUI(player, trick)
			g.uiWaitingForHuman = false
			if restarted {
				g.render()
				return leadPlayer, true
			}
			g.render()
		} else {
			// Show thinking animation for AI
			g.uiThinkingPos = currentPlayer
			g.uiThinking = true
			g.render()
			g.ui.SleepForRedraw(300 * time.Millisecond)
			g.uiThinking = false
			g.render()
			cards = aiPlay(player, trick, g)
		}

		trick.AddPlay(currentPlayer, cards)
		player.RemoveCards(cards)

		// Clear selection after removing cards to prevent stale selection
		if player.IsHuman {
			g.clearSelection()

		}

		// Pause so player can see each play (shorter for human since they chose)
		if player.IsHuman {
			g.ui.SleepForRedraw(250 * time.Millisecond)
		} else {
			g.ui.SleepForRedraw(750 * time.Millisecond)
		}

		currentPlayer = currentPlayer.Next()
	}

	winner := trick.Winner()
	g.TrickWinner = winner
	g.TrickCount++

	return winner, false
}

// humanPlayTUI handles human play via TUI
func (g *Game) humanPlayTUI(player *Player, trick *Trick) ([]Card, bool) {
	level := g.DealerLevel()
	g.setPhase(baseui.PhasePlaying)

	// Build other players' hands for 甩牌 validation
	otherHands := make([][]Card, 0, 3)
	for i := range g.Players {
		if i != int(player.Position) {
			otherHands = append(otherHands, g.Players[i].Hand)
		}
	}

	for {
		action, restarted := g.waitAction()
		if restarted {
			return nil, true
		}
		if action.Type == baseui.ActionPlay && len(action.CardIdx) > 0 {
			var cards []Card
			for _, idx := range action.CardIdx {
				if idx >= 0 && idx < len(g.drawOrder) {
					cards = append(cards, g.drawOrder[idx])
				}
			}

			var leadCards []Card
			if trick.PlayerCount() > 0 {
				leadCards = trick.LeadCards()
			}

			if ValidatePlay(cards, leadCards, player.Hand, otherHands, g.TrumpSuit, level) {
				// If leading with multiple cards (throw), resolve: only throw if all max, otherwise play smallest
				if len(leadCards) == 0 && len(cards) > 1 {
					cards = ResolveThrow(cards, otherHands, g.TrumpSuit, level)
				}
				g.clearSelection()
				return cards, false
			}
			// Invalid play - show error and reset selection
			g.clearSelection()
			errMsg := g.explainInvalidPlay(cards, leadCards, player.Hand, otherHands, g.TrumpSuit, level)
			g.showMessage(errMsg, []baseui.ButtonSpec{{ID: string(baseui.ActionConfirm), Label: "[Enter:确认]", Enabled: true}})
			if _, restarted := g.waitAction(); restarted {
				return nil, true
			}
			g.showMessage("", nil)
		}
	}
}

// PlayHand plays a complete hand (25 tricks)
func (g *Game) PlayHand() bool {
	g.TrickCount = 0
	g.TeamScore = [2]int{0, 0}

	leadPlayer := g.Dealer

	for g.TrickCount < 25 {
		g.setPhase(baseui.PhasePlaying)
		winner, restarted := g.PlayTrickFromLead(leadPlayer)
		if restarted {
			return true
		}

		points := g.CurrentTrick.Points()
		winnerTeam := PlayerTeam(winner)
		g.TeamScore[winnerTeam] += points

		// Show trick result briefly before continuing
		g.uiTrickWinner = winner
		g.uiTrickPoints = points
		g.uiMessage = fmt.Sprintf("本轮 %s 赢得 %d 分", formatPosition(winner), points)
		g.setPhase(baseui.PhaseWaitTrick)
		g.ui.SleepForRedraw(1500 * time.Millisecond)

		// Clear trick display and continue
		g.uiMessage = ""
		g.CurrentTrick = nil
		leadPlayer = winner
	}

	g.HandleBottomScore()
	if g.HandleUpgrade() {
		return true
	}
	return false
}

// HandleBottomScore handles the bottom card scoring
func (g *Game) HandleBottomScore() {
	lastWinnerTeam := PlayerTeam(g.TrickWinner)
	dealerTeam := g.DealerTeam()

	if lastWinnerTeam != dealerTeam {
		bottomPoints := 0
		for _, c := range g.BottomCards {
			bottomPoints += c.Points()
		}

		if bottomPoints > 0 {
			multiplier := CalculateBottomMultiplier(g.CurrentTrick.Plays[g.TrickWinner], g.TrumpSuit, g.DealerLevel())
			totalBottom := bottomPoints * multiplier
			g.TeamScore[lastWinnerTeam] += totalBottom
		}
	}
}

// HandleUpgrade determines the upgrade result
func (g *Game) HandleUpgrade() bool {
	dealerTeam := g.DealerTeam()
	opponentTeam := g.OpponentTeam()
	opponentScore := g.TeamScore[opponentTeam]

	var upgradeTeam Team
	var upgradeCount int
	newDealer := g.Dealer.Next()

	var resultMsg string

	switch {
	case opponentScore == 0:
		upgradeTeam = dealerTeam
		upgradeCount = 3
		resultMsg = "大光！庄家方连升3级！"
	case opponentScore < 40:
		upgradeTeam = dealerTeam
		upgradeCount = 2
		resultMsg = "庄家方连升2级！"
	case opponentScore < 80:
		upgradeTeam = dealerTeam
		upgradeCount = 1
		resultMsg = "庄家方升1级！"
	case opponentScore < 120:
		upgradeCount = 0
		newDealer = g.Dealer.Next()
		resultMsg = "闲家上台！换庄！"
	case opponentScore < 160:
		upgradeTeam = opponentTeam
		upgradeCount = 1
		newDealer = g.Dealer.Next()
		resultMsg = "闲家上台并升1级！"
	case opponentScore < 200:
		upgradeTeam = opponentTeam
		upgradeCount = 2
		newDealer = g.Dealer.Next()
		resultMsg = "闲家上台并连升2级！"
	default:
		upgradeTeam = opponentTeam
		upgradeCount = 2 + (opponentScore-200)/40
		newDealer = g.Dealer.Next()
		resultMsg = fmt.Sprintf("闲家上台并连升%d级！", upgradeCount)
	}

	if upgradeCount > 0 {
		for i := 0; i < upgradeCount; i++ {
			g.Level[upgradeTeam] = NextLevel(g.Level[upgradeTeam])
		}
		resultMsg += fmt.Sprintf("\n%s方升级至 %s", formatTeam(upgradeTeam), LevelDisplayName(g.Level[upgradeTeam]))
	}

	// Show result via TUI
	msg := fmt.Sprintf("本局结束！\n庄家方（%s）得分：%d\n闲家方（%s）得分：%d\n%s\n下一局庄家：%s",
		formatTeam(dealerTeam), g.TeamScore[dealerTeam],
		formatTeam(opponentTeam), opponentScore,
		resultMsg, formatPosition(newDealer))

	if g.Level[Team0] >= RankGameWon {
		msg += "\n\n南北方获胜！"
		g.Phase = PhaseGameOver
	} else if g.Level[Team1] >= RankGameWon {
		msg += "\n\n东西方获胜！"
		g.Phase = PhaseGameOver
	}

	g.setPhase(baseui.PhaseHandResult)
	g.showMessage(msg, nil)
	_, _, restarted := g.waitActionOrTimeout(5 * time.Second)
	if restarted {
		return true
	}
	g.showMessage("", nil)

	g.Dealer = newDealer
	g.CurrentBid = nil
	g.CurrentTrick = nil
	g.TrumpSuit = SuitJoker
	if g.Phase != PhaseGameOver {
		g.Phase = PhaseDealing
	}
	return false
}

// Run is the main game loop driven by the active UI
func (g *Game) Run() {
	for {
		g.resetForNewMatch()
		g.setPhase(baseui.PhaseWelcome)
		g.showMessage("升级（拖拉机）纸牌游戏\n两副牌 · 从2开始打 · 4人对战",
			[]baseui.ButtonSpec{{ID: string(baseui.ActionStart), Label: "[Enter:开始游戏]", Enabled: true}})
		started := false
		for {
			action, restarted := g.waitAction()
			if restarted {
				break
			}
			if action.Type == baseui.ActionStart {
				started = true
				break
			}
		}
		if !started {
			continue
		}
		g.showMessage("", nil)

		g.Phase = PhaseDealing
		restarted := false
		for g.Phase != PhaseGameOver {
			if g.DealAnimated() {
				restarted = true
				break
			}

			g.setPhase(baseui.PhaseBidding)
			if g.RunBiddingPhase() {
				restarted = true
				break
			}

			g.setPhase(baseui.PhaseDiscard)
			if g.DiscardBottom() {
				restarted = true
				break
			}

			if g.PlayHand() {
				restarted = true
				break
			}
		}
		if restarted {
			continue
		}

		// Game over
		g.setPhase(baseui.PhaseGameOver)
		g.showMessage("游戏结束！\n感谢游玩！", []baseui.ButtonSpec{{ID: string(baseui.ActionConfirm), Label: "[Enter:退出]", Enabled: true}})
		if _, restarted := g.waitAction(); restarted {
			continue
		}
		return
	}
}

func (g *Game) setPhase(phase baseui.UIPhase) {
	g.uiPhase = phase
	g.ui.SetPhase(phase)
	g.render()
}

func (g *Game) showMessage(msg string, buttons []baseui.ButtonSpec) {
	g.uiMessage = msg
	g.uiButtons = buttons
	g.ui.ShowMessage(msg, buttons)
	g.render()
}

func (g *Game) setBidOptions(bids []Bid) {
	g.uiBidChoices = nil
	sorted := append([]Bid(nil), bids...)
	suitOrder := map[Suit]int{SuitSpade: 0, SuitHeart: 1, SuitDiamond: 2, SuitClub: 3, SuitJoker: 4}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		return suitOrder[sorted[i].Suit] < suitOrder[sorted[j].Suit]
	})
	for _, bid := range sorted {
		suitName := bid.Suit.String()
		if bid.Suit == SuitJoker {
			suitName = "无主"
		}
		g.uiBidChoices = append(g.uiBidChoices, baseui.BidChoice{Type: bid.Type.String(), Suit: bid.Suit.String(), Text: bid.Type.String() + " " + suitName})
	}
	g.render()
}

func (g *Game) matchBidAction(validBids []Bid, action baseui.UIAction) *Bid {
	for _, bid := range validBids {
		if bid.Type.String() == action.BidType && bid.Suit.String() == action.BidSuit {
			b := bid
			return &b
		}
	}
	if len(validBids) > 0 {
		b := validBids[0]
		return &b
	}
	return nil
}

func (g *Game) clearSelection() {
	g.uiSelectedIdx = map[int]bool{}
	g.render()
}

func (g *Game) render() {
	if g.ui == nil {
		return
	}
	g.ui.Render(g.UISnapshot())
}

func (g *Game) BuildDrawOrder(pos PlayerPosition) []Card {
	player := g.Players[pos]
	if player == nil {
		return nil
	}
	level := g.DealerLevel()
	player.SortHand(g.TrumpSuit, level)
	drawOrder := make([]Card, len(player.Hand))
	copy(drawOrder, player.Hand)
	return drawOrder
}

func (g *Game) tractorMarks(cards []Card) map[Card]bool {
	marks := map[Card]bool{}
	groups := GroupBySuit(cards, g.TrumpSuit, g.DealerLevel())
	for _, group := range groups {
		for _, tractor := range findTractorsInCards(group, g.TrumpSuit, g.DealerLevel()) {
			for _, c := range tractor {
				marks[c] = true
			}
		}
	}
	return marks
}

func (g *Game) buildHintCardIdx() []int {
	if g.uiPhase != baseui.PhasePlaying || !g.uiWaitingForHuman || g.CurrentTrick == nil {
		return nil
	}
	human := g.Players[PositionSouth]
	if human == nil || len(human.Hand) == 0 {
		return nil
	}
	clone := &Game{TrumpSuit: g.TrumpSuit, Level: g.Level, Dealer: g.Dealer}
	for i, p := range g.Players {
		if p == nil {
			continue
		}
		handCopy := append([]Card(nil), p.Hand...)
		clone.Players[i] = &Player{Position: p.Position, Name: p.Name, IsHuman: p.IsHuman, Hand: handCopy}
	}
	suggested := aiPlay(clone.Players[PositionSouth], g.CurrentTrick, clone)
	if len(suggested) == 0 {
		return nil
	}
	indices := make([]int, 0, len(suggested))
	used := map[int]bool{}
	for _, want := range suggested {
		matched := false
		for idx, card := range g.drawOrder {
			if used[idx] {
				continue
			}
			if card == want {
				indices = append(indices, idx)
				used[idx] = true
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
	}
	sort.Ints(indices)
	return indices
}

func (g *Game) UISnapshot() baseui.TableView {
	view := baseui.TableView{
		Phase:             g.uiPhase,
		TrumpSuit:         g.TrumpSuit.String(),
		Dealer:            formatPosition(g.Dealer),
		DealerLevel:       LevelDisplayName(g.DealerLevel()),
		OpponentLevel:     LevelDisplayName(g.Level[g.OpponentTeam()]),
		TeamScore:         g.TeamScore,
		TrickCount:        g.TrickCount,
		Message:           g.uiMessage,
		Buttons:           append([]baseui.ButtonSpec(nil), g.uiButtons...),
		BidChoices:        append([]baseui.BidChoice(nil), g.uiBidChoices...),
		WaitingForHuman:   g.uiWaitingForHuman,
		SelectedIdx:       map[int]bool{},
		DiscardCount:      g.uiDiscardCount,
		TrickWinner:       formatPosition(g.uiTrickWinner),
		TrickPoints:       g.uiTrickPoints,
		DealCounts:        g.uiDealCounts,
		CurrentTrick:      map[string][]baseui.CardView{},
		CanToggleHandView: true,
	}
	for k, v := range g.uiSelectedIdx {
		view.SelectedIdx[k] = v
	}
	for i, p := range g.Players {
		if p == nil {
			continue
		}
		pv := baseui.PlayerView{
			Position:   formatPosition(p.Position),
			Name:       p.Name,
			IsHuman:    p.IsHuman,
			HandCount:  len(p.Hand),
			IsDealer:   p.Position == g.Dealer,
			IsThinking: g.uiThinking && p.Position == g.uiThinkingPos,
		}
		if p.IsHuman {
			drawOrder := g.BuildDrawOrder(p.Position)
			tractorMarks := g.tractorMarks(drawOrder)
			if p.Position == PositionSouth {
				g.drawOrder = append(g.drawOrder[:0], drawOrder...)
			}
			for _, c := range drawOrder {
				pv.HandCards = append(pv.HandCards, baseui.CardView{
					Label:         c.String(),
					Suit:          c.Suit.Symbol(),
					Rank:          c.Rank.String(),
					SuitNum:       int(c.Suit),
					RankNum:       int(c.Rank),
					FaceUp:        true,
					Trump:         IsTrump(c, g.TrumpSuit, g.DealerLevel()),
					EffectiveSuit: EffectiveSuit(c, g.TrumpSuit, g.DealerLevel()).String(),
					IsTractor:     tractorMarks[c],
				})
			}
		}
		if g.CurrentTrick != nil {
			if cards, ok := g.CurrentTrick.Plays[p.Position]; ok {
				for _, c := range cards {
					card := baseui.CardView{Label: c.String(), Suit: c.Suit.Symbol(), Rank: c.Rank.String(), SuitNum: int(c.Suit), RankNum: int(c.Rank), FaceUp: true, Trump: IsTrump(c, g.TrumpSuit, g.DealerLevel())}
					pv.PlayedCards = append(pv.PlayedCards, card)
					view.CurrentTrick[formatPosition(p.Position)] = append(view.CurrentTrick[formatPosition(p.Position)], card)
				}
			}
		}
		view.Players[i] = pv
	}
	for _, c := range g.BottomCards {
		view.BottomCards = append(view.BottomCards, baseui.CardView{Label: c.String(), Suit: c.Suit.Symbol(), Rank: c.Rank.String(), SuitNum: int(c.Suit), RankNum: int(c.Rank), FaceUp: true, Trump: IsTrump(c, g.TrumpSuit, g.DealerLevel())})
	}
	view.HintCardIdx = g.buildHintCardIdx()
	view.CanHint = len(view.HintCardIdx) > 0
	return view
}

func formatTeam(team Team) string {
	if team == Team0 {
		return "南北"
	}
	return "东西"
}

// cardsToString formats cards for display
func cardsToString(cards []Card) string {
	suitGroups := make(map[Suit][]Card)
	for _, c := range cards {
		suitGroups[c.Suit] = append(suitGroups[c.Suit], c)
	}

	// Sort each group
	for suit := range suitGroups {
		sort.Slice(suitGroups[suit], func(i, j int) bool {
			return suitGroups[suit][i].Rank > suitGroups[suit][j].Rank
		})
	}

	result := ""
	suitOrder := []Suit{SuitSpade, SuitHeart, SuitDiamond, SuitClub, SuitJoker}
	for _, suit := range suitOrder {
		if cards, ok := suitGroups[suit]; ok && len(cards) > 0 {
			result += suit.Symbol() + ": "
			for _, c := range cards {
				result += c.Rank.String() + " "
			}
			result += "\n"
		}
	}
	return result
}

// explainInvalidPlay returns a human-readable explanation of why the play is invalid
func (g *Game) explainInvalidPlay(cards []Card, leadCards []Card, hand []Card, allHands [][]Card, trumpSuit Suit, level Rank) string {
	cardStr := cardsToString(cards)
	if len(cards) == 0 {
		return "请选择要出的牌"
	}

	if len(leadCards) == 0 {
		// Leading
		firstSuit := EffectiveSuit(cards[0], trumpSuit, level)
		for _, c := range cards[1:] {
			if EffectiveSuit(c, trumpSuit, level) != firstSuit {
				return fmt.Sprintf("出的牌花色不一致\n%s\n甩牌必须是同一花色", cardStr)
			}
		}
		groups := AnalyzePlay(cards, trumpSuit, level)
		totalCards := 0
		for _, g2 := range groups {
			totalCards += len(g2.Cards)
		}
		if totalCards != len(cards) {
			return fmt.Sprintf("出的牌不是有效组合\n%s\n需要完整的对子或拖拉机", cardStr)
		}
		if len(groups) > 1 {
			var notMax []string
			for _, g2 := range groups {
				if !isMaxGroup(g2, allHands, trumpSuit, level) {
					label := "单牌"
					if g2.IsPair {
						label = "对子"
					} else if g2.IsTractor {
						label = "拖拉机"
					}
					notMax = append(notMax, fmt.Sprintf("%s(%s)", label, cardsToString(g2.Cards)))
				}
			}
			if len(notMax) > 0 {
				return fmt.Sprintf("不能甩牌 - 不是最大的:\n%s\n以下组合可以被压过:\n%s", cardStr, strings.Join(notMax, ", "))
			}
		}
		return fmt.Sprintf("出牌无效\n%s", cardStr)
	}

	// Following
	if len(cards) != len(leadCards) {
		return fmt.Sprintf("出牌数量不对(需出%d张,选了%d张)", len(leadCards), len(cards))
	}
	leadSuit := EffectiveSuit(leadCards[0], trumpSuit, level)
	leadSuitInHand := 0
	for _, c := range hand {
		if EffectiveSuit(c, trumpSuit, level) == leadSuit {
			leadSuitInHand++
		}
	}
	leadSuitPlayed := 0
	for _, c := range cards {
		if EffectiveSuit(c, trumpSuit, level) == leadSuit {
			leadSuitPlayed++
		}
	}
	if leadSuitInHand > 0 && leadSuitPlayed < min(leadSuitInHand, len(leadCards)) {
		return fmt.Sprintf("必须跟出%s花色的牌(手中有%d张)", leadSuit.Symbol(), leadSuitInHand)
	}

	return fmt.Sprintf("出牌无效\n%s", cardStr)
}
