package gui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

var (
	bgColor             = color.RGBA{0x16, 0x5f, 0x3b, 0xff}
	tableColor          = color.RGBA{0x0d, 0x4b, 0x2f, 0xff}
	cardFaceColor       = color.RGBA{0xf6, 0xf1, 0xe7, 0xff}
	cardBackColor       = color.RGBA{0x2c, 0x4f, 0x94, 0xff}
	outlineColor        = color.RGBA{0x22, 0x22, 0x22, 0xff}
	hiliteColor         = color.RGBA{0xff, 0xd6, 0x54, 0xff}
	tractorColor        = color.RGBA{0x6e, 0xc7, 0xff, 0xa0}
	trumpGroupColor     = color.RGBA{0x12, 0x25, 0x42, 0xb8}
	suitGroupColor      = color.RGBA{0x0d, 0x17, 0x22, 0x74}
	messageBgColor      = color.RGBA{0x0d, 0x17, 0x22, 0xdd}
	disabledColor       = color.RGBA{0x6b, 0x6b, 0x6b, 0xff}
	secondaryButtonFill = color.RGBA{0x85, 0x9f, 0xa3, 0xff}
)

type southSlot struct {
	idx int
	x   int
	y   int
	w   int
	h   int
}

func (g *GUI) drawText(dst *ebiten.Image, s string, x, y int, clr color.Color) {
	text.Draw(dst, s, uiFontFace(), x, y, clr)
}

func (g *GUI) drawTextBadge(dst *ebiten.Image, s string, x, y, padX, padY int, fill color.Color, border color.Color, textColor color.Color) {
	width := len([]rune(s))*12 + padX*2
	height := 18 + padY*2
	top := y - 14 - padY
	vector.DrawFilledRect(dst, float32(x-padX), float32(top), float32(width), float32(height), fill, false)
	vector.StrokeRect(dst, float32(x-padX), float32(top), float32(width), float32(height), 1, border, false)
	g.drawText(dst, s, x, y, textColor)
}

func (g *GUI) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)
	g.st.mu.Lock()
	view := g.st.view
	g.st.cardRects = nil
	g.st.buttonRects = nil
	selected := map[int]bool{}
	for k, v := range g.st.selected {
		selected[k] = v
	}
	if view.Phase == baseui.PhaseBidding {
		g.ensureBidSelectionLocked()
	}
	g.st.mu.Unlock()

	vector.DrawFilledRect(screen, float32(TableX), float32(TableY), float32(TableW), float32(TableH), tableColor, false)
	vector.StrokeRect(screen, float32(TableX), float32(TableY), float32(TableW), float32(TableH), 2, outlineColor, false)
	vector.DrawFilledRect(screen, float32(InfoBarX), float32(InfoBarY), float32(InfoBarW), float32(InfoBarH), color.RGBA{0x12, 0x2a, 0x1d, 0xff}, false)
	vector.StrokeRect(screen, float32(InfoBarX), float32(InfoBarY), float32(InfoBarW), float32(InfoBarH), 1, outlineColor, false)
	g.drawInfoBar(screen, view, selected)

	g.drawNorth(screen, view)
	g.drawWest(screen, view)
	g.drawEast(screen, view)
	g.drawCenter(screen, view)
	g.drawSouth(screen, view, selected)
	g.drawMenuBar(screen, view)
	g.drawButtons(screen, view, selected)
	g.drawOverlay(screen, view, selected)
}

func (g *GUI) drawInfoBar(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	leftW := 246
	midW := 176
	rightW := 144
	g.drawStatusBox(screen, InfoBarX+6, InfoBarY+6, leftW, 30, color.RGBA{0x17, 0x35, 0x26, 0xff}, fmt.Sprintf("庄家 %s · 主花色 %s", view.Dealer, view.TrumpSuit), color.White)
	g.drawLevelPair(screen, InfoBarX+16+leftW, InfoBarY+6, midW, 30, view.DealerLevel, view.OpponentLevel)
	g.drawStatusBox(screen, InfoBarX+26+leftW+midW, InfoBarY+6, rightW, 30, color.RGBA{0x18, 0x24, 0x34, 0xff}, fmt.Sprintf("得分 %d : %d", view.TeamScore[0], view.TeamScore[1]), hiliteColor)
	line2 := fmt.Sprintf("第 %d / 25 轮 · 手牌 南%d 西%d 北%d 东%d · %s", view.TrickCount+1, view.Players[0].HandCount, view.Players[1].HandCount, view.Players[2].HandCount, view.Players[3].HandCount, g.statusLine(view, selected))
	g.drawTextBadge(screen, line2, InfoBarX+12, InfoBarY+41, 8, 4, color.RGBA{0x0e, 0x1a, 0x12, 0xcc}, color.RGBA{0x6b, 0x88, 0x73, 0xff}, color.RGBA{0xd5, 0xe5, 0xd5, 0xff})
}

func (g *GUI) drawStatusBox(screen *ebiten.Image, x, y, w, h int, fill color.Color, label string, textColor color.Color) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), fill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{0xe9, 0xe9, 0xe9, 0xff}, false)
	g.drawText(screen, label, x+10, y+h/2+4, textColor)
}

func (g *GUI) drawLevelPair(screen *ebiten.Image, x, y, w, h int, ourLevel, oppLevel string) {
	half := (w - 10) / 2
	g.drawStatusBox(screen, x, y, half, h, color.RGBA{0xf0, 0xf4, 0xfc, 0xff}, "我方 "+ourLevel, color.RGBA{0x2d, 0x65, 0xc4, 0xff})
	g.drawStatusBox(screen, x+half+10, y, half, h, color.RGBA{0xfb, 0xf1, 0xf1, 0xff}, "敌方 "+oppLevel, color.RGBA{0xc8, 0x39, 0x39, 0xff})
}

func (g *GUI) statusLine(view baseui.TableView, selected map[int]bool) string {
	switch view.Phase {
	case baseui.PhaseDealing:
		return "发牌中... 若拿到级牌可立即亮主"
	case baseui.PhaseBidding:
		return "请选择亮主花色，然后点击“继续”确认 / P 不亮"
	case baseui.PhaseDiscard:
		return fmt.Sprintf("请垫底牌（已选 %d/8） / Enter 扣底", len(selected))
	case baseui.PhasePlaying:
		if view.WaitingForHuman {
			hint := ""
			if view.CanHint {
				hint = " / H 提示"
			}
			return fmt.Sprintf("请你出牌（已选 %d 张）/ 双击或 Enter 出牌%s", len(selected), hint)
		}
		return "等待其他玩家出牌..."
	case baseui.PhaseWaitTrick:
		if view.TrickWinner != "" {
			return fmt.Sprintf("本轮 %s 赢得 %d 分", view.TrickWinner, view.TrickPoints)
		}
	}
	return strings.ReplaceAll(strings.TrimSpace(view.Message), "\n", " ")
}

func (g *GUI) drawNorth(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[2]
	g.drawSeatLabel(screen, pv, 272, 68)
	for i := 0; i < pv.HandCount && i < 12; i++ {
		x := 440 - NorthHandGap*i
		g.drawCardWithAlpha(screen, x, NorthHandY, baseui.CardView{Label: "", FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawWest(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[1]
	g.drawSeatLabel(screen, pv, 10, 138)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := WestHandY + WestHandGap*i
		g.drawCardWithAlpha(screen, WestHandX, y, baseui.CardView{FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawEast(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[3]
	g.drawSeatLabel(screen, pv, 520, 138)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := EastHandY - EastHandGap*i
		g.drawCardWithAlpha(screen, EastHandX, y, baseui.CardView{FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawSeatLabel(screen *ebiten.Image, pv baseui.PlayerView, x, y int) {
	label := pv.Name
	if pv.IsDealer {
		label += " ★"
	}
	if pv.IsThinking {
		dots := strings.Repeat(".", int(time.Now().UnixMilli()/350)%3+1)
		label += " 思考中" + dots
	}
	g.drawTextBadge(screen, label, x+8, y, 8, 3, color.RGBA{0x0d, 0x17, 0x22, 0xc8}, color.RGBA{0x9f, 0xb8, 0xc4, 0xff}, color.White)
}

func (g *GUI) drawCenter(screen *ebiten.Image, view baseui.TableView) {
	positions := []struct {
		key string
		x   int
		y   int
	}{{"上(AI)", 250, 150},
		{"左(AI)", 176, 210},
		{"右(AI)", 388, 210},
		{"下(你)", 250, 268},
	}
	for _, p := range positions {
		cards := view.CurrentTrick[p.key]
		for i, c := range cards {
			g.drawCardWithAlpha(screen, p.x+i*15, p.y, c, false, 1)
		}
	}
	showBottom := len(view.BottomCards) > 0 && (view.Phase == baseui.PhaseDiscard || view.Phase == baseui.PhaseHandResult)
	if showBottom {
		for i, c := range view.BottomCards {
			g.drawCardWithAlpha(screen, BottomX+i*BottomGap, BottomY, c, false, 1)
		}
	}
}

func (g *GUI) drawSouth(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	pv := view.Players[0]
	biddingRaise := map[int]bool{}
	if view.Phase == baseui.PhaseBidding {
		biddingRaise = g.biddingRaisedCards(view)
	}
	g.drawText(screen, pv.Name, 236, 328, color.White)
	var slots []southSlot
	slots = g.southSlots(pv.HandCards, selected, biddingRaise, view.TrumpSuit)
	g.drawSouthGroupBackdrops(screen, pv.HandCards, slots, view.TrumpSuit)
	for _, slot := range slots {
		c := pv.HandCards[slot.idx]
		g.drawCardWithAlpha(screen, slot.x, slot.y, c, selected[slot.idx] || biddingRaise[slot.idx], 1)
	}
	rects := make([]rect, len(slots))
	for i, slot := range slots {
		rects[i] = rect{x: slot.x, y: slot.y, w: slot.w, h: slot.h}
	}
	g.st.mu.Lock()
	g.st.cardRects = rects
	g.st.mu.Unlock()
}

func (g *GUI) southSlots(cards []baseui.CardView, selected map[int]bool, biddingRaise map[int]bool, trumpSuit string) []southSlot {
	order := g.suitRowOrder(cards, trumpSuit)

	// Group card indices by effective suit
	suitIndices := map[string][]int{}
	for idx, c := range cards {
		key := effectiveSuit(c)
		suitIndices[key] = append(suitIndices[key], idx)
	}

	// Build flat ordered indices: trump group first, then other suits
	var ordered []int
	for _, suit := range order {
		ordered = append(ordered, suitIndices[suit]...)
	}
	if len(ordered) == 0 {
		return nil
	}

	const withinSuitGap = 15		// was 10, 1.5x
	const betweenSuitGap = 22		// was 15, 1.5x

	// Calculate total width for centering the entire hand
	totalWidth := CardW
	for i := 1; i < len(ordered); i++ {
		g := withinSuitGap
		if effectiveSuit(cards[ordered[i-1]]) != effectiveSuit(cards[ordered[i]]) {
			g = betweenSuitGap
		}
		totalWidth += g
	}

	startX := TableX + (TableW-totalWidth)/2
	if startX < SouthHandX {
		startX = SouthHandX
	}

	y := 335				// was 340, moved up for taller cards
	var slots []southSlot

	for pos, idx := range ordered {
		var slotX int
		if pos == 0 {
			slotX = startX
		} else {
			g := withinSuitGap
			if effectiveSuit(cards[ordered[pos-1]]) != effectiveSuit(cards[idx]) {
				g = betweenSuitGap
			}
			slotX = slots[pos-1].x + g
		}

		slotY := y
		if selected[idx] || biddingRaise[idx] {
			slotY -= 18				// was 12, 1.5x
		}

		slots = append(slots, southSlot{idx: idx, x: slotX, y: slotY, w: CardW, h: CardH})
	}

	sort.Slice(slots, func(i, j int) bool { return slots[i].idx < slots[j].idx })
	return slots
}

func effectiveSuit(c baseui.CardView) string {
	if c.EffectiveSuit != "" {
		return c.EffectiveSuit
	}
	return c.Suit
}

func (g *GUI) suitRowOrder(cards []baseui.CardView, trumpSuit string) []string {
	base := []string{}
	seen := map[string]bool{}
	push := func(s string) {
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		base = append(base, s)
	}
	// Non-trump suits in fixed order: 方块→梅花→红桃→黑桃
	for _, s := range []string{"方块", "梅花", "红心", "黑桃"} {
		if s != trumpSuit {
			push(s)
		}
	}
	// Trump suit goes last (rightmost)
	push(trumpSuit)
	// Any remaining suits not yet seen (e.g. 王 in no-trump mode)
	for _, c := range cards {
		key := c.EffectiveSuit
		if key == "" {
			key = c.Suit
		}
		push(key)
	}
	return base
}

func (g *GUI) drawSouthGroupBackdrops(screen *ebiten.Image, cards []baseui.CardView, slots []southSlot, trumpSuit string) {
	if len(slots) == 0 {
		return
	}
	bySuit := map[string][]southSlot{}
	for _, slot := range slots {
		key := cards[slot.idx].EffectiveSuit
		if key == "" {
			key = cards[slot.idx].Suit
		}
		bySuit[key] = append(bySuit[key], slot)
	}
	for _, suit := range g.suitRowOrder(cards, trumpSuit) {
		group := bySuit[suit]
		if len(group) == 0 {
			continue
		}
		minX, maxX, y := group[0].x, group[0].x+CardW, group[0].y
		for _, slot := range group[1:] {
			if slot.x < minX {
				minX = slot.x
			}
			if slot.x+CardW > maxX {
				maxX = slot.x + CardW
			}
			if slot.y < y {
				y = slot.y
			}
		}
		fill := suitGroupColor
		var labelColor color.Color = color.White
		if suit == trumpSuit {
			fill = trumpGroupColor
			labelColor = hiliteColor
		}
		isTrumpSuit := suit == trumpSuit
		vector.DrawFilledRect(screen, float32(minX-26), float32(y+22), float32(maxX-minX+34), 22, fill, false)
		// Trump label on the RIGHT of the group, non-trump labels on the LEFT
		var labelX int
		if isTrumpSuit {
			labelX = maxX + 4
		} else {
			labelX = minX - 20
		}
		g.drawText(screen, bidSuitSymbol(suit, isTrumpSuit), labelX, y+38, labelColor)
	}
}

// bidSuitSymbol returns the display symbol for a suit group label.
// For trump: returns "主牌".
// For non-trump: returns Chinese suit name characters (not Unicode symbols)
// to ensure the text renders correctly even when the font lacks Unicode suit glyphs.
func bidSuitSymbol(suit string, isTrump bool) string {
	if isTrump {
		return "主牌"
	}
	switch suit {
	case "黑桃":
		return "黑"
	case "红心":
		return "红"
	case "方块":
		return "方"
	case "梅花":
		return "梅"
	case "王":
		return "王"
	default:
		return suit
	}
}

func (g *GUI) drawMenuBar(screen *ebiten.Image, view baseui.TableView) {
	vector.DrawFilledRect(screen, 0, 0, LogicalWidth, MenuBarH, color.RGBA{0xe7, 0xe7, 0xe7, 0xff}, false)
	vector.StrokeRect(screen, 0, 0, LogicalWidth, MenuBarH, 1, color.RGBA{0x88, 0x88, 0x88, 0xff}, false)
	startEnabled := view.Phase == baseui.PhaseWelcome || view.Phase == baseui.PhaseGameOver
	restartEnabled := view.Phase != baseui.PhaseWelcome
	g.drawTopMenuButton(screen, 8, 3, 92, 18, "开始游戏", baseui.UIAction{Type: baseui.ActionStart}, startEnabled)
	g.drawTopMenuButton(screen, 106, 3, 92, 18, "重新开始", baseui.UIAction{Type: baseui.ActionRestart}, restartEnabled)
}

func (g *GUI) drawTopMenuButton(screen *ebiten.Image, x, y, w, h int, label string, action baseui.UIAction, enabled bool) {
	fill := color.RGBA{0xff, 0xff, 0xff, 0xff}
	var textColor color.Color = color.Black
	if !enabled {
		fill = color.RGBA{0xd4, 0xd4, 0xd4, 0xff}
		textColor = color.RGBA{0x66, 0x66, 0x66, 0xff}
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), fill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, outlineColor, false)
	g.drawText(screen, label, x+12, y+h/2+4, textColor)
	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{rect: rect{x: x, y: y, w: w, h: h}, action: action, enabled: enabled})
	g.st.mu.Unlock()
}

func (g *GUI) drawButtons(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	if view.Phase == baseui.PhaseWelcome && len(view.Buttons) > 0 {
		btn := view.Buttons[0]
		g.drawActionButton(screen, LogicalWidth/2-120, LogicalHeight/2+60, 240, 50, btn.Label, baseui.UIAction{Type: baseui.ActionType(btn.ID)}, btn.Enabled, false)
		return
	}

	if view.Phase == baseui.PhaseBidding {
		g.drawBidButtons(screen, view)
		return
	}

	if view.Phase == baseui.PhasePlaying && view.WaitingForHuman {
		g.drawActionButton(screen, 120, ActionBtnY, ActionBtnW, ActionBtnH, "提示", baseui.UIAction{Type: baseui.ActionHint}, view.CanHint, true)
		g.drawActionButton(screen, 226, ActionBtnY, ActionBtnW, ActionBtnH, "出牌", baseui.UIAction{Type: baseui.ActionPlay}, len(selected) > 0, false)
		g.drawActionButton(screen, 332, ActionBtnY, ActionBtnW, ActionBtnH, "取消", baseui.UIAction{Type: baseui.ActionCancel}, len(selected) > 0, true)

	}

	if view.Phase == baseui.PhaseDiscard {

		g.drawActionButton(screen, 254, ActionBtnY, ActionBtnW, ActionBtnH, "扣底", baseui.UIAction{Type: baseui.ActionPlay}, len(selected) == view.DiscardCount, false)
		g.drawActionButton(screen, 360, ActionBtnY, ActionBtnW, ActionBtnH, "取消", baseui.UIAction{Type: baseui.ActionCancel}, len(selected) > 0, true)
	}

	for i, b := range view.Buttons {
		x := 430 + i*100
		if x+ActionBtnW > LogicalWidth {
			x = 430
		}
		g.drawActionButton(screen, x, ActionBtnY, ActionBtnW, ActionBtnH, sanitizeButtonLabel(b.Label), baseui.UIAction{Type: baseui.ActionType(b.ID)}, b.Enabled, false)
	}
}



func sanitizeButtonLabel(label string) string {
	label = strings.TrimSpace(label)
	if idx := strings.LastIndex(label, ":"); idx >= 0 && idx+1 < len(label) {
		return strings.TrimSuffix(strings.TrimSpace(label[idx+1:]), "]")
	}
	return strings.Trim(label, "[]")
}

func (g *GUI) bidSuitButtons(view baseui.TableView) []baseui.BidChoice {
	choices := make([]baseui.BidChoice, 0, 4)
	seen := map[string]bool{}
	for _, suit := range []string{"黑桃", "红心", "方块", "梅花"} {
		best, ok := g.bestBidChoiceForSuit(view.BidChoices, suit)
		if !ok {
			continue
		}
		if seen[best.Suit] {
			continue
		}
		seen[best.Suit] = true
		choices = append(choices, best)
	}
	return choices
}

func (g *GUI) bidSpecialChoice(view baseui.TableView) (baseui.BidChoice, bool) {
	return g.bestBidChoiceForSuit(view.BidChoices, "王")
}

func (g *GUI) bestBidChoiceForSuit(choices []baseui.BidChoice, suit string) (baseui.BidChoice, bool) {
	var best baseui.BidChoice
	ok := false
	for _, choice := range choices {
		if choice.Suit != suit {
			continue
		}
		if !ok || bidPriority(choice.Type) > bidPriority(best.Type) {
			best = choice
			ok = true
		}
	}
	return best, ok
}

func bidPriority(kind string) int {
	switch kind {
	case "对王(无主)":
		return 4
	case "三张级牌":
		return 3
	case "对级牌":
		return 2
	case "单张级牌":
		return 1
	default:
		return 0
	}
}

func (g *GUI) drawBidSuitButton(screen *ebiten.Image, x, y int, choice baseui.BidChoice, selected bool, enabled bool) {
	fill := color.RGBA{0xcf, 0xcf, 0xcf, 0xff}
	stroke := color.RGBA{0x55, 0x5d, 0x66, 0xff}
	var textColor color.Color = color.RGBA{0x5a, 0x5a, 0x5a, 0xff}
	if enabled {
		fill = color.RGBA{0xf2, 0xf2, 0xf2, 0xff}
		textColor = color.Black
	}
	if selected {
		fill = color.RGBA{0x4d, 0x8f, 0xff, 0xff}
		stroke = hiliteColor
		textColor = color.White
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(BidSymbolSize), float32(BidSymbolSize), fill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(BidSymbolSize), float32(BidSymbolSize), 2, stroke, false)
	symbol := choice.Suit
	if choice.Suit != "王" {
		symbol = bidSuitSymbol(choice.Suit, false)
	}
	g.drawText(screen, symbol, x+10, y+21, textColor)
	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{rect: rect{x: x, y: y, w: BidSymbolSize, h: BidSymbolSize}, action: baseui.UIAction{Type: baseui.ActionSelectBid, BidType: choice.Type, BidSuit: choice.Suit}, enabled: enabled})
	g.st.mu.Unlock()
}

func (g *GUI) ensureBidSelectionLocked() {
	if len(g.st.view.BidChoices) == 0 {
		g.st.selectedBidType = ""
		g.st.selectedBidSuit = ""
		g.st.selectedBidChoice = ""
		return
	}
	for _, choice := range g.st.view.BidChoices {
		if choice.Type+"|"+choice.Suit == g.st.selectedBidChoice {
			return
		}
	}
	for _, choice := range g.st.view.BidChoices {
		if choice.Suit == g.st.selectedBidSuit && choice.Type == g.st.selectedBidType {
			g.st.selectedBidChoice = choice.Type + "|" + choice.Suit
			return
		}
	}
	preferred := g.bidSuitButtons(g.st.view)
	if len(preferred) == 0 {
		if special, ok := g.bidSpecialChoice(g.st.view); ok {
			g.st.selectedBidType = special.Type
			g.st.selectedBidSuit = special.Suit
			g.st.selectedBidChoice = special.Type + "|" + special.Suit
		}
		return
	}
	choice := preferred[0]
	g.st.selectedBidType = choice.Type
	g.st.selectedBidSuit = choice.Suit
	g.st.selectedBidChoice = choice.Type + "|" + choice.Suit
}

func (g *GUI) biddingRaisedCards(view baseui.TableView) map[int]bool {
	result := map[int]bool{}
	g.st.mu.RLock()
	suit := g.st.selectedBidSuit
	g.st.mu.RUnlock()
	if suit == "" || len(view.Players) == 0 {
		return result
	}
	for idx, c := range view.Players[0].HandCards {
		if c.Rank == view.DealerLevel && c.Suit == suit {
			result[idx] = true
		}
	}
	return result
}

func (g *GUI) drawBidButtons(screen *ebiten.Image, view baseui.TableView) {
	vector.DrawFilledRect(screen, float32(BidPanelX), float32(BidPanelY), float32(BidPanelW), float32(BidPanelH), color.RGBA{0x1e, 0x26, 0x31, 0xf0}, false)
	vector.StrokeRect(screen, float32(BidPanelX), float32(BidPanelY), float32(BidPanelW), float32(BidPanelH), 2, hiliteColor, false)
	g.drawText(screen, "亮主", BidPanelX+18, BidPanelY+20, hiliteColor)
	g.drawText(screen, "选花色后点继续", BidPanelX+54, BidPanelY+20, color.White)

	g.st.mu.RLock()
	selectedKey := g.st.selectedBidChoice
	g.st.mu.RUnlock()

	available := map[string]baseui.BidChoice{}
	for _, choice := range g.bidSuitButtons(view) {
		available[choice.Suit] = choice
	}
	order := []string{"黑桃", "红心", "梅花", "方块"}
	startX := BidPanelX + 18
	y := BidPanelY + 38
	for i, suit := range order {
		choice, ok := available[suit]
		if !ok {
			choice = baseui.BidChoice{Type: "", Suit: suit}
		}
		x := startX + i*(BidSymbolSize+BidSymbolGap)
		selected := ok && choice.Type+"|"+choice.Suit == selectedKey
		g.drawBidSuitButton(screen, x, y, choice, selected, ok)
	}
	if special, ok := g.bidSpecialChoice(view); ok {
		g.drawActionButton(screen, BidPanelX+18, BidPanelY+76, 76, 26, "无主", baseui.UIAction{Type: baseui.ActionSelectBid, BidType: special.Type, BidSuit: special.Suit}, true, special.Type+"|"+special.Suit != selectedKey)
	}
	g.drawActionButton(screen, BidPanelX+BidPanelW-168, BidPanelY+76, BidPrimaryBtnW, 26, "继续", baseui.UIAction{Type: baseui.ActionConfirm}, selectedKey != "", false)
	g.drawActionButton(screen, BidPanelX+BidPanelW-84, BidPanelY+76, BidSecondaryW, 26, "不亮", baseui.UIAction{Type: baseui.ActionPass}, true, true)
}

func (g *GUI) drawActionButton(screen *ebiten.Image, x, y, w, h int, label string, action baseui.UIAction, enabled bool, secondary bool) {
	fill := color.RGBA{0xdb, 0xb7, 0x43, 0xff}
	var textColor color.Color = color.Black
	if secondary {
		fill = secondaryButtonFill
		textColor = color.White
	}
	if !enabled {
		fill = disabledColor
		textColor = color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), fill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, outlineColor, false)
	g.drawText(screen, label, x+10, y+h/2+4, textColor)
	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{rect: rect{x: x, y: y, w: w, h: h}, action: action, enabled: enabled})
	g.st.mu.Unlock()
}

func (g *GUI) drawOverlay(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	if view.Phase == baseui.PhaseWelcome {
		vector.DrawFilledRect(screen, 80, 110, 480, 190, messageBgColor, false)
		vector.StrokeRect(screen, 80, 110, 480, 190, 2, hiliteColor, false)
		title := "升级（拖拉机）纸牌游戏"
		subtitle := "两副牌 · 从2开始打 · 4人对战"
		g.drawText(screen, title, LogicalWidth/2-(len(title)*7)/2, 145, color.RGBA{0xff, 0xd6, 0x54, 0xff})
		g.drawText(screen, subtitle, LogicalWidth/2-(len(subtitle)*7)/2, 175, color.White)
		hint := "点击左上角按钮开始游戏"
		g.drawText(screen, hint, LogicalWidth/2-(len(hint)*7)/2, 250, color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
		return
	}

	if view.Phase == baseui.PhaseWaitTrick && view.TrickWinner != "" {
		vector.DrawFilledRect(screen, 168, 172, 304, 78, messageBgColor, false)
		vector.StrokeRect(screen, 168, 172, 304, 78, 1, hiliteColor, false)
		g.drawText(screen, fmt.Sprintf("本轮赢家：%s", view.TrickWinner), 205, 202, color.White)
		g.drawText(screen, fmt.Sprintf("本轮得分：%d", view.TrickPoints), 205, 226, hiliteColor)
		return
	}

	if view.Phase == baseui.PhaseHandResult {
		vector.DrawFilledRect(screen, 150, 148, 350, 132, messageBgColor, false)
		vector.StrokeRect(screen, 150, 148, 350, 132, 2, hiliteColor, false)
		g.drawText(screen, "本局结算", 286, 176, hiliteColor)
		for i, line := range strings.Split(view.Message, "\n") {
			g.drawText(screen, line, 176, 206+i*20, color.White)
		}
		return
	}

	if view.Phase == baseui.PhaseDiscard {
		vector.DrawFilledRect(screen, 176, 84, 290, 38, messageBgColor, false)
		vector.StrokeRect(screen, 176, 84, 290, 38, 1, hiliteColor, false)
		g.drawText(screen, fmt.Sprintf("请垫底牌（已选 %d/8）", len(selected)), 205, 108, color.White)
	}

	if view.Phase == baseui.PhaseDealing {
		vector.DrawFilledRect(screen, 208, 84, 224, 38, messageBgColor, false)
		vector.StrokeRect(screen, 208, 84, 224, 38, 1, hiliteColor, false)
		g.drawText(screen, "发牌中...", 286, 108, color.White)
	}

	if view.Message == "" || view.Phase == baseui.PhaseBidding {
		return
	}
	vector.DrawFilledRect(screen, 150, 170, 340, 80, messageBgColor, false)
	vector.StrokeRect(screen, 150, 170, 340, 80, 1, hiliteColor, false)
	for i, line := range strings.Split(view.Message, "\n") {
		g.drawText(screen, line, 170, 195+i*18, color.White)
	}
}

func (g *GUI) drawCardWithAlpha(screen *ebiten.Image, x, y int, c baseui.CardView, selected bool, alpha float32) {
	if c.IsTractor {
		vector.DrawFilledRect(screen, float32(x-2), float32(y-4), float32(CardW+4), 16, tractorColor, false)
	}
	if IsImageLoaded() && c.FaceUp && c.RankNum > 0 {
		img := CardFaceImage(c.SuitNum, c.RankNum)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(CardW)/float64(img.Bounds().Dx()), float64(CardH)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.ScaleAlpha(alpha)
			screen.DrawImage(img, op)
			stroke := outlineColor
			width := float32(2)
			if c.Trump {
				stroke = color.RGBA{0x51, 0x7a, 0xcf, 0xff}
			}
			if selected {
				stroke = hiliteColor
				width = 3
			}
			vector.StrokeRect(screen, float32(x-1), float32(y-1), float32(CardW+2), float32(CardH+2), width, stroke, false)
			return
		}
	}
	if IsImageLoaded() && !c.FaceUp {
		img := CardBackImage(0)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(CardW)/float64(img.Bounds().Dx()), float64(CardH)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.ScaleAlpha(alpha)
			screen.DrawImage(img, op)
			return
		}
	}
	fill := cardFaceColor
	if !c.FaceUp {
		fill = cardBackColor
	}
	fill = withAlpha(fill, alpha)
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(CardW), float32(CardH), fill, false)
	stroke := outlineColor
	if c.Trump {
		stroke = color.RGBA{0x51, 0x7a, 0xcf, 0xff}
	}
	if selected {
		stroke = hiliteColor
	}
	vector.StrokeRect(screen, float32(x), float32(y), float32(CardW), float32(CardH), 2, stroke, false)
	if !c.FaceUp {
		g.drawText(screen, "###", x+20, y+52, color.White)
		return
	}
	title := cardTitle(c)
	lines := []string{title}
	if c.Trump {
		lines = append(lines, "主")
	}
	sort.Strings(lines)
	for i, line := range lines {
		g.drawText(screen, line, x+10, y+24+i*18, withAlpha(color.RGBA{0x00, 0x00, 0x00, 0xff}, alpha))
	}
}

func withAlpha(clr color.RGBA, alpha float32) color.RGBA {
	clr.A = uint8(float32(clr.A) * alpha)
	return clr
}

func cardTitle(c baseui.CardView) string {
	if c.Rank != "" && c.Suit != "" {
		return c.Suit + c.Rank
	}
	return c.Label
}
