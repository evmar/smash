package smash

import "github.com/evmar/gocairo/cairo"

type Font struct {
	Name string
	Size int

	// Character width and height.
	cw, ch int
	// Adjustment from drawing baseline to bottom of character.
	descent int
}

func NewMonoFont() *Font {
	return &Font{
		Name: "monospace",
		Size: 16,
	}
}

// fakeMetrics fills in the font metrics with plausible fake values.
// Useful in tests.
func (f *Font) fakeMetrics() {
	f.cw = 10
	f.ch = 18
	f.descent = 5
}

// metricsFromCairo fills in the font metrics with the font metrics as
// measured on a cairo Context.
func (f *Font) metricsFromCairo(cr *cairo.Context) {
	ext := cairo.FontExtents{}
	cr.FontExtents(&ext)
	f.cw = int(ext.MaxXAdvance)
	f.ch = int(ext.Height)
	f.descent = int(ext.Descent)
}

func (f *Font) Use(cr *cairo.Context, bold bool) {
	weight := cairo.FontWeightNormal
	if bold {
		weight = cairo.FontWeightBold
	}
	cr.SelectFontFace(f.Name, cairo.FontSlantNormal, weight)
	cr.SetFontSize(float64(f.Size))
	if f.cw == 0 {
		f.metricsFromCairo(cr)
	}
}
