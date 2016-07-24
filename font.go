package smash

import "github.com/evmar/gocairo/cairo"

type MonoFont struct {
	Name string
	Size int

	// Character width and height.
	cw, ch int
	// Adjustment from drawing baseline to bottom of character.
	descent int
}

func NewMonoFont() *MonoFont {
	return &MonoFont{
		Name: "monospace",
		Size: 16,
	}
}

// fakeMetrics fills in the font metrics with plausible fake values.
// Useful in tests.
func (m *MonoFont) fakeMetrics() {
	m.cw = 10
	m.ch = 18
	m.descent = 5
}

// metricsFromCairo fills in the font metrics with the font metrics as
// measured on a cairo Context.
func (m *MonoFont) metricsFromCairo(cr *cairo.Context) {
	ext := cairo.FontExtents{}
	cr.FontExtents(&ext)
	m.cw = int(ext.MaxXAdvance)
	m.ch = int(ext.Height)
	m.descent = int(ext.Descent)
}

func (m *MonoFont) Use(cr *cairo.Context, bold bool) {
	weight := cairo.FontWeightNormal
	if bold {
		weight = cairo.FontWeightBold
	}
	cr.SelectFontFace(m.Name, cairo.FontSlantNormal, weight)
	cr.SetFontSize(float64(m.Size))
	if m.cw == 0 {
		m.metricsFromCairo(cr)
	}
}
