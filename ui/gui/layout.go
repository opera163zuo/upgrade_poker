package gui

const (
	MenuBarH      = 24
	LogicalWidth  = 640
	LogicalHeight = 480

	InfoBarX = 25
	InfoBarY = 28
	InfoBarW = 590
	InfoBarH = 38
	TableX   = 25
	TableY   = 70
	TableW   = 590
	TableH   = 350
	CardW    = 71
	CardH    = 96

	SouthHandX   = 32
	SouthHandY   = 376
	SouthHandGap = 18
	NorthHandY   = 78
	NorthHandGap = 13
	WestHandX    = 16
	WestHandY    = 150
	WestHandGap  = 4
	EastHandX    = 554
	EastHandY    = 246
	EastHandGap  = 4
	BottomX      = 230
	BottomY      = 186
	BottomGap    = 14

	ActionBtnW = 96
	ActionBtnH = 28
	ActionBtnY = 438

	BidPanelX = 170
	BidPanelY = 126
	BidPanelW = 300
	BidBtnH   = 30
)

type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }
