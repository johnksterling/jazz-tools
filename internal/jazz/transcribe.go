package jazz

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// TranscribeOptions holds options for chart transcription.
type TranscribeOptions struct {
	APIKey string
	Model  string // Defaults to "gemini-2.5-flash"
}

// RawExtractedChord represents a single chord extraction from the Vision LLM.
type RawExtractedChord struct {
	Symbol        string  `json:"symbol"`
	Beat          float64 `json:"beat"`
	DurationBeats float64 `json:"durationBeats"`
}

// RawExtractedMeasure represents a measure extraction from the Vision LLM.
type RawExtractedMeasure struct {
	Number int                 `json:"number"`
	Chords []RawExtractedChord `json:"chords"`
}

// RawExtractedLeadSheet represents the full lead sheet extraction.
type RawExtractedLeadSheet struct {
	Title         string                `json:"title"`
	Composer      string                `json:"composer"`
	Key           string                `json:"key"`
	TimeSignature []int                 `json:"timeSignature"`
	Measures      []RawExtractedMeasure `json:"measures"`
}

var leadSheetSchema = map[string]interface{}{
	"type": "OBJECT",
	"properties": map[string]interface{}{
		"title":    map[string]interface{}{"type": "STRING"},
		"composer": map[string]interface{}{"type": "STRING"},
		"key":      map[string]interface{}{"type": "STRING", "description": "e.g. F, Bb, C, A minor, etc."},
		"timeSignature": map[string]interface{}{
			"type":        "ARRAY",
			"items":       map[string]interface{}{"type": "INTEGER"},
			"description": "Time signature as [numerator, denominator], e.g. [4, 4] or [3, 4]",
		},
		"measures": map[string]interface{}{
			"type": "ARRAY",
			"items": map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"number": map[string]interface{}{"type": "INTEGER"},
					"chords": map[string]interface{}{
						"type": "ARRAY",
						"items": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"symbol":        map[string]interface{}{"type": "STRING", "description": "Standard jazz chord symbol"},
								"beat":          map[string]interface{}{"type": "NUMBER", "description": "1-indexed starting beat in the measure (1.0 for beat 1, 3.0 for beat 3)"},
								"durationBeats": map[string]interface{}{"type": "NUMBER", "description": "Duration in beats (e.g. 2.0 or 4.0)"},
							},
							"required": []string{"symbol", "beat", "durationBeats"},
						},
					},
				},
				"required": []string{"number", "chords"},
			},
		},
	},
	"required": []string{"title", "measures"},
}

// TranscribeChart takes image or PDF bytes and transcribes the jazz chart using Gemini Vision.
func TranscribeChart(fileBytes []byte, mimeType string, opts TranscribeOptions) (*Tune, error) {
	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("empty file data")
	}

	if os.Getenv("MOCK_TRANSCRIBE") == "1" || strings.EqualFold(opts.APIKey, "mock") {
		return mockTranscribeTune(), nil
	}

	rawAuth := cleanAPIKey(opts.APIKey)
	if rawAuth == "" {
		rawAuth = cleanAPIKey(os.Getenv("GEMINI_API_KEY"))
	}

	var isBearer bool
	var authKey string

	if strings.HasPrefix(rawAuth, "ya29.") {
		// User explicitly provided an OAuth2 / ADC access token
		isBearer = true
		authKey = rawAuth
	} else if rawAuth != "" {
		// User provided an API key (e.g. AIzaSy...)
		isBearer = false
		authKey = rawAuth
	} else {
		// Attempt Application Default Credentials (ADC) resolution
		adcToken, err := ResolveADCAccessToken()
		if err == nil && adcToken != "" {
			isBearer = true
			authKey = adcToken
		} else {
			return nil, fmt.Errorf("no Gemini API key or Google Application Default Credentials found. Please provide an API key (AIzaSy...), an OAuth token (ya29...), or configure ADC via 'gcloud auth application-default login'")
		}
	}

	model := normalizeModel(opts.Model)

	if mimeType == "" {
		mimeType = http.DetectContentType(fileBytes)
	}
	// Normalize common types
	if strings.HasPrefix(mimeType, "application/pdf") {
		mimeType = "application/pdf"
	} else if strings.HasPrefix(mimeType, "image/png") {
		mimeType = "image/png"
	} else if strings.HasPrefix(mimeType, "image/jpeg") || strings.HasPrefix(mimeType, "image/jpg") {
		mimeType = "image/jpeg"
	} else if strings.HasPrefix(mimeType, "image/webp") {
		mimeType = "image/webp"
	}

	b64Data := base64.StdEncoding.EncodeToString(fileBytes)

	prompt := `You are an expert jazz musician, copyist, and musicologist.
Analyze this jazz lead sheet / chart image or PDF carefully.
Transcribe the tune's title, composer, key, time signature, and chord progression measure by measure from start to finish.

Guidelines:
1. Title and Composer: Extract the song title and composer if visible.
2. Key and Time Signature: Determine the key signature and time signature (default to [4, 4] if standard common time).
3. Measures and Chords:
   - Walk through the chart system by system, measure by measure from left to right.
   - For each measure, provide the 1-indexed measure number.
   - Extract every chord symbol. Standardize jazz notation:
     * Convert triangle 'Δ' or '^7' or 'ma7' to 'maj7' (e.g. 'Cmaj7', 'Ebmaj7').
     * Convert half-diminished 'ø' or 'ø7' or '-7b5' to 'm7b5' (e.g. 'Am7b5').
     * Convert diminished circle 'o' or 'o7' to 'dim7' (e.g. 'Bdim7').
     * Convert '-' or 'min' to 'm' (e.g. 'Am7', 'Dm7', 'Cm9').
     * Preserve alterations and extensions (e.g. '7#9', '7b9', '7alt', '7#11', '13', 'm11').
     * Preserve bass note slash chords (e.g. 'C/E', 'Db/F', 'Bb/D').
   - For each chord, specify 'beat': starting beat in the measure (1-indexed: 1.0 for beat 1, 3.0 for beat 3, etc.).
   - For each chord, specify 'durationBeats': duration in beats (e.g. 4.0 for a whole bar of 4/4; 2.0 each if two chords share a bar).
4. Measure Repeats & Endings:
   - Transcribe all measures through the full form. If there are 1st and 2nd endings, transcribe the progression sequentially.`

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"inlineData": map[string]interface{}{
							"mimeType": mimeType,
							"data":     b64Data,
						},
					},
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema":   leadSheetSchema,
			"temperature":      0.1,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var apiURL string
	if isBearer {
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	} else {
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, url.QueryEscape(authKey))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if isBearer {
		httpReq.Header.Set("Authorization", "Bearer "+authKey)
	} else {
		httpReq.Header.Set("x-goog-api-key", authKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respStr := string(respBytes)
		if strings.Contains(respStr, "API_KEY_INVALID") {
			return nil, fmt.Errorf("invalid Gemini API Key. Google AI Studio keys typically start with 'AIzaSy'. Please verify your key at https://aistudio.google.com/app/apikey (ensure you are using a Gemini API key from AI Studio, not an OpenAI/Anthropic key or GCP service account)")
		}
		return nil, fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, respStr)
	}

	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini API response wrapper: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini API returned no content candidates")
	}

	extractedJSON := apiResp.Candidates[0].Content.Parts[0].Text

	var rawLeadSheet RawExtractedLeadSheet
	if err := json.Unmarshal([]byte(extractedJSON), &rawLeadSheet); err != nil {
		return nil, fmt.Errorf("failed to parse extracted lead sheet JSON: %w (raw: %s)", err, extractedJSON)
	}

	return convertRawLeadSheetToTune(&rawLeadSheet)
}

func convertRawLeadSheetToTune(raw *RawExtractedLeadSheet) (*Tune, error) {
	tune := &Tune{
		Title:         strings.TrimSpace(raw.Title),
		Composer:      strings.TrimSpace(raw.Composer),
		Key:           strings.TrimSpace(raw.Key),
		TimeSignature: [2]int{4, 4},
	}
	if tune.Title == "" {
		tune.Title = "Transcribed Lead Sheet"
	}
	if len(raw.TimeSignature) == 2 && raw.TimeSignature[0] > 0 && raw.TimeSignature[1] > 0 {
		tune.TimeSignature = [2]int{raw.TimeSignature[0], raw.TimeSignature[1]}
	}

	beatsPerBar := float64(tune.TimeSignature[0])

	for _, mRaw := range raw.Measures {
		m := TimedMeasure{
			Number: mRaw.Number,
		}

		for _, cRaw := range mRaw.Chords {
			sym := strings.TrimSpace(cRaw.Symbol)
			if sym == "" || sym == "%" || sym == "/" {
				continue
			}

			chord, err := ParseChord(sym)
			if err != nil {
				normalized := normalizeChordSymbol(sym)
				if c2, err2 := ParseChord(normalized); err2 == nil {
					chord = c2
				} else {
					continue
				}
			}

			beatOffset := cRaw.Beat - 1.0
			if beatOffset < 0 {
				beatOffset = 0
			}
			dur := cRaw.DurationBeats
			if dur <= 0 {
				dur = beatsPerBar - beatOffset
			}
			if beatOffset+dur > beatsPerBar {
				dur = beatsPerBar - beatOffset
			}

			m.Chords = append(m.Chords, TimedChord{
				Chord:         chord,
				MeasureNumber: m.Number,
				BeatOffset:    beatOffset,
				DurationBeats: dur,
			})
		}

		sort.Slice(m.Chords, func(i, j int) bool {
			return m.Chords[i].BeatOffset < m.Chords[j].BeatOffset
		})

		tune.Measures = append(tune.Measures, m)
	}

	if len(tune.Measures) == 0 {
		return nil, fmt.Errorf("no valid measures or chords could be parsed from chart")
	}

	// Run full jazz-tools pipeline: voice leading, tonal centers, and Berklee harmonic devices
	allChords := tune.AllChords()
	_, _ = VoiceLeadChords(allChords)
	AnalyzeTonalCenters(tune)
	_ = DetectHarmonicDevices(tune)

	return tune, nil
}

func normalizeChordSymbol(sym string) string {
	s := sym
	s = strings.ReplaceAll(s, "Δ", "maj7")
	s = strings.ReplaceAll(s, "^7", "maj7")
	s = strings.ReplaceAll(s, "^", "maj7")
	s = strings.ReplaceAll(s, "ø7", "m7b5")
	s = strings.ReplaceAll(s, "ø", "m7b5")
	s = strings.ReplaceAll(s, "o7", "dim7")
	s = strings.ReplaceAll(s, "°", "dim7")
	s = strings.ReplaceAll(s, "ma7", "maj7")
	s = strings.ReplaceAll(s, "M7", "maj7")
	s = strings.ReplaceAll(s, "min7", "m7")
	s = strings.ReplaceAll(s, "minor7", "m7")
	return s
}

// mockTranscribeTune returns a reliable standard for test environments.
func mockTranscribeTune() *Tune {
	tune, err := LoadTune("Autumn Leaves")
	if err == nil {
		tune.Title = "Autumn Leaves (Transcribed)"
		allChords := tune.AllChords()
		_, _ = VoiceLeadChords(allChords)
		AnalyzeTonalCenters(tune)
		_ = DetectHarmonicDevices(tune)
		return tune
	}

	// Minimal fallback
	t := &Tune{
		Title:         "Autumn Leaves (Transcribed)",
		Composer:      "Joseph Kosma",
		Key:           "G minor",
		TimeSignature: [2]int{4, 4},
		Measures:      make([]TimedMeasure, 8),
	}
	chords := []string{"Cm7", "F7", "Bbmaj7", "Ebmaj7", "Am7b5", "D7", "Gm", "Gm"}
	for i, sym := range chords {
		c, _ := ParseChord(sym)
		t.Measures[i] = TimedMeasure{
			Number: i + 1,
			Chords: []TimedChord{
				{Chord: c, MeasureNumber: i + 1, BeatOffset: 0.0, DurationBeats: 4.0},
			},
		}
	}
	allChords := t.AllChords()
	_, _ = VoiceLeadChords(allChords)
	AnalyzeTonalCenters(t)
	_ = DetectHarmonicDevices(t)
	return t
}

func cleanAPIKey(raw string) string {
	k := strings.TrimSpace(raw)
	// If user copied 'export GEMINI_API_KEY=...' or 'GEMINI_API_KEY=...'
	if idx := strings.Index(k, "="); idx != -1 && strings.Contains(strings.ToUpper(k[:idx]), "KEY") {
		k = strings.TrimSpace(k[idx+1:])
	}
	// Strip quotes and extra spaces
	k = strings.Trim(k, "\"'` \t\r\n")
	return k
}

func normalizeModel(model string) string {
	m := strings.TrimSpace(model)
	if m == "" || m == "gemini-2.5-flash" || m == "gemini-2.0-flash" {
		return "gemini-2.0-flash"
	}
	if m == "gemini-2.5-pro" || m == "gemini-2.0-pro" || m == "gemini-1.5-pro" {
		return "gemini-1.5-pro"
	}
	return m
}
