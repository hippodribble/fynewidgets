package pz4

import (
	"errors"
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// Pyramid decomposition of an image.Image
//   - Each level is half the width and height of the previous level
//   - A projection PyramidDatum stores datum and scale for the top level of the pyramid, as well as the active level of the pyramid
type Pyramid struct {
	images []*image.NRGBA
	level  int
}

func (p Pyramid) String() string {
	s := fmt.Sprintf("\n\n------------------------\nPyramid\n%d\tLevels:\n", p.Height())
	// log.Println(s)

	for i := 0; i < p.Height(); i++ {
		if p.images[i] == nil {
			s += fmt.Sprintf("Level %2d : nil\n", i)
		} else {
			w := p.images[i].Bounds().Dx()
			h := p.images[i].Bounds().Dy()
			// log.Println(w,h)
			s += fmt.Sprintf("Level %2d : %4d x %4d", i, w, h)
			if i==p.level {
				s+=" (Active)"
			}
			s += "\n"
		}
	}

	return s

}

// creates a pyramid from an image,Image, with the smallest possible size given. Active level will be the last one (lowest resolution)
func NewPyramid(img image.Image, smallestsize image.Point) (*Pyramid, error) {
	newpyramid := Pyramid{}
	W := img.Bounds().Dx()
	H := img.Bounds().Dy()
	ww := smallestsize.X
	hh := smallestsize.Y
	newpyramid.images = make([]*image.NRGBA, 0)
	newpyramid.images = append(newpyramid.images, imaging.Clone(img))
	W, H = W/2, H/2 // next level down the pyramid

	for W > ww && H > hh { // add level and scale down by 2 in x,y, while remaining above the minimum required size
		newpyramid.images = append(newpyramid.images, imaging.Resize(newpyramid.images[len(newpyramid.images)-1], W, H, imaging.Gaussian))
		W, H = W/2, H/2
	}
	newpyramid.level = len(newpyramid.images) - 1 // for safety in case someone tries to display a massive image

	return &newpyramid, nil

}

// returns the current level of the pyramid to use
func (p *Pyramid) Level() int {

	return p.level
}

// returns the number of levels in the pyramid, whcih is at least 1 if the normal constructor was used.
func (p Pyramid) Height() int {

	return len(p.images)

}

// image at level i of the pyramid. Level 0 corresponds to the full image, with level 1 at half width and height, etc.
func (p Pyramid) CurrentImage() (*image.NRGBA, error) {

	if len(p.images) == 0 {
		return nil, errors.New("no images in pyramid")
	}
	if p.level < 0 || p.level >= len(p.images) {
		return nil, errors.New("level out of range")
	}
	return p.images[p.level], nil

}
