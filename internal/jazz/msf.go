package jazz

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
)

// MSFSong represents a single song extracted from a MobileSheets .msf file.
type MSFSong struct {
	Title    string
	Filename string
	FilePath string
	PDFBytes []byte
}

// MSFArchive represents a parsed MobileSheets .msf bundle.
type MSFArchive struct {
	SetlistName string
	Songs       []MSFSong
}

// ParseMSF parses a MobileSheets .msf binary file and extracts its setlist name
// and all embedded PDF charts.
func ParseMSF(data []byte) (*MSFArchive, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("invalid .msf file: too short (%d bytes)", len(data))
	}

	archive := &MSFArchive{
		SetlistName: extractMSFSetlistName(data),
	}

	pdfSig := []byte("%PDF-")
	offset := 0
	for {
		idx := bytes.Index(data[offset:], pdfSig)
		if idx == -1 {
			break
		}
		pos := offset + idx

		// Scan backwards up to 512 bytes for: [uint32 BE pdfLen][uint16 BE pathLen][pathBytes]
		matched := false
		maxPathLen := 512
		if pos-6 < maxPathLen {
			maxPathLen = pos - 6
		}
		for pathLen := 4; pathLen <= maxPathLen; pathLen++ {
			u16 := int(binary.BigEndian.Uint16(data[pos-pathLen-2 : pos-pathLen]))
			if u16 != pathLen {
				continue
			}
			fileLen := int(binary.BigEndian.Uint32(data[pos-pathLen-6 : pos-pathLen-2]))
			if fileLen <= 0 || pos+fileLen > len(data) {
				continue
			}
			rawPath := string(data[pos-pathLen : pos])
			if !strings.HasSuffix(strings.ToLower(rawPath), ".pdf") {
				continue
			}
			// Normalize Windows/Android separators and extract filename
			normPath := strings.ReplaceAll(rawPath, "\\", "/")
			base := path.Base(normPath)
			pdfCopy := make([]byte, fileLen)
			copy(pdfCopy, data[pos:pos+fileLen])

			archive.Songs = append(archive.Songs, MSFSong{
				Title:    CleanMSFSongTitle(base),
				Filename: base,
				FilePath: rawPath,
				PDFBytes: pdfCopy,
			})
			offset = pos + fileLen
			matched = true
			break
		}

		if !matched {
			offset = pos + len(pdfSig)
		}
	}

	if len(archive.Songs) == 0 {
		return nil, fmt.Errorf("no embedded PDF charts found in .msf file")
	}

	return archive, nil
}

func extractMSFSetlistName(data []byte) string {
	// MobileSheets stores the setlist/collection name near the header after 0xff 0xff 0x01 0x01 [uint16 BE len]
	marker := []byte{0xff, 0xff, 0x01, 0x01}
	limit := len(data)
	if limit > 512 {
		limit = 512
	}
	if idx := bytes.Index(data[:limit], marker); idx != -1 && idx+6 < limit {
		strLen := int(binary.BigEndian.Uint16(data[idx+4 : idx+6]))
		if strLen > 0 && idx+6+strLen <= limit {
			name := strings.TrimSpace(string(data[idx+6 : idx+6+strLen]))
			if name != "" {
				return name
			}
		}
	}
	return "MobileSheets Setlist"
}

var trailingKeySuffixRe = regexp.MustCompile(`\s*-\s*(C|Db|D|Eb|E|F|Gb|F#|G|Ab|A|Bb|B|Bb Inst|Eb Inst|Bass Clef|Concert)$`)

// CleanMSFSongTitle normalizes a PDF filename from MobileSheets into a clean song title.
func CleanMSFSongTitle(filename string) string {
	s := strings.TrimSpace(filename)
	if idx := strings.LastIndex(strings.ToLower(s), ".pdf"); idx != -1 {
		s = s[:idx]
	}
	s = trailingKeySuffixRe.ReplaceAllString(s, "")
	// Real Book PDF filenames often replace apostrophes with '_' (e.g., Jardin-d_Hiver, New Year_s)
	s = strings.ReplaceAll(s, "_", "'")
	s = strings.ReplaceAll(s, "-", " ")

	// Split CamelCase filenames like "FunkInDeepFreeze" or "BerniesTune"
	if !strings.Contains(s, " ") {
		var out strings.Builder
		runes := []rune(s)
		for i, r := range runes {
			if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				out.WriteRune(' ')
			}
			out.WriteRune(r)
		}
		s = out.String()
	}

	return strings.Join(strings.Fields(s), " ")
}

func normalizeTitleForMatch(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	return strings.TrimSpace(reg.ReplaceAllString(s, " "))
}

// MatchStandardForMSF attempts to match a cleaned MobileSheets song title against
// the built-in 1,400 standards catalog.
func MatchStandardForMSF(title string) (*StandardTune, bool) {
	if std, ok := FindStandardByExactTitle(title); ok {
		return std, true
	}

	norm := normalizeTitleForMatch(title)
	if norm == "" {
		return nil, false
	}

	all := GetAllStandards()
	var best *StandardTune
	for i := range all {
		s := &all[i]
		sNorm := normalizeTitleForMatch(s.Title)
		if sNorm == norm {
			if s.HasMelody {
				return s, true
			}
			if best == nil {
				best = s
			}
		}
		// Handle leading articles like "On The Sunny Side Of The Street" vs "Sunny Side Of The Street"
		// or "Bernie's Tune" vs "Bernies Tune"
		sNoPunct := strings.ReplaceAll(sNorm, " ", "")
		tNoPunct := strings.ReplaceAll(norm, " ", "")
		if sNoPunct == tNoPunct {
			if s.HasMelody {
				return s, true
			}
			if best == nil {
				best = s
			}
		}
		if strings.TrimPrefix(sNorm, "on the ") == norm || strings.TrimPrefix(sNorm, "the ") == norm {
			if s.HasMelody {
				return s, true
			}
			if best == nil {
				best = s
			}
		}
	}

	if best != nil {
		return best, true
	}

	// Subsequence word match for minor title variations (e.g., "Softly As A Morning Sunrise" -> "Softly, As In A Morning Sunrise")
	tWords := strings.Fields(norm)
	if len(tWords) >= 3 {
		for i := range all {
			s := &all[i]
			sWords := strings.Fields(normalizeTitleForMatch(s.Title))
			if len(sWords) > len(tWords) && len(sWords)-len(tWords) <= 2 && wordsSubsequence(tWords, sWords) {
				if s.HasMelody {
					return s, true
				}
				if best == nil {
					best = s
				}
			}
		}
	}

	if best != nil {
		return best, true
	}
	return nil, false
}

func wordsSubsequence(sub, full []string) bool {
	j := 0
	for _, w := range full {
		if j < len(sub) && w == sub[j] {
			j++
		}
	}
	return j == len(sub)
}
