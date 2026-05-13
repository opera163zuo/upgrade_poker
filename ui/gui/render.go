package gui

import (
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
	g.drawText(screen, strings.TrimSpace(view.Message), InfoBarX+220, InfoBarY+16, hiliteColor)

	g.drawNorth(screen, view)
	g.drawWest(screen, view)
	g.drawEast(screen, view)
	g.drawCenter(screen, view)
	g.drawSouth(screen, view, selected)
	g.drawMenuBar(screen)
	g.drawButtons(screen, view)
	g.drawOverlay(screen, view)
}

func (g *GUI) drawNorth(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[2]
	g.drawText(screen, pv.Name, 280, 20, color.White)
	for i := 0; i < pv.HandCount && i < 12; i++ {
		x := 437 - NorthHandGap*i
		g.drawCard(screen, x, NorthHandY, baseui.CardView{Label: "", FaceUp: false}, false)
	}
}

func (g *GUI) drawWest(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[1]
	g.drawText(screen, pv.Name, 10, 140, color.White)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := WestHandY + WestHandGap*i
		g.drawCard(screen, WestHandX, y, baseui.CardView{FaceUp: false}, false)
	}
}

func (g *GUI) drawEast(screen *ebiten.Image, view baseui.TableView) {
	pv := view.Players[3]
	g.drawText(screen, pv.Name, 560, 140, color.White)
	for i := 0; i < pv.HandCount && i < 10; i++ {
		y := EastHandY - EastHandGap*i
		g.drawCard(screen, EastHandX, y, baseui.CardView{FaceUp: false}, false)
	}
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
	for i, c := range view.BottomCards {
		g.drawCard(screen, BottomX+i*BottomGap, BottomY, c, false)
	}
}

func (g *GUI) drawSouth(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	pv := view.Players[0]
	g.drawText(screen, pv.Name, 285, 360, color.White)
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
	// 菜单栏背景
	vector.DrawFilledRect(screen, 0, 0, LogicalWidth, MenuBarH, color.RGBA{0x0d, 0x17, 0x22, 0xff}, false)
	g.drawText(screen, "游戏(G)  功能(F)  设定(S)  帮助(H)", 10, 14, color.RGBA{0xaa, 0xcc, 0xaa, 0xff})
}

func (g *GUI) drawButtons(screen *ebiten.Image, view baseui.TableView) {
	if len(view.Buttons) == 0 {
		return
	}
	// 欢迎界面：中央大按钮
	if view.Phase == baseui.PhaseWelcome {
		btn := view.Buttons[0]
		bx := LogicalWidth/2 - 120
		by := LogicalHeight/2 + 60
		vector.DrawFilledRect(screen, float32(bx), float32(by), 240, 50, color.RGBA{0xdb, 0xb7, 0x43, 0xff}, false)
		vector.StrokeRect(screen, float32(bx), float32(by), 240, 50, 2, outlineColor, false)
		textW := len(btn.Label) * 7
		g.drawText(screen, btn.Label, bx+(240-textW)/2, by+30, color.Black)
		g.st.mu.Lock()
		g.st.buttonRects = append(g.st.buttonRects, buttonRect{rect: rect{x: bx, y: by, w: 240, h: 50}, action: baseui.ActionType(btn.ID)})
		g.st.mu.Unlock()
		return
	}
	// 常规底部按钮
	for i, b := range view.Buttons {
		x := 210 + i*110
		y := 440
		vector.DrawFilledRect(screen, float32(x), float32(y), 96, 28, color.RGBA{0xdb, 0xb7, 0x43, 0xff}, false)
		vector.StrokeRect(screen, float32(x), float32(y), 96, 28, 1, outlineColor, false)
		g.drawText(screen, b.Label, x+12, y+18, color.Black)
		g.st.mu.Lock()
		g.st.buttonRects = append(g.st.buttonRects, buttonRect{rect: rect{x: x, y: y, w: 96, h: 28}, action: baseui.ActionType(b.ID)})
		g.st.mu.Unlock()
	}
}

func (g *GUI) drawOverlay(screen *ebiten.Image, view baseui.TableView) {
	if view.Message == "" {
		return
	}
	// 欢迎界面：中央带标题的消息区域
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
	// 普通消息弹窗
	vector.DrawFilledRect(screen, 150, 170, 340, 80, messageBgColor, false)
	vector.StrokeRect(screen, 150, 170, 340, 80, 1, hiliteColor, false)
	for i, line := range strings.Split(view.Message, "\n") {
		g.drawText(screen, line, 170, 195+i*18, color.White)
	}
}

func (g *GUI) drawCard(screen *ebiten.Image, x, y int, c baseui.CardView, selected bool) {
	// 优先绘制真实牌面图片
	if IsImageLoaded() && c.FaceUp && c.RankNum > 0 {
		img := CardFaceImage(c.SuitNum, c.RankNum)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(CardW)/float64(img.Bounds().Dx()), float64(CardH)/float64(img.Bounds().Dy()))
			op.GeoM.Translate(float64(x), float64(y))
			if selected {
				op.GeoM.Translate(0, -20)
			}
			screen.DrawImage(img, op)
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
	// 备选：纯色方块渲染
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
