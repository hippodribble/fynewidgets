package tree

import (
	"errors"
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Tree[T comparable] struct {
	root *Node[T]
}

func NewTree[T comparable](value T) *Tree[T] {
	return &Tree[T]{
		root: &Node[T]{
			value:    value,
			children: []*Node[T]{},
			parent:   nil,
		},
	}
}

func (t *Tree[T]) Root() *Node[T] {
	return t.root
}

func (t *Tree[T]) FindNode(value T) (*Node[T], error) {
	if t.root.value == value {
		return t.root, nil
	}
	return t.root.FindChild(value)
}

func (t *Tree[T]) MaxDepth() int {
	if t.root == nil {
		return 0
	}
	return t.root.maxDepth()
}

type Node[T comparable] struct {
	value    T
	children []*Node[T]
	parent   *Node[T]
}

func NewNode[T comparable](value T) *Node[T] {
	return &Node[T]{
		value:    value,
		children: []*Node[T]{},
		parent:   nil,
	}
}

func (n *Node[T]) Value() T {
	return n.value
}

func (n *Node[T]) Children() []*Node[T] {
	if n == nil {
		fmt.Println("nil")
	}
	return n.children
}

func (n *Node[T]) Parent() *Node[T] {
	return n.parent
}

func (n *Node[T]) AddChild(value T) (*Node[T], error) {
	if n.HasChild(value) {
		return nil, errors.New("child with this value already exists")
	}

	child := &Node[T]{
		value:    value,
		children: []*Node[T]{},
		parent:   n,
	}
	n.children = append(n.children, child)
	return child, nil
}

func (n *Node[T]) RemoveChild(child *Node[T]) {
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			child.parent = nil
			return
		}
	}
}

func (n *Node[T]) FindChild(value T) (*Node[T], error) {
	for _, child := range n.children {
		if child.value == value {
			return child, nil
		}
		// Recursively search in the child's children
		if foundChild, err := child.FindChild(value); err == nil {
			return foundChild, nil
		}
	}
	return nil, errors.New("child not found")
}

func (n *Node[T]) HasChild(value T) bool {
	for _, child := range n.children {
		if child.value == value {
			return true
		}
	}
	return false
}

func (n *Node[T]) CountDescendants() int {
	count := len(n.children)
	for _, child := range n.children {
		count += child.CountDescendants()
	}
	return count
}
func (n *Node[T]) maxDepth() int {
	if n == nil {
		return 0
	}
	maxChildDepth := 0
	for _, child := range n.children {
		depth := child.maxDepth()
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}
	return maxChildDepth + 1
}

func (n *Node[T]) Lineage() []*Node[T] {
	lineage := []*Node[T]{}
	current := n
	for current != nil {
		lineage = append([]*Node[T]{current}, lineage...)
		current = current.parent
	}
	return lineage
}

func (n *Node[T]) IsLeaf() bool {
	return n != nil && len(n.Children()) == 0
}

func (n *Node[T]) Leaves() []*Node[T] {
	if n == nil {
		return nil
	}

	leaves := make([]*Node[T], 0)
	stack := []*Node[T]{n}

	for len(stack) > 0 {
		last := len(stack) - 1
		n := stack[last]
		stack = stack[:last]

		if len(n.Children()) == 0 {
			leaves = append(leaves, n)
			continue
		}

		// Push right-to-left so visitation is left-to-right.
		for _, v := range slices.Backward(n.Children()) {
			stack = append(stack, v)
		}
	}

	return leaves
}

type NodeVis[T comparable] struct {
	widget.BaseWidget
	node      *Node[T]
	rectangle *canvas.Rectangle
	label     *canvas.Text
}

func NewNodeVis[T comparable](node *Node[T]) *NodeVis[T] {

	nv := &NodeVis[T]{
		node:      node,
		rectangle: canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
		label:     canvas.NewText(fmt.Sprintf("%v", node.Value()), theme.Color(theme.ColorNameForeground)),
	}
	nv.label.Alignment = fyne.TextAlignCenter
	nv.ExtendBaseWidget(nv)
	return nv
}

func (n *NodeVis[T]) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(n.rectangle, n.label)
	return widget.NewSimpleRenderer(c)
}

func (n *NodeVis[T]) MouseIn(event *desktop.MouseEvent) {
	n.rectangle.FillColor = theme.Color(theme.ColorNameHover)
	n.rectangle.Refresh()
}

func (n *NodeVis[T]) MouseOut() {
	n.rectangle.FillColor = theme.Color(theme.ColorNameBackground)
	n.rectangle.Refresh()
}

func (n *NodeVis[T]) MouseMoved(event *desktop.MouseEvent) {
	// No-op
}

type TreeVis[T comparable] struct {
	widget.BaseWidget
	tree   *Tree[T]
	shapes []fyne.CanvasObject
}

func NewTreeVis[T comparable](tree *Tree[T]) *TreeVis[T] {
	tv := &TreeVis[T]{
		tree:   tree,
		shapes: []fyne.CanvasObject{},
	}
	tv.ExtendBaseWidget(tv)
	return tv
}

func (t *TreeVis[T]) CreateRenderer() fyne.WidgetRenderer {
	t.shapes = []fyne.CanvasObject{}
	t.buildShapes(t.tree.Root(), 0, 0)
	layout := TreeVisLayout[T]{v: t}
	c := container.New(&layout, t.shapes...)
	return widget.NewSimpleRenderer(c)
}

func (t *TreeVis[T]) buildShapes(node *Node[T], depth int, index int) {
	nodeVis := NewNodeVis(node)
	nodeVis.Move(fyne.NewPos(float32(depth*150), float32(index*50)))
	t.shapes = append(t.shapes, nodeVis)

	for i, child := range node.Children() {
		t.buildShapes(child, depth+1, index+i)
	}
}

type TreeVisLayout[T comparable] struct {
	v *TreeVis[T]
}

func (l *TreeVisLayout[T]) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	maxdepth := l.v.tree.MaxDepth()
	// leaves := l.v.tree.Root().Leaves()
	for _, obj := range objects {
		if nv, ok := obj.(*NodeVis[T]); ok {
			lineage := nv.node.Lineage()
			depth := len(lineage) - 1

			// index := 0
			parent := lineage[depth].Parent()
			if parent == nil {
				nv.Move(fyne.NewPos(size.Width/2, 50))
				nv.rectangle.Resize(fyne.NewSize(100, 30))
				continue
			}

			// for i, sibling := range parent.Children() {
			// 	if sibling == nv.node {
			// 		index = i
			// 		break
			// 	}
			// }

			fmt.Println("Level", depth)

			// dx := size.Width / (float32(len(leaves) + 2))

			var mid float32 = 100
			pos := fyne.NewPos(mid, float32(depth*150))
			// pos := fyne.NewPos(float32(index*150), float32(depth*150))

			fmt.Println(depth, pos, maxdepth-depth, len(lineage))

			nv.Move(pos)
			nv.rectangle.Resize(fyne.NewSize(100, 30))
		}
	}
}

func (l *TreeVisLayout[T]) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(400, 400)
}
