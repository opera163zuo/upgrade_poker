package gui

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

// ════════════════════════════════════════════════════════════════════════════
// 方案二：扑克牌目标尺寸缓存 + 文字高质量链路
//
// 1. 牌面图：不每帧临时缩放 500×726 到小尺寸，而是按当前 DPI/窗口尺寸
//    预生成目标尺寸缓存牌图，运行时直接绘制缓存，避免二次缩放。
// 2. 字体：按 Windows 实际 DPI/物理字号直接创建字体面，文字绘制像素对齐。
// ════════════════════════════════════════════════════════════════════════════

// ----- 牌面缓存 -----------------------------------------------------------

var (
	cardImages        [57]*ebiten.Image // 原始高清牌图（500×726）
	cardImagesLoaded  bool

	// 物理尺寸缓存：key = (physW << 16) | physH
	cardPhysCache     map[uint32]*[57]*ebiten.Image
	cardPhysCacheMu   sync.Mutex
	cardPhysCacheKey  uint32
)

func physCacheKey(w, h int) uint32 {
	return (uint32(w) << 16) | uint32(h)
}

// EnsureImagesLoaded 加载原始高清牌图
func EnsureImagesLoaded() error {
	if cardImagesLoaded {
		return nil
	}
	dir := assetDir()

	// 先尝试高清牌图（500×726）
	hiresDir := filepath.Join(dir, "cards", "hires")
	if info, err := os.Stat(hiresDir); err == nil && info.IsDir() {
		for i := 0; i < 54; i++ {
			filename := cardNumberToHiresName(i)
			path := filepath.Join(hiresDir, filename)
			img, err := loadImage(path)
			if err != nil {
				path = filepath.Join(dir, "cards", "faces", fmt.Sprintf("%d.png", i))
				img, err = loadImage(path)
				if err != nil {
					return fmt.Errorf("load card %d: %w", i, err)
				}
			}
			cardImages[i] = img
		}
	} else {
		for i := 0; i < 54; i++ {
			path := filepath.Join(dir, "cards", "faces", fmt.Sprintf("%d.png", i))
			img, err := loadImage(path)
			if err != nil {
				return fmt.Errorf("load card %d: %w", i, err)
			}
			cardImages[i] = img
		}
	}
	// 牌背
	backNames := []string{"back.png", "back2.png", "back3.png"}
	for i, name := range backNames {
		path := filepath.Join(dir, "cards", "backs", name)
		img, err := loadImage(path)
		if err != nil {
			return err
		}
		cardImages[54+i] = img
	}
	cardImagesLoaded = true
	cardPhysCache = make(map[uint32]*[57]*ebiten.Image)
	return nil
}

// ensurePhysCardCache 确保指定物理尺寸的牌面缓存已就绪
// 将原始 500×726 牌图缩放到目标尺寸并缓存，每次窗口/DPI 变化时重建
func ensurePhysCardCache(physW, physH int) {
	if !cardImagesLoaded || physW <= 0 || physH <= 0 {
		return
	}
	key := physCacheKey(physW, physH)
	cardPhysCacheMu.Lock()
	defer cardPhysCacheMu.Unlock()

	if _, ok := cardPhysCache[key]; ok {
		return // 已有缓存
	}

	// 重建该尺寸缓存
	cache := &[57]*ebiten.Image{}
	for i := 0; i < len(cardImages); i++ {
		src := cardImages[i]
		if src == nil {
			continue
		}
		// 使用 FilterNearest 或 FilterLinear？
		// 扑克牌是矢量风格的矩形图，用 FilterLinear 做高质量下采样，
		// 但因为要消除"先在低分辨率画布缩放再整体放大"的问题，
		// 这里直接用 FilterLinear 一次性缩放到物理目标尺寸。
		// 后续绘制时直接 blit，不再二次缩放 —— 一举消除模糊。
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()
		op.GeoM.Scale(float64(physW)/float64(srcW), float64(physH)/float64(srcH))
		dst := ebiten.NewImage(physW, physH)
		dst.DrawImage(src, op)
		cache[i] = dst
	}
	cardPhysCache[key] = cache
	cardPhysCacheKey = key
}

// CardFaceImagePhys 返回指定花色的物理尺寸牌面缓存图
func CardFaceImagePhys(suit, rank, physW, physH int) *ebiten.Image {
	key := physCacheKey(physW, physH)
	cardPhysCacheMu.Lock()
	cache, ok := cardPhysCache[key]
	cardPhysCacheMu.Unlock()
	if !ok {
		// 尝试惰性加载
		ensurePhysCardCache(physW, physH)
		cardPhysCacheMu.Lock()
		cache, ok = cardPhysCache[key]
		cardPhysCacheMu.Unlock()
		if !ok {
			return nil
		}
	}
	idx := GoCardToCSharpNumber(suit, rank)
	if idx < 0 || idx >= 54 {
		return nil
	}
	return cache[idx]
}

// CardBackImagePhys 返回指定索引的物理尺寸牌背缓存图
func CardBackImagePhys(index, physW, physH int) *ebiten.Image {
	key := physCacheKey(physW, physH)
	cardPhysCacheMu.Lock()
	cache, ok := cardPhysCache[key]
	cardPhysCacheMu.Unlock()
	if !ok {
		ensurePhysCardCache(physW, physH)
		cardPhysCacheMu.Lock()
		cache, ok = cardPhysCache[key]
		cardPhysCacheMu.Unlock()
		if !ok {
			return nil
		}
	}
	if index < 0 || index > 2 {
		index = 0
	}
	return cache[54+index]
}

// InvalidatePhysCache 清除所有物理尺寸缓存（窗口大小/DPI 变化时调用）
func InvalidatePhysCache() {
	cardPhysCacheMu.Lock()
	defer cardPhysCacheMu.Unlock()
	cardPhysCache = make(map[uint32]*[57]*ebiten.Image)
	cardPhysCacheKey = 0
}

// ----- 字体 ---------------------------------------------------------------

var (
	uiFont     font.Face
	uiFontOnce sync.Once
	// 当前字体参数，用于检测是否需要重新创建
	currFontSize float64
	currFontDPI  float64
	uiFontMu     sync.Mutex
)

func loadFont(path string) font.Face {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// 先尝试作为 TTF/OTF 解析
	tt, err := opentype.Parse(data)
	if err == nil {
		f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    float64(RefFontSize),
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if f != nil {
			return f
		}
	}
	// 尝试作为 TTC 解析
	col, err := opentype.ParseCollection(data)
	if err == nil && col.NumFonts() > 0 {
		for i := 0; i < col.NumFonts(); i++ {
			tt, err := col.Font(i)
			if err == nil {
				f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
					Size:    float64(RefFontSize),
					DPI:     72,
					Hinting: font.HintingFull,
				})
				if f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// uiFontFace 返回字体面。
// 方案二：按物理 DPI/字号创建，确保中文有足够像素清晰渲染。
//
// 注意：我们只创建一个参考字体面用于 text.Draw 中的度量计算。
// 实际文本绘制由 drawText 内部创建匹配当前 ScaleCtx 的临时字体。
func uiFontFace() font.Face {
	uiFontOnce.Do(func() {
		// 优先加载项目自带字体
		dir := assetDir()
		fontPath := filepath.Join(dir, "fonts", "wqy-microhei.ttc")
		if f := loadFont(fontPath); f != nil {
			uiFont = f
			return
		}
		// 系统字体
		var sysPaths []string
		switch runtime.GOOS {
		case "windows":
			sysPaths = []string{
				`C:\Windows\Fonts\msyh.ttc`,
				`C:\Windows\Fonts\simsun.ttc`,
				`C:\Windows\Fonts\simhei.ttf`,
			}
		case "linux":
			sysPaths = []string{
				"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
				"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
				"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
			}
		}
		for _, p := range sysPaths {
			if f := loadFont(p); f != nil {
				uiFont = f
				return
			}
		}
		// 兜底：basicfont（仅ASCII）
		uiFont = basicfont.Face7x13
	})
	return uiFont
}

// NewFontFaceForScale 根据当前缩放参数创建物理尺寸的字体面
func NewFontFaceForScale(sc ScaleCtx) font.Face {
	fsize := sc.FontSize()
	fdpi := sc.FontDPI()

	uiFontMu.Lock()
	defer uiFontMu.Unlock()

	if currFontSize == fsize && currFontDPI == fdpi && uiFont != nil {
		return uiFont // 重用
	}

	dir := assetDir()
	fontPath := filepath.Join(dir, "fonts", "wqy-microhei.ttc")
	f := loadFontAtSize(fontPath, fsize, fdpi)
	if f != nil {
		currFontSize = fsize
		currFontDPI = fdpi
		uiFont = f
		return f
	}

	// 尝试系统字体
	var sysPaths []string
	switch runtime.GOOS {
	case "windows":
		sysPaths = []string{
			`C:\Windows\Fonts\msyh.ttc`,
			`C:\Windows\Fonts\simsun.ttc`,
			`C:\Windows\Fonts\simhei.ttf`,
		}
	case "linux":
		sysPaths = []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
			"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		}
	}
	for _, p := range sysPaths {
		if f := loadFontAtSize(p, fsize, fdpi); f != nil {
			currFontSize = fsize
			currFontDPI = fdpi
			uiFont = f
			return f
		}
	}

	// 兜底：仍用基础字体
	return uiFont
}

func loadFontAtSize(path string, size, dpi float64) font.Face {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	tt, err := opentype.Parse(data)
	if err == nil {
		f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		})
		if f != nil {
			return f
		}
	}
	col, err := opentype.ParseCollection(data)
	if err == nil && col.NumFonts() > 0 {
		for i := 0; i < col.NumFonts(); i++ {
			tt, err := col.Font(i)
			if err == nil {
				f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
					Size:    size,
					DPI:     dpi,
					Hinting: font.HintingFull,
				})
				if f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// ----- 工具函数 -----------------------------------------------------------

func assetDir() string {
	exe, _ := os.Executable()
	base := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(base, "assets"),
		filepath.Join(base, "..", "assets"),
		"assets",
	}
	for _, d := range candidates {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d
		}
	}
	return "assets"
}

func loadImage(path string) (*ebiten.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func CardFaceImage(suit, rank int) *ebiten.Image {
	if !cardImagesLoaded {
		return nil
	}
	idx := GoCardToCSharpNumber(suit, rank)
	if idx < 0 || idx >= 54 {
		return nil
	}
	return cardImages[idx]
}

func CardBackImage(index int) *ebiten.Image {
	if !cardImagesLoaded || index < 0 || index > 2 {
		return cardImages[54]
	}
	return cardImages[54+index]
}

func IsImageLoaded() bool { return cardImagesLoaded }

// GoCardToCSharpNumber converts Go card suit/rank to C# card number 0-53
func GoCardToCSharpNumber(suit, rank int) int {
	if rank >= 15 {
		if rank == 15 {
			return 52
		}
		if rank == 16 {
			return 53
		}
	}
	csharpSuit := []int{2, 1, 3, 4}[suit]
	csharpRank := rank - 2
	return (csharpSuit-1)*13 + csharpRank
}

// cardNumberToHiresName converts C# card number 0-53 to high-res filename
func cardNumberToHiresName(n int) string {
	if n >= 52 {
		if n == 52 {
			return "BJ.png"
		}
		if n == 53 {
			return "RJ.png"
		}
	}
	suitIdx := n / 13
	rankIdx := n % 13
	suits := []string{"♥", "♠", "♦", "♣"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	return ranks[rankIdx] + suits[suitIdx] + ".png"
}

func FreeImages() {
	for i := range cardImages {
		cardImages[i] = nil
	}
	cardImagesLoaded = false
	cardPhysCacheMu.Lock()
	cardPhysCache = nil
	cardPhysCacheMu.Unlock()
}
