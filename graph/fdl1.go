package graph

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Sub(v2 Vec2) Vec2 {
	return Vec2{v.X - v2.X, v.Y - v2.Y}
}

func (v Vec2) Add(v2 Vec2) Vec2 {
	return Vec2{v.X + v2.X, v.Y + v2.Y}
}

func (v Vec2) Scale(f float64) Vec2 {
	return Vec2{v.X * f, v.Y * f}
}

func (v Vec2) Norm() Vec2 {
	r := math.Sqrt(v.X*v.X + v.Y*v.Y)
	return Vec2{v.X / r, v.Y / r}
}

func (v Vec2) Mag() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// A thing
//   - it has a name, generally unique
//   - can be laid out using force-directed methods
type Node struct {
	Name         string
	EdgeList     []*Edge
	Value        float64
	Displacement Vec2
	Position     Vec2
	Force        Vec2
}

// returns a standard Node pointer
func NewNode(name string, x, y float64) *Node {
	return &Node{
		Name:     name,
		Position: Vec2{X: x, Y: y},
	}
}

func (n Node) String() string {
	return fmt.Sprintf("%s  (%.3g,%.3g)", n.Name, n.Position.X, n.Position.Y)
}

// A visible representation of a Node, with a circle and a label
type NodeVis struct {
	widget.BaseWidget
	node   *Node
	circle *canvas.Circle
	label  *canvas.Text
}

func NewNodeVis(node *Node) *NodeVis {
	n := &NodeVis{
		node:   node,
		circle: canvas.NewCircle(theme.Color(theme.ColorNameBackground)),
		label:  canvas.NewText(node.Name, color.Black),
	}
	n.circle.StrokeWidth = 1
	n.circle.StrokeColor = theme.Color(theme.ColorNameForeground)
	n.circle.Resize(fyne.NewSize(20, 20))
	n.label.Text = node.Name
	n.label.TextSize = 14
	n.label.Color = theme.Color(theme.ColorNameForeground)

	n.label.Alignment = fyne.TextAlignCenter
	n.ExtendBaseWidget(n)
	return n
}

func (n *NodeVis) SetStrokeWidth(w float32) {
	n.circle.StrokeWidth = w
}

func (n *NodeVis) SetTextSize(s float32) {
	n.label.TextSize = s
}

func (n *NodeVis) SetStrokeColor(c color.Color) {
	n.circle.StrokeColor = c
}

func (n *NodeVis) SetFillColor(c color.Color) {
	n.circle.FillColor = c
}

func (n *NodeVis) SetFontColor(c color.Color) {
	n.label.Color = c
}

func (n *NodeVis) MinSize() fyne.Size {
	return fyne.NewSize(50, 10)
}

func (n *NodeVis) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(n.circle, n.label)
	return widget.NewSimpleRenderer(c)
}

func (n *NodeVis) MouseDown(_ *desktop.MouseEvent) {
	n.circle.FillColor = colornames.Orange
	n.label.Color = colornames.Black
	n.circle.Refresh()
}

func (n *NodeVis) MouseUp(_ *desktop.MouseEvent) {
	n.circle.FillColor = theme.Color(theme.ColorNameBackground)
	n.label.Color = theme.Color(theme.ColorNameForeground)
	n.circle.Refresh()
}

func (n *NodeVis) MouseIn(_ *desktop.MouseEvent) {
	n.circle.FillColor = colornames.Orange
	n.label.Color = colornames.Black
	n.circle.Refresh()
}

func (n *NodeVis) MouseOut() {
	n.circle.FillColor = theme.Color(theme.ColorNameBackground)
	n.label.Color = theme.Color(theme.ColorNameForeground)
	n.circle.Refresh()
}

func (n *NodeVis) MouseMoved(_ *desktop.MouseEvent) {
	// n.circle.FillColor = theme.Color(theme.ColorNameBackground)
	// n.circle.Refresh()
}

// Represents a connection between two Nodes
//   - it has start and end nodes - useful for directed graphs
//   - it has a value associated with the edge that can be used to assist in layout, represent costs, etc
type Edge struct {
	Start, End *Node
	Value      float64
}

// returns a standard Edge pointer
func NewEdge(start, end *Node, value float64) *Edge {
	return &Edge{
		Start: start,
		End:   end,
		Value: value,
	}
}

func (e Edge) String() string {
	return fmt.Sprintf("%s - %s (%.3g) (%.3g,%.3g)-(%.3g,%.3g)", e.Start.Name, e.End.Name, e.Value, e.Start.Position.X, e.Start.Position.Y, e.End.Position.X, e.End.Position.Y)
}

// Visible graph edge, with a line drawn between the two nodes
type EdgeVis struct {
	widget.BaseWidget
	edge *Edge
	line *canvas.Line
}

func NewEdgeVis(e *Edge) {
	ev := &EdgeVis{
		edge: e,
		line: canvas.NewLine(theme.Color(theme.ColorNameForeground)),
	}

	ev.line.StrokeWidth = 1
	ev.ExtendBaseWidget(ev)
}

func (ev *EdgeVis) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewStack(ev.line)
	return widget.NewSimpleRenderer(c)
}

// Represents a set of *Nodes and the *Edges that connect them
type Graph struct {
	Nodes      []*Node
	Edges      []*Edge
	relA, relB float64
}

func NewGraph(nodes []*Node, edges []*Edge) *Graph {
	g := &Graph{
		Nodes: nodes,
		Edges: edges,
		relA:  0.95,
		relB:  0.95,
	}
	return g
}

func (g *Graph) String() string {
	s := ""
	s += fmt.Sprintf("Graph - %d nodes, %d edges\n", len(g.Nodes), len(g.Edges))
	for _, n := range g.Nodes {
		s += fmt.Sprintf("%+v\n", n)
		for j, edge := range n.EdgeList {
			s += fmt.Sprintf(" edge %d %+v\n", j, edge)
		}
	}
	for k, e := range g.Edges {
		s += fmt.Sprintf("Edge %d %+v\n", k, e)
	}
	return s
}

func (d *Graph) SetRelaxParams(a, b float64) {
	d.relA = a
	d.relB = b
}

func (g *Graph) Recentre(w, h float64) {

	// rescale to screen centre
	bignum := 1.0e+20
	minx := bignum
	miny := bignum
	maxx := -bignum
	maxy := -bignum
	for _, v := range g.Nodes {
		if v.Position.X < minx {
			minx = v.Position.X
		}
		if v.Position.Y < miny {
			miny = v.Position.Y
		}
		if v.Position.X > maxx {
			maxx = v.Position.X
		}
		if v.Position.Y > maxy {
			maxy = v.Position.Y
		}
	}

	dx := maxx - minx
	dy := maxy - miny

	cx := minx + maxx
	cx /= 2
	cy := miny + maxy
	cy /= 2

	W := w / 2
	H := h / 2

	sx := W / dx
	sy := H / dy

	for _, v := range g.Nodes {
		p := v.Position
		x := p.X - cx
		y := p.Y - cy
		x *= sx * 1.9
		y *= sy * 1.9
		x += W
		y += H

		v.Position = Vec2{x, y}
	}
}

// relaxes a graph network's node positions using Eades' method
//
//   - M - number of iterations
//   - w,h - size of canvas for display
//   - cooling - annealing factor (exponential decay, ao maybe 0.99 for 100 iterations)
func (g *Graph) RelaxEades(M int, w, h, cooling, c1, c2, c3, c4 float64) {

	T := min(w, h) / 10

	Fr := func(d float64) float64 {
		return c3 / d / d
	}

	Fa := func(d float64) float64 {
		return c1 * math.Log10(d/c2)
	}

	for iteration := 0; iteration < M; iteration++ {

		for i, v := range g.Nodes {
			v.Displacement = Vec2{}
			for j, u := range g.Nodes {
				if i == j {
					continue
				} // repel
				δ := v.Position.Sub(u.Position)
				v.Displacement = v.Displacement.Add(δ.Norm().Scale(Fr(δ.Mag())))
			}
		}

		for _, e := range g.Edges { // attract
			δ := e.Start.Position.Sub(e.End.Position)
			h := δ.Norm().Scale(Fa(δ.Mag()))
			e.Start.Displacement = e.Start.Displacement.Sub(h)
			e.End.Displacement = e.End.Displacement.Add(h)
		}

		for _, v := range g.Nodes { // limit displacement
			v.Position = v.Position.Add(v.Displacement.Norm().Scale(min(v.Displacement.Mag(), T)))
			// v.Position.X = min(w, max(0, v.Position.X))
			// v.Position.Y = min(h, max(0, v.Position.Y))
		}

		T *= cooling // slow it down a bit
	}
	g.Recentre(w, h) // fit it on a page

}

// relaxes a graph network's node positions using Fruchterman / Rheingold
//
//   - M - number of iterations
//   - w,h - size of canvas for display
//   - cooling - annealing factor (exponential decay, ao maybe 0.99 for 100 iterations)
//   - scale - sets the k-value in the force calculations, where k = c x sqrt(w*h/N) for N nodes
func (g *Graph) RelaxFR(M int, w, h, cooling, scale float64) {
	T := min(w, h) / 30

	k := math.Sqrt(w*h/float64(len(g.Nodes))) * scale

	Fr := func(d float64) float64 {
		return k * k / d
	}

	Fa := func(d float64) float64 {
		return d * d / k
	}

	for iteration := 0; iteration < M; iteration++ {

		for i, v := range g.Nodes {
			v.Displacement = Vec2{}
			for j, u := range g.Nodes {
				if i == j {
					continue
				}
				δ := v.Position.Sub(u.Position)
				v.Displacement = v.Displacement.Add(δ.Norm().Scale(Fr(δ.Mag())))
			}
		}

		for _, e := range g.Edges {
			δ := e.Start.Position.Sub(e.End.Position)
			h := δ.Norm().Scale(Fa(δ.Mag()))
			e.Start.Displacement = e.Start.Displacement.Sub(h)
			e.End.Displacement = e.End.Displacement.Add(h)
		}

		for _, v := range g.Nodes {
			v.Position = v.Position.Add(v.Displacement.Norm().Scale(min(v.Displacement.Mag(), T)))
			// v.Position.X = min(w, max(0, v.Position.X))
			// v.Position.Y = min(h, max(0, v.Position.Y))
		}
		g.Recentre(w, h)

		T *= cooling
	}

}

// relaxes a graph network's node positions using Fruchterman / Rheingold
//
//   - M - number of iterations
//   - w,h - size of canvas for display
//   - cooling - annealing factor (exponential decay, ao maybe 0.99 for 100 iterations)
//   - scale - sets the k-value in the force calculations, where k = c x sqrt(w*h/N) for N nodes
func (g *Graph) RelaxFRNodeWeighted(Niterations int, w, h, cooling, scale float64) {
	T := min(w, h) / 30

	k := math.Sqrt(w*h/float64(len(g.Nodes))) * scale

	Fr := func(d float64) float64 {
		return k * k / d
	}

	Fa := func(d float64) float64 {
		return d * d / k
	}

	for range Niterations {

		for i, v := range g.Nodes {
			v.Displacement = Vec2{}
			m1 := v.Value
			for j, u := range g.Nodes {
				if i == j {
					continue // same node - do nothing
				}
				m2 := u.Value
				distanceVector := v.Position.Sub(u.Position)                          // distance between nodes
				distanceNorm := distanceVector.Norm()                                 // unit vector in direction of distance
				fRepulsionDistanceMagnitude := Fr(distanceVector.Mag())               // magnitude of repulsion force due to distance
				dRepulsionDistance := distanceNorm.Scale(fRepulsionDistanceMagnitude) // repulsion force vector due to distance

				r := distanceVector.Mag() * scale
				mag := m1 * m2 / r / r
				vrd := distanceNorm.Scale(mag) // repulsion force vector due to node mass

				v.Displacement = v.Displacement.Add(dRepulsionDistance).Add(vrd) // updated displacement for node v due to repulsion from node u

				// v.Displacement = v.Displacement.Add(dRepulsionDistance)               // updated displacement for node v due to repulsion from node u

			}
		}

		for _, e := range g.Edges {
			δ := e.Start.Position.Sub(e.End.Position)
			h := δ.Norm().Scale(Fa(δ.Mag()))
			e.Start.Displacement = e.Start.Displacement.Sub(h)
			e.End.Displacement = e.End.Displacement.Add(h)
		}

		for _, v := range g.Nodes {
			v.Position = v.Position.Add(v.Displacement.Norm().Scale(min(v.Displacement.Mag(), T)))
		}
		g.Recentre(w, h)

		T *= cooling
	}

}

// representation of a graph, with nodes and edges, for display in a fyne canvas
type GraphDrawing struct {
	widget.BaseWidget
	Graph  *Graph
	shapes []fyne.CanvasObject
}

func NewGraphDrawing(g *Graph) fyne.CanvasObject {
	d := &GraphDrawing{
		Graph: g,
	}

	for _, edge := range g.Edges {
		// fmt.Println("Adding edge")
		l := &EdgeVis{edge: edge}
		l.line = canvas.NewLine(colornames.Lightgray)
		d.shapes = append(d.shapes, l)
	}
	for _, node := range g.Nodes {
		// fmt.Println("New node")
		nv := NewNodeVis(node)
		d.shapes = append(d.shapes, nv)
	}

	d.ExtendBaseWidget(d)
	return d
}

func (d *GraphDrawing) CreateRenderer() fyne.WidgetRenderer {
	c := container.New(NewGraphLayout(d.Graph), d.shapes...)
	return widget.NewSimpleRenderer(c)
}

func (d *GraphDrawing) Relax(N int) {
	// B := int(math.Sqrt(float64(N)))
	d.Graph.Recentre(float64(d.Size().Width), float64(d.Size().Height))
	for i := 0; i < 1; i++ {
		d.Graph.RelaxFR(1, float64(d.Size().Width), float64(d.Size().Height), .9, 1)
		d.Refresh()
		time.Sleep(time.Millisecond * 100)
	}
}

type GraphLayout struct {
	Graph *Graph
}

func NewGraphLayout(g *Graph) fyne.Layout {
	l := &GraphLayout{Graph: g}
	return l
}

func (r *GraphLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(500,500)
}

func (r *GraphLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {

	fmt.Println(r.Graph.relA, r.Graph.relB)

	// r.Graph.RelaxFR(1000, float64(size.Width), float64(size.Height), r.Graph.relA, r.Graph.relB)
	r.Graph.RelaxFRNodeWeighted(1000, float64(size.Width), float64(size.Height), r.Graph.relA, r.Graph.relB)
	// r.Graph.RelaxEades(50,float64(size.Width), float64(size.Height),.5,.95,.95,.95 ,.95)

	for _, o := range objects {

		if v, ok := o.(*NodeVis); ok {
			v.Move(fyne.NewPos(float32(v.node.Position.X-float64(v.circle.Size().Height)/2), float32(v.node.Position.Y-float64(v.circle.Size().Height)/2)))
			v.Resize(fyne.NewSize(50, 50))
		}

		if v, ok := o.(*EdgeVis); ok {
			v.line.Position1.X = float32(v.edge.Start.Position.X)
			v.line.Position1.Y = float32(v.edge.Start.Position.Y)
			v.line.Position2.X = float32(v.edge.End.Position.X)
			v.line.Position2.Y = float32(v.edge.End.Position.Y)
			v.line.Refresh()
		}

	}
}
