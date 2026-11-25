package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	Data             string                `json:"data,omitempty"`
}

// parseTuneFromRequest extracts and decodes a Tune from JSON or multipart form.
func (s *Server) parseTuneFromRequest(r *http.Request) (*jazz.Tune, error) {
	contentType := r.Header.Get("Content-Type")

	// 1. Multipart Form Upload
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB limit
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

		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read uploaded file: %w", err)
			}

			mime := ""
			if header != nil {
				mime = header.Header.Get("Content-Type")
				if mime == "" || mime == "application/octet-stream" {
					ext := strings.ToLower(filepath.Ext(header.Filename))
					switch ext {
					case ".pdf":
						mime = "application/pdf"
					case ".png":
						mime = "image/png"
					case ".jpg", ".jpeg":
						mime = "image/jpeg"
					case ".webp":
						mime = "image/webp"
					}
				}
			}
			if mime == "" || mime == "application/octet-stream" {
				mime = http.DetectContentType(data)
			}

			if strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "application/pdf") {
				apiKey := r.FormValue("apiKey")
				if apiKey == "" {
					apiKey = r.Header.Get("X-Gemini-API-Key")
				}
				model := r.FormValue("model")
				return jazz.TranscribeChart(data, mime, jazz.TranscribeOptions{APIKey: apiKey, Model: model})
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
				if t, err := jazz.LoadTune("testdata/Jordu.musicxml"); err == nil {
					return t, nil
				}
				return jazz.LoadTune("musicxml://standards_xml/Jordu.musicxml")
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
	if results == nil {
		results = []jazz.StandardTune{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := buildAnalyzeResponse(tune)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleTranscribe(w http.ResponseWriter, r *http.Request) {
	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := buildAnalyzeResponse(tune)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func buildAnalyzeResponse(tune *jazz.Tune) AnalyzeResponse {
	// Run voice leading
	allChords := tune.AllChords()
	pairs, _ := jazz.VoiceLeadChords(allChords)

	// Run tonal center analysis and device detection
	jazz.AnalyzeTonalCenters(tune)
	devices := jazz.DetectHarmonicDevices(tune)
	if devices == nil {
		devices = []jazz.HarmonicDevice{}
	}
	journeySpans := jazz.GetHarmonicJourneySpans(tune)
	if journeySpans == nil {
		journeySpans = []jazz.JourneySpan{}
	}
	journeyStr := jazz.HarmonicJourney(tune)

	measuresDTO := make([]MeasureDTO, 0, len(tune.Measures))
	pairIdx := 0

	for _, m := range tune.Measures {
		mDTO := MeasureDTO{
			Number: m.Number,
			Chords: make([]ChordDTO, 0, len(m.Chords)),
		}
		for _, tc := range m.Chords {
			var p jazz.GuideTonePair
			if pairIdx < len(pairs) {
				p = pairs[pairIdx]
				pairIdx++
			}
			colorHex := tc.TonalCenterHex
			if colorHex == "" {
				tcInfo := jazz.GetTuneTonalCenterInfo(tune, tc.TonalCenter)
				colorHex = tcInfo.ColorHex
			}
			mDTO.Chords = append(mDTO.Chords, ChordDTO{
				Symbol:        tc.Chord.String(),
				BeatOffset:    tc.BeatOffset,
				DurationBeats: tc.DurationBeats,
				Voice1:        p.Voice1.String(),
				Voice2:        p.Voice2.String(),
				TonalCenter:   tc.TonalCenter,
				ColorHex:      colorHex,
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

	xmlStr, _ := jazz.GenerateMusicXMLGuideTones(tune)

	return AnalyzeResponse{
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
		Data:             xmlStr,
	}
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
	var fullReport bool
	var apiKey, model string

	if qFull := r.URL.Query().Get("full"); qFull == "1" || strings.EqualFold(qFull, "true") {
		fullReport = true
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") && r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			var meta struct {
				Full   bool   `json:"full"`
				APIKey string `json:"apiKey"`
				Model  string `json:"model"`
			}
			if json.Unmarshal(bodyBytes, &meta) == nil {
				if meta.Full {
					fullReport = true
				}
				apiKey = meta.APIKey
				model = meta.Model
			}
		}
	}

	tune, err := s.parseTuneFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if fVal := r.FormValue("full"); fVal == "1" || strings.EqualFold(fVal, "true") {
			fullReport = true
		}
		if apiKey == "" {
			apiKey = r.FormValue("apiKey")
		}
		if model == "" {
			model = r.FormValue("model")
		}
	}
	if apiKey == "" {
		apiKey = r.Header.Get("X-Gemini-API-Key")
	}

	pdfOpts := jazz.CompanionPDFOptions{
		BarsPerLine:           0, // auto
		AnnotationConfig:      jazz.DefaultAnnotationConfig(),
		IncludeAnalysisReport: fullReport,
	}
	if fullReport {
		summary, isLLM, sumErr := jazz.GenerateHarmonicSummary(tune, jazz.TranscribeOptions{
			APIKey: apiKey,
			Model:  model,
		})
		if sumErr != nil {
			log.Printf("warning: LLM harmonic summary failed (using deterministic fallback): %v", sumErr)
		}
		if isLLM {
			w.Header().Set("X-LLM-Summary", "true")
		} else {
			w.Header().Set("X-LLM-Summary", "false")
		}
		pdfOpts.HarmonicSummary = summary
		pdfOpts.IsLLMSummary = isLLM
	}

	lyStr, err := jazz.GenerateLilyPondFullCompanion(tune, pdfOpts)
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

// CompanionSVGResponse represents the list of SVG pages compiled by LilyPond.
type CompanionSVGResponse struct {
	Pages []string `json:"pages"`
}

func (s *Server) handleCompanionSVG(w http.ResponseWriter, r *http.Request) {
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

	tmpDir, err := os.MkdirTemp("", "jazz-svg-*")
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

	pages, err := jazz.CompileLilyPondToSVG(lyPath, tmpDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("LilyPond SVG compilation failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CompanionSVGResponse{Pages: pages})
}

func (s *Server) handleBatchMSF(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB limit for .msf bundles
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing .msf 'file' in form upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	msfBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read .msf file: %v", err), http.StatusBadRequest)
		return
	}

	fullReport := r.FormValue("full") == "1" || strings.EqualFold(r.FormValue("full"), "true")
	forceVision := r.FormValue("transcribe") == "1" || strings.EqualFold(r.FormValue("transcribe"), "true")
	apiKey := r.FormValue("apiKey")
	if apiKey == "" {
		apiKey = r.Header.Get("X-Gemini-API-Key")
	}
	model := r.FormValue("model")

	zipBytes, setlistName, results, err := jazz.ProcessMSFToZip(msfBytes, jazz.BatchOptions{
		FullReport:       fullReport,
		ForceVision:      forceVision,
		AnnotationConfig: jazz.DefaultAnnotationConfig(),
		TranscribeOpts: jazz.TranscribeOptions{
			APIKey: apiKey,
			Model:  model,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process .msf setlist: %v", err), http.StatusBadRequest)
		return
	}

	succeeded := 0
	for _, res := range results {
		if res.Source != "error" {
			succeeded++
		}
	}
	log.Printf("batch .msf %q: %d/%d songs generated", setlistName, succeeded, len(results))

	safeName := strings.NewReplacer(" ", "_", "/", "-", "\\", "-").Replace(setlistName)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_companion_bundle.zip\"", safeName))
	w.Header().Set("X-Batch-Total", strconv.Itoa(len(results)))
	w.Header().Set("X-Batch-Succeeded", strconv.Itoa(succeeded))
	_, _ = w.Write(zipBytes)
}

type batchJob struct {
	ID          string `json:"jobId"`
	Status      string `json:"status"` // "running", "done", "error"
	SetlistName string `json:"setlistName"`
	Completed   int    `json:"completed"`
	Total       int    `json:"total"`
	Succeeded   int    `json:"succeeded"`
	CurrentSong string `json:"currentSong,omitempty"`
	Error       string `json:"error,omitempty"`
	zipBytes    []byte
	createdAt   time.Time
}

var (
	batchJobsMu sync.RWMutex
	batchJobs   = make(map[string]*batchJob)
)

func newBatchJobID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func pruneOldBatchJobs() {
	cutoff := time.Now().Add(-20 * time.Minute)
	batchJobsMu.Lock()
	for id, j := range batchJobs {
		if j.createdAt.Before(cutoff) {
			delete(batchJobs, id)
		}
	}
	batchJobsMu.Unlock()
}

func (s *Server) handleBatchMSFStart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing .msf 'file' in form upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	msfBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read .msf file: %v", err), http.StatusBadRequest)
		return
	}

	archive, err := jazz.ParseMSF(msfBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid .msf file: %v", err), http.StatusBadRequest)
		return
	}

	fullReport := r.FormValue("full") == "1" || strings.EqualFold(r.FormValue("full"), "true")
	forceVision := r.FormValue("transcribe") == "1" || strings.EqualFold(r.FormValue("transcribe"), "true")
	apiKey := r.FormValue("apiKey")
	if apiKey == "" {
		apiKey = r.Header.Get("X-Gemini-API-Key")
	}
	model := r.FormValue("model")

	pruneOldBatchJobs()
	job := &batchJob{
		ID:          newBatchJobID(),
		Status:      "running",
		SetlistName: archive.SetlistName,
		Total:       len(archive.Songs),
		createdAt:   time.Now(),
	}

	batchJobsMu.Lock()
	batchJobs[job.ID] = job
	batchJobsMu.Unlock()

	go func() {
		zipBytes, setlistName, results, err := jazz.ProcessMSFToZip(msfBytes, jazz.BatchOptions{
			FullReport:       fullReport,
			ForceVision:      forceVision,
			AnnotationConfig: jazz.DefaultAnnotationConfig(),
			TranscribeOpts: jazz.TranscribeOptions{
				APIKey: apiKey,
				Model:  model,
			},
			Progress: func(completed, total int, song jazz.MSFSong, status string) {
				batchJobsMu.Lock()
				job.Completed = completed
				job.Total = total
				job.CurrentSong = song.Title
				batchJobsMu.Unlock()
			},
		})

		batchJobsMu.Lock()
		defer batchJobsMu.Unlock()
		if err != nil {
			job.Status = "error"
			job.Error = err.Error()
			return
		}
		succeeded := 0
		for _, res := range results {
			if res.Source != "error" {
				succeeded++
			}
		}
		job.SetlistName = setlistName
		job.Completed = len(results)
		job.Succeeded = succeeded
		job.zipBytes = zipBytes
		job.Status = "done"
		log.Printf("async batch .msf %q (%s): %d/%d songs generated", setlistName, job.ID, succeeded, len(results))
	}()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(job)
}

func (s *Server) handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batchJobsMu.RLock()
	job, ok := batchJobs[id]
	if !ok {
		batchJobsMu.RUnlock()
		http.Error(w, "batch job not found", http.StatusNotFound)
		return
	}
	snap := *job
	batchJobsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

func (s *Server) handleBatchDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batchJobsMu.RLock()
	job, ok := batchJobs[id]
	if !ok {
		batchJobsMu.RUnlock()
		http.Error(w, "batch job not found", http.StatusNotFound)
		return
	}
	status := job.Status
	setlistName := job.SetlistName
	zipBytes := job.zipBytes
	errMsg := job.Error
	batchJobsMu.RUnlock()

	if status == "error" {
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	if status != "done" || len(zipBytes) == 0 {
		http.Error(w, "batch job not ready", http.StatusConflict)
		return
	}

	safeName := strings.NewReplacer(" ", "_", "/", "-", "\\", "-").Replace(setlistName)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_companion_bundle.zip\"", safeName))
	_, _ = w.Write(zipBytes)
}
