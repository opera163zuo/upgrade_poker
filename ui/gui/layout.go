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
	CardW    = 53		// was 35, 1.5x
	CardH    = 72		// was 48, 1.5x

	SouthHandX   = 32
	SouthHandY   = 396
	SouthHandGap = 14	// was 9, 1.5x
	NorthHandY   = 86
	NorthHandGap = 9	// was 6, 1.5x
	WestHandX    = 16
	WestHandY    = 175
	WestHandGap  = 3	// was 2, 1.5x
	EastHandX    = 554
	EastHandY    = 223
	EastHandGap  = 3	// was 2, 1.5x
	BottomX      = 230
	BottomY      = 186
	BottomGap    = 11	// was 7, ~1.5x

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
