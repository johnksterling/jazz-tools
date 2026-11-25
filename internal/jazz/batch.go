package jazz

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BatchOptions controls how a MobileSheets .msf setlist is batch-processed.
type BatchOptions struct {
	FullReport       bool
	ForceVision      bool // Always transcribe embedded PDFs via Vision LLM instead of matching built-in standards
	TranscribeOpts   TranscribeOptions
	AnnotationConfig AnnotationConfig
	BarsPerLine      int
	Progress         func(idx, total int, song MSFSong, status string)
}

// BatchResultItem records the outcome for one song in a batch run.
type BatchResultItem struct {
	Index            int    `json:"index"`
	OriginalTitle    string `json:"originalTitle"`
	OriginalFilename string `json:"originalFilename"`
	MatchedStandard  string `json:"matchedStandard,omitempty"`
	Source           string `json:"source"` // "standard", "vision", or "error"
	OutputFilename   string `json:"outputFilename,omitempty"`
	Composer         string `json:"composer,omitempty"`
	Key              string `json:"key,omitempty"`
	Error            string `json:"error,omitempty"`
}

var nonFileCharsRe = regexp.MustCompile(`[^a-zA-Z0-9_\-]+`)

func sanitizeFilePart(s string) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), " ", "_")
	s = nonFileCharsRe.ReplaceAllString(s, "")
	if s == "" {
		return "tune"
	}
	return s
}

func resolveSongTune(song MSFSong, opts BatchOptions) (*Tune, string, string, error) {
	if !opts.ForceVision {
		if std, ok := MatchStandardForMSF(song.Title); ok {
			tune, err := LoadTune(std.URL)
			if err == nil {
				return tune, "standard", std.Title, nil
			}
		}
	}

	tune, err := TranscribeChart(song.PDFBytes, "application/pdf", opts.TranscribeOpts)
	if err != nil {
		return nil, "error", "", err
	}
	return tune, "vision", "", nil
}

// ProcessMSFToDir extracts all songs from an .msf archive, generates companion PDFs,
// and writes them alongside a MobileSheets-compatible setlist.csv and manifest.txt.
func ProcessMSFToDir(msfData []byte, outDir string, opts BatchOptions) (string, []BatchResultItem, error) {
	archive, err := ParseMSF(msfData)
	if err != nil {
		return "", nil, err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	tmpWorkDir, err := os.MkdirTemp("", "jazz-batch-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpWorkDir)

	results := make([]BatchResultItem, 0, len(archive.Songs))
	var csvRows [][]string
	csvRows = append(csvRows, []string{"Title", "Artists", "Keys", "Setlists", "Files"})

	companionSetlist := archive.SetlistName + " (Companion)"

	for idx, song := range archive.Songs {
		item := BatchResultItem{
			Index:            idx + 1,
			OriginalTitle:    song.Title,
			OriginalFilename: song.Filename,
		}

		if opts.Progress != nil {
			opts.Progress(idx+1, len(archive.Songs), song, "analyzing")
		}

		tune, source, matchedStd, err := resolveSongTune(song, opts)
		if err != nil {
			item.Source = "error"
			item.Error = err.Error()
			results = append(results, item)
			continue
		}

		item.Source = source
		item.MatchedStandard = matchedStd
		item.Composer = tune.Composer
		item.Key = tune.KeyName()

		pdfOpts := CompanionPDFOptions{
			BarsPerLine:           opts.BarsPerLine,
			AnnotationConfig:      opts.AnnotationConfig,
			IncludeAnalysisReport: opts.FullReport,
		}
		if opts.FullReport {
			summary, isLLM, _ := GenerateHarmonicSummary(tune, opts.TranscribeOpts)
			pdfOpts.HarmonicSummary = summary
			pdfOpts.IsLLMSummary = isLLM
		}

		lyStr, err := GenerateLilyPondFullCompanion(tune, pdfOpts)
		if err != nil {
			item.Source = "error"
			item.Error = fmt.Sprintf("LilyPond generation failed: %v", err)
			results = append(results, item)
			continue
		}

		base := fmt.Sprintf("%02d_%s", idx+1, sanitizeFilePart(song.Title))
		lyPath := filepath.Join(tmpWorkDir, base+".ly")
		if err := os.WriteFile(lyPath, []byte(lyStr), 0644); err != nil {
			item.Source = "error"
			item.Error = err.Error()
			results = append(results, item)
			continue
		}

		if err := CompileLilyPondToPDF(lyPath, tmpWorkDir); err != nil {
			item.Source = "error"
			item.Error = fmt.Sprintf("LilyPond compile failed: %v", err)
			results = append(results, item)
			continue
		}

		suffix := "_companion.pdf"
		if opts.FullReport {
			suffix = "_full_companion.pdf"
		}
		outFileName := base + suffix
		compiledPDFPath := filepath.Join(tmpWorkDir, base+".pdf")
		pdfBytes, err := os.ReadFile(compiledPDFPath)
		if err != nil {
			item.Source = "error"
			item.Error = err.Error()
			results = append(results, item)
			continue
		}

		if err := os.WriteFile(filepath.Join(outDir, outFileName), pdfBytes, 0644); err != nil {
			item.Source = "error"
			item.Error = err.Error()
			results = append(results, item)
			continue
		}

		item.OutputFilename = outFileName
		results = append(results, item)

		csvRows = append(csvRows, []string{
			song.Title + " (Companion)",
			tune.Composer,
			tune.KeyName(),
			companionSetlist,
			outFileName,
		})
	}

	// Write MobileSheets CSV import file
	var csvBuf bytes.Buffer
	cw := csv.NewWriter(&csvBuf)
	_ = cw.WriteAll(csvRows)
	cw.Flush()
	csvName := sanitizeFilePart(archive.SetlistName) + "_mobilesheets.csv"
	if err := os.WriteFile(filepath.Join(outDir, csvName), csvBuf.Bytes(), 0644); err != nil {
		return archive.SetlistName, results, err
	}

	// Write human-readable manifest.txt
	var manBuf bytes.Buffer
	manBuf.WriteString(fmt.Sprintf("MobileSheets Setlist: %s\n", archive.SetlistName))
	manBuf.WriteString(fmt.Sprintf("Total Songs: %d\n", len(results)))
	manBuf.WriteString(fmt.Sprintf("MobileSheets CSV Import File: %s\n\n", csvName))
	for _, r := range results {
		if r.Source == "error" {
			manBuf.WriteString(fmt.Sprintf("[%02d] %-35s  ERROR: %s\n", r.Index, r.OriginalTitle, r.Error))
		} else {
			manBuf.WriteString(fmt.Sprintf("[%02d] %-35s  %-8s -> %s\n", r.Index, r.OriginalTitle, r.Source, r.OutputFilename))
		}
	}
	_ = os.WriteFile(filepath.Join(outDir, "manifest.txt"), manBuf.Bytes(), 0644)

	return archive.SetlistName, results, nil
}

// ProcessMSFToZip batch-processes an .msf archive into an in-memory .zip bundle
// containing all companion PDFs, MobileSheets CSV import file, and manifest.txt.
func ProcessMSFToZip(msfData []byte, opts BatchOptions) ([]byte, string, []BatchResultItem, error) {
	outDir, err := os.MkdirTemp("", "jazz-batch-zip-*")
	if err != nil {
		return nil, "", nil, err
	}
	defer os.RemoveAll(outDir)

	setlistName, results, err := ProcessMSFToDir(msfData, outDir, opts)
	if err != nil {
		return nil, "", results, err
	}

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	entries, err := os.ReadDir(outDir)
	if err != nil {
		return nil, setlistName, results, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileBytes, err := os.ReadFile(filepath.Join(outDir, entry.Name()))
		if err != nil {
			continue
		}
		fw, err := zw.Create(entry.Name())
		if err != nil {
			continue
		}
		_, _ = fw.Write(fileBytes)
	}
	if err := zw.Close(); err != nil {
		return nil, setlistName, results, err
	}

	return zipBuf.Bytes(), setlistName, results, nil
}
