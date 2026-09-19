package jazz

import (
	"testing"
)

func TestGetAllStandards(t *testing.T) {
	standards := GetAllStandards()
	if len(standards) < 1300 {
		t.Fatalf("expected at least 1300 standards, got %d", len(standards))
	}
}

func TestSearchStandards(t *testing.T) {
	results := SearchStandards("Autumn", 10)
	if len(results) == 0 {
		t.Fatalf("expected search results for 'Autumn', got 0")
	}

	foundLeaves := false
	for _, s := range results {
		if s.Title == "Autumn Leaves" {
			foundLeaves = true
			if !s.HasMelody {
				t.Errorf("expected prioritized Autumn Leaves to have HasMelody == true")
			}
			break
		}
	}
	if !foundLeaves {
		t.Errorf("Autumn Leaves not found in search results for 'Autumn'")
	}
}

func TestLoadTuneByStandardTitle(t *testing.T) {
	tune, err := LoadTune("Autumn Leaves")
	if err != nil {
		t.Fatalf("LoadTune by title failed: %v", err)
	}
	if tune.Title != "Autumn Leaves" {
		t.Errorf("expected title 'Autumn Leaves', got '%s'", tune.Title)
	}
	if len(tune.Measures) == 0 {
		t.Errorf("expected measures in Autumn Leaves, got 0")
	}
	if !tune.HasMelody() {
		t.Errorf("expected Autumn Leaves loaded standard to have melody notes")
	}
}
