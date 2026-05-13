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

func loadFont(path string) font.Face {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// 先尝试作为TTF/OTF解析
	tt, err := opentype.Parse(data)
	if err == nil {
		f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
			Size: 14, DPI: 72, Hinting: font.HintingFull,
		})
		if f != nil {
			return f
		}
	}
	// 尝试作为TTC（字体集）解析
	col, err := opentype.ParseCollection(data)
	if err == nil && col.NumFonts() > 0 {
		for i := 0; i < col.NumFonts(); i++ {
			tt, err := col.Font(i)
			if err == nil {
				f, _ := opentype.NewFace(tt, &opentype.FaceOptions{
					Size: 14, DPI: 72, Hinting: font.HintingFull,
				})
				if f != nil {
					return f
				}
			}
		}
	}
	return nil
}

func uiFontFace() font.Face {
	uiFontOnce.Do(func() {
		// 优先加载项目自带字体（assets/fonts/）
		dir := assetDir()
		fontPath := filepath.Join(dir, "fonts", "wqy-microhei.ttc")
		if f := loadFont(fontPath); f != nil {
			uiFont = f
			return
		}
		// 方案二：系统字体
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

	// 先尝试加载高清牌图（500x726，文件名如 10♠.png）
	hiresDir := filepath.Join(dir, "cards", "hires")
	if info, err := os.Stat(hiresDir); err == nil && info.IsDir() {
		// 两副牌需要两个副本，直接加载0-53
		for i := 0; i < 54; i++ {
			filename := cardNumberToHiresName(i)
			path := filepath.Join(hiresDir, filename)
			img, err := loadImage(path)
			if err != nil {
				// 回退到faces目录
				path = filepath.Join(dir, "cards", "faces", fmt.Sprintf("%d.png", i))
				img, err = loadImage(path)
				if err != nil {
					return fmt.Errorf("load card %d: %w", i, err)
				}
			}
			cardImages[i] = img
		}
	} else {
		// 旧版faces目录编号牌图
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
	return nil
}

// cardNumberToHiresName converts C# card number 0-53 to high-res filename
// 0-12=♥, 13-25=♠, 26-38=♦, 39-51=♣, 52=小王, 53=大王
func cardNumberToHiresName(n int) string {
	if n >= 52 {
		if n == 52 { return "BJ.png" } // black joker
		if n == 53 { return "RJ.png" } // red joker
	}
	suitIdx := n / 13
	rankIdx := n % 13
	suits := []string{"♥", "♠", "♦", "♣"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	return ranks[rankIdx] + suits[suitIdx] + ".png"
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

func FreeImages() {
	for i := range cardImages {
		cardImages[i] = nil
	}
	cardImagesLoaded = false
}
