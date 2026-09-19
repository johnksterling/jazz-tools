package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
