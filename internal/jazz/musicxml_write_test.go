package jazz

import (
	"encoding/xml"
	"os"
	"testing"
)

func TestGenerateMusicXMLGuideTones(t *testing.T) {
	tune, err := ParseMusicXMLFile("testdata/Jordu.musicxml")
	if err != nil {
		t.Fatalf("Failed to parse testdata/Jordu.musicxml: %v", err)
	}

	xmlStr, err := GenerateMusicXMLGuideTones(tune)
	if err != nil {
		t.Fatalf("GenerateMusicXMLGuideTones failed: %v", err)
	}

	if len(xmlStr) == 0 {
		t.Fatalf("Generated XML is empty")
	}

	// Verify the generated XML is valid XML syntax
	var parsedScore xmlScorePartwise
	if err := xml.Unmarshal([]byte(xmlStr), &parsedScore); err != nil {
		t.Fatalf("Generated XML failed to unmarshal: %v", err)
	}

	// Save to temporary companion file
	outFile := "test_companion.musicxml"
	defer os.Remove(outFile)

	if err := os.WriteFile(outFile, []byte(xmlStr), 0644); err != nil {
		t.Fatalf("Failed to write companion file: %v", err)
	}

	t.Logf("Successfully generated companion MusicXML score (%d bytes)", len(xmlStr))
}
