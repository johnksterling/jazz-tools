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
		if len(tunes) == 0 {
			return nil, fmt.Errorf("no tunes found in iReal URL")
		}
		return tunes[0], nil
	}

	// Check if file exists
	if _, err := os.Stat(target); err != nil {
		// Check if target matches a built-in jazz standard title
		if std, ok := FindStandardByExactTitle(target); ok {
			return LoadTune(std.URL)
		}
		if matches := SearchStandards(target, 2); len(matches) == 1 {
			return LoadTune(matches[0].URL)
		}
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
		if len(tunes) == 0 {
			return nil, fmt.Errorf("no tunes found in html")
		}
		return tunes[0], nil
	}

	if ext == ".musicxml" || ext == ".xml" {
		return ParseMusicXMLFile(target)
	}

	return nil, fmt.Errorf("unrecognized file format or content in %s", target)
}

// ParseTuneData parses tune data from raw string content (iReal URL or MusicXML).
func ParseTuneData(data string) (*Tune, error) {
	trimmed := strings.TrimSpace(data)
	if strings.HasPrefix(trimmed, "irealb://") || strings.HasPrefix(trimmed, "irealbook://") {
		tunes, err := ParseIRealURL(trimmed)
		if err != nil {
			return nil, err
		}
		if len(tunes) == 0 {
			return nil, fmt.Errorf("no tunes found in iReal URL")
		}
		return tunes[0], nil
	}

	if strings.Contains(trimmed, "irealb://") || strings.Contains(trimmed, "irealbook://") {
		tunes, err := ParseIRealURL(trimmed)
		if err == nil && len(tunes) > 0 {
			return tunes[0], nil
		}
	}

	if strings.HasPrefix(trimmed, "<?xml") || strings.Contains(trimmed, "<score-partwise") {
		return ParseMusicXML(strings.NewReader(trimmed))
	}

	if std, ok := FindStandardByExactTitle(trimmed); ok {
		return LoadTune(std.URL)
	}
	if matches := SearchStandards(trimmed, 2); len(matches) == 1 {
		return LoadTune(matches[0].URL)
	}

	return nil, fmt.Errorf("unrecognized format: expected iReal URL, standard song name, or MusicXML content")
}
