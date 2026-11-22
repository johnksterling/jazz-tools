package jazz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// BuildDeterministicHarmonicSummary builds a concise, zero-network summary of the
// tune's harmonic structure from its tonal center spans and Berklee devices.
func BuildDeterministicHarmonicSummary(tune *Tune) string {
	if tune == nil || len(tune.Measures) == 0 {
		return ""
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)
	devices := DetectHarmonicDevices(tune)

	homeRoot, homeMode := extractHomeKey(tune)
	homeDesc := homeRoot + " " + homeMode

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s is a %d-measure tune in %s (%d/%d time). ",
		tune.Title, len(tune.Measures), homeDesc, tune.TimeSignature[0], tune.TimeSignature[1]))

	if len(spans) == 1 {
		sb.WriteString(fmt.Sprintf("The harmony remains centered in %s throughout the form. ", spans[0].Center))
	} else if len(spans) > 1 {
		var parts []string
		for _, sp := range spans {
			if sp.StartBar == sp.EndBar {
				parts = append(parts, fmt.Sprintf("%s (bar %d)", sp.Center, sp.StartBar))
			} else {
				parts = append(parts, fmt.Sprintf("%s (bars %d-%d)", sp.Center, sp.StartBar, sp.EndBar))
			}
		}
		sb.WriteString("Its tonal trajectory moves through ")
		if len(parts) <= 6 {
			sb.WriteString(strings.Join(parts, " -> "))
		} else {
			sb.WriteString(strings.Join(parts[:5], " -> "))
			sb.WriteString(fmt.Sprintf(" -> ... (%d tonal regions total)", len(parts)))
		}
		sb.WriteString(". ")
	}

	if len(devices) > 0 {
		counts := make(map[DeviceType]int)
		var order []DeviceType
		for _, d := range devices {
			if counts[d.Type] == 0 {
				order = append(order, d.Type)
			}
			counts[d.Type]++
		}
		var devParts []string
		for _, dt := range order {
			devParts = append(devParts, fmt.Sprintf("%dx %s", counts[dt], string(dt)))
		}
		sb.WriteString(fmt.Sprintf("Key harmonic devices include %s.", strings.Join(devParts, ", ")))
	}

	return strings.TrimSpace(sb.String())
}

// GenerateHarmonicSummary generates a brief LLM narrative of the tune's harmonic structure
// using Gemini when credentials are available, falling back to a deterministic summary.
// Returns (summaryText, isLLMGenerated, error).
func GenerateHarmonicSummary(tune *Tune, opts TranscribeOptions) (string, bool, error) {
	if tune == nil || len(tune.Measures) == 0 {
		return "", false, fmt.Errorf("empty tune")
	}

	AnalyzeTonalCenters(tune)
	spans := GetHarmonicJourneySpans(tune)
	devices := DetectHarmonicDevices(tune)

	if os.Getenv("MOCK_TRANSCRIBE") == "1" || strings.EqualFold(opts.APIKey, "mock") {
		return BuildDeterministicHarmonicSummary(tune) + " (Mock LLM narrative)", true, nil
	}

	rawAuth := cleanAPIKey(opts.APIKey)
	if rawAuth == "" {
		rawAuth = cleanAPIKey(os.Getenv("GEMINI_API_KEY"))
	}

	var isBearer bool
	var authKey string
	if strings.HasPrefix(rawAuth, "ya29.") {
		isBearer = true
		authKey = rawAuth
	} else if rawAuth != "" {
		isBearer = false
		authKey = rawAuth
	} else if adcToken, err := ResolveADCAccessToken(); err == nil && adcToken != "" {
		isBearer = true
		authKey = adcToken
	} else {
		// No LLM credentials available; return deterministic summary cleanly.
		return BuildDeterministicHarmonicSummary(tune), false, nil
	}

	// Build compact harmonic context for the LLM
	var progLines []string
	for _, m := range tune.Measures {
		var cs []string
		for _, c := range m.Chords {
			cs = append(cs, fmt.Sprintf("%s [%s]", c.Chord.String(), c.TonalCenter))
		}
		if len(cs) > 0 {
			progLines = append(progLines, fmt.Sprintf("Bar %d: %s", m.Number, strings.Join(cs, ", ")))
		}
	}

	var devLines []string
	for _, d := range devices {
		devLines = append(devLines, fmt.Sprintf("- Bars %d-%d: %s (target: %s)", d.StartMeasure+1, d.EndMeasure+1, d.Label, d.TargetKey))
	}

	var spanLines []string
	for _, sp := range spans {
		spanLines = append(spanLines, fmt.Sprintf("%s (bars %d-%d)", sp.Center, sp.StartBar, sp.EndBar))
	}

	prompt := fmt.Sprintf(`You are an expert jazz theorist and Berklee harmony instructor.
Write a very brief, insightful harmonic overview (3 to 5 sentences, plain prose only—no markdown, no bullet points, no headings) for the jazz standard "%s" (%s, key: %s, %d/%d, %d bars).
Explain the overall form, tonal centers/modulations, and the most notable harmonic devices (e.g., ii-V chains, secondary dominants, modal interchange, tritone subs, passing diminished chords).

Harmonic Journey:
%s

Identified Devices:
%s

Measure-by-measure Chords:
%s`,
		tune.Title,
		tune.Composer,
		tune.KeyName(),
		tune.TimeSignature[0], tune.TimeSignature[1],
		len(tune.Measures),
		strings.Join(spanLines, " -> "),
		strings.Join(devLines, "\n"),
		strings.Join(progLines, "\n"),
	)

	model := normalizeModel(opts.Model)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.3,
			"maxOutputTokens": 300,
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return BuildDeterministicHarmonicSummary(tune), false, err
	}

	var apiURL string
	if isBearer {
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	} else {
		apiURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, url.QueryEscape(authKey))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqJSON))
	if err != nil {
		return BuildDeterministicHarmonicSummary(tune), false, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if isBearer {
		httpReq.Header.Set("Authorization", "Bearer "+authKey)
	} else {
		httpReq.Header.Set("x-goog-api-key", authKey)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return BuildDeterministicHarmonicSummary(tune), false, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return BuildDeterministicHarmonicSummary(tune), false, fmt.Errorf("gemini summary status %d", resp.StatusCode)
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
	if err := json.Unmarshal(respBytes, &apiResp); err != nil ||
		len(apiResp.Candidates) == 0 ||
		len(apiResp.Candidates[0].Content.Parts) == 0 {
		return BuildDeterministicHarmonicSummary(tune), false, fmt.Errorf("empty gemini summary response")
	}

	text := strings.TrimSpace(apiResp.Candidates[0].Content.Parts[0].Text)
	if text == "" {
		return BuildDeterministicHarmonicSummary(tune), false, nil
	}
	return text, true, nil
}
