package jazz

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func buildSyntheticMSF(setlistName string, files []struct {
	path string
	pdf  []byte
}) []byte {
	var buf bytes.Buffer
	// Header with 0xff 0xff 0x01 0x01 [uint16 BE len] [setlistName]
	buf.Write([]byte{0xff, 0x0f, 0x98, 0x34, 0xff, 0xff, 0x01, 0x01})
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(setlistName)))
	buf.WriteString(setlistName)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00})

	for _, f := range files {
		_ = binary.Write(&buf, binary.BigEndian, uint32(len(f.pdf)))
		_ = binary.Write(&buf, binary.BigEndian, uint16(len(f.path)))
		buf.WriteString(f.path)
		buf.Write(f.pdf)
		buf.Write([]byte{0xff, 0xff, 0x05, 0x29})
	}
	return buf.Bytes()
}

func TestCleanMSFSongTitle(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Winter Wonderland - C.pdf", "Winter Wonderland"},
		{"Jardin-d_Hiver - C.pdf", "Jardin d'Hiver"},
		{"What Are You Doing New Year_s Eve - C.pdf", "What Are You Doing New Year's Eve"},
		{"FunkInDeepFreeze.pdf", "Funk In Deep Freeze"},
		{"BerniesTune.pdf", "Bernies Tune"},
		{"Softly As A Morning Sunrise - C.pdf", "Softly As A Morning Sunrise"},
	}

	for _, tc := range tests {
		got := CleanMSFSongTitle(tc.in)
		if got != tc.want {
			t.Errorf("CleanMSFSongTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMatchStandardForMSF(t *testing.T) {
	for _, title := range []string{
		"Beatrice",
		"Softly As A Morning Sunrise",
		"Sunny Side Of The Street",
		"Bernies Tune",
		"Blue Bossa",
		"There Will Never Be Another You",
	} {
		if _, ok := MatchStandardForMSF(title); !ok {
			t.Errorf("expected MatchStandardForMSF(%q) to match a built-in standard", title)
		}
	}
}

func TestParseMSF_Synthetic(t *testing.T) {
	pdfA := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n%%EOF\n")
	pdfB := []byte("%PDF-1.3\n1 0 obj\n<<>>\nendobj\n%%EOF\n")
	msf := buildSyntheticMSF("12/3/2026", []struct {
		path string
		pdf  []byte
	}{
		{"/storage/emulated/0/Android/data/com.zubersoft.mobilesheetspro/files/Autumn Leaves - C.pdf", pdfA},
		{"/storage/emulated/0/Android/data/com.zubersoft.mobilesheetspro/files/Beatrice - C.pdf", pdfB},
	})

	archive, err := ParseMSF(msf)
	if err != nil {
		t.Fatalf("ParseMSF failed: %v", err)
	}
	if archive.SetlistName != "12/3/2026" {
		t.Errorf("expected SetlistName '12/3/2026', got %q", archive.SetlistName)
	}
	if len(archive.Songs) != 2 {
		t.Fatalf("expected 2 songs, got %d", len(archive.Songs))
	}
	if archive.Songs[0].Title != "Autumn Leaves" || !bytes.Equal(archive.Songs[0].PDFBytes, pdfA) {
		t.Errorf("unexpected song 0: %+v", archive.Songs[0])
	}
	if archive.Songs[1].Title != "Beatrice" || !bytes.Equal(archive.Songs[1].PDFBytes, pdfB) {
		t.Errorf("unexpected song 1: %+v", archive.Songs[1])
	}
}

func TestProcessMSFToZip_Synthetic(t *testing.T) {
	if _, err := exec.LookPath("lilypond"); err != nil {
		t.Skip("lilypond not installed; skipping PDF zip test")
	}

	pdfA := []byte("%PDF-1.4\n%%EOF\n")
	pdfB := []byte("%PDF-1.4\n%%EOF\n")
	msf := buildSyntheticMSF("Gig Setlist", []struct {
		path string
		pdf  []byte
	}{
		{"/files/Autumn Leaves - C.pdf", pdfA},
		{"/files/Beatrice - C.pdf", pdfB},
	})

	zipBytes, setlist, results, err := ProcessMSFToZip(msf, BatchOptions{
		FullReport:       true,
		AnnotationConfig: DefaultAnnotationConfig(),
	})
	if err != nil {
		t.Fatalf("ProcessMSFToZip failed: %v", err)
	}
	if setlist != "Gig Setlist" {
		t.Errorf("expected setlist 'Gig Setlist', got %q", setlist)
	}
	if len(results) != 2 || results[0].Source != "standard" || results[1].Source != "standard" {
		t.Fatalf("expected 2 standard results, got %+v", results)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader failed: %v", err)
	}

	var names []string
	var csvContent string
	for _, f := range zr.File {
		names = append(names, f.Name)
		if strings.HasSuffix(f.Name, ".csv") {
			rc, _ := f.Open()
			var b bytes.Buffer
			_, _ = b.ReadFrom(rc)
			_ = rc.Close()
			csvContent = b.String()
		}
	}
	if len(names) != 4 {
		t.Errorf("expected 4 files in zip (2 PDFs, 1 CSV, 1 manifest), got %v", names)
	}
	if !strings.Contains(csvContent, "Autumn Leaves (Companion)") || !strings.Contains(csvContent, "Gig Setlist (Companion)") {
		t.Errorf("unexpected CSV content:\n%s", csvContent)
	}
}

func TestParseMSF_LocalFixtureIfPresent(t *testing.T) {
	data, err := os.ReadFile("../../12 3 2026.msf")
	if err != nil {
		t.Skip("local 12 3 2026.msf not present")
	}

	archive, err := ParseMSF(data)
	if err != nil {
		t.Fatalf("ParseMSF failed: %v", err)
	}
	if len(archive.Songs) != 25 {
		t.Errorf("expected 25 songs in 12 3 2026.msf, got %d", len(archive.Songs))
	}
	matched := 0
	for _, s := range archive.Songs {
		if _, ok := MatchStandardForMSF(s.Title); ok {
			matched++
		}
	}
	t.Logf("12 3 2026.msf (%s): %d songs, %d matched built-in standards", archive.SetlistName, len(archive.Songs), matched)
}
