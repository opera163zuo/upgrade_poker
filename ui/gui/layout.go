package gui

const (
	MenuBarH      = 24
	LogicalWidth  = 640
	LogicalHeight = 480

	InfoBarX = 25
	InfoBarY = 8
	InfoBarW = 590
	InfoBarH = 24
	TableX   = 25
	TableY   = 40
	TableW   = 590
	TableH   = 400
	CardW    = 71
	CardH    = 96

	SouthHandX   = 30
	SouthHandY   = 375
	SouthHandGap = 18
	NorthHandY   = 25
	NorthHandGap = 13
	WestHandX    = 6
	WestHandY    = 145
	WestHandGap  = 4
	EastHandX    = 554
	EastHandY    = 241
	EastHandGap  = 4
	BottomX      = 230
	BottomY      = 186
	BottomGap    = 14

	ActionBtnW = 96
	ActionBtnH = 28
	ActionBtnY = 440

	BidPanelX = 170
	BidPanelY = 126
	BidPanelW = 300
	BidBtnH   = 30
)

type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }
