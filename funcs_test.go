package fynewidgets

import (
	"testing"

)

func TestBestRowsColumns(t *testing.T) {
	t.Log("ghgh")
	for N := 1; N <= 29; N++ {
		a,b:=BestRowsColumns(N)
		if a>100*b{t.Fail()}
		// BestRowsColumns(N, 0.75)
	}
}

func TestTicksScale(t *testing.T) {

	t.Log("Test Ticks to Scale")
	for i:=-10;i<11;i++{
		s:=TickScaleToFloatScale(i, 5)
		ticks:=FloatScaleToTicks(s, 5)
		t.Logf("tick: %d, scale: %5.2f retick: %d @ 5 ticks per octave", i, s,ticks)
	}
}