package jazz

import (
	"fmt"
	"strings"
)

// DeviceType classifies the jazz harmonic device.
type DeviceType string

const (
	DeviceMajorTwoFive      DeviceType = "ii-V"
	DeviceMinorTwoFive      DeviceType = "iiø-V"
	DeviceTritoneSub        DeviceType = "subV"
	DeviceSecondaryDominant DeviceType = "V/x"
	DeviceTurnaround        DeviceType = "Turnaround"
	DeviceChromaticNeighbor DeviceType = "Chromatic Neighbor"
	DeviceModalInterchange  DeviceType = "Modal Interchange"
	DeviceBackdoorDominant  DeviceType = "Backdoor Resolution"
)

// HarmonicDevice describes an identified harmonic device spanning one or more chords.
type HarmonicDevice struct {
	Type          DeviceType
	Label         string  // Display text, e.g. "ii - V", "iiø - V", "subV", "V / vi"
	StartMeasure  int     // 0-indexed measure index where the device begins
	StartBeat     float64 // Beat offset within start measure
	EndMeasure    int     // 0-indexed measure index where device ends / resolves
	EndBeat       float64 // Beat offset within end measure
	Resolves      bool    // True if resolving to target chord (arrow = true)
	TargetKey     string  // Key name for color coding (e.g. "Bb", "G min")
	LilyPondColor string  // e.g. "(rgb-color 0.85 0.47 0.02)"
	IsDashed      bool    // True for tritone substitutions
}

// FlatChord is an indexed chord reference with absolute position.
type FlatChord struct {
	MeasureIndex     int
	ChordIndex       int
	Chord            Chord
	BeatOffset       float64
	DurationBeats    float64
	TonalCenter      string
	TonalCenterColor string
}

// semitoneDistance calculates pitch class difference (target - source + 12) % 12.
func semitoneDistance(fromPitch, toPitch string) int {
	fromPC := pitchNameToClass(fromPitch)
	toPC := pitchNameToClass(toPitch)
	return (toPC - fromPC + 12) % 12
}

// pitchClassToName returns standard flat-oriented pitch names for jazz keys.
func pitchClassToName(pc int) string {
	names := []string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}
	return names[(pc%12+12)%12]
}

// DetectHarmonicDevices scans a Tune and extracts common jazz harmonic devices.
func DetectHarmonicDevices(tune *Tune) []HarmonicDevice {
	// Ensure tonal center analysis is run
	AnalyzeTonalCenters(tune)

	// Flatten all timed chords
	var chords []FlatChord
	for mIdx, m := range tune.Measures {
		for cIdx, tc := range m.Chords {
			chords = append(chords, FlatChord{
				MeasureIndex:     mIdx,
				ChordIndex:       cIdx,
				Chord:            tc.Chord,
				BeatOffset:       tc.BeatOffset,
				DurationBeats:    tc.DurationBeats,
				TonalCenter:      tc.TonalCenter,
				TonalCenterColor: tc.TonalCenterColor,
			})
		}
	}

	if len(chords) < 2 {
		return nil
	}

	homeRoot, homeMode := extractHomeKey(tune)
	homeKeyName := homeRoot
	if homeMode == "minor" {
		homeKeyName += " min"
	}

	var devices []HarmonicDevice

	i := 0
	for i < len(chords) {
		c1 := chords[i]
		var c2, c3 *FlatChord
		if i+1 < len(chords) {
			c2 = &chords[i+1]
		}
		if i+2 < len(chords) {
			c3 = &chords[i+2]
		}

		// ---------------------------------------------------------
		// 1. Check ii - V pairs (Major or Minor)
		// ---------------------------------------------------------
		if c2 != nil && (c1.Chord.Quality == QualityMinor7 ||
			c1.Chord.Quality == QualityMinorTriad ||
			c1.Chord.Quality == QualityMinor6 ||
			c1.Chord.Quality == QualityHalfDiminished) &&
			c2.Chord.Quality == QualityDominant7 {

			dist1to2 := semitoneDistance(c1.Chord.Root.Name(), c2.Chord.Root.Name())
			if dist1to2 == 5 { // Perfect 4th up (+5 semitones)
				isMinor := c1.Chord.Quality == QualityHalfDiminished
				targetPC := (pitchNameToClass(c2.Chord.Root.Name()) + 5) % 12
				targetName := pitchClassToName(targetPC)

				// Expand backward for consecutive identical ii chords
				startChord := c1
				for prev := i - 1; prev >= 0; prev-- {
					if chords[prev].Chord.Root.Name() == c1.Chord.Root.Name() && chords[prev].Chord.Quality == c1.Chord.Quality {
						startChord = chords[prev]
					} else {
						break
					}
				}

				// Expand forward for consecutive identical V chords
				lastV := c2
				nextIdx := i + 1
				for nextIdx+1 < len(chords) &&
					chords[nextIdx+1].Chord.Root.Name() == c2.Chord.Root.Name() &&
					chords[nextIdx+1].Chord.Quality == c2.Chord.Quality {
					nextIdx++
					lastV = &chords[nextIdx]
				}

				var actualC3 *FlatChord
				if nextIdx+1 < len(chords) {
					actualC3 = &chords[nextIdx+1]
				}

				resolves := false
				endMeasure := lastV.MeasureIndex
				endBeat := lastV.BeatOffset + lastV.DurationBeats

				if actualC3 != nil && pitchNameToClass(actualC3.Chord.Root.Name()) == targetPC {
					resolves = true
					endMeasure = actualC3.MeasureIndex
					endBeat = actualC3.BeatOffset
					if actualC3.Chord.Quality == QualityMinor7 ||
						actualC3.Chord.Quality == QualityMinorTriad ||
						actualC3.Chord.Quality == QualityMinor6 {
						isMinor = true
					}
				} else if actualC3 != nil && (actualC3.Chord.Root.Name() == homeRoot || (actualC3.TonalCenter != "" && actualC3.Chord.Root.Name() == strings.Fields(actualC3.TonalCenter)[0])) &&
					semitoneDistance(actualC3.Chord.Root.Name(), lastV.Chord.Root.Name()) == 10 {
					// Backdoor resolution! (e.g. Fm7 Bb7 -> Cmaj7 in Lady Bird)
					resolves = true
					endMeasure = actualC3.MeasureIndex
					endBeat = actualC3.BeatOffset
					targetName = actualC3.TonalCenter
					if targetName == "" {
						targetName = homeKeyName
					}
				}

				label := "ii - V"
				devType := DeviceMajorTwoFive
				if isMinor {
					label = "iiø - V"
					devType = DeviceMinorTwoFive
					targetName += " min"
				} else if resolves && actualC3 != nil && semitoneDistance(actualC3.Chord.Root.Name(), lastV.Chord.Root.Name()) == 10 {
					label = "iv - bVII7 (backdoor)"
					devType = DeviceBackdoorDominant
				}

				tcInfo := GetTonalCenterInfo(targetName)

				devices = append(devices, HarmonicDevice{
					Type:          devType,
					Label:         label,
					StartMeasure:  startChord.MeasureIndex,
					StartBeat:     startChord.BeatOffset,
					EndMeasure:    endMeasure,
					EndBeat:       endBeat,
					Resolves:      resolves,
					TargetKey:     targetName,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      false,
				})

				i = nextIdx + 1
				continue
			}
		}

		// ---------------------------------------------------------
		// 2. Check Tritone Substitution: ii - subV (e.g. Dm7 -> Db7 -> C)
		// ---------------------------------------------------------
		if c2 != nil && c3 != nil &&
			(c1.Chord.Quality == QualityMinor7 || c1.Chord.Quality == QualityHalfDiminished) &&
			c2.Chord.Quality == QualityDominant7 {

			dist1to2 := semitoneDistance(c1.Chord.Root.Name(), c2.Chord.Root.Name())
			dist2to3 := semitoneDistance(c3.Chord.Root.Name(), c2.Chord.Root.Name())
			if dist1to2 == 11 && dist2to3 == 1 { // Dm7 -> Db7 -> C
				targetName := pitchClassToName(pitchNameToClass(c3.Chord.Root.Name()))
				if c3.Chord.Quality == QualityMinor7 || c3.Chord.Quality == QualityMinorTriad {
					targetName += " min"
				}
				tcInfo := GetTonalCenterInfo(targetName)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceTritoneSub,
					Label:         "ii - subV",
					StartMeasure:  c1.MeasureIndex,
					StartBeat:     c1.BeatOffset,
					EndMeasure:    c3.MeasureIndex,
					EndBeat:       c3.BeatOffset,
					Resolves:      true,
					TargetKey:     targetName,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      true,
				})

				i += 2
				continue
			}
		}

		// ---------------------------------------------------------
		// 3. Check Isolated subV -> I (e.g. Db7 -> C)
		// ---------------------------------------------------------
		if c2 != nil && c1.Chord.Quality == QualityDominant7 {
			dist1to2 := semitoneDistance(c2.Chord.Root.Name(), c1.Chord.Root.Name())
			if dist1to2 == 1 { // Root moves down 1 semitone (e.g. Db7 -> C)
				targetName := pitchClassToName(pitchNameToClass(c2.Chord.Root.Name()))
				if c2.Chord.Quality == QualityMinor7 || c2.Chord.Quality == QualityMinorTriad {
					targetName += " min"
				}
				tcInfo := GetTonalCenterInfo(targetName)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceTritoneSub,
					Label:         "subV",
					StartMeasure:  c1.MeasureIndex,
					StartBeat:     c1.BeatOffset,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetName,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      true,
				})

				i++
				continue
			}
		}

		// ---------------------------------------------------------
		// 4. Check Secondary Dominants / Dominant Resolving down 5th
		// ---------------------------------------------------------
		if c2 != nil && c1.Chord.Quality == QualityDominant7 {
			dist1to2 := semitoneDistance(c1.Chord.Root.Name(), c2.Chord.Root.Name())
			if dist1to2 == 5 { // Resolves down a 5th (+5 semitones)
				targetName := pitchClassToName(pitchNameToClass(c2.Chord.Root.Name()))
				if c2.Chord.Quality == QualityMinor7 || c2.Chord.Quality == QualityMinorTriad || c2.Chord.Quality == QualityMinor6 {
					targetName += " min"
				}

				// Expand backward if previous chords are identical dominant chords
				startChord := c1
				for prev := i - 1; prev >= 0; prev-- {
					if chords[prev].Chord.Root.Name() == c1.Chord.Root.Name() && chords[prev].Chord.Quality == c1.Chord.Quality {
						startChord = chords[prev]
					} else {
						break
					}
				}

				label := romanNumeralSecondaryDominant(c2.Chord.Root.Name(), c2.Chord.Quality, homeKeyName)

				tcInfo := GetTonalCenterInfo(targetName)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceSecondaryDominant,
					Label:         label,
					StartMeasure:  startChord.MeasureIndex,
					StartBeat:     startChord.BeatOffset,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetName,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      false,
				})

				i++
				continue
			}
		}

		// ---------------------------------------------------------
		// 5. Check Phrygian Chromatic Neighbor (Major 7th chord 1 semitone above prev or next)
		// ---------------------------------------------------------
		if c1.Chord.Quality == QualityMajor7 {
			var prev *FlatChord
			if i > 0 {
				prev = &chords[i-1]
			}

			// Pattern A: Neighbor embellishment returning to prev chord (prev -> c1 -> prev)
			// e.g. F^7 -> Gb^7 -> F^7 (major) or D-7 -> Eb^7 -> D-7 (minor)
			if prev != nil && semitoneDistance(prev.Chord.Root.Name(), c1.Chord.Root.Name()) == 1 &&
				c2 != nil && c2.Chord.Root.Name() == prev.Chord.Root.Name() {
				targetKey := c2.TonalCenter
				if targetKey == "" {
					targetKey = homeKeyName
				}
				tcInfo := GetTonalCenterInfo(targetKey)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceChromaticNeighbor,
					Label:         "bIIΔ (Phrygian neighbor)",
					StartMeasure:  c1.MeasureIndex,
					StartBeat:     c1.BeatOffset,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetKey,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      false,
				})

				i++
				continue
			}

			// Pattern B: Chromatic neighbor resolving down a half-step to tonic
			// e.g. Gb^7 -> F^7 or Gb^7 -> Fm7
			resolvesToTonic := c2 != nil && (c2.Chord.Root.Name() == homeRoot || (c2.TonalCenter != "" && c2.Chord.Root.Name() == strings.Fields(c2.TonalCenter)[0]))
			if c2 != nil && semitoneDistance(c2.Chord.Root.Name(), c1.Chord.Root.Name()) == 1 && resolvesToTonic {
				targetKey := c2.TonalCenter
				if targetKey == "" {
					targetKey = homeKeyName
				}
				tcInfo := GetTonalCenterInfo(targetKey)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceChromaticNeighbor,
					Label:         "bIIΔ (Phrygian neighbor)",
					StartMeasure:  c1.MeasureIndex,
					StartBeat:     c1.BeatOffset,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetKey,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      false,
				})

				i++
				continue
			}
		}

		// ---------------------------------------------------------
		// 6. Check Modal Interchange Chords (bVIIΔ, bVIΔ, bIIIΔ, iv-7, v-7, i-7)
		// ---------------------------------------------------------
		if homeMode == "major" {
			dHome := semitoneDistance(homeRoot, c1.Chord.Root.Name())

			if c1.Chord.Quality == QualityMajor7 {
				label := ""
				switch dHome {
				case 10:
					label = "bVIIΔ (modal interchange)"
				case 8:
					label = "bVIΔ (modal interchange)"
				case 3:
					label = "bIIIΔ (modal interchange)"
				}

				if label != "" {
					tcInfo := GetTonalCenterInfo(homeKeyName)
					devices = append(devices, HarmonicDevice{
						Type:          DeviceModalInterchange,
						Label:         label,
						StartMeasure:  c1.MeasureIndex,
						StartBeat:     c1.BeatOffset,
						EndMeasure:    c1.MeasureIndex,
						EndBeat:       c1.BeatOffset + c1.DurationBeats,
						Resolves:      false,
						TargetKey:     homeKeyName,
						LilyPondColor: tcInfo.LilyPondColor,
						IsDashed:      false,
					})

					i++
					continue
				}
			} else if c1.Chord.Quality == QualityMinor7 || c1.Chord.Quality == QualityMinorTriad {
				label := ""
				switch dHome {
				case 5:
					label = "iv-7 (minor subdominant)"
				case 7:
					label = "v-7 (minor v)"
				case 0:
					label = "i-7 (tonic minor)"
				}

				if label != "" {
					tcInfo := GetTonalCenterInfo(homeKeyName)
					devices = append(devices, HarmonicDevice{
						Type:          DeviceModalInterchange,
						Label:         label,
						StartMeasure:  c1.MeasureIndex,
						StartBeat:     c1.BeatOffset,
						EndMeasure:    c1.MeasureIndex,
						EndBeat:       c1.BeatOffset + c1.DurationBeats,
						Resolves:      false,
						TargetKey:     homeKeyName,
						LilyPondColor: tcInfo.LilyPondColor,
						IsDashed:      false,
					})

					i++
					continue
				}
			}
		}

		// ---------------------------------------------------------
		// 7. Check Minor Key Modal Interchange / Diatonic Devices (bIIΔ Phrygian, bVIΔ, iv-7, IΔ)
		// ---------------------------------------------------------
		if homeMode == "minor" {
			dHome := semitoneDistance(homeRoot, c1.Chord.Root.Name())
			if c1.Chord.Quality == QualityMajor7 {
				label := ""
				switch dHome {
				case 1:
					label = "bIIΔ (Phrygian chord)"
				case 8:
					label = "bVIΔ (submediant)"
				case 0:
					label = "IΔ (Picardy third)"
				}
				if label != "" {
					tcInfo := GetTonalCenterInfo(homeKeyName)
					devices = append(devices, HarmonicDevice{
						Type:          DeviceModalInterchange,
						Label:         label,
						StartMeasure:  c1.MeasureIndex,
						StartBeat:     c1.BeatOffset,
						EndMeasure:    c1.MeasureIndex,
						EndBeat:       c1.BeatOffset + c1.DurationBeats,
						Resolves:      false,
						TargetKey:     homeKeyName,
						LilyPondColor: tcInfo.LilyPondColor,
						IsDashed:      false,
					})
					i++
					continue
				}
			} else if c1.Chord.Quality == QualityMinor7 || c1.Chord.Quality == QualityMinorTriad {
				if dHome == 5 {
					tcInfo := GetTonalCenterInfo(homeKeyName)
					devices = append(devices, HarmonicDevice{
						Type:          DeviceModalInterchange,
						Label:         "iv-7 (subdominant minor)",
						StartMeasure:  c1.MeasureIndex,
						StartBeat:     c1.BeatOffset,
						EndMeasure:    c1.MeasureIndex,
						EndBeat:       c1.BeatOffset + c1.DurationBeats,
						Resolves:      false,
						TargetKey:     homeKeyName,
						LilyPondColor: tcInfo.LilyPondColor,
						IsDashed:      false,
					})
					i++
					continue
				}
			}
		}

		// ---------------------------------------------------------
		// 8. Check Backdoor Dominant Resolution (bVII7 -> I) & Tritone Sub to Minor (bII7 -> i)
		// ---------------------------------------------------------
		if c1.Chord.Quality == QualityDominant7 && c2 != nil {
			targetRoot := c2.Chord.Root.Name()
			dTarget := semitoneDistance(targetRoot, c1.Chord.Root.Name())

			// Backdoor dominant resolution (dTarget == 10, i.e. bVII7 resolving up whole step to I)
			if dTarget == 10 && (targetRoot == homeRoot || (c2.TonalCenter != "" && targetRoot == strings.Fields(c2.TonalCenter)[0])) {
				targetKey := c2.TonalCenter
				if targetKey == "" {
					targetKey = homeKeyName
				}
				tcInfo := GetTonalCenterInfo(targetKey)

				label := "bVII7 (backdoor)"
				startMeas := c1.MeasureIndex
				startBeat := c1.BeatOffset
				if i > 0 && (chords[i-1].Chord.Quality == QualityMinor7 || chords[i-1].Chord.Quality == QualityMinorTriad) &&
					semitoneDistance(chords[i-1].Chord.Root.Name(), c1.Chord.Root.Name()) == 5 {
					label = "iv - bVII7 (backdoor)"
					startMeas = chords[i-1].MeasureIndex
					startBeat = chords[i-1].BeatOffset
				}

				devices = append(devices, HarmonicDevice{
					Type:          DeviceBackdoorDominant,
					Label:         label,
					StartMeasure:  startMeas,
					StartBeat:     startBeat,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetKey,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      true,
				})

				i++
				continue
			}

			// Tritone sub to tonic minor / Phrygian dominant (dTarget == 1, i.e. bII7 resolving down half-step to i)
			if dTarget == 1 && (targetRoot == homeRoot || (c2.TonalCenter != "" && targetRoot == strings.Fields(c2.TonalCenter)[0])) {
				targetKey := c2.TonalCenter
				if targetKey == "" {
					targetKey = homeKeyName
				}
				tcInfo := GetTonalCenterInfo(targetKey)

				devices = append(devices, HarmonicDevice{
					Type:          DeviceTritoneSub,
					Label:         "subV7 / i (tritone sub)",
					StartMeasure:  c1.MeasureIndex,
					StartBeat:     c1.BeatOffset,
					EndMeasure:    c2.MeasureIndex,
					EndBeat:       c2.BeatOffset,
					Resolves:      true,
					TargetKey:     targetKey,
					LilyPondColor: tcInfo.LilyPondColor,
					IsDashed:      true,
				})

				i++
				continue
			}
		}

		i++
	}

	return devices
}

// romanNumeralSecondaryDominant determines the Roman label (e.g. "V / vi", "V / ii", "V - I").
func romanNumeralSecondaryDominant(targetRoot string, targetQuality ChordQuality, tonicCenter string) string {
	parts := strings.Fields(tonicCenter)
	tonicRoot := parts[0]
	isMinor := len(parts) > 1 && strings.EqualFold(parts[1], "min")

	diff := semitoneDistance(tonicRoot, targetRoot)

	if !isMinor {
		switch diff {
		case 0:
			return "V - I"
		case 2:
			return "V / ii"
		case 4:
			return "V / iii"
		case 5:
			return "V / IV"
		case 7:
			return "V / V"
		case 9:
			return "V / vi"
		case 11:
			return "V / vii"
		}
	} else {
		switch diff {
		case 0:
			return "V - i"
		case 2:
			return "V / ii°"
		case 3:
			return "V / III"
		case 5:
			return "V / iv"
		case 7:
			return "V / v"
		case 8:
			return "V / VI"
		case 10:
			return "V / VII"
		}
	}

	return fmt.Sprintf("V / %s", targetRoot)
}
