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

var cardImages [57]*ebiten.Image
var cardImagesLoaded bool
var (
	uiFont     font.Face
	uiFontOnce sync.Once
)

func uiFontFace() font.Face {
	uiFontOnce.Do(func() {
		// 尝试加载系统自带中文字体
		var fontPath string
		switch runtime.GOOS {
		case "windows":
			// Windows 中文字体
			candidates := []string{
				"C:\\Windows\\Fonts\\msyh.ttc",    // 微软雅黑
				"C:\\Windows\\Fonts\\simsun.ttc",   // 宋体
				"C:\\Windows\\Fonts\\simhei.ttf",   // 黑体
				"C:\\Windows\\Fonts\\yahei.ttf",
			}
			for _, p := range candidates {
				if _, err := os.Stat(p); err == nil {
					fontPath = p
					break
				}
			}
		case "linux":
			candidates := []string{
				"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
				"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
				"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc", // 文泉驿
			}
			for _, p := range candidates {
				if _, err := os.Stat(p); err == nil {
					fontPath = p
					break
				}
			}
		}

		if fontPath != "" {
			data, err := os.ReadFile(fontPath)
			if err == nil {
				tt, err := opentype.Parse(data)
				if err == nil {
					f, err := opentype.NewFace(tt, &opentype.FaceOptions{
						Size:    14,
						DPI:     72,
						Hinting: font.HintingFull,
					})
					if err == nil {
						uiFont = f
						return
					}
				}
			}
		}
		// 方案二：加载本地assets/fonts/下的字体
		dir := assetDir()
		fontCandidates := []string{
			filepath.Join(dir, "fonts", "NotoSansCJK-Regular.ttc"),
			filepath.Join(dir, "fonts", "msyh.ttf"),
			filepath.Join(dir, "fonts", "simsun.ttc"),
		}
		for _, p := range fontCandidates {
			data, err := os.ReadFile(p)
			if err == nil {
				tt, err := opentype.Parse(data)
				if err == nil {
					f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
						Size: 14, DPI: 72, Hinting: font.HintingFull,
					})
					if f != nil {
						uiFont = f
						return
					}
				}
			}
		}
		// 兜底：basicfont（不支持中文）
		uiFont = basicfont.Face7x13
	})
	return uiFont
}
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

func EnsureImagesLoaded() error {
	if cardImagesLoaded {
		return nil
	}
	dir := assetDir()
	for i := 0; i < 54; i++ {
		path := filepath.Join(dir, "cards", "faces", fmt.Sprintf("%d.png", i))
		img, err := loadImage(path)
		if err != nil {
			return fmt.Errorf("load card %d: %w", i, err)
		}
		cardImages[i] = img
	}
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
	return nil
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
	if !cardImagesLoaded { return nil }
	idx := GoCardToCSharpNumber(suit, rank)
	if idx < 0 || idx >= 54 { return nil }
	return cardImages[idx]
}

func CardBackImage(index int) *ebiten.Image {
	if !cardImagesLoaded || index < 0 || index > 2 { return cardImages[54] }
	return cardImages[54+index]
}

func IsImageLoaded() bool { return cardImagesLoaded }

func GoCardToCSharpNumber(suit, rank int) int {
	if rank >= 15 {
		if rank == 15 { return 52 }
		if rank == 16 { return 53 }
	}
	csharpSuit := []int{2, 1, 3, 4}[suit]
	csharpRank := rank - 2
	return (csharpSuit-1)*13 + csharpRank
}

func FreeImages() {
	for i := range cardImages { cardImages[i] = nil }
	cardImagesLoaded = false
}
