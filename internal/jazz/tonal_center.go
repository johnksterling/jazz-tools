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

// TonalCenterColor holds color formatting tokens for sheet music engraving and web display.
type TonalCenterColor struct {
	ColorHex      string // e.g. "#1D4ED8"
	LilyPondColor string // e.g. "(rgb-color 0.11 0.31 0.85)"
	AnsiColor     string // ANSI 24-bit terminal color escape code
}

// CuratedTonalCenterColors provides a small, high-contrast, visually distinct set of colors
// designed for clear legibility and variety on printed sheet music and vector scores.
var CuratedTonalCenterColors = []TonalCenterColor{
	{ColorHex: "#1D4ED8", LilyPondColor: "(rgb-color 0.11 0.31 0.85)", AnsiColor: "\033[38;2;29;78;216m"},  // Royal Blue (primary)
	{ColorHex: "#D97706", LilyPondColor: "(rgb-color 0.85 0.47 0.02)", AnsiColor: "\033[38;2;217;119;6m"},  // Warm Amber / Gold
	{ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)", AnsiColor: "\033[38;2;21;128;61m"},  // Forest Green
	{ColorHex: "#7E22CE", LilyPondColor: "(rgb-color 0.49 0.13 0.81)", AnsiColor: "\033[38;2;126;34;206m"}, // Rich Purple
	{ColorHex: "#DC2626", LilyPondColor: "(rgb-color 0.86 0.15 0.15)", AnsiColor: "\033[38;2;220;38;38m"},  // Crimson Red
	{ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)", AnsiColor: "\033[38;2;13;148;136m"}, // Deep Teal
}

// TonalCenterPalette maps keys to visually distinct, high-contrast colors from CuratedTonalCenterColors.
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
		ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)",
		AnsiColor: "\033[38;2;21;128;61m",
	},
	"Db": {
		Name: "Db", Fifths: -5, Mode: "major",
		ColorHex: "#DC2626", LilyPondColor: "(rgb-color 0.86 0.15 0.15)",
		AnsiColor: "\033[38;2;220;38;38m",
	},
	"Gb": {
		Name: "Gb", Fifths: -6, Mode: "major",
		ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)",
		AnsiColor: "\033[38;2;13;148;136m",
	},
	// Sharp / Natural Keys
	"C": {
		Name: "C", Fifths: 0, Mode: "major",
		ColorHex: "#1D4ED8", LilyPondColor: "(rgb-color 0.11 0.31 0.85)",
		AnsiColor: "\033[38;2;29;78;216m",
	},
	"G": {
		Name: "G", Fifths: 1, Mode: "major",
		ColorHex: "#D97706", LilyPondColor: "(rgb-color 0.85 0.47 0.02)",
		AnsiColor: "\033[38;2;217;119;6m",
	},
	"D": {
		Name: "D", Fifths: 2, Mode: "major",
		ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)",
		AnsiColor: "\033[38;2;21;128;61m",
	},
	"A": {
		Name: "A", Fifths: 3, Mode: "major",
		ColorHex: "#7E22CE", LilyPondColor: "(rgb-color 0.49 0.13 0.81)",
		AnsiColor: "\033[38;2;126;34;206m",
	},
	"E": {
		Name: "E", Fifths: 4, Mode: "major",
		ColorHex: "#DC2626", LilyPondColor: "(rgb-color 0.86 0.15 0.15)",
		AnsiColor: "\033[38;2;220;38;38m",
	},
	"B": {
		Name: "B", Fifths: 5, Mode: "major",
		ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)",
		AnsiColor: "\033[38;2;13;148;136m",
	},
	// Common Minor Tonal Centers (paired to contrast with their relative major)
	"D min": {
		Name: "D min", Fifths: -1, Mode: "minor",
		ColorHex: "#D97706", LilyPondColor: "(rgb-color 0.85 0.47 0.02)",
		AnsiColor: "\033[38;2;217;119;6m",
	},
	"G min": {
		Name: "G min", Fifths: -2, Mode: "minor",
		ColorHex: "#7E22CE", LilyPondColor: "(rgb-color 0.49 0.13 0.81)",
		AnsiColor: "\033[38;2;126;34;206m",
	},
	"C min": {
		Name: "C min", Fifths: -3, Mode: "minor",
		ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)",
		AnsiColor: "\033[38;2;21;128;61m",
	},
	"F min": {
		Name: "F min", Fifths: -4, Mode: "minor",
		ColorHex: "#DC2626", LilyPondColor: "(rgb-color 0.86 0.15 0.15)",
		AnsiColor: "\033[38;2;220;38;38m",
	},
	"Bb min": {
		Name: "Bb min", Fifths: -5, Mode: "minor",
		ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)",
		AnsiColor: "\033[38;2;13;148;136m",
	},
	"A min": {
		Name: "A min", Fifths: 0, Mode: "minor",
		ColorHex: "#15803D", LilyPondColor: "(rgb-color 0.08 0.50 0.24)",
		AnsiColor: "\033[38;2;21;128;61m",
	},
	"E min": {
		Name: "E min", Fifths: 1, Mode: "minor",
		ColorHex: "#DC2626", LilyPondColor: "(rgb-color 0.86 0.15 0.15)",
		AnsiColor: "\033[38;2;220;38;38m",
	},
	"B min": {
		Name: "B min", Fifths: 2, Mode: "minor",
		ColorHex: "#0D9488", LilyPondColor: "(rgb-color 0.05 0.58 0.53)",
		AnsiColor: "\033[38;2;13;148;136m",
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

// GetTuneTonalCenterInfo retrieves color and metadata for a given key name,
// preferring the tune's dynamic high-variety palette if available.
func GetTuneTonalCenterInfo(tune *Tune, name string) TonalCenterInfo {
	if tune != nil && tune.TonalCenterColors != nil {
		if info, ok := tune.TonalCenterColors[name]; ok {
			return info
		}
		trimmed := strings.TrimSuffix(strings.TrimSuffix(name, " major"), " minor")
		if info, ok := tune.TonalCenterColors[trimmed]; ok {
			return info
		}
	}
	return GetTonalCenterInfo(name)
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

	key = strings.TrimSpace(key)
	root := key
	mode := "major"

	if strings.HasSuffix(key, "-") {
		root = strings.TrimSuffix(key, "-")
		mode = "minor"
	} else if strings.HasSuffix(key, "m") && len(key) > 1 && key[len(key)-2] != 'b' && key[len(key)-2] != '#' {
		root = strings.TrimSuffix(key, "m")
		mode = "minor"
	} else {
		parts := strings.Fields(key)
		if len(parts) > 0 {
			root = parts[0]
		}
		if len(parts) > 1 && (strings.EqualFold(parts[1], "minor") || strings.EqualFold(parts[1], "min") || strings.EqualFold(parts[1], "m")) {
			mode = "minor"
		}
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

		// 3. Phrygian Chromatic Neighbor & Modal Interchange Checks
		// A chord should NOT trigger a new key center if it is an embellishing chromatic neighbor
		// or borrowed modal interchange chord within the active/home key framework.
		isNeighborOrModalInterchange := false

		// Check A: Phrygian Chromatic Neighbor (Major 7th chord 1 semitone above previous chord or resolving to tonic)
		// e.g. F^7 -> Gb^7 -> F^7 (Major), D-7 -> Eb^7 -> D-7 (Minor), or F^7 -> E-7 (resolving to tonic)
		if tc.Chord.Quality == QualityMajor7 {
			if i > 0 {
				prevTc := timedChords[i-1]
				if semitoneDistance(prevTc.Chord.Root.Name(), tc.Chord.Root.Name()) == 1 {
					// Returns to previous chord or resolves down a half-step
					if nextTc != nil && nextTc.Chord.Root.Name() == prevTc.Chord.Root.Name() {
						isNeighborOrModalInterchange = true
						newCenter = activeCenter
					} else if nextTc != nil && semitoneDistance(nextTc.Chord.Root.Name(), tc.Chord.Root.Name()) == 1 &&
						(nextTc.Chord.Root.Name() == homeRoot || nextTc.Chord.Root.Name() == prevTc.Chord.Root.Name()) {
						isNeighborOrModalInterchange = true
						if nextTc.Chord.Root.Name() == homeRoot {
							newCenter = homeKeyName
						} else {
							newCenter = activeCenter
						}
					}
				}
			}
			// Resolving down a half-step directly to home tonic (e.g. F^7 -> E-7 in Nardis, or Gb^7 -> F^7)
			if !isNeighborOrModalInterchange && nextTc != nil &&
				semitoneDistance(nextTc.Chord.Root.Name(), tc.Chord.Root.Name()) == 1 &&
				nextTc.Chord.Root.Name() == homeRoot {
				isNeighborOrModalInterchange = true
				newCenter = homeKeyName
			}
		}

		// Check B: Modal Interchange in Home Major Key (e.g. bVIIΔ, bVIΔ, bIIIΔ, iv-7, v-7, i-7)
		// Applies when currently in the home key framework
		if !isNeighborOrModalInterchange && activeCenter == homeKeyName && homeMode == "major" {
			dHome := semitoneDistance(homeRoot, tc.Chord.Root.Name())
			if tc.Chord.Quality == QualityMajor7 {
				// bVIIΔ (subtonic, d=10), bVIΔ (submediant, d=8), bIIIΔ (mediant, d=3), bIIΔ (Phrygian, d=1)
				if dHome == 10 || dHome == 8 || dHome == 3 || dHome == 1 {
					// Check if this is part of a sequential cycle of 4ths modulation (e.g. Eb^7 -> Ab^7 -> Db^7)
					isCycleOfFourths := nextTc != nil && nextTc.Chord.Quality == QualityMajor7 &&
						semitoneDistance(tc.Chord.Root.Name(), nextTc.Chord.Root.Name()) == 5
					if !isCycleOfFourths {
						isNeighborOrModalInterchange = true
						newCenter = homeKeyName
					}
				}
			} else if tc.Chord.Quality == QualityMinor7 || tc.Chord.Quality == QualityMinorTriad {
				// iv-7 (minor subdominant, d=5), v-7 (minor v, d=7), i-7 (tonic minor, d=0)
				if dHome == 5 || dHome == 7 || dHome == 0 {
					isNeighborOrModalInterchange = true
					newCenter = homeKeyName
				}
			}
		}

		// Check C: Modal Interchange / Diatonic Chords in Home Minor Key (e.g. bIIΔ Phrygian, bVIΔ, iv-7, IΔ)
		// Applies when home key is minor (e.g. Nardis in E minor)
		if !isNeighborOrModalInterchange && activeCenter == homeKeyName && homeMode == "minor" {
			dHome := semitoneDistance(homeRoot, tc.Chord.Root.Name())
			if tc.Chord.Quality == QualityMajor7 {
				// bIIΔ (Phrygian, d=1), bVIΔ (Aeolian submediant, d=8), IΔ (Picardy major, d=0)
				if dHome == 1 || dHome == 8 || dHome == 0 {
					isNeighborOrModalInterchange = true
					newCenter = homeKeyName
				}
			} else if tc.Chord.Quality == QualityMinor7 || tc.Chord.Quality == QualityMinorTriad {
				// iv-7 (minor subdominant, d=5)
				if dHome == 5 {
					isNeighborOrModalInterchange = true
					newCenter = homeKeyName
				}
			}
		}

		// Check D: Dominant Resolutions without Modulation (Backdoor bVII7 -> I, and Tritone Sub bII7 -> i / I)
		// e.g. Lady Bird bars 3-5: Fm7 Bb7 -> Cmaj7 (bVII7 resolving to I)
		// e.g. A Night in Tunisia: Eb7 -> Dm or F7 -> Em (bII7 resolving to i)
		// e.g. The Night Has a Thousand Eyes bar 12: F7 -> Gmaj7 (bVII7 resolving to I)
		if !isNeighborOrModalInterchange && tc.Chord.Quality == QualityDominant7 && nextTc != nil {
			nextRoot := nextTc.Chord.Root.Name()
			dNext := semitoneDistance(nextRoot, tc.Chord.Root.Name())
			// dNext == 10 is bVII7 resolving up whole step to nextRoot (e.g. Bb7 -> C, F7 -> G)
			// dNext == 1 is bII7 resolving down half step to nextRoot (e.g. Eb7 -> Dm, F7 -> Em)
			if (dNext == 10 || dNext == 1) && (nextRoot == homeRoot || nextRoot == strings.Fields(activeCenter)[0]) {
				isNeighborOrModalInterchange = true
				if nextRoot == homeRoot {
					newCenter = homeKeyName
				} else {
					newCenter = activeCenter
				}
			}
		}

		// Check E: Minor Tonic chords with major 6th / major 7th (e.g. Dm6 in D minor, Em6 in E minor)
		if !isNeighborOrModalInterchange && (tc.Chord.Quality == QualityMinor6 || tc.Chord.Quality == QualityMinorMajor7) {
			if tc.Chord.Root.Name() == homeRoot && homeMode == "minor" {
				isNeighborOrModalInterchange = true
				newCenter = homeKeyName
			} else if tc.Chord.Root.Name() == strings.Fields(activeCenter)[0] && strings.HasSuffix(activeCenter, "min") {
				isNeighborOrModalInterchange = true
				newCenter = activeCenter
			}
		}

		// 4. Explicit Major 7th tonicization (if not neighbor or modal interchange)
		if newCenter == activeCenter && !isNeighborOrModalInterchange && tc.Chord.Quality == QualityMajor7 {
			rootName := tc.Chord.Root.Name()
			if rootName != activeCenter {
				if rootName == homeKeyName {
					newCenter = homeKeyName
				} else if activeCenter != homeKeyName || !fitsHome {
					newCenter = rootName
				}
			}
		}

		// 5. Explicit Minor 7th modulation (only if not neighbor or modal interchange)
		if newCenter == activeCenter && !isNeighborOrModalInterchange && (tc.Chord.Quality == QualityMinor7 || tc.Chord.Quality == QualityMinorTriad) && !fitsActive && !fitsHome {
			minKey := tc.Chord.Root.Name() + " min"
			if minScale := scaleSets[minKey]; minScale != nil && chordFitsScale(tc.Chord, minScale) {
				newCenter = minKey
			}
		}

		// 6. Return to Home Key or Active Key Continuity
		if newCenter == activeCenter && !isNeighborOrModalInterchange {
			prevWasTarget := i > 0 && timedChords[i-1].Chord.Root.Name() == strings.Fields(activeCenter)[0]
			if activeCenter != homeKeyName && fitsHome && prevWasTarget {
				newCenter = homeKeyName
			} else if fitsActive {
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

		// Assign key center name
		tc.TonalCenter = newCenter
		activeCenter = newCenter
	}

	// Dynamic high-variety color assignment for this tune:
	// Select distinct, widely-spaced colors from CuratedTonalCenterColors for each unique
	// tonal center present in the tune (home key first, then modulations in order of appearance).
	var uniqueCenters []string
	seenCenters := make(map[string]bool)

	if homeKeyName != "" {
		uniqueCenters = append(uniqueCenters, homeKeyName)
		seenCenters[homeKeyName] = true
	}

	for _, tc := range timedChords {
		if tc.TonalCenter != "" && !seenCenters[tc.TonalCenter] {
			uniqueCenters = append(uniqueCenters, tc.TonalCenter)
			seenCenters[tc.TonalCenter] = true
		}
	}

	tune.TonalCenterColors = make(map[string]TonalCenterInfo)
	for idx, center := range uniqueCenters {
		baseInfo := GetTonalCenterInfo(center)
		paletteColor := CuratedTonalCenterColors[idx%len(CuratedTonalCenterColors)]
		baseInfo.ColorHex = paletteColor.ColorHex
		baseInfo.LilyPondColor = paletteColor.LilyPondColor
		baseInfo.AnsiColor = paletteColor.AnsiColor
		tune.TonalCenterColors[center] = baseInfo
	}

	for i, tc := range timedChords {
		info, ok := tune.TonalCenterColors[tc.TonalCenter]
		if !ok {
			info = GetTonalCenterInfo(tc.TonalCenter)
		}
		tc.TonalCenter = info.Name
		tc.TonalCenterColor = info.LilyPondColor
		tc.TonalCenterHex = info.ColorHex

		if i == 0 || tc.TonalCenter != timedChords[i-1].TonalCenter {
			tc.IsTonalCenterChange = true
		} else {
			tc.IsTonalCenterChange = false
		}
	}
}

// JourneySpan describes a continuous tonal center span across measures.
type JourneySpan struct {
	Center   string `json:"center"`
	StartBar int    `json:"startBar"`
	EndBar   int    `json:"endBar"`
	ColorHex string `json:"colorHex"`
}

// GetHarmonicJourneySpans extracts all tonal center spans across measures.
func GetHarmonicJourneySpans(tune *Tune) []JourneySpan {
	var spans []JourneySpan

	for _, m := range tune.Measures {
		for _, tc := range m.Chords {
			if tc.TonalCenter == "" {
				continue
			}
			if len(spans) == 0 || spans[len(spans)-1].Center != tc.TonalCenter {
				colorHex := tc.TonalCenterHex
				if colorHex == "" {
					info := GetTuneTonalCenterInfo(tune, tc.TonalCenter)
					colorHex = info.ColorHex
				}
				spans = append(spans, JourneySpan{
					Center:   tc.TonalCenter,
					StartBar: tc.MeasureNumber,
					EndBar:   tc.MeasureNumber,
					ColorHex: colorHex,
				})
			} else {
				spans[len(spans)-1].EndBar = tc.MeasureNumber
			}
		}
	}
	return spans
}

// HarmonicJourney creates a human-readable summary of tonal centers.
func HarmonicJourney(tune *Tune) string {
	spans := GetHarmonicJourneySpans(tune)
	if len(spans) == 0 {
		return ""
	}

	var parts []string
	for _, sp := range spans {
		info := GetTuneTonalCenterInfo(tune, sp.Center)
		barLabel := fmt.Sprintf("bars %d-%d", sp.StartBar, sp.EndBar)
		if sp.StartBar == sp.EndBar {
			barLabel = fmt.Sprintf("bar %d", sp.StartBar)
		}
		coloredName := fmt.Sprintf("%s[ %s ]\033[0m (%s)", info.AnsiColor, sp.Center, barLabel)
		parts = append(parts, coloredName)
	}

	return strings.Join(parts, " ➔ ")
}
