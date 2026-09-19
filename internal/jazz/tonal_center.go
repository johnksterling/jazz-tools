package jazz

import (
	"fmt"
	"strings"
)

// TonalCenterInfo holds metadata and formatting tokens for a tonal center.
type TonalCenterInfo struct {
	Name          string // e.g. "F", "Bb", "Eb", "G min"
	Fifths        int    // circle of fifths position (-1 for F, -2 for Bb, etc.)
	Mode          string // "major" or "minor"
	ColorHex      string // e.g. "#1D4ED8" for MusicXML and web
	LilyPondColor string // e.g. "(rgb-color 0.11 0.31 0.85)"
	AnsiColor     string // ANSI 24-bit terminal color escape code
}

// TonalCenterPalette maps keys to visually distinct, harmonious colors.
var TonalCenterPalette = map[string]TonalCenterInfo{
	// Flat Keys
	"F": {
		Name: "F", Fifths: -1, Mode: "major",
		ColorHex: "#1D4ED8", LilyPondColor: "(rgb-color 0.11 0.31 0.85)",
		AnsiColor: "\033[38;2;29;78;216m",
	},
	"Bb": {
		Name: "Bb", Fifths: -2, Mode: "major",
		ColorHex: "#D97706", LilyPondColor: "(rgb-color 0.85 0.47 0.02)",
		AnsiColor: "\033[38;2;217;119;6m",
	},
	"Eb": {
		Name: "Eb", Fifths: -3, Mode: "major",
		ColorHex: "#7E22CE", LilyPondColor: "(rgb-color 0.49 0.13 0.81)",
		AnsiColor: "\033[38;2;126;34;206m",
	},
	"Ab": {
		Name: "Ab", Fifths: -4, Mode: "major",
		ColorHex: "#BE185D", LilyPondColor: "(rgb-color 0.75 0.09 0.36)",
		AnsiColor: "\033[38;2;190;24;93m",
	},
	"Db": {
		Name: "Db", Fifths: -5, Mode: "major",
		ColorHex: "#EA580C", LilyPondColor: "(rgb-color 0.92 0.35 0.05)",
		AnsiColor: "\033[38;2;234;88;12m",
	},
	"Gb": {
		Name: "Gb", Fifths: -6, Mode: "major",
		ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)",
		AnsiColor: "\033[38;2;13;148;136m",
	},
	// Sharp / Natural Keys
	"C": {
		Name: "C", Fifths: 0, Mode: "major",
		ColorHex: "#2563EB", LilyPondColor: "(rgb-color 0.15 0.39 0.92)",
		AnsiColor: "\033[38;2;37;99;235m",
	},
	"G": {
		Name: "G", Fifths: 1, Mode: "major",
		ColorHex: "#16A34A", LilyPondColor: "(rgb-color 0.09 0.64 0.29)",
		AnsiColor: "\033[38;2;22;163;74m",
	},
	"D": {
		Name: "D", Fifths: 2, Mode: "major",
		ColorHex: "#4F46E5", LilyPondColor: "(rgb-color 0.31 0.27 0.90)",
		AnsiColor: "\033[38;2;79;70;229m",
	},
	"A": {
		Name: "A", Fifths: 3, Mode: "major",
		ColorHex: "#C026D3", LilyPondColor: "(rgb-color 0.75 0.15 0.83)",
		AnsiColor: "\033[38;2;192;38;211m",
	},
	"E": {
		Name: "E", Fifths: 4, Mode: "major",
		ColorHex: "#65A30D", LilyPondColor: "(rgb-color 0.40 0.64 0.05)",
		AnsiColor: "\033[38;2;101;163;13m",
	},
	"B": {
		Name: "B", Fifths: 5, Mode: "major",
		ColorHex: "#059669", LilyPondColor: "(rgb-color 0.02 0.59 0.41)",
		AnsiColor: "\033[38;2;5;150;105m",
	},
	// Common Minor Tonal Centers
	"G min": {
		Name: "G min", Fifths: -2, Mode: "minor",
		ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)",
		AnsiColor: "\033[38;2;21;128;61m",
	},
	"D min": {
		Name: "D min", Fifths: -1, Mode: "minor",
		ColorHex: "#4338CA", LilyPondColor: "(rgb-color 0.26 0.22 0.79)",
		AnsiColor: "\033[38;2;67;56;202m",
	},
	"C min": {
		Name: "C min", Fifths: -3, Mode: "minor",
		ColorHex: "#6B21A8", LilyPondColor: "(rgb-color 0.42 0.13 0.66)",
		AnsiColor: "\033[38;2;107;33;168m",
	},
	"F min": {
		Name: "F min", Fifths: -4, Mode: "minor",
		ColorHex: "#9F1239", LilyPondColor: "(rgb-color 0.62 0.07 0.22)",
		AnsiColor: "\033[38;2;159;18;57m",
	},
	"Bb min": {
		Name: "Bb min", Fifths: -5, Mode: "minor",
		ColorHex: "#9A3412", LilyPondColor: "(rgb-color 0.60 0.20 0.07)",
		AnsiColor: "\033[38;2;154;52;18m",
	},
	"A min": {
		Name: "A min", Fifths: 0, Mode: "minor",
		ColorHex: "#1E3A8A", LilyPondColor: "(rgb-color 0.12 0.23 0.54)",
		AnsiColor: "\033[38;2;30;58;138m",
	},
	"E min": {
		Name: "E min", Fifths: 1, Mode: "minor",
		ColorHex: "#047857", LilyPondColor: "(rgb-color 0.02 0.47 0.34)",
		AnsiColor: "\033[38;2;4;120;87m",
	},
	"B min": {
		Name: "B min", Fifths: 2, Mode: "minor",
		ColorHex: "#374151", LilyPondColor: "(rgb-color 0.22 0.25 0.32)",
		AnsiColor: "\033[38;2;55;65;81m",
	},
}

// GetTonalCenterInfo retrieves color and metadata for a given key name.
func GetTonalCenterInfo(name string) TonalCenterInfo {
	if info, ok := TonalCenterPalette[name]; ok {
		return info
	}
	// Enharmonic substitutions
	switch name {
	case "F#":
		return GetTonalCenterInfo("Gb")
	case "C#":
		return GetTonalCenterInfo("Db")
	case "D#":
		return GetTonalCenterInfo("Eb")
	case "G#":
		return GetTonalCenterInfo("Ab")
	case "A#":
		return GetTonalCenterInfo("Bb")
	case "F# min":
		return GetTonalCenterInfo("Gb min")
	case "C# min":
		return GetTonalCenterInfo("Db min")
	}

	// Clean any trailing "major" / "minor"
	trimmed := strings.TrimSuffix(strings.TrimSuffix(name, " major"), " minor")
	if info, ok := TonalCenterPalette[trimmed]; ok {
		return info
	}

	// Default fallback
	return TonalCenterInfo{
		Name:          name,
		Fifths:        0,
		Mode:          "major",
		ColorHex:      "#1E293B",
		LilyPondColor: "(rgb-color 0.12 0.16 0.23)",
		AnsiColor:     "\033[38;2;30;41;59m",
	}
}

// pitchNameToClass maps pitch names to standard pitch classes 0-11 for comparisons.
func pitchNameToClass(name string) int {
	switch name {
	case "C", "B#":
		return 0
	case "C#", "Db":
		return 1
	case "D":
		return 2
	case "D#", "Eb":
		return 3
	case "E", "Fb":
		return 4
	case "F", "E#":
		return 5
	case "F#", "Gb":
		return 6
	case "G":
		return 7
	case "G#", "Ab":
		return 8
	case "A":
		return 9
	case "A#", "Bb":
		return 10
	case "B", "Cb":
		return 11
	default:
		return 0
	}
}

// buildScalePitchSet returns the 7 pitch classes (0-11) of a key's scale.
func buildScalePitchSet(rootName, mode string) map[int]bool {
	rootClass := pitchNameToClass(rootName)
	steps := []int{0, 2, 4, 5, 7, 9, 11} // major
	if mode == "minor" {
		steps = []int{0, 2, 3, 5, 7, 8, 10} // natural minor
	}

	set := make(map[int]bool)
	for _, s := range steps {
		set[(rootClass+s)%12] = true
	}
	if mode == "minor" {
		// Include harmonic minor leading tone
		set[(rootClass+11)%12] = true
	}
	return set
}

// chordFitsScale checks if all chord tones belong to the given scale pitch set.
func chordFitsScale(c Chord, scaleSet map[int]bool) bool {
	tones := c.ChordTones()
	for _, p := range tones {
		pc := pitchNameToClass(p.Name())
		if !scaleSet[pc] {
			return false
		}
	}
	return true
}

// chordScaleOverlap counts how many chord tones belong to the scale.
func chordScaleOverlap(c Chord, scaleSet map[int]bool) int {
	overlap := 0
	for _, p := range c.ChordTones() {
		pc := pitchNameToClass(p.Name())
		if scaleSet[pc] {
			overlap++
		}
	}
	return overlap
}

// extractHomeKey parses the primary key name from the tune metadata.
func extractHomeKey(t *Tune) (string, string) {
	key := t.Key
	if key == "" && len(t.Measures) > 0 {
		m := t.Measures[0]
		k, err := FifthsToKey(m.KeyFifths, m.KeyMode)
		if err == nil {
			key = k
		}
	}
	if key == "" {
		return "C", "major"
	}

	parts := strings.Fields(key)
	root := parts[0]
	mode := "major"
	if len(parts) > 1 && strings.EqualFold(parts[1], "minor") {
		mode = "minor"
	}
	return root, mode
}

// AnalyzeTonalCenters determines the local diatonic/tonal center for each chord in the tune.
func AnalyzeTonalCenters(tune *Tune) {
	homeRoot, homeMode := extractHomeKey(tune)
	homeKeyName := homeRoot
	if homeMode == "minor" {
		homeKeyName = homeRoot + " min"
	}

	// Cache major scale pitch sets
	majorKeys := []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	scaleSets := make(map[string]map[int]bool)
	for _, k := range majorKeys {
		scaleSets[k] = buildScalePitchSet(k, "major")
	}

	// Cache common minor centers
	minorKeys := []string{"G", "D", "C", "F", "Bb", "A", "E", "B"}
	for _, k := range minorKeys {
		scaleSets[k+" min"] = buildScalePitchSet(k, "minor")
	}

	// Collect sequential pointers to all TimedChords
	var timedChords []*TimedChord
	for mIdx := range tune.Measures {
		for cIdx := range tune.Measures[mIdx].Chords {
			timedChords = append(timedChords, &tune.Measures[mIdx].Chords[cIdx])
		}
	}

	if len(timedChords) == 0 {
		return
	}

	activeCenter := homeKeyName

	for i := 0; i < len(timedChords); i++ {
		tc := timedChords[i]
		var nextTc *TimedChord
		if i+1 < len(timedChords) {
			nextTc = timedChords[i+1]
		}

		currScale := scaleSets[activeCenter]
		homeScale := scaleSets[homeKeyName]

		fitsActive := currScale != nil && chordFitsScale(tc.Chord, currScale)
		fitsHome := homeScale != nil && chordFitsScale(tc.Chord, homeScale)

		newCenter := activeCenter

		// 1. Cadential Lookahead: Secondary Dominants (V7 -> target)
		if tc.Chord.Quality == QualityDominant7 && nextTc != nil {
			targetPitch := tc.Chord.Root.AddInterval(3, 5) // perfect 4th up (+5 st)
			targetName := targetPitch.Name()
			if nextTc.Chord.Root.Name() == targetName {
				// V7 resolving to target!
				if nextTc.Chord.Quality == QualityMinor7 || nextTc.Chord.Quality == QualityMinorTriad {
					targetName += " min"
				}
				if !fitsActive || targetName != activeCenter {
					newCenter = targetName
				}
			}
		}

		// 2. Cadential Lookahead: ii-V pairs (e.g. Am7 -> D7 or Fm7 -> Bb7)
		if (tc.Chord.Quality == QualityMinor7 || tc.Chord.Quality == QualityHalfDiminished) && nextTc != nil && nextTc.Chord.Quality == QualityDominant7 {
			vTarget := tc.Chord.Root.AddInterval(3, 5) // +5 semitones
			if nextTc.Chord.Root.Name() == vTarget.Name() {
				// ii-V detected! Resolution target is a 4th up from V
				finalTargetPitch := nextTc.Chord.Root.AddInterval(3, 5)
				finalTargetName := finalTargetPitch.Name()
				if tc.Chord.Quality == QualityHalfDiminished {
					finalTargetName += " min"
				} else if i+2 < len(timedChords) {
					targetChord := timedChords[i+2]
					if targetChord.Chord.Root.Name() == finalTargetName &&
						(targetChord.Chord.Quality == QualityMinor7 || targetChord.Chord.Quality == QualityMinorTriad) {
						finalTargetName += " min"
					}
				}

				if !fitsActive || finalTargetName != activeCenter {
					newCenter = finalTargetName
				}
			}
		}

		// 3. Explicit Major 7th tonicization (e.g. Ebmaj7, Abmaj7, Dbmaj7)
		if tc.Chord.Quality == QualityMajor7 {
			rootName := tc.Chord.Root.Name()
			if rootName != activeCenter {
				if rootName == homeKeyName {
					newCenter = homeKeyName
				} else if activeCenter != homeKeyName || !fitsHome {
					newCenter = rootName
				}
			}
		}

		// 4. Minor 7th modal / subdominant shift (e.g. Bbm7 -> Bb min)
		if (tc.Chord.Quality == QualityMinor7 || tc.Chord.Quality == QualityMinorTriad) && !fitsActive {
			minKey := tc.Chord.Root.Name() + " min"
			if minScale := scaleSets[minKey]; minScale != nil && chordFitsScale(tc.Chord, minScale) {
				newCenter = minKey
			}
		}

		// 5. Return to Home Key or Active Key Continuity
		if newCenter == activeCenter {
			if fitsActive {
				// Stay in active center
				newCenter = activeCenter
			} else if fitsHome {
				// Chords fit home key
				newCenter = homeKeyName
			} else {
				// Non-cadential / remote modulation: find best fitting scale
				bestKey := activeCenter
				bestOverlap := -1
				for k, s := range scaleSets {
					overlap := chordScaleOverlap(tc.Chord, s)
					if overlap > bestOverlap {
						bestOverlap = overlap
						bestKey = k
					}
				}
				newCenter = bestKey
			}
		}

		// Pull back to home key on dominant resolution to home tonic
		if tc.Chord.Quality == QualityDominant7 && fitsHome {
			targetPitch := tc.Chord.Root.AddInterval(3, 5)
			if targetPitch.Name() == homeRoot {
				newCenter = homeKeyName
			}
		}

		// Assign metadata to TimedChord
		info := GetTonalCenterInfo(newCenter)
		tc.TonalCenter = info.Name
		tc.TonalCenterColor = info.LilyPondColor
		tc.TonalCenterHex = info.ColorHex

		if i == 0 || tc.TonalCenter != timedChords[i-1].TonalCenter {
			tc.IsTonalCenterChange = true
		} else {
			tc.IsTonalCenterChange = false
		}

		activeCenter = newCenter
	}
}

// HarmonicJourney formats a readable trajectory of tonal centers across the tune.
func HarmonicJourney(tune *Tune) string {
	var spans []struct {
		Center   string
		StartBar int
		EndBar   int
	}

	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			if tc.TonalCenter == "" {
				continue
			}
			if len(spans) == 0 || spans[len(spans)-1].Center != tc.TonalCenter {
				spans = append(spans, struct {
					Center   string
					StartBar int
					EndBar   int
				}{
					Center:   tc.TonalCenter,
					StartBar: tc.MeasureNumber,
					EndBar:   tc.MeasureNumber,
				})
			} else {
				spans[len(spans)-1].EndBar = tc.MeasureNumber
			}
		}
	}

	if len(spans) == 0 {
		return ""
	}

	var parts []string
	for _, sp := range spans {
		info := GetTonalCenterInfo(sp.Center)
		barLabel := fmt.Sprintf("bars %d-%d", sp.StartBar, sp.EndBar)
		if sp.StartBar == sp.EndBar {
			barLabel = fmt.Sprintf("bar %d", sp.StartBar)
		}
		coloredName := fmt.Sprintf("%s[ %s ]\033[0m (%s)", info.AnsiColor, sp.Center, barLabel)
		parts = append(parts, coloredName)
	}

	return strings.Join(parts, " ➔ ")
}
