package main

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/hippodribble/fynewidgets"
	"github.com/hippodribble/fynewidgets/graph"
)

func main() {
	a := app.NewWithID("com.github.hippodribble.graph")
	w := a.NewWindow("Test Graph Nodes")
	w.Resize(fyne.NewSize(800,600))
	w.SetContent(gui())
	w.ShowAndRun()
}

func gui() fyne.CanvasObject {

	some_nodes, some_edges := make_random_nodes_and_edges()
	g := graph.NewGraph(some_nodes, some_edges)
	graphdraw := graph.NewGraphDrawing(g)

	var a, b float64
	a = .95
	b = .95
	sliderA := fynewidgets.NewSliderLabel(.01, .99, .01)
	sliderA.SetOnChanged(func(f float64) { a = f; g.SetRelaxParams(a, b); graphdraw.Refresh() })
	sliderB := fynewidgets.NewSliderLabel(.01, .99, .01)
	sliderB.SetOnChanged(func(f float64) { b = f; g.SetRelaxParams(a, b); graphdraw.Refresh() })
	sliderA.SetVertical(true)
	sliderB.SetVertical(true)

	left := container.NewGridWithColumns(1, sliderA, sliderB)
	return container.NewBorder(nil, nil, left, nil, graphdraw)
}

func make_text_nodes_and_edges() ([]*graph.Node, []*graph.Edge) {
	some_nodes := []*graph.Node{}
	some_edges := []*graph.Edge{}

	some_nodes = append(some_nodes, graph.NewNode("N01", 1, 10))
	some_nodes = append(some_nodes, graph.NewNode("N02", 100, 100))
	some_nodes = append(some_nodes, graph.NewNode("N03", 0, 1))
	some_nodes = append(some_nodes, graph.NewNode("N04", 10, 1))
	some_nodes = append(some_nodes, graph.NewNode("N05", 100, 1))

	some_edges = append(some_edges, graph.NewEdge(some_nodes[0], some_nodes[1], 1))
	some_edges = append(some_edges, graph.NewEdge(some_nodes[0], some_nodes[2], 1))
	some_edges = append(some_edges, graph.NewEdge(some_nodes[2], some_nodes[3], 1))
	some_edges = append(some_edges, graph.NewEdge(some_nodes[2], some_nodes[4], 1))

	return some_nodes, some_edges

}

func make_random_nodes_and_edges() ([]*graph.Node, []*graph.Edge) {
	some_nodes := []*graph.Node{}
	some_edges := []*graph.Edge{}
	N := 20
	for i := range N {
		n := graph.NewNode(fmt.Sprintf("N%02d", i+1), rand.Float64()*90+10, rand.Float64()*900+100)
		n.Value = float64(2 * (i + 1))
		// n.Value=100
		some_nodes = append(some_nodes, n)
	}

	for range 1 {
		for i := range N {
			var edge *graph.Edge
			switch i {
			case 0:
				continue
			case 1:
				edge = graph.NewEdge(some_nodes[i], some_nodes[i-1], 10)
			case 2:
				edge = graph.NewEdge(some_nodes[i], some_nodes[i-1], 10)

			default:
				if rand.Float32() > 0.5 {
					edge = graph.NewEdge(some_nodes[i], some_nodes[i-2], 10)
				} else {
					edge = graph.NewEdge(some_nodes[i], some_nodes[i-1], 10)
				}
			}

			some_edges = append(some_edges, edge)
		}
	}
	return some_nodes, some_edges
}
