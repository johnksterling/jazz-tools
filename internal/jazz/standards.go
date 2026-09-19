package jazz

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

//go:embed jazz1400.txt
var jazz1400Data string

// StandardTune represents an entry in the built-in standards catalog.
type StandardTune struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Composer string `json:"composer"`
	Style    string `json:"style"`
	Key      string `json:"key"`
	URL      string `json:"url"`
}

var (
	standardsCache []StandardTune
	standardsOnce  sync.Once
)

// GetAllStandards parses and returns all built-in jazz standards.
func GetAllStandards() []StandardTune {
	standardsOnce.Do(func() {
		raw := strings.TrimSpace(jazz1400Data)
		if idx := strings.Index(raw, "irealb://"); idx != -1 {
			raw = raw[idx+len("irealb://"):]
		}

		chunks := strings.Split(raw, "===")
		var list []StandardTune
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
						ID:       fmt.Sprintf("std-%d", i+1),
						Title:    title,
						Composer: composer,
						Style:    style,
						Key:      key,
						URL:      singleURL,
					})
				}
			}
		}
		standardsCache = list
	})
	return standardsCache
}

// SearchStandards searches standards by title or composer.
func SearchStandards(query string, limit int) []StandardTune {
	all := GetAllStandards()
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		if limit > 0 && len(all) > limit {
			return all[:limit]
		}
		return all
	}

	var exactMatches []StandardTune
	var prefixMatches []StandardTune
	var substringMatches []StandardTune

	for _, s := range all {
		tLower := strings.ToLower(s.Title)
		cLower := strings.ToLower(s.Composer)

		if tLower == q {
			exactMatches = append(exactMatches, s)
		} else if strings.HasPrefix(tLower, q) {
			prefixMatches = append(prefixMatches, s)
		} else if strings.Contains(tLower, q) || strings.Contains(cLower, q) {
			substringMatches = append(substringMatches, s)
		}
	}

	results := append(exactMatches, prefixMatches...)
	results = append(results, substringMatches...)

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// FindStandardByExactTitle searches for an exact title match (case-insensitive).
func FindStandardByExactTitle(title string) (*StandardTune, bool) {
	q := strings.ToLower(strings.TrimSpace(title))
	for _, s := range GetAllStandards() {
		if strings.ToLower(s.Title) == q {
			return &s, true
		}
	}
	return nil, false
}
