package main

import "github.com/martine/gocairo/cairo"

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

func (m *MonoFont) MetricsFromCairo(cr *cairo.Context) {
	ext := cairo.FontExtents{}
	cr.FontExtents(&ext)
	m.cw = int(ext.MaxXAdvance)
	m.ch = int(ext.Height)
	m.descent = int(ext.Descent)
}

func (m *MonoFont) Use(cr *cairo.Context) {
	cr.SelectFontFace(m.Name, cairo.FontSlantNormal, cairo.FontWeightNormal)
	cr.SetFontSize(float64(m.Size))
	if m.cw == 0 {
		m.MetricsFromCairo(cr)
	}
}
