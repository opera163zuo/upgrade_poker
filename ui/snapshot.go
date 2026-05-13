package ui

type SnapshotSource interface {
	UISnapshot() TableView
}

func BuildTableView(g SnapshotSource) TableView {
	return g.UISnapshot()
}
