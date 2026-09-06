package main

import (
	"flag"
	"image/color"
	"log"
	"strings"

	"github.com/hippodribble/fynewidgets"
)

var (
	w           int
	h           int
	major       int
	minor       int
	hue         string
	logarithmic bool
)

func main() {

	flag.IntVar(&w, "w", 401, "width of graph paper")
	flag.IntVar(&h, "h", 601, "height of graph paper")
	flag.IntVar(&major, "major", 50, "major grid spacing")
	flag.IntVar(&minor, "minor", 5, "minor grid spacing")
	flag.StringVar(&hue, "hue", "g", "hue of the graph paper (r,g,b)")
	flag.BoolVar(&logarithmic, "log", false, "log rather than linear spacing")

	flag.Parse()

	var dark, mid, light color.Color

	switch strings.ToLower(hue) {
	case "r":
		dark = color.RGBA{255, 0, 0, 255}
		mid = color.RGBA{255, 0, 0, 128}
		light = color.RGBA{255, 0, 0, 64}
	case "g":
		dark = color.RGBA{0, 192, 0, 255}
		mid = color.RGBA{0, 192, 0, 128}
		light = color.RGBA{0, 192, 0, 64}
	case "b":
		dark = color.RGBA{0, 0, 255, 255}
		mid = color.RGBA{0, 0, 255, 128}
		light = color.RGBA{0, 0, 255, 64}
	}

	// fmt.Println(dark, mid, light)

	gp, err := fynewidgets.NewGraphPaper(major, minor, w, h, dark, light, mid, color.White, logarithmic)
	if err != nil {
		log.Fatalln(err)
	}

	gp.SaveImage()
}
