package gui

// ════════════════════════════════════════════════════════════════════════════
// 方案二：统一坐标体系
//
// 设计分辨率（reference）：640×480 设计像素
// 物理渲染：按实际窗口 DPI / 缩放比将设计坐标映射到物理像素坐标，
//           不再先在低分辨率逻辑画布栅格化再整体放大。
// ════════════════════════════════════════════════════════════════════════════

const (
	// 设计尺寸（reference size）—— 用于布局比例参考，非实际渲染尺寸
	RefWidth  = 640
	RefHeight = 480

	// 牌设计尺寸
	RefCardW = 53
	RefCardH = 72

	// 菜单栏
	RefMenuBarH = 24

	// 右侧状态面板（原版纵列风格）
	RefInfoBarX = 535
	RefInfoBarY = 28
	RefInfoBarW = 100
	RefInfoBarH = 200

	// 牌桌（缩窄留出右侧状态面板）
	RefTableX = 25
	RefTableY = 74
	RefTableW = 506
	RefTableH = 346

	// 南家手牌（底部）
	RefSouthHandX   = 32
	RefSouthHandY   = 396
	RefSouthHandGap = 14

	// 北家手牌（顶部）
	RefNorthHandY   = 86
	RefNorthHandGap = 9

	// 西家手牌（左侧）
	RefWestHandX   = 16
	RefWestHandY   = 175
	RefWestHandGap = 3

	// 东家手牌（右侧，随牌桌左移）
	RefEastHandX   = 478
	RefEastHandY   = 223
	RefEastHandGap = 3

	// 底牌
	RefBottomX      = 230
	RefBottomY      = 186
	RefBottomGap    = 11

	// 操作按钮
	RefActionBtnW = 96
	RefActionBtnH = 28
	RefActionBtnY = 446

	// 亮主面板
	RefBidPanelX      = 396
	RefBidPanelY      = 170
	RefBidPanelW      = 194
	RefBidPanelH      = 120
	RefBidSymbolSize  = 24
	RefBidSymbolGap   = 6
	RefBidPrimaryBtnW = 76
	RefBidSecondaryW  = 76

	// 设计字号（参考点）
	RefFontSize = 10 // pt，比旧 8pt 略大，保证中文有足够像素

	// 亮主面板文字偏移
	RefBidTitleX = 18
	RefBidTitleY = 20
)

// Rect 表示一个矩形区域（物理像素坐标）
type Rect struct{ X, Y, W, H int }

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// ScaleCtx 包含所有物理缩放参数
type ScaleCtx struct {
	// 设备独立像素下的窗口尺寸
	WindowW int
	WindowH int

	// DPI 缩放因子（100%=1.0, 125%=1.25, 150%=1.5, 200%=2.0）
	DevScale float64

	// 有效缩放比（取整约束后的缩放）
	Scale float64

	// 物理游戏区域（letterbox 后）
	GameW int
	GameH int
	OffX  int
	OffY  int
}

// BuildScaleCtx 根据窗口尺寸和 DPI 计算缩放参数
// 策略：取能完整容纳设计尺寸的最大整数缩放比，不满足时允许非整数但尽量靠近整数
func BuildScaleCtx(windowW, windowH int, devScale float64) ScaleCtx {
	s := ScaleCtx{
		WindowW:  windowW,
		WindowH:  windowH,
		DevScale: devScale,
	}

	// 基础缩放 = 最小轴缩放
	if windowW <= 0 || windowH <= 0 {
		s.Scale = 1.0
		return s
	}
	base := float64(windowW) / float64(RefWidth)
	if hScale := float64(windowH) / float64(RefHeight); hScale < base {
		base = hScale
	}

	// 整数约束：如果 > 1.25 且接近整数，取整数
	if base > 1.25 {
		intScale := float64(int(base + 0.5))
		s.Scale = intScale
	} else {
		s.Scale = base
	}

	// 计算游戏区域（letterbox）
	s.GameW = int(float64(RefWidth) * s.Scale)
	s.GameH = int(float64(RefHeight) * s.Scale)
	s.OffX = (windowW - s.GameW) / 2
	s.OffY = (windowH - s.GameH) / 2
	if s.OffX < 0 {
		s.OffX = 0
	}
	if s.OffY < 0 {
		s.OffY = 0
	}

	return s
}

// PX 将设计坐标转换为物理像素坐标（含 letterbox 偏移）
func (s ScaleCtx) PX(design int) int {
	return s.OffX + int(float64(design)*s.Scale+0.5)
}

// PXAbsolute 将设计坐标转换为物理像素坐标（不含 letterbox 偏移）
func (s ScaleCtx) PXAbsolute(design int) int {
	return int(float64(design)*s.Scale + 0.5)
}

// FontSize 返回物理字号（点），考虑 DPI 缩放
func (s ScaleCtx) FontSize() float64 {
	return float64(RefFontSize) * s.Scale
}

// FontDPI 返回物理字体 DPI = 72 * DevScale
func (s ScaleCtx) FontDPI() float64 {
	dpi := 72.0 * s.DevScale
	if dpi < 72 {
		dpi = 72
	}
	return dpi
}

// CardPhysSize 返回牌的实际物理像素尺寸
func (s ScaleCtx) CardPhysSize() (w, h int) {
	return s.PXAbsolute(RefCardW), s.PXAbsolute(RefCardH)
}

// CardScale 返回牌图从源图到物理尺寸的缩放比
func (s ScaleCtx) CardScale(srcW, srcH int) (float64, float64) {
	dstW, dstH := s.CardPhysSize()
	return float64(dstW) / float64(srcW), float64(dstH) / float64(srcH)
}
