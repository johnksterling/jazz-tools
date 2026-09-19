package jazz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadTune loads a tune from either a MusicXML file or an iReal Pro URL / file.
func LoadTune(target string) (*Tune, error) {
	if strings.HasPrefix(target, "irealb://") || strings.HasPrefix(target, "irealbook://") {
		tunes, err := ParseIRealURL(target)
		if err != nil {
			return nil, err
		}
		return tunes[0], nil
	}

	// Check if file exists
	if _, err := os.Stat(target); err != nil {
		return nil, fmt.Errorf("file not found: %s", target)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "irealb://") || strings.Contains(contentStr, "irealbook://") {
		tunes, err := ParseIRealURL(contentStr)
		if err == nil && len(tunes) > 0 {
			return tunes[0], nil
		}
	}

	ext := strings.ToLower(filepath.Ext(target))
	if ext == ".html" || ext == ".htm" {
		tunes, err := ParseIRealURL(contentStr)
		if err != nil {
			return nil, err
		}
		return tunes[0], nil
	}

	if ext == ".musicxml" || ext == ".xml" {
		return ParseMusicXMLFile(target)
	}

	return nil, fmt.Errorf("unrecognized file format or content in %s", target)
}
