package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jazz-tools/internal/jazz"
)

func newTestServer(t *testing.T) *Server {
	s, err := NewServer(Config{Port: 8080})
	if err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}
	return s
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("expected status ok, got %s", rec.Body.String())
	}
}

func TestSamplesEndpoint(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/samples", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var samples []SampleTune
	if err := json.Unmarshal(rec.Body.Bytes(), &samples); err != nil {
		t.Fatalf("failed to decode samples: %v", err)
	}
	if len(samples) < 2 {
		t.Errorf("expected at least 2 samples, got %d", len(samples))
	}
}

func TestAnalyzeEndpoint_SampleWaltzForDebby(t *testing.T) {
	s := newTestServer(t)
	body := strings.NewReader(`{"sample":"waltz_for_debby"}`)
	req := httptest.NewRequest("POST", "/api/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp AnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal analyze response: %v", err)
	}

	if resp.Title != "Waltz For Debby" {
		t.Errorf("expected Waltz For Debby, got %s", resp.Title)
	}
	if len(resp.Measures) == 0 {
		t.Errorf("expected measures in response, got 0")
	}
	if len(resp.JourneySpans) == 0 {
		t.Errorf("expected journey spans, got 0")
	}
}

func TestCompanionXMLEndpoint(t *testing.T) {
	s := newTestServer(t)
	body := strings.NewReader(`{"sample":"waltz_for_debby"}`)
	req := httptest.NewRequest("POST", "/api/companion/xml", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	xmlStr := rec.Body.String()
	if !strings.Contains(xmlStr, "<score-partwise") {
		t.Errorf("expected valid MusicXML score-partwise, got: %s", xmlStr)
	}
}

func TestStandardsEndpoint(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/standards?q=Autumn", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []jazz.StandardTune
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to decode standards: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least 1 match for 'Autumn', got 0")
	}

	found := false
	for _, tune := range results {
		if tune.Title == "Autumn Leaves" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find Autumn Leaves in search results")
	}
}

func TestAnalyzeEndpoint_StandardAutumnLeaves(t *testing.T) {
	s := newTestServer(t)
	body := strings.NewReader(`{"standard":"Autumn Leaves"}`)
	req := httptest.NewRequest("POST", "/api/analyze", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp AnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal analyze response: %v", err)
	}

	if resp.Title != "Autumn Leaves" {
		t.Errorf("expected Autumn Leaves, got %s", resp.Title)
	}
	if len(resp.Measures) == 0 {
		t.Errorf("expected measures in Autumn Leaves, got 0")
	}

	rawJSON := rec.Body.String()
	if strings.Contains(rawJSON, `"chords":null`) {
		t.Errorf("expected non-null chords in JSON, but found '\"chords\":null'")
	}
	if resp.Measures[0].Chords == nil {
		t.Errorf("expected Measure[0].Chords to be non-nil empty slice, got nil")
	}
	if resp.Devices == nil {
		t.Errorf("expected Devices to be non-nil slice, got nil")
	}
	if resp.JourneySpans == nil {
		t.Errorf("expected JourneySpans to be non-nil slice, got nil")
	}
}
