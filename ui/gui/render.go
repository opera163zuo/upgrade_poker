package gui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

var (
	bgColor        = color.RGBA{0x16, 0x5f, 0x3b, 0xff}
	tableColor     = color.RGBA{0x0d, 0x4b, 0x2f, 0xff}
	cardFaceColor  = color.RGBA{0xf6, 0xf1, 0xe7, 0xff}
	cardBackColor  = color.RGBA{0x2c, 0x4f, 0x94, 0xff}
	outlineColor   = color.RGBA{0x22, 0x22, 0x22, 0xff}
	hiliteColor    = color.RGBA{0xff, 0xd6, 0x54, 0xff}
	messageBgColor = color.RGBA{0x0d, 0x17, 0x22, 0xdd}
	disabledColor  = color.RGBA{0x6b, 0x6b, 0x6b, 0xff}
)

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
	g.drawText(screen, strings.TrimSpace(strings.Join([]string{"庄家", view.Dealer, "主", view.TrumpSuit}, "  ")), InfoBarX+8, InfoBarY+16, color.White)
	g.drawText(screen, g.statusLine(view, selected), InfoBarX+220, InfoBarY+16, hiliteColor)

	g.drawNorth(screen, view)
	g.drawWest(screen, view)
	g.drawEast(screen, view)
	g.drawCenter(screen, view)
	g.drawSouth(screen, view, selected)
	g.drawMenuBar(screen)
	g.drawButtons(screen, view, selected)
	g.drawOverlay(screen, view, selected)
}

func (g *GUI) statusLine(view baseui.TableView, selected map[int]bool) string {
	switch view.Phase {
	case baseui.PhaseBidding:
		return "发牌中可亮主 / B 亮主 / P 不亮"
	case baseui.PhaseDiscard:
		return fmt.Sprintf("请选择 8 张扣底（已选 %d/8）", len(selected))
	case baseui.PhasePlaying:
		if view.WaitingForHuman {
			return fmt.Sprintf("点击选牌，双击或 Enter 出牌（已选 %d 张）", len(selected))
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
	g.drawSeatLabel(screen, pv, 280, 20)
	for i := 0; i < pv.HandCount && i < 12; i++ {
		x := 437 - NorthHandGap*i
		g.drawCard(screen, x, NorthHandY, baseui.CardView{Label: "", FaceUp: false}, false)
	}
}

func (g *GUI) drawWest(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[1]
	g.drawSeatLabel(screen, pv, 10, 140)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := WestHandY + WestHandGap*i
		g.drawCard(screen, WestHandX, y, baseui.CardView{FaceUp: false}, false)
	}
}

func (g *GUI) drawEast(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[3]
	g.drawSeatLabel(screen, pv, 520, 140)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := EastHandY - EastHandGap*i
		g.drawCard(screen, EastHandX, y, baseui.CardView{FaceUp: false}, false)
	}
}

func (g *GUI) drawSeatLabel(screen *ebiten.Image, pv baseui.PlayerView, x, y int) {
	label := pv.Name
	if pv.IsDealer {
		label += " ★"
	}
	if pv.IsThinking {
		label += " 思考中…"
	}
	g.drawText(screen, label, x, y, color.White)
}

func (g *GUI) drawCenter(screen *ebiten.Image, view baseui.TableView) {
	positions := []struct {
		key string
		x   int
		y   int
	}{{"北", 250, 150}, {"西", 180, 215}, {"东", 390, 215}, {"南", 250, 280}}
	for _, p := range positions {
		cards := view.CurrentTrick[p.key]
		for i, c := range cards {
			g.drawCard(screen, p.x+i*22, p.y, c, false)
		}
	}
	showBottom := len(view.BottomCards) > 0 && (view.Phase == baseui.PhaseDealing || view.Phase == baseui.PhaseDiscard || view.Phase == baseui.PhaseWaitTrick || view.Phase == baseui.PhaseHandResult)
	if showBottom {
		for i, c := range view.BottomCards {
			g.drawCard(screen, BottomX+i*BottomGap, BottomY, c, false)
		}
	}
}

func (g *GUI) drawSouth(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	pv := view.Players[0]
	g.drawSeatLabel(screen, pv, 250, 360)
	cards := pv.HandCards
	for i, c := range cards {
		x := SouthHandX + i*SouthHandGap
		y := SouthHandY
		if selected[i] {
			y -= 12
		}
		g.drawCard(screen, x, y, c, selected[i])
		g.st.mu.Lock()
		g.st.cardRects = append(g.st.cardRects, rect{x: x, y: y, w: CardW, h: CardH})
		g.st.mu.Unlock()
	}
}

func (g *GUI) drawMenuBar(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, LogicalWidth, MenuBarH, color.RGBA{0x0d, 0x17, 0x22, 0xff}, false)
	g.drawText(screen, "游戏(G)  功能(F)  设定(S)  帮助(H)", 10, 14, color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
}

func (g *GUI) drawButtons(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	if view.Phase == baseui.PhaseWelcome && len(view.Buttons) > 0 {
		btn := view.Buttons[0]
		g.drawActionButton(screen, LogicalWidth/2-120, LogicalHeight/2+60, 240, 50, btn.Label, baseui.UIAction{Type: baseui.ActionType(btn.ID)}, btn.Enabled)
		return
	}

	if view.Phase == baseui.PhaseBidding {
		g.drawBidButtons(screen, view)
		return
	}

	if view.Phase == baseui.PhasePlaying && view.WaitingForHuman {
		g.drawActionButton(screen, 210, ActionBtnY, ActionBtnW, ActionBtnH, "出牌", baseui.UIAction{Type: baseui.ActionPlay}, len(selected) > 0)
		g.drawActionButton(screen, 320, ActionBtnY, ActionBtnW, ActionBtnH, "取消", baseui.UIAction{Type: baseui.ActionCancel}, len(selected) > 0)
	}

	if view.Phase == baseui.PhaseDiscard {
		g.drawActionButton(screen, 210, ActionBtnY, ActionBtnW, ActionBtnH, "扣底", baseui.UIAction{Type: baseui.ActionPlay}, len(selected) == view.DiscardCount)
		g.drawActionButton(screen, 320, ActionBtnY, ActionBtnW, ActionBtnH, "取消", baseui.UIAction{Type: baseui.ActionCancel}, len(selected) > 0)
	}

	for i, b := range view.Buttons {
		x := 430 + i*100
		if x+ActionBtnW > LogicalWidth {
			x = 430
		}
		g.drawActionButton(screen, x, ActionBtnY, ActionBtnW, ActionBtnH, sanitizeButtonLabel(b.Label), baseui.UIAction{Type: baseui.ActionType(b.ID)}, b.Enabled)
	}
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
		g.drawActionButton(screen, BidPanelX+20, y, BidPanelW-40, BidBtnH, choice.Text, baseui.UIAction{Type: baseui.ActionBid, BidType: choice.Type, BidSuit: choice.Suit}, true)
	}
	g.drawActionButton(screen, BidPanelX+20, BidPanelY+34+len(view.BidChoices)*36, BidPanelW-40, BidBtnH, "不亮", baseui.UIAction{Type: baseui.ActionPass}, true)
}

func (g *GUI) drawActionButton(screen *ebiten.Image, x, y, w, h int, label string, action baseui.UIAction, enabled bool) {
	fill := color.RGBA{0xdb, 0xb7, 0x43, 0xff}
	var textColor color.Color = color.Black
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
		vector.DrawFilledRect(screen, 80, 100, 480, 200, messageBgColor, false)
		vector.StrokeRect(screen, 80, 100, 480, 200, 2, hiliteColor, false)
		title := "升级（拖拉机）纸牌游戏"
		subtitle := "两副牌 · 2为常主 · 4人对战"
		g.drawText(screen, title, LogicalWidth/2-(len(title)*7)/2, 135, color.RGBA{0xff, 0xd6, 0x54, 0xff})
		g.drawText(screen, subtitle, LogicalWidth/2-(len(subtitle)*7)/2, 165, color.White)
		hint := "点击下方按钮开始游戏"
		g.drawText(screen, hint, LogicalWidth/2-(len(hint)*7)/2, 240, color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
		return
	}

	if view.Phase == baseui.PhaseWaitTrick && view.TrickWinner != "" {
		vector.DrawFilledRect(screen, 170, 175, 300, 70, messageBgColor, false)
		vector.StrokeRect(screen, 170, 175, 300, 70, 1, hiliteColor, false)
		g.drawText(screen, fmt.Sprintf("本轮赢家：%s", view.TrickWinner), 210, 202, color.White)
		g.drawText(screen, fmt.Sprintf("本轮得分：%d", view.TrickPoints), 210, 224, hiliteColor)
		return
	}

	if view.Phase == baseui.PhaseDiscard {
		vector.DrawFilledRect(screen, 180, 84, 280, 38, messageBgColor, false)
		vector.StrokeRect(screen, 180, 84, 280, 38, 1, hiliteColor, false)
		g.drawText(screen, fmt.Sprintf("请选择 8 张扣底（已选 %d/8）", len(selected)), 196, 108, color.White)
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

func (g *GUI) drawCard(screen *ebiten.Image, x, y int, c baseui.CardView, selected bool) {
	if IsImageLoaded() && c.FaceUp && c.RankNum > 0 {
		img := CardFaceImage(c.SuitNum, c.RankNum)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(CardW)/float64(img.Bounds().Dx()), float64(CardH)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(img, op)
			if selected {
				vector.StrokeRect(screen, float32(x-2), float32(y-2), float32(CardW+4), float32(CardH+4), 3, hiliteColor, false)
			}
			return
		}
	}
	if IsImageLoaded() && !c.FaceUp {
		img := CardBackImage(0)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(CardW)/float64(img.Bounds().Dx()), float64(CardH)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(img, op)
			return
		}
	}
	fill := cardFaceColor
	if !c.FaceUp {
		fill = cardBackColor
	}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(CardW), float32(CardH), fill, false)
	stroke := outlineColor
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
		g.drawText(screen, line, x+10, y+24+i*18, color.Black)
	}
}

func cardTitle(c baseui.CardView) string {
	if c.Rank != "" && c.Suit != "" {
		return c.Suit + c.Rank
	}
	return c.Label
}
