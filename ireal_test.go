package main

import (
	"testing"
)

func TestPyRealParserUnscrambleCases(t *testing.T) {
	cases := []struct {
		scrambled   string
		unscrambled string
	}{
		{
			scrambled:   "[T44A BLZC DLZE FLZG, ALZA BLZC DLZE FLZG ALZA B | ",
			unscrambled: "[T44A BLZC DLZE FLZG, ALZA BLZC DLZE FLZG ALZA B | ",
		},
		{
			scrambled:   "| B A BLZCZLF EZLD CZLB ZALA ,GZLF EZLD G ALZA44T[C ",
			unscrambled: "[T44A BLZC DLZE FLZG, ALZA BLZC DLZE FLZG ALZA B |C ",
		},
		{
			scrambled:   "ZLB A BLZCZLF EZLD CZLB ZALA ,GZLF EZLD G ALZA44T[C- DLZE FLZG A,LZA BLZC DLZE FLZG ALZA BLZC- D |E ",
			unscrambled: "[T44A BLZC DLZE FLZG, ALZA BLZC DLZE FLZG ALZA BLZC- DLZE FLZG A,LZA BLZC DLZE FLZG ALZA BLZC- D |E ",
		},
		{
			scrambled:   "ZLB A BLZCZLF EZLD CZLB ZALA ,GZLF EZLD G ALZA44T[ EZLDZE FLB AZLA GZLF EZDL CZLB AZL,A GZLZC- LD -CF ",
			unscrambled: "[T44A BLZC DLZE FLZG, ALZA BLZC DLZE FLZG ALZA BLZC- DLZE FLZG A,LZA BLZC DLZE FLZG ALZA BLZC- DLZE F ",
		},
	}

	for i, c := range cases {
		got := UnscrambleIReal(c.scrambled)
		if got != c.unscrambled {
			t.Errorf("Case %d failed:\n got  %q\n want %q", i+1, got, c.unscrambled)
		}
	}
}

func TestParseIRealTune(t *testing.T) {
	musicStr := "*A{T44D- |Eh7 A7b9|G-7 C7|F^7 }*B[F^7 |G-7 C7|F^7 |Eh7 A7b9 ]"
	tune, err := ParseIRealTune(musicStr, "Dear Old Stockholm", "Traditional", "Medium Swing", "D-")
	if err != nil {
		t.Fatalf("ParseIRealTune failed: %v", err)
	}

	if tune.Title != "Dear Old Stockholm" {
		t.Errorf("Title = %s; want Dear Old Stockholm", tune.Title)
	}
	if tune.TimeSignature != [2]int{4, 4} {
		t.Errorf("TimeSignature = %v; want [4 4]", tune.TimeSignature)
	}

	allChords := tune.AllChords()
	t.Logf("Total chords found in iReal tune: %d across %d measures", len(allChords), len(tune.Measures))

	if len(allChords) == 0 {
		t.Fatalf("Expected chords from iReal string, got 0")
	}

	// Verify measure 2 has 2 chords: Eh7 (Em7b5) and A7b9
	m2 := tune.Measures[1]
	if len(m2.Chords) != 2 {
		t.Fatalf("Measure 2 expected 2 chords, got %d", len(m2.Chords))
	}
	if m2.Chords[0].Chord.Root.Name() != "E" {
		t.Errorf("Measure 2 chord 1 root = %s; want E", m2.Chords[0].Chord.Root.Name())
	}
	if m2.Chords[1].Chord.Root.Name() != "A" {
		t.Errorf("Measure 2 chord 2 root = %s; want A", m2.Chords[1].Chord.Root.Name())
	}
	if m2.Chords[0].DurationBeats != 2.0 || m2.Chords[1].DurationBeats != 2.0 {
		t.Errorf("Measure 2 chord durations expected 2.0 each, got %.1f, %.1f",
			m2.Chords[0].DurationBeats, m2.Chords[1].DurationBeats)
	}

	// Test voice leading and guide tone generation on the iReal tune!
	xmlScore, err := GenerateMusicXMLGuideTones(tune)
	if err != nil {
		t.Fatalf("GenerateMusicXMLGuideTones failed on iReal tune: %v", err)
	}
	if len(xmlScore) == 0 {
		t.Fatalf("Generated empty MusicXML score from iReal tune")
	}
	t.Logf("Generated %d bytes of MusicXML for iReal tune", len(xmlScore))
}
