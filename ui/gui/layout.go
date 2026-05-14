package gui

const (
	MenuBarH      = 24
	LogicalWidth  = 640
	LogicalHeight = 480

	InfoBarX = 25
	InfoBarY = 28
	InfoBarW = 590
	InfoBarH = 42
	TableX   = 25
	TableY   = 74
	TableW   = 590
	TableH   = 346
	CardW    = 35
	CardH    = 48

	SouthHandX   = 32
	SouthHandY   = 396
	SouthHandGap = 9
	NorthHandY   = 86
	NorthHandGap = 6
	WestHandX    = 16
	WestHandY    = 175
	WestHandGap  = 2
	EastHandX    = 554
	EastHandY    = 223
	EastHandGap  = 2
	BottomX      = 230
	BottomY      = 186
	BottomGap    = 7

	ActionBtnW = 96
	ActionBtnH = 28
	ActionBtnY = 446

	BidPanelX      = 396
	BidPanelY      = 170
	BidPanelW      = 194
	BidPanelH      = 120
	BidSymbolSize  = 24
	BidSymbolGap   = 6
	BidPrimaryBtnW = 76
	BidSecondaryW  = 76
)

type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }
