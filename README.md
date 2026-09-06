# jazz-tools

Command-line utility and music theory engine to analyze jazz tunes, import lead sheets from **MusicXML** and **iReal Pro**, and generate voice-led **Guide Tone (3rd & 7th) Companion Sheets** for performance and combo rehearsal.

## Features

- **Multi-Format Ingestion**:
  - **MusicXML**: Parses partwise MusicXML lead sheets (from iReal Pro, MuseScore, Finale, Sibelius), tracking chord timing and hold durations accurately.
  - **iReal Pro**: Directly imports songs from `irealb://` or `irealbook://` URL links, exported `.html` playlist files, or forum chart text (using the symmetric `obfusc50` deobfuscation algorithm).
- **Accurate Jazz Theory Engine**:
  - Exact enharmonic spelling for all 12 keys (e.g. $E\flat^{\Delta 7}$ produces $G$ & $D$; $F\sharp^7$ produces $A\sharp$ & $E$; $A\flat^7$ produces $C$ & $G\flat$).
  - Full support for jazz chord qualities: Major 7th, Dominant 7th, Minor 7th, Half-diminished ($m7\flat 5$), Diminished 7th, Minor-Major 7th, 6th chords, Altered dominants, and Suspended chords.
- **Voice-Led Guide Tone Companion Generator**:
  - Writes out the **3rd and 7th** of every chord on the treble staff for the exact duration of each chord hold.
  - Generates **two distinct polyphonic voices** (Voice 1 upper stems up, Voice 2 lower stems down), automatically connecting each voice to whichever note is closest to the previous note for smooth, stepwise voice leading.
  - Exports standard **MusicXML** companion scores and compiles print-ready **PDFs** via LilyPond.

---

## Installation & Requirements

Ensure you have [Go](https://go.dev) (1.20+) installed.

To compile PDFs directly, install [LilyPond](https://lilypond.org):
```bash
brew install lilypond
```

Build the binary:
```bash
go build -o jazz-tools .
```

---

## Usage & Commands

### 1. Analyze a Tune
Inspect tune structure, key, harmonic rhythm, chord hold durations, and guide tones:

```bash
# Analyze a MusicXML file:
./jazz-tools analyze Jordu.musicxml

# Or analyze an iReal Pro URL or exported HTML file:
./jazz-tools analyze "irealb://Dear%20Old%20Stockholm=Traditional===Medium%20Swing=D-==1r34LbKcu7*A{T44D- |Eh7 A7b9|G-7 C7|F^7 |Eh7 A7b9|D- |Eh7 |A7b9 |D-7 |D-6 |D-7 |D-6 }*B[F^7 |G-7 C7|F^7 |Eh7 A7b9 ]*C[D- |Eh7 A7b9|G-7 C7|F^7 |Eh7 A7b9|D- |C7sus |x |C7sus |x |x |x |C7sus A7b9|D- |x]="
```

### 2. Generate Guide Tone Companion Sheets
Create standard MusicXML companion scores and publication-quality PDFs:

```bash
# Generate MusicXML companion score:
./jazz-tools companion Jordu.musicxml -o Jordu_guide_tones.musicxml

# Generate both MusicXML and print-ready PDF:
./jazz-tools companion Jordu.musicxml --pdf

# Generate companion sheet directly from an iReal Pro link or exported HTML playlist:
./jazz-tools companion "irealb://..." --pdf
./jazz-tools companion "MyPlaylist.html" --pdf
```

---

## Running Tests

Run the full test suite covering pitch class math, chord qualities, voice leading, MusicXML read/write, and iReal Pro deobfuscation:

```bash
go test -v ./...
```
