package gui

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	baseui "github.com/smallnest/upgrade_poker/ui"
)

var (
	roomBaseColor        = color.RGBA{0x09, 0x1c, 0x1b, 0xff}
	roomGlowColor        = color.RGBA{0x11, 0x36, 0x30, 0x88}
	tableShadowColor     = color.RGBA{0x04, 0x08, 0x08, 0x90}
	tableRailDarkColor   = color.RGBA{0x3b, 0x25, 0x14, 0xff}
	tableRailLightColor  = color.RGBA{0x76, 0x4d, 0x23, 0xff}
	tableFeltOuterColor  = color.RGBA{0x0e, 0x5c, 0x45, 0xff}
	tableFeltInnerColor  = color.RGBA{0x0a, 0x49, 0x3d, 0xff}
	tableFeltLineColor   = color.RGBA{0xc0, 0xe0, 0xc8, 0x40}
	tableCenterGlowColor = color.RGBA{0xbd, 0xe6, 0xd0, 0x20}
	panelBorderColor     = color.RGBA{0xd6, 0xb9, 0x72, 0xff}
	panelFillColor       = color.RGBA{0x10, 0x1f, 0x25, 0xdc}
	panelMutedColor      = color.RGBA{0x12, 0x2f, 0x30, 0xd8}
	panelTextMutedColor  = color.RGBA{0xc0, 0xcc, 0xca, 0xff}
	seatFillColor        = color.RGBA{0x0b, 0x18, 0x20, 0xe0}
	seatBadgeColor       = color.RGBA{0x17, 0x31, 0x38, 0xf0}
	seatHumanColor       = color.RGBA{0xf0, 0xd8, 0x97, 0xff}
	seatDealerColor      = color.RGBA{0xd6, 0xa3, 0x37, 0xff}
	seatBidderColor      = color.RGBA{0x5d, 0xb8, 0xff, 0xff}
	seatThinkingColor    = color.RGBA{0x7d, 0xd3, 0x8e, 0xff}
	statusRibbonFill     = color.RGBA{0x08, 0x12, 0x19, 0xd4}
	statusRibbonBorder   = color.RGBA{0x9a, 0xb4, 0xb2, 0x80}
)

func (g *GUI) tableOuterBounds() Rect {
	return Rect{
		X: g.sc.PX(18),
		Y: g.sc.PX(44),
		W: g.sc.PXAbsolute(604),
		H: g.sc.PXAbsolute(348),
	}
}

func (g *GUI) tableInnerBounds() Rect {
	outer := g.tableOuterBounds()
	insetX := g.sc.PXAbsolute(18)
	insetY := g.sc.PXAbsolute(16)
	return Rect{
		X: outer.X + insetX,
		Y: outer.Y + insetY,
		W: outer.W - insetX*2,
		H: outer.H - insetY*2,
	}
}

func (g *GUI) southHandY() int {
	table := g.tableOuterBounds()
	_, cardH := g.sc.CardPhysSize()
	return table.Y + table.H - cardH - g.sc.PXAbsolute(18)
}

func (g *GUI) textWidth(s string) int {
	charW := int(g.sc.FontSize() * 0.58)
	if charW < 5 {
		charW = 5
	}
	return len([]rune(s)) * charW
}

func tintWithAlpha(clr color.RGBA, alpha uint8) color.RGBA {
	clr.A = alpha
	return clr
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fillRoundedRect(screen *ebiten.Image, rect Rect, radius int, clr color.Color) {
	if rect.W <= 0 || rect.H <= 0 {
		return
	}
	if radius <= 0 {
		resetFillRect(screen, rect.X, rect.Y, rect.W, rect.H, clr)
		return
	}
	maxRadius := minInt(rect.W/2, rect.H/2)
	if radius > maxRadius {
		radius = maxRadius
	}
	resetFillRect(screen, rect.X+radius, rect.Y, rect.W-radius*2, rect.H, clr)
	resetFillRect(screen, rect.X, rect.Y+radius, radius, rect.H-radius*2, clr)
	resetFillRect(screen, rect.X+rect.W-radius, rect.Y+radius, radius, rect.H-radius*2, clr)
	vector.DrawFilledCircle(screen, float32(rect.X+radius), float32(rect.Y+radius), float32(radius), clr, false)
	vector.DrawFilledCircle(screen, float32(rect.X+rect.W-radius), float32(rect.Y+radius), float32(radius), clr, false)
	vector.DrawFilledCircle(screen, float32(rect.X+radius), float32(rect.Y+rect.H-radius), float32(radius), clr, false)
	vector.DrawFilledCircle(screen, float32(rect.X+rect.W-radius), float32(rect.Y+rect.H-radius), float32(radius), clr, false)
}

func fillCapsule(screen *ebiten.Image, rect Rect, clr color.Color) {
	if rect.W <= 0 || rect.H <= 0 {
		return
	}
	radius := rect.H / 2
	if radius <= 0 {
		return
	}
	if radius > rect.W/2 {
		radius = rect.W / 2
	}
	centerWidth := rect.W - radius*2
	if centerWidth > 0 {
		resetFillRect(screen, rect.X+radius, rect.Y, centerWidth, rect.H, clr)
	}
	vector.DrawFilledCircle(screen, float32(rect.X+radius), float32(rect.Y+radius), float32(radius), clr, false)
	vector.DrawFilledCircle(screen, float32(rect.X+rect.W-radius), float32(rect.Y+radius), float32(radius), clr, false)
}

func drawRoundedPanel(screen *ebiten.Image, rect Rect, radius, border int, fill, borderColor color.RGBA) {
	if border > 0 {
		fillRoundedRect(screen, rect, radius, borderColor)
		rect = Rect{
			X: rect.X + border,
			Y: rect.Y + border,
			W: rect.W - border*2,
			H: rect.H - border*2,
		}
		radius -= border
		if radius < 0 {
			radius = 0
		}
	}
	fillRoundedRect(screen, rect, radius, fill)
}

func (g *GUI) drawScene(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	g.drawRoomBackdrop(screen)
	g.drawTableSurface(screen)
	g.drawTableHUD(screen, view)
	if view.Phase != baseui.PhaseWelcome {
		g.drawStatusRibbon(screen, view, selected)
	}
}

func (g *GUI) drawRoomBackdrop(screen *ebiten.Image) {
	gameArea := Rect{
		X: g.sc.OffX,
		Y: g.sc.OffY,
		W: g.sc.GameW,
		H: g.sc.GameH,
	}
	resetFillRect(screen, gameArea.X, gameArea.Y, gameArea.W, gameArea.H, roomBaseColor)

	glowRadius := g.sc.PXAbsolute(180)
	vector.DrawFilledCircle(screen,
		float32(gameArea.X+g.sc.PXAbsolute(92)),
		float32(gameArea.Y+g.sc.PXAbsolute(76)),
		float32(glowRadius),
		roomGlowColor, false)
	vector.DrawFilledCircle(screen,
		float32(gameArea.X+gameArea.W-g.sc.PXAbsolute(84)),
		float32(gameArea.Y+g.sc.PXAbsolute(92)),
		float32(glowRadius),
		tintWithAlpha(roomGlowColor, 0x5a), false)
	vector.DrawFilledCircle(screen,
		float32(gameArea.X+gameArea.W/2),
		float32(gameArea.Y+gameArea.H-g.sc.PXAbsolute(32)),
		float32(g.sc.PXAbsolute(220)),
		tintWithAlpha(roomGlowColor, 0x44), false)
}

func (g *GUI) drawTableSurface(screen *ebiten.Image) {
	outer := g.tableOuterBounds()
	shadow := outer
	shadow.Y += g.sc.PXAbsolute(10)
	fillCapsule(screen, shadow, tableShadowColor)

	fillCapsule(screen, outer, tableRailDarkColor)

	railInset := g.sc.PXAbsolute(8)
	rail := Rect{
		X: outer.X + railInset,
		Y: outer.Y + railInset,
		W: outer.W - railInset*2,
		H: outer.H - railInset*2,
	}
	fillCapsule(screen, rail, tableRailLightColor)

	inner := g.tableInnerBounds()
	fillCapsule(screen, inner, tableFeltOuterColor)

	innerInset := g.sc.PXAbsolute(8)
	core := Rect{
		X: inner.X + innerInset,
		Y: inner.Y + innerInset,
		W: inner.W - innerInset*2,
		H: inner.H - innerInset*2,
	}
	fillCapsule(screen, core, tableFeltInnerColor)

	lineInsetX := g.sc.PXAbsolute(50)
	lineY := core.Y + core.H/2
	vector.StrokeLine(screen,
		float32(core.X+lineInsetX),
		float32(lineY),
		float32(core.X+core.W-lineInsetX),
		float32(lineY),
		float32(maxInt(1, g.sc.PXAbsolute(1))),
		tableFeltLineColor, false)

	vector.StrokeCircle(screen,
		float32(core.X+core.W/2),
		float32(core.Y+core.H/2),
		float32(g.sc.PXAbsolute(72)),
		float32(maxInt(1, g.sc.PXAbsolute(2))),
		tableFeltLineColor, false)
	vector.DrawFilledCircle(screen,
		float32(core.X+core.W/2),
		float32(core.Y+core.H/2),
		float32(g.sc.PXAbsolute(76)),
		tableCenterGlowColor, false)
}

func (g *GUI) drawTableHUD(screen *ebiten.Image, view baseui.TableView) {
	inner := g.tableInnerBounds()
	chipY := inner.Y + g.sc.PXAbsolute(18)
	x := inner.X + g.sc.PXAbsolute(34)
	gap := g.sc.PXAbsolute(10)

	chips := []struct {
		text   string
		fill   color.RGBA
		border color.RGBA
	}{
		{text: "主牌 " + g.trumpStatusText(view), fill: tintWithAlpha(panelMutedColor, 0xf0), border: tintWithAlpha(panelBorderColor, 0xb8)},
		{text: "叫主 " + g.bidderStatusText(view), fill: tintWithAlpha(seatBadgeColor, 0xf0), border: tintWithAlpha(seatBidderColor, 0xb8)},
		{text: "己级 " + g.levelStatusText(view.DealerLevel), fill: tintWithAlpha(color.RGBA{0x39, 0x28, 0x16, 0xf0}, 0xf0), border: tintWithAlpha(seatDealerColor, 0xb8)},
		{text: "敌级 " + g.levelStatusText(view.OpponentLevel), fill: tintWithAlpha(color.RGBA{0x1f, 0x23, 0x34, 0xf0}, 0xf0), border: tintWithAlpha(color.RGBA{0x8b, 0xa9, 0xd9, 0xb8}, 0xb8)},
		{text: "闲分 " + g.scoreStatusText(view.NonDealerScore), fill: tintWithAlpha(color.RGBA{0x2c, 0x20, 0x16, 0xf0}, 0xf0), border: tintWithAlpha(hiliteColor, 0xb0)},
		{text: "第 " + g.scoreStatusText(view.TrickCount+1) + " 墩", fill: tintWithAlpha(color.RGBA{0x18, 0x24, 0x29, 0xf0}, 0xf0), border: tintWithAlpha(color.RGBA{0x8c, 0xa9, 0xae, 0xb8}, 0xb8)},
	}

	for _, chip := range chips {
		w := g.textWidth(chip.text) + g.sc.PXAbsolute(26)
		h := g.sc.PXAbsolute(24)
		drawRoundedPanel(screen, Rect{X: x, Y: chipY, W: w, H: h},
			g.sc.PXAbsolute(10), 1, chip.fill, chip.border)
		textY := chipY + h/2 + int(g.sc.FontSize()*0.28)
		g.physText(screen, chip.text, x+g.sc.PXAbsolute(12), textY, color.White)
		x += w + gap
	}
}

func (g *GUI) drawStatusRibbon(screen *ebiten.Image, view baseui.TableView, selected map[int]bool) {
	text := g.statusLine(view, selected)
	if strings.TrimSpace(text) == "" {
		return
	}
	w := minInt(g.sc.PXAbsolute(470), g.textWidth(text)+g.sc.PXAbsolute(34))
	h := g.sc.PXAbsolute(28)
	x := g.sc.PX(RefWidth/2) - w/2
	y := g.sc.PX(408)
	drawRoundedPanel(screen, Rect{X: x, Y: y, W: w, H: h},
		g.sc.PXAbsolute(12), 1, statusRibbonFill, statusRibbonBorder)
	textX := x + (w-g.textWidth(text))/2
	textY := y + h/2 + int(g.sc.FontSize()*0.3)
	g.physText(screen, text, textX, textY, color.White)
}

func (g *GUI) drawCardRowCentered(screen *ebiten.Image, cards []baseui.CardView, centerX, y, gap int) {
	if len(cards) == 0 {
		return
	}
	cardW, _ := g.sc.CardPhysSize()
	totalW := cardW + gap*(len(cards)-1)
	x := centerX - totalW/2
	for _, c := range cards {
		g.drawCardPhys(screen, x, y, c, false, 1)
		x += gap
	}
}

func (g *GUI) trumpStatusText(view baseui.TableView) string {
	switch view.BidderSuitSymbol {
	case "":
		return "未定"
	case "🃏":
		return "无主"
	default:
		return view.BidderSuitSymbol
	}
}

func (g *GUI) bidderStatusText(view baseui.TableView) string {
	if view.BidderDirection == "" || view.BidderDirection == "---" {
		return "未亮"
	}
	return view.BidderDirection
}

func (g *GUI) levelStatusText(level string) string {
	if level == "" {
		return "2"
	}
	return level
}

func (g *GUI) scoreStatusText(score int) string {
	return strconv.Itoa(score)
}
