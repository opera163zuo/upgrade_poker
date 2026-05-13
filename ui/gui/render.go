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
	g.drawMenuBar(screen)
	g.drawButtons(screen, view, selected)
	g.drawOverlay(screen, view, selected)
}

func (g *GUI) drawInfoBar(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	line1 := fmt.Sprintf("庄家 %s    主花色 %s    级别 %s / %s    比分 南北 %d : 东西 %d", view.Dealer, view.TrumpSuit, view.DealerLevel, view.OpponentLevel, view.TeamScore[0], view.TeamScore[1])
	line2 := fmt.Sprintf("当前第 %d / 25 轮    手牌 南%d 西%d 北%d 东%d    %s", view.TrickCount+1, view.Players[0].HandCount, view.Players[1].HandCount, view.Players[2].HandCount, view.Players[3].HandCount, g.statusLine(view, selected))
	g.drawText(screen, line1, InfoBarX+8, InfoBarY+15, color.White)
	g.drawText(screen, line2, InfoBarX+8, InfoBarY+32, hiliteColor)
}

func (g *GUI) statusLine(view baseui.TableView, selected map[int]bool) string {
	switch view.Phase {
	case baseui.PhaseBidding:
		return "发牌中可亮主 / B 亮主 / P 不亮"
	case baseui.PhaseDiscard:
		return fmt.Sprintf("请选择 8 张扣底（已选 %d/8） / Tab 切视图", len(selected))
	case baseui.PhasePlaying:
		if view.WaitingForHuman {
			hint := ""
			if view.CanHint {
				hint = " / H 提示"
			}
			return fmt.Sprintf("点击选牌，双击或 Enter 出牌（已选 %d 张）%s / Tab 切视图", len(selected), hint)
		}
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
	g.drawText(screen, label, x, y, color.White)
}

func (g *GUI) drawCenter(screen *ebiten.Image, view baseui.TableView) {
	positions := []struct {
		key string
		x   int
		y   int
	}{{"北", 250, 160}, {"西", 176, 222}, {"东", 388, 222}, {"南", 250, 286}}
	for _, p := range positions {
		cards := view.CurrentTrick[p.key]
		for i, c := range cards {
			g.drawCardWithAlpha(screen, p.x+i*22, p.y, c, false, 0.72)
		}
	}
	showBottom := len(view.BottomCards) > 0 && (view.Phase == baseui.PhaseDealing || view.Phase == baseui.PhaseDiscard || view.Phase == baseui.PhaseWaitTrick || view.Phase == baseui.PhaseHandResult)
	if showBottom {
		for i, c := range view.BottomCards {
			g.drawCardWithAlpha(screen, BottomX+i*BottomGap, BottomY, c, false, 0.88)
		}
	}
}

func (g *GUI) drawSouth(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	pv := view.Players[0]
	mode := view.HandViewMode
	if mode == "" {
		mode = "combo"
	}
	label := fmt.Sprintf("%s · %s", pv.Name, map[string]string{"combo": "牌型视图", "flat": "普通排序"}[mode])
	g.drawText(screen, label, 236, 350, color.White)
	var slots []southSlot
	if mode == "combo" {
		slots = g.comboSouthSlots(pv.HandCards, selected, view.TrumpSuit)
		g.drawSouthGroupBackdrops(screen, pv.HandCards, slots, view.TrumpSuit)
	} else {
		slots = g.flatSouthSlots(pv.HandCards, selected)
	}
	for _, slot := range slots {
		c := pv.HandCards[slot.idx]
		g.drawCardWithAlpha(screen, slot.x, slot.y, c, selected[slot.idx], 1)
	}
	rects := make([]rect, len(slots))
	for i, slot := range slots {
		rects[i] = rect{x: slot.x, y: slot.y, w: slot.w, h: slot.h}
	}
	g.st.mu.Lock()
	g.st.cardRects = rects
	g.st.mu.Unlock()
}

func (g *GUI) flatSouthSlots(cards []baseui.CardView, selected map[int]bool) []southSlot {
	slots := make([]southSlot, 0, len(cards))
	for i := range cards {
		x := SouthHandX + i*SouthHandGap
		y := SouthHandY
		if selected[i] {
			y -= 14
		}
		w := SouthHandGap
		if i == len(cards)-1 {
			w = CardW
		}
		slots = append(slots, southSlot{idx: i, x: x, y: y, w: w, h: CardH})
	}
	return slots
}

func (g *GUI) comboSouthSlots(cards []baseui.CardView, selected map[int]bool, trumpSuit string) []southSlot {
	order := g.suitRowOrder(cards, trumpSuit)
	rowY := map[string]int{}
	baseY := 308
	for i, suit := range order {
		rowY[suit] = baseY + i*18
	}
	rowCards := map[string][]int{}
	for idx, c := range cards {
		key := c.EffectiveSuit
		if key == "" {
			key = c.Suit
		}
		rowCards[key] = append(rowCards[key], idx)
	}
	var slots []southSlot
	for _, suit := range order {
		indices := rowCards[suit]
		if len(indices) == 0 {
			continue
		}
		x := 58
		for pos, idx := range indices {
			if pos > 0 {
				prev := cards[indices[pos-1]]
				curr := cards[idx]
				gap := 18
				if curr.IsTractor && prev.IsTractor && curr.EffectiveSuit == prev.EffectiveSuit {
					gap = 12
				} else if curr.EffectiveSuit != prev.EffectiveSuit {
					gap = 30
				}
				x += gap
			}
			y := rowY[suit]
			if selected[idx] {
				y -= 12
			}
			w := 18
			if pos == len(indices)-1 {
				w = CardW
			}
			slots = append(slots, southSlot{idx: idx, x: x, y: y, w: w, h: CardH})
		}
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].idx < slots[j].idx })
	return slots
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
	push(trumpSuit)
	for _, s := range []string{"黑桃", "红心", "方块", "梅花", "王"} {
		if s != trumpSuit {
			push(s)
		}
	}
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
		vector.DrawFilledRect(screen, float32(minX-26), float32(y+22), float32(maxX-minX+34), 22, fill, false)
		g.drawText(screen, suitLabel(suit, suit == trumpSuit), minX-20, y+38, labelColor)
	}
}

func suitLabel(suit string, isTrump bool) string {
	if isTrump {
		return "主牌"
	}
	switch suit {
	case "黑桃":
		return "♠"
	case "红心":
		return "♥"
	case "方块":
		return "♦"
	case "梅花":
		return "♣"
	case "王":
		return "王"
	default:
		return suit
	}
}

func (g *GUI) drawMenuBar(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, LogicalWidth, MenuBarH, color.RGBA{0x0d, 0x17, 0x22, 0xff}, false)
	g.drawText(screen, "游戏(G)  功能(F)  设定(S)  帮助(H)", 10, 14, color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
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
		g.drawActionButton(screen, 438, ActionBtnY, 124, ActionBtnH, toggleLabel(view.HandViewMode), baseui.UIAction{Type: baseui.ActionToggleView}, true, true)
	}

	if view.Phase == baseui.PhaseDiscard {
		g.drawActionButton(screen, 120, ActionBtnY, 124, ActionBtnH, toggleLabel(view.HandViewMode), baseui.UIAction{Type: baseui.ActionToggleView}, true, true)
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

func toggleLabel(mode string) string {
	if mode == "combo" {
		return "切到普通排序"
	}
	return "切到牌型视图"
}

func sanitizeButtonLabel(label string) string {
	label = strings.TrimSpace(label)
	if idx := strings.LastIndex(label, ":"); idx >= 0 && idx+1 < len(label) {
		return strings.TrimSuffix(strings.TrimSpace(label[idx+1:]), "]")
	}
	return strings.Trim(label, "[]")
}

func (g *GUI) drawBidButtons(screen *ebiten.Image, view baseui.TableView) {
	vector.DrawFilledRect(screen, BidPanelX, BidPanelY, BidPanelW, float32(44+len(view.BidChoices)*36), messageBgColor, false)
	vector.StrokeRect(screen, BidPanelX, BidPanelY, BidPanelW, float32(44+len(view.BidChoices)*36), 2, hiliteColor, false)
	g.drawText(screen, "请选择亮主方式", BidPanelX+90, BidPanelY+22, color.White)
	for i, choice := range view.BidChoices {
		y := BidPanelY + 34 + i*36
		g.drawActionButton(screen, BidPanelX+20, y, BidPanelW-40, BidBtnH, choice.Text, baseui.UIAction{Type: baseui.ActionBid, BidType: choice.Type, BidSuit: choice.Suit}, true, false)
	}
	g.drawActionButton(screen, BidPanelX+20, BidPanelY+34+len(view.BidChoices)*36, BidPanelW-40, BidBtnH, "不亮", baseui.UIAction{Type: baseui.ActionPass}, true, true)
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
		subtitle := "两副牌 · 2为常主 · 4人对战"
		g.drawText(screen, title, LogicalWidth/2-(len(title)*7)/2, 145, color.RGBA{0xff, 0xd6, 0x54, 0xff})
		g.drawText(screen, subtitle, LogicalWidth/2-(len(subtitle)*7)/2, 175, color.White)
		hint := "点击下方按钮开始游戏"
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
		g.drawText(screen, fmt.Sprintf("请选择 8 张扣底（已选 %d/8）", len(selected)), 193, 108, color.White)
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
