package elevationImage

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"slices"
)

// A grayscale image where the gray level represents z(x,y) height. The image is rescaled to the 1% and 99% quantiles of the pixel values.
type ElevationImage struct {
	*image.Gray16
	format string
}

func NewElevationImage(path string) (*ElevationImage, error) {

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	i := img.(*image.Gray16)

	im := &ElevationImage{
		Gray16: i,
		format: "PNG",
	}
	im.Rescale(.99)

	return im, nil

}

func (e *ElevationImage) Rescale(perc float64) error {
	if e.Gray16 == nil {
		return errors.New("no image")
	}
	var vmin uint16 = 65535
	var vmax uint16
	for i := range e.Gray16.Bounds().Dx() {
		for j := range e.Gray16.Bounds().Dy() {
			v := e.Gray16.Gray16At(i, j).Y
			vmin = min(vmin, v)
			vmax = max(vmax, v)
		}
	}

	zoid := make([]uint16, 1000)

	for n := range 1000 {
		i := rand.Intn(e.Gray16.Bounds().Dx())
		j := rand.Intn(e.Gray16.Bounds().Dy())
		zoid[n] = e.Gray16.Gray16At(i, j).Y
	}
	slices.Sort(zoid)

	vmin = zoid[100]
	vmax = zoid[900]

	dv := float64(vmax - vmin)

	for i := range e.Gray16.Bounds().Dx() {
		for j := range e.Gray16.Bounds().Dy() {
			v := e.Gray16.Gray16At(i, j).Y
			if v < vmin {
				e.Gray16.SetGray16(i, j, color.Gray16{0})
			} else if v > vmax {
				e.Gray16.SetGray16(i, j, color.Gray16{65535})
			} else {
				v = uint16(float64(v-vmin) / dv * 65535)
				e.Gray16.SetGray16(i, j, color.Gray16{v})
			}
		}
	}

	fmt.Println(vmin, vmax)
	return nil
}
