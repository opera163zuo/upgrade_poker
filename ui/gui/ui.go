package gui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	baseui "github.com/smallnest/upgrade_poker/ui"
)

// ════════════════════════════════════════════════════════════════════════════
// 方案二：GUI 核心
//
// - 不再使用 640×480 离屏缓冲 + 二次放大
// - Layout() 返回实际窗口尺寸，Draw() 直接渲染到 screen
// - 所有坐标经 ScaleCtx 映射到物理像素
// - 字体按物理 DPI 创建，牌图按物理尺寸缓存
// ════════════════════════════════════════════════════════════════════════════

type GUI struct {
	st        *state
	mouseDown bool
	sc        ScaleCtx // 当前缩放上下文

	// 外部 letterbox 区域填充色
	letterboxColor color.RGBA
}

func New() *GUI {
	return &GUI{
		st:             &state{actionCh: make(chan baseui.UIAction, 16), selected: map[int]bool{}},
		letterboxColor: color.RGBA{0x0a, 0x0a, 0x0a, 0xff}, // 深灰
	}
}

func (g *GUI) Init() error {
	// 窗口初始大小：设计尺寸 * 2（1280×960），适配常见 100%~200% 缩放
	ebiten.SetWindowSize(RefWidth*2, RefHeight*2)
	// 通知 Ebitengine 我们需要高 DPI 渲染
	ebiten.SetWindowTitle("upgrade_poker - Ebitengine GUI")
	if err := EnsureImagesLoaded(); err != nil {
		// 牌图加载失败不阻断，会降级到矢量绘制
	}
	if err := EnsureSuitIconsLoaded(); err != nil {
		// 花色图标加载失败不阻断，会回退到文字渲染
	}
	return nil
}

func (g *GUI) Run(loop func()) error {
	go loop()
	return ebiten.RunGame(g)
}

func (g *GUI) Close() error { return nil }

func (g *GUI) Render(view baseui.TableView) {
	g.st.mu.Lock()
	view.UpdatedAt = time.Now()
	g.st.view = view
	if view.SelectedIdx != nil {
		g.st.selected = map[int]bool{}
		for k, v := range view.SelectedIdx {
			g.st.selected[k] = v
		}
	}
	if view.Phase != baseui.PhaseBidding {
		g.st.selectedBidType = ""
		g.st.selectedBidSuit = ""
		g.st.selectedBidChoice = ""
	}
	g.st.mu.Unlock()
}

func (g *GUI) WaitAction() baseui.UIAction {
	return <-g.st.actionCh
}

func (g *GUI) WaitActionOrTimeout(d time.Duration) (baseui.UIAction, bool) {
	select {
	case act := <-g.st.actionCh:
		return act, false
	case <-time.After(d):
		return baseui.UIAction{Type: baseui.ActionTimeout}, true
	}
}

func (g *GUI) ShowMessage(msg string, buttons []baseui.ButtonSpec) {
	g.st.mu.Lock()
	g.st.view.Message = msg
	g.st.view.Buttons = buttons
	g.st.mu.Unlock()
}

func (g *GUI) ClearMessage() {
	g.st.mu.Lock()
	g.st.view.Message = ""
	g.st.view.Buttons = nil
	g.st.mu.Unlock()
}

func (g *GUI) SetPhase(phase baseui.UIPhase) {
	g.st.mu.Lock()
	g.st.view.Phase = phase
	g.st.selected = map[int]bool{}
	g.st.mu.Unlock()
}

func (g *GUI) SleepForRedraw(d time.Duration) { time.Sleep(d) }

func (g *GUI) Update() error {
	g.updateInput()
	return nil

}

// Layout 返回实际窗口尺寸，禁止 Ebitengine 内置缩放。
// 我们在 Draw() 中直接以物理像素渲染，无任何逻辑画布→物理窗口的二次缩放。
func (g *GUI) Layout(outsideWidth, outsideHeight int) (int, int) {
	devScale := ebiten.DeviceScaleFactor()
	if devScale < 1.0 {
		devScale = 1.0
	}

	// 更新缩放上下文
	oldKey := physCacheKey(g.sc.PX(RefCardW), g.sc.PX(RefCardH))
	g.sc = BuildScaleCtx(outsideWidth, outsideHeight, devScale)

	// 检测牌图缓存是否失效，需要重建
	newKey := physCacheKey(g.sc.PX(RefCardW), g.sc.PX(RefCardH))
	if oldKey != newKey {
		InvalidatePhysCache()
	}

	return outsideWidth, outsideHeight
}

// Draw 直接渲染到 screen，无中间画布。
// 所有坐标均经 ScaleCtx 映射到物理像素。
func (g *GUI) Draw(screen *ebiten.Image) {
	// --- letterbox fill ---
	screen.Fill(g.letterboxColor)

	// 确保牌图缓存
	physW, physH := g.sc.CardPhysSize()
	ensurePhysCardCache(physW, physH)

	// 更新字体
	_ = NewFontFaceForScale(g.sc)

	// --- 渲染游戏内容到物理像素坐标 ---
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

	g.drawScene(screen, view, selected)

	g.drawNorth(screen, view)
	g.drawWest(screen, view)
	g.drawEast(screen, view)
	g.drawCenter(screen, view)
	g.drawSouth(screen, view, selected)
	g.drawMenuBar(screen, view)
	g.drawButtons(screen, view, selected)
	g.drawOverlay(screen, view, selected)
}
