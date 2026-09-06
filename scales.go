package main

import (
	"fmt"
	"strings"
)

// ScaleFormula represents interval steps in degree deltas and semitone deltas.
var scaleFormulas = map[string][]struct {
	degDelta  int
	semiDelta int
}{
	"major": {
		{0, 0}, {1, 2}, {2, 4}, {3, 5}, {4, 7}, {5, 9}, {6, 11},
	},
	"natural_minor": {
		{0, 0}, {1, 2}, {2, 3}, {3, 5}, {4, 7}, {5, 8}, {6, 10},
	},
	"harmonic_minor": {
		{0, 0}, {1, 2}, {2, 3}, {3, 5}, {4, 7}, {5, 8}, {6, 11},
	},
	"melodic_minor": {
		{0, 0}, {1, 2}, {2, 3}, {3, 5}, {4, 7}, {5, 9}, {6, 11},
	},
	"dorian": {
		{0, 0}, {1, 2}, {2, 3}, {3, 5}, {4, 7}, {5, 9}, {6, 10},
	},
	"mixolydian": {
		{0, 0}, {1, 2}, {2, 4}, {3, 5}, {4, 7}, {5, 9}, {6, 10},
	},
}

// ScaleGenerator produces enharmonically correct scales.
type ScaleGenerator struct{}

// NewScaleGenerator creates a new ScaleGenerator instance.
func NewScaleGenerator() *ScaleGenerator {
	return &ScaleGenerator{}
}

// Generate calculates the notes of a scale given a root note and scale type.
func (sg *ScaleGenerator) Generate(rootNote, scaleType string) ([]string, error) {
	root, err := ParsePitch(rootNote)
	if err != nil {
		return nil, fmt.Errorf("invalid root note: %s", rootNote)
	}

	formula, ok := scaleFormulas[strings.ToLower(scaleType)]
	if !ok {
		return nil, fmt.Errorf("unknown scale type: %s", scaleType)
	}

	scale := make([]string, len(formula))
	for i, step := range formula {
		p := root.AddInterval(step.degDelta, step.semiDelta)
		scale[i] = p.Name()
	}
	return scale, nil
}
