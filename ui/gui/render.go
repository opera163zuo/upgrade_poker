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

// ════════════════════════════════════════════════════════════════════════════
// 方案二：渲染实现
//
// 所有坐标均为物理像素。
// 字体按物理 DPI/字号创建，无低分辨率栅格化后放大的问题。
// 牌图从物理尺寸缓存直接绘制，不每帧缩放。
// ════════════════════════════════════════════════════════════════════════════

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

// ----- 绘制辅助函数：基于 ScaleCtx 的几何绘制 -----------------------------

// resetFillRect 在游戏区域绘制填充矩形
func resetFillRect(screen *ebiten.Image, x, y, w, h int, clr color.Color) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), clr, false)
}

// resetStrokeRect 在游戏区域绘制描边矩形
func resetStrokeRect(screen *ebiten.Image, x, y, w, h int, strokeWidth float32, clr color.Color) {
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), strokeWidth, clr, false)
}

// physText 绘制文字到物理像素坐标
// 使用当前缩放上下文创建对应尺寸的临时字体面
func (g *GUI) physText(dst *ebiten.Image, s string, x, y int, clr color.Color) {
	fnt := NewFontFaceForScale(g.sc)
	if fnt == nil {
		fnt = uiFontFace()
	}
	text.Draw(dst, s, fnt, x, y, clr)
}

// physTextBadge 绘制带背景的 badge
func (g *GUI) physTextBadge(dst *ebiten.Image, s string, x, y, padX, padY int,
	fill color.Color, border color.Color, textColor color.Color) {

	charW := int(g.sc.FontSize() * 0.55)
	if charW < 4 {
		charW = 4
	}
	lineH := int(g.sc.FontSize() * 1.2)
	if lineH < 10 {
		lineH = 10
	}
	width := len([]rune(s))*charW + padX*2
	height := lineH + padY*2
	top := y - lineH/2 - padY
	resetFillRect(dst, x-padX, top, width, height, fill)
	resetStrokeRect(dst, x-padX, top, width, height, 1, border)
	g.physText(dst, s, x, y, textColor)
}

// southSlot 南家牌槽位（物理坐标）
type southSlot struct {
	idx int
	x   int
	y   int
	w   int
	h   int
}

// ----- 信息栏 -------------------------------------------------------------

func (g *GUI) drawInfoBar(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	baseX, baseY := g.sc.PX(RefInfoBarX), g.sc.PX(RefInfoBarY)
	padX := g.sc.PXAbsolute(6)
	padY := g.sc.PXAbsolute(6)

	leftW := g.sc.PXAbsolute(246)
	midW := g.sc.PXAbsolute(176)
	rightW := g.sc.PXAbsolute(144)
	h := g.sc.PXAbsolute(30)

	// 庄家/主花色
	leftX := baseX + padX
	g.drawStatusBoxPhys(screen, leftX, baseY+padY, leftW, h,
		color.RGBA{0x17, 0x35, 0x26, 0xff},
		fmt.Sprintf("庄家 %s · 主花色 %s", view.Dealer, view.TrumpSuit), color.White)

	// 等级
	g.drawLevelPairPhys(screen, leftX+padX+leftW, baseY+padY, midW, h,
		view.DealerLevel, view.OpponentLevel)

	// 得分
	rightX := leftX + padX + leftW + padX + midW
	g.drawStatusBoxPhys(screen, rightX, baseY+padY, rightW, h,
		color.RGBA{0x18, 0x24, 0x34, 0xff},
		fmt.Sprintf("得分 %d : %d", view.TeamScore[0], view.TeamScore[1]), hiliteColor)

	// 第二行状态
	line2 := fmt.Sprintf("第 %d / 25 轮 · 手牌 南%d 西%d 北%d 东%d · %s",
		view.TrickCount+1,
		view.Players[0].HandCount, view.Players[1].HandCount,
		view.Players[2].HandCount, view.Players[3].HandCount,
		g.statusLine(view, selected))

	badgePadX := g.sc.PXAbsolute(8)
	badgePadY := g.sc.PXAbsolute(4)
	g.physTextBadge(screen, line2,
		g.sc.PX(RefInfoBarX)+g.sc.PXAbsolute(12),
		g.sc.PX(RefInfoBarY)+g.sc.PXAbsolute(41),
		badgePadX, badgePadY,
		color.RGBA{0x0e, 0x1a, 0x12, 0xcc},
		color.RGBA{0x6b, 0x88, 0x73, 0xff},
		color.RGBA{0xd5, 0xe5, 0xd5, 0xff})
}

func (g *GUI) drawStatusBoxPhys(screen *ebiten.Image, x, y, w, h int,
	fill color.Color, label string, textColor color.Color) {

	resetFillRect(screen, x, y, w, h, fill)
	resetStrokeRect(screen, x, y, w, h, 1,
		color.RGBA{0xe9, 0xe9, 0xe9, 0xff})
	textOffset := g.sc.PXAbsolute(10)
	textY := y + h/2 + int(g.sc.FontSize()*0.3)
	g.physText(screen, label, x+textOffset, textY, textColor)
}

func (g *GUI) drawLevelPairPhys(screen *ebiten.Image, x, y, w, h int,
	ourLevel, oppLevel string) {

	gap := g.sc.PXAbsolute(10)
	half := (w - gap) / 2
	g.drawStatusBoxPhys(screen, x, y, half, h,
		color.RGBA{0xf0, 0xf4, 0xfc, 0xff},
		"我方 "+ourLevel, color.RGBA{0x2d, 0x65, 0xc4, 0xff})
	g.drawStatusBoxPhys(screen, x+half+gap, y, half, h,
		color.RGBA{0xfb, 0xf1, 0xf1, 0xff},
		"敌方 "+oppLevel, color.RGBA{0xc8, 0x39, 0x39, 0xff})
}

func (g *GUI) statusLine(view baseui.TableView, selected map[int]bool) string {
	switch view.Phase {
	case baseui.PhaseDealing:
		return "发牌中..."
	case baseui.PhaseBidding:
		return "请选择亮主花色 / 反主 / B确认 P不亮"
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

// ----- 四方牌面 -----------------------------------------------------------

func (g *GUI) drawNorth(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[2]
	labelX, labelY := g.sc.PX(272), g.sc.PX(68)
	g.drawSeatLabelPhys(screen, pv, labelX, labelY)
	for i := 0; i < pv.HandCount && i < 12; i++ {
		x := g.sc.PX(440) - g.sc.PXAbsolute(RefNorthHandGap)*i
		g.drawCardPhys(screen, x, g.sc.PX(RefNorthHandY),
			baseui.CardView{Label: "", FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawWest(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[1]
	g.drawSeatLabelPhys(screen, pv, g.sc.PX(10), g.sc.PX(138))
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := g.sc.PX(RefWestHandY) + g.sc.PXAbsolute(RefWestHandGap)*i
		g.drawCardPhys(screen, g.sc.PX(RefWestHandX), y,
			baseui.CardView{FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawEast(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[3]
	g.drawSeatLabelPhys(screen, pv, g.sc.PX(520), g.sc.PX(138))
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := g.sc.PX(RefEastHandY) - g.sc.PXAbsolute(RefEastHandGap)*i
		g.drawCardPhys(screen, g.sc.PX(RefEastHandX), y,
			baseui.CardView{FaceUp: false}, false, 1)
	}
}

func (g *GUI) drawSeatLabelPhys(screen *ebiten.Image, pv baseui.PlayerView, x, y int) {
	label := pv.Name
	if pv.IsDealer {
		label += " ★"
	}
	if pv.IsBidder {
		label += " 🃏" // bidder indicator
	}
	if pv.IsThinking {
		dots := strings.Repeat(".", int(time.Now().UnixMilli()/350)%3+1)
		label += " 思考中" + dots
	}
	padX := g.sc.PXAbsolute(8)
	padY := g.sc.PXAbsolute(3)

	if pv.IsBidder {
		// White background, red text for bidder
		g.physTextBadge(screen, label, x+padX, y, padX, padY,
			color.RGBA{0xff, 0xff, 0xff, 0xff},
			color.RGBA{0xcc, 0x00, 0x00, 0xff},
			color.RGBA{0xcc, 0x00, 0x00, 0xff})
	} else {
		g.physTextBadge(screen, label, x+padX, y, padX, padY,
			color.RGBA{0x0d, 0x17, 0x22, 0xc8},
			color.RGBA{0x9f, 0xb8, 0xc4, 0xff},
			color.White)
	}
}

func (g *GUI) drawCenter(screen *ebiten.Image, view baseui.TableView) {
	positions := []struct {
		key string
		x   int
		y   int
	}{
		{"上(AI)", g.sc.PX(250), g.sc.PX(150)},
		{"左(AI)", g.sc.PX(176), g.sc.PX(210)},
		{"右(AI)", g.sc.PX(388), g.sc.PX(210)},
		{"下(你)", g.sc.PX(250), g.sc.PX(268)},
	}
	gap := g.sc.PXAbsolute(15)
	for _, p := range positions {
		cards := view.CurrentTrick[p.key]
		for i, c := range cards {
			g.drawCardPhys(screen, p.x+i*gap, p.y, c, false, 1)
		}
	}

	showBottom := len(view.BottomCards) > 0 &&
		(view.Phase == baseui.PhaseDiscard || view.Phase == baseui.PhaseHandResult)
	if showBottom {
		for i, c := range view.BottomCards {
			g.drawCardPhys(screen,
				g.sc.PX(RefBottomX)+i*g.sc.PXAbsolute(RefBottomGap),
				g.sc.PX(RefBottomY), c, false, 1)
		}
	}
}

func (g *GUI) drawSouth(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	pv := view.Players[0]
	biddingRaise := map[int]bool{}
	if view.Phase == baseui.PhaseBidding {
		biddingRaise = g.biddingRaisedCards(view)
	}
	g.physText(screen, pv.Name, g.sc.PX(236), g.sc.PX(328), color.White)

	var slots []southSlot
	slots = g.southSlots(pv.HandCards, selected, biddingRaise, view.TrumpSuit)
	g.drawSouthGroupBackdrops(screen, pv.HandCards, slots, view.TrumpSuit)
	for _, slot := range slots {
		c := pv.HandCards[slot.idx]
		g.drawCardPhys(screen, slot.x, slot.y, c, selected[slot.idx] || biddingRaise[slot.idx], 1)
	}
	rects := make([]Rect, len(slots))
	for i, slot := range slots {
		rects[i] = Rect{X: slot.x, Y: slot.y, W: slot.w, H: slot.h}
	}
	g.st.mu.Lock()
	g.st.cardRects = rects
	g.st.mu.Unlock()
}

func (g *GUI) southSlots(cards []baseui.CardView, selected map[int]bool,
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

	withinSuitGap := g.sc.PXAbsolute(RefSouthHandGap)
	betweenSuitGap := g.sc.PXAbsolute(22) // was RefSouthHandGap + 8

	physCardW, _ := g.sc.CardPhysSize()

	// 计算总宽以居中
	totalWidth := physCardW
	for i := 1; i < len(ordered); i++ {
		gap := withinSuitGap
		if effectiveSuit(cards[ordered[i-1]]) != effectiveSuit(cards[ordered[i]]) {
			gap = betweenSuitGap
		}
		totalWidth += gap
	}

	tableX := g.sc.PX(RefTableX)
	tableW := g.sc.PXAbsolute(RefTableW)
	startX := tableX + (tableW-totalWidth)/2
	southHandX := g.sc.PX(RefSouthHandX)
	if startX < southHandX {
		startX = southHandX
	} else if startX+totalWidth > tableX+tableW {
		startX = tableX + tableW - totalWidth
	}

	y := g.sc.PX(RefSouthHandY)
	_, physCardH := g.sc.CardPhysSize()

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
		slots = append(slots, southSlot{idx: idx, x: slotX, y: slotY,
			w: physCardW, h: physCardH})
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
	for _, s := range []string{"方块", "梅花", "红心", "黑桃"} {
		if s != trumpSuit {
			push(s)
		}
	}
	push(trumpSuit)
	for _, c := range cards {
		key := c.EffectiveSuit
		if key == "" {
			key = c.Suit
		}
		push(key)
	}
	return base
}

func (g *GUI) drawSouthGroupBackdrops(screen *ebiten.Image,
	cards []baseui.CardView, slots []southSlot, trumpSuit string) {

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

	physCardW, _ := g.sc.CardPhysSize()
	for _, suit := range g.suitRowOrder(cards, trumpSuit) {
		group := bySuit[suit]
		if len(group) == 0 {
			continue
		}
		minX, maxX, y := group[0].x, group[0].x+physCardW, group[0].y
		for _, slot := range group[1:] {
			if slot.x < minX {
				minX = slot.x
			}
			if slot.x+physCardW > maxX {
				maxX = slot.x + physCardW
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

		// 背景条
		bgPadX := g.sc.PXAbsolute(26)
		bgPadW := g.sc.PXAbsolute(34)
		bgH := g.sc.PXAbsolute(22)
		bgY := y + g.sc.PXAbsolute(22)
		resetFillRect(screen, minX-bgPadX, bgY, maxX-minX+bgPadW, bgH, fill)

		// 花色标签
		var labelX int
		if isTrumpSuit {
			labelX = maxX + g.sc.PXAbsolute(4)
		} else {
			labelX = minX - g.sc.PXAbsolute(20)
		}
		labelY := bgY + g.sc.PXAbsolute(16)
		g.physText(screen, bidSuitSymbol(suit, isTrumpSuit), labelX, labelY, labelColor)
	}
}

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

// ----- 菜单栏 -------------------------------------------------------------

func (g *GUI) drawMenuBar(screen *ebiten.Image, view baseui.TableView) {
	menuH := g.sc.PXAbsolute(RefMenuBarH)
	resetFillRect(screen, g.sc.PX(0), g.sc.PX(0),
		g.sc.PXAbsolute(RefWidth), menuH,
		color.RGBA{0xe7, 0xe7, 0xe7, 0xff})
	resetStrokeRect(screen, g.sc.PX(0), g.sc.PX(0),
		g.sc.PXAbsolute(RefWidth), menuH, 1,
		color.RGBA{0x88, 0x88, 0x88, 0xff})

	startEnabled := view.Phase == baseui.PhaseWelcome || view.Phase == baseui.PhaseGameOver
	restartEnabled := view.Phase != baseui.PhaseWelcome

	g.drawTopMenuButton(screen, g.sc.PX(8), g.sc.PX(3),
		g.sc.PXAbsolute(92), g.sc.PXAbsolute(18),
		"开始游戏", baseui.UIAction{Type: baseui.ActionStart}, startEnabled)
	g.drawTopMenuButton(screen, g.sc.PX(106), g.sc.PX(3),
		g.sc.PXAbsolute(92), g.sc.PXAbsolute(18),
		"重新开始", baseui.UIAction{Type: baseui.ActionRestart}, restartEnabled)
}

func (g *GUI) drawTopMenuButton(screen *ebiten.Image, x, y, w, h int,
	label string, action baseui.UIAction, enabled bool) {

	fill := color.RGBA{0xff, 0xff, 0xff, 0xff}
	var textColor color.Color = color.Black
	if !enabled {
		fill = color.RGBA{0xd4, 0xd4, 0xd4, 0xff}
		textColor = color.RGBA{0x66, 0x66, 0x66, 0xff}
	}
	resetFillRect(screen, x, y, w, h, fill)
	resetStrokeRect(screen, x, y, w, h, 1, outlineColor)

	textOffset := g.sc.PXAbsolute(12)
	textY := y + h/2 + int(g.sc.FontSize()*0.3)
	g.physText(screen, label, x+textOffset, textY, textColor)

	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{
		Rect: Rect{X: x, Y: y, W: w, H: h},
		action:  action,
		enabled: enabled,
	})
	g.st.mu.Unlock()
}

// ----- 操作按钮 -----------------------------------------------------------

func (g *GUI) drawButtons(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	if view.Phase == baseui.PhaseWelcome && len(view.Buttons) > 0 {
		btn := view.Buttons[0]
		centerX := g.sc.PX(RefWidth/2) - g.sc.PXAbsolute(120)
		btnY := g.sc.PX(RefHeight/2) + g.sc.PXAbsolute(60)
		g.drawActionButtonPhys(screen, centerX, btnY,
			g.sc.PXAbsolute(240), g.sc.PXAbsolute(50),
			btn.Label, baseui.UIAction{Type: baseui.ActionType(btn.ID)},
			btn.Enabled, false)
		return
	}

	if view.Phase == baseui.PhaseBidding {
		g.drawBidButtonsPhys(screen, view)
		return
	}

	btnW := g.sc.PXAbsolute(RefActionBtnW)
	btnH := g.sc.PXAbsolute(RefActionBtnH)
	btnY := g.sc.PX(RefActionBtnY)

	if view.Phase == baseui.PhasePlaying && view.WaitingForHuman {
		g.drawActionButtonPhys(screen, g.sc.PX(120), btnY, btnW, btnH,
			"提示", baseui.UIAction{Type: baseui.ActionHint}, view.CanHint, true)
		g.drawActionButtonPhys(screen, g.sc.PX(226), btnY, btnW, btnH,
			"出牌", baseui.UIAction{Type: baseui.ActionPlay}, len(selected) > 0, false)
		g.drawActionButtonPhys(screen, g.sc.PX(332), btnY, btnW, btnH,
			"取消", baseui.UIAction{Type: baseui.ActionCancel}, len(selected) > 0, true)
	}

	if view.Phase == baseui.PhaseDiscard {
		g.drawActionButtonPhys(screen, g.sc.PX(254), btnY, btnW, btnH,
			"扣底", baseui.UIAction{Type: baseui.ActionPlay},
			len(selected) == view.DiscardCount, false)
		g.drawActionButtonPhys(screen, g.sc.PX(360), btnY, btnW, btnH,
			"取消", baseui.UIAction{Type: baseui.ActionCancel},
			len(selected) > 0, true)
	}

	// 通用按钮（从 view.Buttons）
	for i, b := range view.Buttons {
		x := g.sc.PX(430) + i*g.sc.PXAbsolute(100)
		if x+btnW > g.sc.PXAbsolute(RefWidth) {
			x = g.sc.PX(430)
		}
		g.drawActionButtonPhys(screen, x, btnY, btnW, btnH,
			sanitizeButtonLabel(b.Label),
			baseui.UIAction{Type: baseui.ActionType(b.ID)},
			b.Enabled, false)
	}
}

func sanitizeButtonLabel(label string) string {
	label = strings.TrimSpace(label)
	if idx := strings.LastIndex(label, ":"); idx >= 0 && idx+1 < len(label) {
		return strings.TrimSuffix(strings.TrimSpace(label[idx+1:]), "]")
	}
	return strings.Trim(label, "[]")
}

func (g *GUI) drawActionButtonPhys(screen *ebiten.Image, x, y, w, h int,
	label string, action baseui.UIAction, enabled bool, secondary bool) {

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
	resetFillRect(screen, x, y, w, h, fill)
	resetStrokeRect(screen, x, y, w, h, 1, outlineColor)

	textOffset := g.sc.PXAbsolute(10)
	textY := y + h/2 + int(g.sc.FontSize()*0.3)
	g.physText(screen, label, x+textOffset, textY, textColor)

	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{
		Rect: Rect{X: x, Y: y, W: w, H: h},
		action:  action,
		enabled: enabled,
	})
	g.st.mu.Unlock()
}

// ----- 亮主面板 -----------------------------------------------------------

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

func (g *GUI) drawBidButtonsPhys(screen *ebiten.Image, view baseui.TableView) {
	panelX := g.sc.PX(RefBidPanelX)
	panelY := g.sc.PX(RefBidPanelY)
	panelW := g.sc.PXAbsolute(RefBidPanelW)
	panelH := g.sc.PXAbsolute(RefBidPanelH)

	resetFillRect(screen, panelX, panelY, panelW, panelH,
		color.RGBA{0x1e, 0x26, 0x31, 0xf0})
	resetStrokeRect(screen, panelX, panelY, panelW, panelH, 2, hiliteColor)

	titleX := panelX + g.sc.PXAbsolute(RefBidTitleX)
	titleY := panelY + g.sc.PXAbsolute(RefBidTitleY)
	g.physText(screen, "亮主", titleX, titleY, hiliteColor)

	hintX := titleX + g.sc.PXAbsolute(36)
	g.physText(screen, "选花色后点继续", hintX, titleY, color.White)

	g.st.mu.RLock()
	selectedKey := g.st.selectedBidChoice
	g.st.mu.RUnlock()

	available := map[string]baseui.BidChoice{}
	for _, choice := range g.bidSuitButtons(view) {
		available[choice.Suit] = choice
	}
	order := []string{"黑桃", "红心", "梅花", "方块"}

	symSize := g.sc.PXAbsolute(RefBidSymbolSize)
	symGap := g.sc.PXAbsolute(RefBidSymbolGap)
	startX := panelX + g.sc.PXAbsolute(RefBidTitleX)
	y := panelY + g.sc.PXAbsolute(38)

	for i, suit := range order {
		choice, ok := available[suit]
		if !ok {
			choice = baseui.BidChoice{Type: "", Suit: suit}
		}
		x := startX + i*(symSize+symGap)
		selected := ok && choice.Type+"|"+choice.Suit == selectedKey
		g.drawBidSuitButtonPhys(screen, x, y, symSize, choice, selected, ok)
	}

	if special, ok := g.bidSpecialChoice(view); ok {
		g.drawActionButtonPhys(screen,
			panelX+g.sc.PXAbsolute(RefBidTitleX),
			panelY+g.sc.PXAbsolute(76),
			g.sc.PXAbsolute(76),
			g.sc.PXAbsolute(26),
			"无主", baseui.UIAction{Type: baseui.ActionSelectBid,
				BidType: special.Type, BidSuit: special.Suit},
			true, special.Type+"|"+special.Suit != selectedKey)
	}

	btnH := g.sc.PXAbsolute(26)
	g.drawActionButtonPhys(screen,
		panelX+panelW-g.sc.PXAbsolute(168),
		panelY+g.sc.PXAbsolute(76),
		g.sc.PXAbsolute(RefBidPrimaryBtnW), btnH,
		"继续", baseui.UIAction{Type: baseui.ActionConfirm},
		selectedKey != "", false)
	g.drawActionButtonPhys(screen,
		panelX+panelW-g.sc.PXAbsolute(84),
		panelY+g.sc.PXAbsolute(76),
		g.sc.PXAbsolute(RefBidSecondaryW), btnH,
		"不亮", baseui.UIAction{Type: baseui.ActionPass}, true, true)
}

func (g *GUI) drawBidSuitButtonPhys(screen *ebiten.Image, x, y, size int,
	choice baseui.BidChoice, selected bool, enabled bool) {

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
	resetFillRect(screen, x, y, size, size, fill)
	resetStrokeRect(screen, x, y, size, size, 2, stroke)

	symbol := choice.Suit
	if choice.Suit != "王" {
		symbol = bidSuitSymbol(choice.Suit, false)
	}
	textOffset := g.sc.PXAbsolute(10)
	textY := y + size/2 + int(g.sc.FontSize()*0.3)
	g.physText(screen, symbol, x+textOffset, textY, textColor)

	g.st.mu.Lock()
	g.st.buttonRects = append(g.st.buttonRects, buttonRect{
		Rect: Rect{X: x, Y: y, W: size, H: size},
		action:  baseui.UIAction{Type: baseui.ActionSelectBid,
			BidType: choice.Type, BidSuit: choice.Suit},
		enabled: enabled,
	})
	g.st.mu.Unlock()
}

// ----- 覆盖层（欢迎/结算/消息提示）----------------------------------------

func (g *GUI) drawOverlay(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	if view.Phase == baseui.PhaseWelcome {
		g.drawWelcomeOverlay(screen, view)
		return
	}

	if view.Phase == baseui.PhaseWaitTrick && view.TrickWinner != "" {
		g.drawTrickResultOverlay(screen, view)
		return
	}

	if view.Phase == baseui.PhaseHandResult {
		g.drawHandResultOverlay(screen, view)
		return
	}

	if view.Phase == baseui.PhaseDiscard {
		g.drawDiscardOverlay(screen, view, selected)
		return
	}

	if view.Phase == baseui.PhaseDealing {
		g.drawDealingOverlay(screen)
		return
	}

	if view.Message == "" || view.Phase == baseui.PhaseBidding {
		return
	}
	g.drawMessageOverlay(screen, view)
}

func (g *GUI) drawWelcomeOverlay(screen *ebiten.Image, view baseui.TableView) {
	msgW := g.sc.PXAbsolute(480)
	msgH := g.sc.PXAbsolute(190)
	msgX := g.sc.PX(80)
	msgY := g.sc.PX(110)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 2, hiliteColor)

	title := "升级（拖拉机）纸牌游戏"
	subtitle := "两副牌 · 从2开始打 · 4人对战"
	fontSize := g.sc.FontSize()

	centerText := func(s string, y int) int {
		charW := int(fontSize * 0.55)
		if charW < 4 {
			charW = 4
		}
		return g.sc.PX(RefWidth/2) - (charW*len([]rune(s)))/2
	}

	titleY := msgY + g.sc.PXAbsolute(35)
	titleX := centerText(title, titleY)
	g.physText(screen, title, titleX, titleY,
		color.RGBA{0xff, 0xd6, 0x54, 0xff})

	subY := titleY + g.sc.PXAbsolute(30)
	subX := centerText(subtitle, subY)
	g.physText(screen, subtitle, subX, subY, color.White)

	hint := "点击左上角按钮开始游戏"
	hintY := subY + g.sc.PXAbsolute(75)
	hintX := centerText(hint, hintY)
	g.physText(screen, hint, hintX, hintY,
		color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
}

func (g *GUI) drawTrickResultOverlay(screen *ebiten.Image, view baseui.TableView) {
	msgW := g.sc.PXAbsolute(304)
	msgH := g.sc.PXAbsolute(78)
	msgX := g.sc.PX(168)
	msgY := g.sc.PX(172)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 1, hiliteColor)

	winnerX := g.sc.PX(205)
	winnerY := msgY + g.sc.PXAbsolute(30)
	g.physText(screen, fmt.Sprintf("本轮赢家：%s", view.TrickWinner),
		winnerX, winnerY, color.White)

	scoreY := winnerY + g.sc.PXAbsolute(24)
	g.physText(screen, fmt.Sprintf("本轮得分：%d", view.TrickPoints),
		g.sc.PX(205), scoreY, hiliteColor)
}

func (g *GUI) drawHandResultOverlay(screen *ebiten.Image, view baseui.TableView) {
	msgW := g.sc.PXAbsolute(350)
	msgH := g.sc.PXAbsolute(132)
	msgX := g.sc.PX(150)
	msgY := g.sc.PX(148)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 2, hiliteColor)

	titleX := g.sc.PX(286)
	titleY := msgY + g.sc.PXAbsolute(28)
	g.physText(screen, "本局结算", titleX, titleY, hiliteColor)

	lines := strings.Split(view.Message, "\n")
	for i, line := range lines {
		g.physText(screen, line,
			g.sc.PX(176),
			msgY+g.sc.PXAbsolute(58)+i*g.sc.PXAbsolute(20),
			color.White)
	}
}

func (g *GUI) drawDiscardOverlay(screen *ebiten.Image, view baseui.TableView,
	selected map[int]bool) {

	msgW := g.sc.PXAbsolute(290)
	msgH := g.sc.PXAbsolute(38)
	msgX := g.sc.PX(176)
	msgY := g.sc.PX(84)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 1, hiliteColor)

	textX := g.sc.PX(205)
	textY := msgY + g.sc.PXAbsolute(24)
	g.physText(screen, fmt.Sprintf("请垫底牌（已选 %d/8）", len(selected)),
		textX, textY, color.White)
}

func (g *GUI) drawDealingOverlay(screen *ebiten.Image) {
	msgW := g.sc.PXAbsolute(224)
	msgH := g.sc.PXAbsolute(38)
	msgX := g.sc.PX(208)
	msgY := g.sc.PX(84)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 1, hiliteColor)

	g.physText(screen, "发牌中...",
		g.sc.PX(286), msgY+g.sc.PXAbsolute(24), color.White)
}

func (g *GUI) drawMessageOverlay(screen *ebiten.Image, view baseui.TableView) {
	msgW := g.sc.PXAbsolute(340)
	msgH := g.sc.PXAbsolute(80)
	msgX := g.sc.PX(150)
	msgY := g.sc.PX(170)
	resetFillRect(screen, msgX, msgY, msgW, msgH, messageBgColor)
	resetStrokeRect(screen, msgX, msgY, msgW, msgH, 1, hiliteColor)

	lines := strings.Split(view.Message, "\n")
	for i, line := range lines {
		g.physText(screen, line,
			g.sc.PX(170),
			msgY+g.sc.PXAbsolute(25)+i*g.sc.PXAbsolute(18),
			color.White)
	}
}

// ----- 扑克牌绘制（物理像素）----------------------------------------------

// drawCardPhys 绘制一张牌到物理像素坐标。
// 使用预缓存的物理尺寸牌图，不进行每帧缩放。
func (g *GUI) drawCardPhys(screen *ebiten.Image, x, y int, c baseui.CardView,
	selected bool, alpha float32) {

	if c.IsTractor {
		barH := g.sc.PXAbsolute(16)
		resetFillRect(screen, x-g.sc.PXAbsolute(2), y-g.sc.PXAbsolute(4),
			g.sc.PXAbsolute(RefCardW+4), barH, tractorColor)
	}

	physW, physH := g.sc.CardPhysSize()

	if IsImageLoaded() && c.FaceUp && c.RankNum > 0 {
		img := CardFaceImagePhys(c.SuitNum, c.RankNum, physW, physH)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
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
			resetStrokeRect(screen, x-1, y-1, physW+2, physH+2, width, stroke)
			return
		}
	}

	if IsImageLoaded() && !c.FaceUp {
		img := CardBackImagePhys(0, physW, physH)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorScale.ScaleAlpha(alpha)
			screen.DrawImage(img, op)
			return
		}
	}

	// 降级：矢量绘制
	fill := cardFaceColor
	if !c.FaceUp {
		fill = cardBackColor
	}
	fill = withAlpha(fill, alpha)
	resetFillRect(screen, x, y, physW, physH, fill)

	stroke := outlineColor
	if c.Trump {
		stroke = color.RGBA{0x51, 0x7a, 0xcf, 0xff}
	}
	if selected {
		stroke = hiliteColor
	}
	resetStrokeRect(screen, x, y, physW, physH, 2, stroke)

	if !c.FaceUp {
		textX := x + physW/2 - g.sc.PXAbsolute(10)
		textY := y + physH - g.sc.PXAbsolute(20)
		g.physText(screen, "###", textX, textY, color.White)
		return
	}

	title := cardTitle(c)
	lines := []string{title}
	if c.Trump {
		lines = append(lines, "主")
	}
	sort.Strings(lines)
	for i, line := range lines {
		textX := x + g.sc.PXAbsolute(10)
		textY := y + g.sc.PXAbsolute(18) + i*g.sc.PXAbsolute(12)
		g.physText(screen, line, textX, textY,
			withAlpha(color.RGBA{0x00, 0x00, 0x00, 0xff}, alpha))
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

// ----- 保留方法（方案二迁移）----------------------------------------------

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
