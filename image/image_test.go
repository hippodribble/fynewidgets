package elevationImage

import (
	"fmt"
	"testing"
)

func TestImage(t *testing.T) {

	path := "/Users/glenn/Documents/Data/Hidaka/GIS/GEBCO_14_Apr_2026_df8280d493df/geb16.png"
	im, err := NewElevationImage(path)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("FORMAT", im.format)
	fmt.Println(im.Bounds())

}
