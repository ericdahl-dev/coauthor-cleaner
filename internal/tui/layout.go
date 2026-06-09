package tui

const (
	headerLines = 3
	footerLines = 1
	leftPct     = 42
)

type layout struct {
	width      int
	height     int
	bodyHeight int
	leftWidth  int
	rightWidth int
	innerH     int
}

func computeLayout(w, h int) layout {
	bodyH := h - headerLines - footerLines
	if bodyH < 4 {
		bodyH = 4
	}
	leftW := w * leftPct / 100
	if leftW < 28 {
		leftW = 28
	}
	if leftW > w-24 {
		leftW = w - 24
	}
	rightW := w - leftW - 1
	innerH := bodyH - 2
	if innerH < 3 {
		innerH = 3
	}
	return layout{
		width: w, height: h, bodyHeight: bodyH,
		leftWidth: leftW, rightWidth: rightW, innerH: innerH,
	}
}
