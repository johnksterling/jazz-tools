import React, { useState, useEffect } from 'react';
import { AnalyzeResponse, SampleTune } from './types';
import { HarmonicJourney } from './components/HarmonicJourney';
import { DevicesList } from './components/DevicesList';
import { ChordsTable } from './components/ChordsTable';
import { ScoreViewer } from './components/ScoreViewer';
import { Music, FileText, Download, Play, Upload, RefreshCw, AlertCircle } from 'lucide-react';

export const App: React.FC = () => {
  const [samples, setSamples] = useState<SampleTune[]>([]);
  const [activeTab, setActiveTab] = useState<'sample' | 'text' | 'upload'>('sample');
  const [selectedSample, setSelectedSample] = useState<string>('waltz_for_debby');
  const [textContent, setTextContent] = useState<string>('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [analysis, setAnalysis] = useState<AnalyzeResponse | null>(null);
  const [scoreXml, setScoreXml] = useState<string | null>(null);
  const [downloadingPdf, setDownloadingPdf] = useState<boolean>(false);

  useEffect(() => {
    fetch('/api/samples')
      .then((res) => (res.ok ? res.json() : []))
      .then((data: SampleTune[]) => {
        setSamples(data);
        if (data.length > 0) {
          // Auto-load first sample
          triggerAnalyze({ sample: data[1]?.id || data[0].id });
        }
      })
      .catch(() => {});
  }, []);

  const triggerAnalyze = async (payload: { sample?: string; content?: string; file?: File }) => {
    setLoading(true);
    setError(null);

    try {
      let analyzeRes: Response;
      let companionRes: Response;

      if (payload.file) {
        const formData = new FormData();
        formData.append('file', payload.file);

        analyzeRes = await fetch('/api/analyze', {
          method: 'POST',
          body: formData,
        });

        const compForm = new FormData();
        compForm.append('file', payload.file);
        companionRes = await fetch('/api/companion/xml', {
          method: 'POST',
          body: compForm,
        });
      } else {
        const jsonBody = JSON.stringify({
          sample: payload.sample,
          content: payload.content,
        });

        analyzeRes = await fetch('/api/analyze', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: jsonBody,
        });

        companionRes = await fetch('/api/companion/xml', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: jsonBody,
        });
      }

      if (!analyzeRes.ok) {
        const errText = await analyzeRes.text();
        throw new Error(errText || 'Failed to analyze tune');
      }

      const data: AnalyzeResponse = await analyzeRes.json();
      setAnalysis(data);

      if (companionRes.ok) {
        const xml = await companionRes.text();
        setScoreXml(xml);
      } else {
        setScoreXml(null);
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred during analysis');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (activeTab === 'sample') {
      triggerAnalyze({ sample: selectedSample });
    } else if (activeTab === 'text') {
      if (!textContent.trim()) {
        setError('Please enter an iReal Pro URL or MusicXML content.');
        return;
      }
      triggerAnalyze({ content: textContent.trim() });
    } else if (activeTab === 'upload') {
      if (!selectedFile) {
        setError('Please select a MusicXML or iReal file to upload.');
        return;
      }
      triggerAnalyze({ file: selectedFile });
    }
  };

  const handleDownloadXml = () => {
    if (!scoreXml) return;
    const blob = new Blob([scoreXml], { type: 'application/vnd.recordare.musicxml+xml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${(analysis?.title || 'tune').replace(/\s+/g, '_')}_guide_tones.musicxml`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleDownloadPdf = async () => {
    if (!analysis) return;
    setDownloadingPdf(true);
    try {
      let res: Response;
      if (activeTab === 'upload' && selectedFile) {
        const formData = new FormData();
        formData.append('file', selectedFile);
        res = await fetch('/api/companion/pdf', { method: 'POST', body: formData });
      } else {
        const jsonBody = JSON.stringify({
          sample: activeTab === 'sample' ? selectedSample : undefined,
          content: activeTab === 'text' ? textContent.trim() : undefined,
        });
        res = await fetch('/api/companion/pdf', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: jsonBody,
        });
      }

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'LilyPond compilation error. Ensure LilyPond is installed on the host.');
      }

      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${analysis.title.replace(/\s+/g, '_')}_guide_tones.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err: any) {
      alert(`PDF Error: ${err.message}`);
    } finally {
      setDownloadingPdf(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 flex flex-col">
      {/* Top Navigation Bar */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 flex items-center justify-center text-white shadow-md shadow-blue-500/20">
              <Music className="w-5 h-5" />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight text-slate-900 leading-none">
                jazz-tools
              </h1>
              <p className="text-xs text-slate-500 mt-1">
                Harmonic Journey & Voice-Led Guide Tone Engine
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3 text-xs">
            <span className="hidden sm:inline-flex items-center px-2.5 py-1 rounded-full bg-slate-100 text-slate-700 font-medium">
              CLI &bull; Web &bull; LilyPond
            </span>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 w-full space-y-6">
        {/* Input Card */}
        <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
          <div className="flex border-b border-slate-200 bg-slate-50/75">
            <button
              onClick={() => setActiveTab('sample')}
              className={`px-5 py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
                activeTab === 'sample'
                  ? 'border-blue-600 text-blue-600 bg-white'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              <Play className="w-4 h-4" />
              Demo Tunes
            </button>
            <button
              onClick={() => setActiveTab('text')}
              className={`px-5 py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
                activeTab === 'text'
                  ? 'border-blue-600 text-blue-600 bg-white'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              <FileText className="w-4 h-4" />
              iReal URL / XML Text
            </button>
            <button
              onClick={() => setActiveTab('upload')}
              className={`px-5 py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
                activeTab === 'upload'
                  ? 'border-blue-600 text-blue-600 bg-white'
                  : 'border-transparent text-slate-600 hover:text-slate-900'
              }`}
            >
              <Upload className="w-4 h-4" />
              Upload File
            </button>
          </div>

          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {activeTab === 'sample' && (
              <div className="flex flex-wrap gap-3 items-center">
                <span className="text-xs font-semibold text-slate-500 uppercase">Select Standard:</span>
                {samples.map((s) => (
                  <button
                    key={s.id}
                    type="button"
                    onClick={() => {
                      setSelectedSample(s.id);
                      triggerAnalyze({ sample: s.id });
                    }}
                    className={`px-4 py-2 rounded-xl text-sm font-medium transition-all flex items-center gap-2 ${
                      selectedSample === s.id
                        ? 'bg-blue-600 text-white shadow-sm'
                        : 'bg-slate-100 hover:bg-slate-200 text-slate-800'
                    }`}
                  >
                    <span>{s.title}</span>
                    <span className="text-xs opacity-75">by {s.composer}</span>
                  </button>
                ))}
              </div>
            )}

            {activeTab === 'text' && (
              <div className="space-y-2">
                <textarea
                  value={textContent}
                  onChange={(e) => setTextContent(e.target.value)}
                  placeholder="Paste an irealb:// link, playlist html, or MusicXML score content..."
                  rows={4}
                  className="w-full text-xs font-mono p-3 rounded-xl border border-slate-300 focus:ring-2 focus:ring-blue-500 focus:outline-none"
                />
              </div>
            )}

            {activeTab === 'upload' && (
              <div className="border-2 border-dashed border-slate-300 rounded-xl p-6 text-center hover:bg-slate-50 transition-colors">
                <input
                  type="file"
                  accept=".musicxml,.xml,.html,.htm"
                  onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                  className="hidden"
                  id="file-upload"
                />
                <label htmlFor="file-upload" className="cursor-pointer block">
                  <Upload className="w-8 h-8 mx-auto text-slate-400 mb-2" />
                  <span className="text-sm font-semibold text-blue-600 hover:underline">
                    {selectedFile ? selectedFile.name : 'Click to upload MusicXML or iReal file'}
                  </span>
                  <p className="text-xs text-slate-500 mt-1">Supports .musicxml, .xml, .html</p>
                </label>
              </div>
            )}

            <div className="flex justify-end pt-2">
              <button
                type="submit"
                disabled={loading}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white rounded-xl font-medium text-sm shadow-sm transition-all flex items-center gap-2"
              >
                {loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    Analyzing...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 fill-white" />
                    Analyze & Voice Lead
                  </>
                )}
              </button>
            </div>
          </form>
        </div>

        {/* Error Notification */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-xl flex items-start gap-3">
            <AlertCircle className="w-5 h-5 flex-shrink-0 text-red-500 mt-0.5" />
            <div>
              <p className="text-sm font-semibold">Error analyzing tune</p>
              <p className="text-xs mt-1 text-red-600">{error}</p>
            </div>
          </div>
        )}

        {/* Results Section */}
        {analysis && (
          <div className="space-y-6">
            {/* Tune Header & Quick Actions */}
            <div className="bg-white rounded-2xl shadow-sm border border-slate-200 p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <h2 className="text-2xl font-black tracking-tight text-slate-900">
                  {analysis.title}
                </h2>
                {analysis.composer && (
                  <p className="text-sm text-slate-500 mt-0.5">by {analysis.composer}</p>
                )}
                <div className="flex items-center gap-2 mt-3">
                  <span className="px-2.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-semibold">
                    Key: {analysis.key || 'Modal/C'}
                  </span>
                  <span className="px-2.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-semibold">
                    Time: {analysis.timeSignature[0]}/{analysis.timeSignature[1]}
                  </span>
                  <span className="px-2.5 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-semibold">
                    {analysis.measureCount} Measures
                  </span>
                </div>
              </div>

              {/* Download / Export Buttons */}
              <div className="flex items-center gap-3">
                <button
                  onClick={handleDownloadXml}
                  disabled={!scoreXml}
                  className="px-4 py-2 border border-slate-300 hover:bg-slate-50 text-slate-700 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-colors"
                >
                  <Download className="w-4 h-4 text-slate-500" />
                  MusicXML
                </button>
                <button
                  onClick={handleDownloadPdf}
                  disabled={downloadingPdf}
                  className="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-semibold flex items-center gap-1.5 shadow-sm transition-colors"
                >
                  <Download className="w-4 h-4" />
                  {downloadingPdf ? 'Compiling PDF...' : 'LilyPond PDF'}
                </button>
              </div>
            </div>

            {/* Harmonic Journey Bar */}
            <HarmonicJourney spans={analysis.journeySpans} />

            {/* Identified Devices */}
            <DevicesList devices={analysis.devices} />

            {/* Interactive MusicXML Score Viewer */}
            {scoreXml && <ScoreViewer xmlContent={scoreXml} />}

            {/* Measure & Chords Table */}
            <ChordsTable measures={analysis.measures} />
          </div>
        )}
      </main>
    </div>
  );
};
