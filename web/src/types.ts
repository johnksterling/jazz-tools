export interface SampleTune {
  id: string;
  title: string;
  composer: string;
  type: string;
}

export interface JourneySpan {
  center: string;
  startBar: number;
  endBar: number;
  colorHex: string;
}

export interface ChordDTO {
  symbol: string;
  beatOffset: number;
  durationBeats: number;
  voice1: string;
  voice2: string;
  tonalCenter: string;
  colorHex: string;
}

export interface MeasureDTO {
  number: number;
  chords: ChordDTO[];
}

export interface HarmonicDevice {
  Type: string;
  Label: string;
  StartMeasure: number;
  StartBeat: number;
  EndMeasure: number;
  EndBeat: number;
  Resolves: boolean;
  TargetKey: string;
  LilyPondColor: string;
  IsDashed: boolean;
}

export interface AnalyzeResponse {
  title: string;
  composer: string;
  key: string;
  timeSignature: [number, number];
  measureCount: number;
  harmonicJourney: string;
  journeySpans: JourneySpan[];
  measures: MeasureDTO[];
  devices: HarmonicDevice[];
}
