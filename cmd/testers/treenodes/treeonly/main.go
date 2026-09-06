package main

import (
	"fmt"
	"time"

	"github.com/hippodribble/fynewidgets/tree"
)

func main() {

	// N := 300

	for _, N := range []int{3} {
		thistree := tree.NewTree("root")
		t0 := time.Now()

		for i := range N {
			child, _ := thistree.Root().AddChild(fmt.Sprintf("child %02d", i+1))
			for j := range N {
				grandchild, _ := child.AddChild(fmt.Sprintf("grandchild %02d.%02d", i+1, j+1))
				for k := range N {
					grandchild.AddChild(fmt.Sprintf("great-grandchild %02d.%02d.%02d", i+1, j+1, k+1))
					// for l := range N {
					// 	greatgrandchild.AddChild(fmt.Sprintf("great-great-grandchild %02d.%02d.%02d.%02d", i+1, j+1, k+1, l+1))
					// }
				}
			}
		}

		t1 := time.Now()
		dur := float64(t1.Sub(t0).Milliseconds()) / 1000
		fmt.Printf("%4d\tchildren created in %6f with %7d descendants and %d leaves\n", 
		N, dur, thistree.Root().CountDescendants(), len(thistree.Root().Leaves()))

		for _,ch:=range thistree.Root().Children(){
			fmt.Println(ch.Value(),len(ch.Leaves()))
		}
	}
}
