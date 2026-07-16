package network

import (
	"image/color"
	"net/netip"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/fynewidgets"
)

type TCPEndPoint struct {
	tcp netip.AddrPort
}

func NewTCPEndPoint(ap netip.AddrPort) TCPEndPoint {
	node := TCPEndPoint{
		tcp: ap,
	}
	return node
}

// visualises a Node, showing the IP address and port of the Node
type NodeView struct {
	fynewidgets.DraggableBaseWidget
	node   *TCPEndPoint
	textIP *canvas.Text
	r      *canvas.Rectangle
}

func NewNodeView(node *TCPEndPoint, background color.Color) *NodeView {
	ip := canvas.NewText(node.tcp.String(), theme.Color(theme.ColorNameBackground))
	ip.Alignment = fyne.TextAlignCenter
	ip.TextSize = 24
	ip.TextStyle.Bold = true
	r := canvas.NewRectangle(background)
	r.CornerRadius = 20
	v := &NodeView{node: node, textIP: ip, r: r}
	v.ExtendBaseWidget(v)
	return v
}

func (n *NodeView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(n.r, n.textIP))
}

// func (n *NodeView) MouseDown(e *desktop.MouseEvent) {
// 	n.dragpos = n.Position()
// 	n.mousedragpos=e.AbsolutePosition
// }
// func (n *NodeView) MouseUp(e *desktop.MouseEvent) { fmt.Println("click") }

// func (n *NodeView) Dragged(e *fyne.DragEvent) {
// 	pos := e.AbsolutePosition.Subtract(n.mousedragpos).Add(n.dragpos)
// 	// move:=pos.Subtract(n.startpos)
// 	n.Move(pos)
// }
// func (n *NodeView) DragEnd() {}
