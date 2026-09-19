package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"jazz-tools/internal/jazz"
)

// SampleTune represents a built-in tune available for quick loading.
type SampleTune struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Composer string `json:"composer"`
	Type     string `json:"type"`
	Data     string `json:"data,omitempty"`
}

var sampleTunes = []SampleTune{
	{
		ID:       "jordu",
		Title:    "Jordu",
		Composer: "Duke Jordan",
		Type:     "musicxml",
	},
	{
		ID:       "waltz_for_debby",
		Title:    "Waltz For Debby",
		Composer: "Bill Evans",
		Type:     "ireal",
		Data:     "irealb://Waltz%20For%20Debby=Bill%20Evans=Medium%20Waltz=F==*A{T34F^7 |D-7 |G-7 |C7 |F^7 |D-7 |G-7 |C7 }*B[F^7 |F7 |Bb^7 |Bb-7 |A-7 |D7 |G-7 |C7 ]*C{F^7 |D-7 |G-7 |C7 |F^7 |F7 |Bb^7 |Bb-7 }*D[A-7 |D7 |G-7 |C7 |Eb^7 |Ab^7 |Db^7 |C7 ]*E{F^7 |D-7 |G-7 |C7 |F^7 |G-7 C7|F^7 |C7 Z",
	},
}

// ChordDTO represents a chord event with voice-led guide tones.
type ChordDTO struct {
	Symbol        string  `json:"symbol"`
	BeatOffset    float64 `json:"beatOffset"`
	DurationBeats float64 `json:"durationBeats"`
	Voice1        string  `json:"voice1"`
	Voice2        string  `json:"voice2"`
	TonalCenter   string  `json:"tonalCenter"`
	ColorHex      string  `json:"colorHex"`
}

// MelodyNoteDTO represents a note or rest in the lead sheet melody.
type MelodyNoteDTO struct {
	Pitch         string  `json:"pitch"`
	DurationBeats float64 `json:"durationBeats"`
	BeatOffset    float64 `json:"beatOffset"`
	IsRest        bool    `json:"isRest"`
	Lyric         string  `json:"lyric,omitempty"`
}

// MeasureDTO represents a measure containing chord events and optional melody notes.
type MeasureDTO struct {
	Number int             `json:"number"`
	Chords []ChordDTO      `json:"chords"`
	Melody []MelodyNoteDTO `json:"melody,omitempty"`
}

// AnalyzeResponse contains comprehensive harmonic analysis of a tune.
type AnalyzeResponse struct {
	Title            string                `json:"title"`
	Composer         string                `json:"composer"`
	Key              string                `json:"key"`
	TimeSignature    [2]int                `json:"timeSignature"`
	MeasureCount     int                   `json:"measureCount"`
	HarmonicJourney  string                `json:"harmonicJourney"`
	JourneySpans     []jazz.JourneySpan    `json:"journeySpans"`
	Measures         []MeasureDTO          `json:"measures"`
	Devices          []jazz.HarmonicDevice `json:"devices"`
	HasMelody        bool                  `json:"hasMelody"`
	MelodyNotesCount int                   `json:"melodyNotesCount"`
}

// parseTuneFromRequest extracts and decodes a Tune from JSON or multipart form.
func (s *Server) parseTuneFromRequest(r *http.Request) (*jazz.Tune, error) {
	contentType := r.Header.Get("Content-Type")

	// 1. Multipart Form Upload
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
			return nil, fmt.Errorf("failed to parse multipart form: %w", err)
		}

		if standard := r.FormValue("standard"); standard != "" {
			return jazz.LoadTune(standard)
		}

		if sampleID := r.FormValue("sample"); sampleID != "" {
			return s.loadSampleTune(sampleID)
		}

		if textContent := r.FormValue("content"); strings.TrimSpace(textContent) != "" {
			return jazz.ParseTuneData(textContent)
		}

		file, _, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read uploaded file: %w", err)
			}
			return jazz.ParseTuneData(string(data))
		}
	}

	// 2. JSON Body
	if strings.HasPrefix(contentType, "application/json") || r.Body != nil {
		var req struct {
			Content  string `json:"content"`
			Sample   string `json:"sample"`
			Standard string `json:"standard"`
		}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err == nil {
			if req.Standard != "" {
				return jazz.LoadTune(req.Standard)
			}
			if req.Sample != "" {
				return s.loadSampleTune(req.Sample)
			}
			if strings.TrimSpace(req.Content) != "" {
				return jazz.ParseTuneData(req.Content)
			}
		}
	}

	// 3. Fallback to query params (useful for GET / test queries)
	if standard := r.URL.Query().Get("standard"); standard != "" {
		return jazz.LoadTune(standard)
	}
	if sampleID := r.URL.Query().Get("sample"); sampleID != "" {
		return s.loadSampleTune(sampleID)
	}

	return nil, fmt.Errorf("no tune content provided (provide 'content', upload a 'file', select a 'sample', or search a 'standard')")
}

func (s *Server) loadSampleTune(id string) (*jazz.Tune, error) {
	for _, sample := range sampleTunes {
		if sample.ID == id {
			if sample.Data != "" {
				return jazz.ParseTuneData(sample.Data)
			}
			if sample.ID == "jordu" {
				return jazz.LoadTune("testdata/Jordu.musicxml")
			}
		}
	}
	return nil, fmt.Errorf("unknown sample tune: %s", id)
}

func (s *Server) handleSamples(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sampleTunes)
}

func (s *Server) handleStandards(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results := jazz.SearchStandards(q, limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Run voice leading
	allChords := tune.AllChords()
	pairs, _ := jazz.VoiceLeadChords(allChords)

	// Run tonal center analysis and device detection
	jazz.AnalyzeTonalCenters(tune)
	devices := jazz.DetectHarmonicDevices(tune)
	journeySpans := jazz.GetHarmonicJourneySpans(tune)
	journeyStr := jazz.HarmonicJourney(tune)

	var measuresDTO []MeasureDTO
	pairIdx := 0

	for _, m := range tune.Measures {
		mDTO := MeasureDTO{
			Number: m.Number,
		}
		for _, tc := range m.Chords {
			var p jazz.GuideTonePair
			if pairIdx < len(pairs) {
				p = pairs[pairIdx]
				pairIdx++
			}
			tcInfo := jazz.GetTonalCenterInfo(tc.TonalCenter)
			mDTO.Chords = append(mDTO.Chords, ChordDTO{
				Symbol:        tc.Chord.String(),
				BeatOffset:    tc.BeatOffset,
				DurationBeats: tc.DurationBeats,
				Voice1:        p.Voice1.String(),
				Voice2:        p.Voice2.String(),
				TonalCenter:   tc.TonalCenter,
				ColorHex:      tcInfo.ColorHex,
			})
		}
		for _, mn := range m.Melody {
			mDTO.Melody = append(mDTO.Melody, MelodyNoteDTO{
				Pitch:         mn.PitchName(),
				DurationBeats: mn.DurationBeats,
				BeatOffset:    mn.BeatOffset,
				IsRest:        mn.IsRest,
				Lyric:         mn.Lyric,
			})
		}
		measuresDTO = append(measuresDTO, mDTO)
	}

	resp := AnalyzeResponse{
		Title:            tune.Title,
		Composer:         tune.Composer,
		Key:              tune.Key,
		TimeSignature:    tune.TimeSignature,
		MeasureCount:     len(tune.Measures),
		HarmonicJourney:  journeyStr,
		JourneySpans:     journeySpans,
		Measures:         measuresDTO,
		Devices:          devices,
		HasMelody:        tune.HasMelody(),
		MelodyNotesCount: tune.MelodyNotesCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleCompanionXML(w http.ResponseWriter, r *http.Request) {
	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	xmlStr, err := jazz.GenerateMusicXMLGuideTones(tune)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate MusicXML: %v", err), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("download") == "1" {
		title := strings.ReplaceAll(tune.Title, " ", "_")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_guide_tones.musicxml\"", title))
	}
	w.Header().Set("Content-Type", "application/vnd.recordare.musicxml+xml; charset=utf-8")
	_, _ = w.Write([]byte(xmlStr))
}

func (s *Server) handleCompanionPDF(w http.ResponseWriter, r *http.Request) {
	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := jazz.DefaultAnnotationConfig()
	bars := 0 // auto

	lyStr, err := jazz.GenerateLilyPondScoreWithAnnotationConfig(tune, bars, cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate LilyPond score: %v", err), http.StatusInternalServerError)
		return
	}

	tmpDir, err := os.MkdirTemp("", "jazz-pdf-*")
	if err != nil {
		http.Error(w, "failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	lyPath := filepath.Join(tmpDir, "score.ly")
	if err := os.WriteFile(lyPath, []byte(lyStr), 0644); err != nil {
		http.Error(w, "failed to write lilypond file", http.StatusInternalServerError)
		return
	}

	if err := jazz.CompileLilyPondToPDF(lyPath, tmpDir); err != nil {
		http.Error(w, fmt.Sprintf("LilyPond PDF compilation failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	pdfPath := filepath.Join(tmpDir, "score.pdf")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, "failed to read generated PDF", http.StatusInternalServerError)
		return
	}

	title := strings.ReplaceAll(tune.Title, " ", "_")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_guide_tones.pdf\"", title))
	_, _ = w.Write(pdfBytes)
}
