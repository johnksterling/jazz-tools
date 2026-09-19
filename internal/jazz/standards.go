package jazz

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

//go:embed jazz1400.txt
var jazz1400Data string

//go:embed standards_xml/*.musicxml
var standardsXMLFS embed.FS

// StandardTune represents an entry in the built-in standards catalog.
type StandardTune struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Composer  string `json:"composer"`
	Style     string `json:"style"`
	Key       string `json:"key"`
	URL       string `json:"url"`
	HasMelody bool   `json:"hasMelody"`
	Source    string `json:"source"` // "musicxml" or "ireal"
}

var (
	standardsCache []StandardTune
	standardsOnce  sync.Once
)

// GetMusicXMLStandardData reads an embedded MusicXML standard file.
func GetMusicXMLStandardData(filename string) ([]byte, error) {
	cleanName := strings.TrimPrefix(filename, "musicxml://")
	cleanName = strings.TrimPrefix(cleanName, "standards_xml/")
	return standardsXMLFS.ReadFile("standards_xml/" + cleanName)
}

// GetAllStandards parses and returns all built-in jazz standards (both MusicXML with melodies and iReal).
func GetAllStandards() []StandardTune {
	standardsOnce.Do(func() {
		var list []StandardTune

		// 1. Load curated MusicXML standards with full melodies
		entries, err := standardsXMLFS.ReadDir("standards_xml")
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".musicxml") {
					continue
				}
				data, err := standardsXMLFS.ReadFile("standards_xml/" + entry.Name())
				if err != nil {
					continue
				}
				tune, err := ParseMusicXML(bytes.NewReader(data))
				if err != nil {
					continue
				}
				list = append(list, StandardTune{
					ID:        "xml-" + strings.TrimSuffix(entry.Name(), ".musicxml"),
					Title:     tune.Title,
					Composer:  tune.Composer,
					Style:     "Lead Sheet (Melody & Chords)",
					Key:       tune.KeyName(),
					URL:       "musicxml://" + entry.Name(),
					HasMelody: true,
					Source:    "musicxml",
				})
			}
		}

		// Sort MusicXML standards alphabetically by Title
		sort.Slice(list, func(i, j int) bool {
			return list[i].Title < list[j].Title
		})

		// 2. Load iReal standards
		raw := strings.TrimSpace(jazz1400Data)
		if idx := strings.Index(raw, "irealb://"); idx != -1 {
			raw = raw[idx+len("irealb://"):]
		}

		chunks := strings.Split(raw, "===")
		for i, chunk := range chunks {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" || strings.Contains(chunk, "Jazz 1400") || strings.Contains(chunk, "%4A%61%7A%7A%20%31%34%30%30") {
				continue
			}

			decoded, err := url.QueryUnescape(chunk)
			if err != nil {
				continue
			}

			parts := strings.Split(decoded, "=")
			if len(parts) >= 7 {
				title := strings.TrimSpace(parts[0])
				composer := strings.TrimSpace(parts[1])
				style := ""
				if len(parts) > 3 {
					style = strings.TrimSpace(parts[3])
				}
				key := ""
				if len(parts) > 4 {
					key = strings.TrimSpace(parts[4])
				}
				music := strings.TrimSpace(parts[6])

				if title != "" && music != "" {
					singleURL := "irealb://" + chunk + "==="
					list = append(list, StandardTune{
						ID:        fmt.Sprintf("std-%d", i+1),
						Title:     title,
						Composer:  composer,
						Style:     style,
						Key:       key,
						URL:       singleURL,
						HasMelody: false,
						Source:    "ireal",
					})
				}
			}
		}
		standardsCache = list
	})
	return standardsCache
}

// SearchStandards searches standards by title or composer.
// Matches with HasMelody == true are given priority ranking.
func SearchStandards(query string, limit int) []StandardTune {
	all := GetAllStandards()
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		if limit > 0 && len(all) > limit {
			return all[:limit]
		}
		return all
	}

	var melodyExact []StandardTune
	var otherExact []StandardTune
	var melodyPrefix []StandardTune
	var otherPrefix []StandardTune
	var melodySub []StandardTune
	var otherSub []StandardTune

	for _, s := range all {
		tLower := strings.ToLower(s.Title)
		cLower := strings.ToLower(s.Composer)

		if tLower == q {
			if s.HasMelody {
				melodyExact = append(melodyExact, s)
			} else {
				otherExact = append(otherExact, s)
			}
		} else if strings.HasPrefix(tLower, q) {
			if s.HasMelody {
				melodyPrefix = append(melodyPrefix, s)
			} else {
				otherPrefix = append(otherPrefix, s)
			}
		} else if strings.Contains(tLower, q) || strings.Contains(cLower, q) {
			if s.HasMelody {
				melodySub = append(melodySub, s)
			} else {
				otherSub = append(otherSub, s)
			}
		}
	}

	var results []StandardTune
	results = append(results, melodyExact...)
	results = append(results, otherExact...)
	results = append(results, melodyPrefix...)
	results = append(results, otherPrefix...)
	results = append(results, melodySub...)
	results = append(results, otherSub...)

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// FindStandardByExactTitle searches for an exact title match (case-insensitive),
// prioritizing standards with full melody.
func FindStandardByExactTitle(title string) (*StandardTune, bool) {
	q := strings.ToLower(strings.TrimSpace(title))
	all := GetAllStandards()
	for _, s := range all {
		if s.HasMelody && strings.ToLower(s.Title) == q {
			return &s, true
		}
	}
	for _, s := range all {
		if strings.ToLower(s.Title) == q {
			return &s, true
		}
	}
	return nil, false
}
