import React, { useEffect, useRef, useState } from 'react';
import { OpenSheetMusicDisplay } from 'opensheetmusicdisplay';
import { Sparkles, Music, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';

interface Props {
  xmlContent?: string | null;
  svgPages?: string[] | null;
  hasMelody?: boolean;
  loadingSvg?: boolean;
}

export const ScoreViewer: React.FC<Props> = ({
  xmlContent,
  svgPages,
  hasMelody,
  loadingSvg = false,
}) => {
  const osmdContainerRef = useRef<HTMLDivElement>(null);
  const osmdRef = useRef<OpenSheetMusicDisplay | null>(null);

  // Default to LilyPond engraved view if SVG is available or loading
  const [viewMode, setViewMode] = useState<'lilypond' | 'osmd'>('lilypond');
  const [activePageIndex, setActivePageIndex] = useState<number>(0);
  const [showAllPages, setShowAllPages] = useState<boolean>(true);

  // Initialize/update OSMD when switching to or updating MusicXML
  useEffect(() => {
    if (viewMode !== 'osmd' || !osmdContainerRef.current || !xmlContent) return;

    if (!osmdRef.current) {
      osmdRef.current = new OpenSheetMusicDisplay(osmdContainerRef.current, {
        autoResize: true,
        backend: 'svg',
        drawTitle: true,
        drawSubtitle: false,
        drawComposer: true,
        drawCredits: false,
        drawPartNames: true,
        drawingParameters: 'compacttight',
      });
    }

    try {
      osmdRef.current.clear();
      osmdRef.current.load(xmlContent).then(() => {
        osmdRef.current?.render();
      }).catch(err => {
        console.error("OSMD render error:", err);
      });
    } catch (err) {
      console.error("OSMD load error:", err);
    }
  }, [xmlContent, viewMode]);

  const hasSvg = Boolean(svgPages && svgPages.length > 0);
  const pageCount = svgPages?.length || 0;

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-slate-200 p-6">
      <style>{`
        .lilypond-score-page svg {
          width: 100% !important;
          height: auto !important;
          display: block;
          margin: 0 auto;
        }
      `}</style>

      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-6 pb-4 border-b border-slate-100">
        <div className="flex items-center gap-2 flex-wrap">
          <h3 className="text-base font-bold text-slate-800">
            {hasMelody
              ? 'Lead Sheet & Companion Score (Melody + 3rds & 7ths)'
              : 'Voice-Led Companion Score (3rds & 7ths)'}
          </h3>
          {hasMelody && (
            <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
              Dual Staff System
            </span>
          )}
        </div>

        {/* View Mode & Page Navigation */}
        <div className="flex items-center gap-2 flex-wrap">
          <div className="inline-flex rounded-lg bg-slate-100 p-1 text-xs font-medium">
            <button
              type="button"
              onClick={() => setViewMode('lilypond')}
              className={`flex items-center gap-1.5 px-3 py-1 rounded-md transition-all ${
                viewMode === 'lilypond'
                  ? 'bg-white text-blue-600 font-semibold shadow-xs'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              <Sparkles className="w-3.5 h-3.5 text-blue-500" />
              <span>LilyPond Engraved</span>
            </button>
            <button
              type="button"
              onClick={() => setViewMode('osmd')}
              className={`flex items-center gap-1.5 px-3 py-1 rounded-md transition-all ${
                viewMode === 'osmd'
                  ? 'bg-white text-blue-600 font-semibold shadow-xs'
                  : 'text-slate-600 hover:text-slate-900'
              }`}
            >
              <Music className="w-3.5 h-3.5 text-slate-400" />
              <span>MusicXML</span>
            </button>
          </div>
        </div>
      </div>

      {/* LilyPond Engraved View */}
      {viewMode === 'lilypond' && (
        <div className="w-full">
          {loadingSvg ? (
            <div className="flex flex-col items-center justify-center py-16 text-slate-500 gap-3">
              <RefreshCw className="w-6 h-6 animate-spin text-blue-500" />
              <p className="text-sm font-medium text-slate-700">Engraving publication-grade score with LilyPond...</p>
              <p className="text-xs text-slate-400">Typesetting jazz fonts, harmonic analysis brackets & guide tones</p>
            </div>
          ) : hasSvg ? (
            <div className="space-y-6">
              {/* Multi-page controls if > 1 page */}
              {pageCount > 1 && (
                <div className="flex items-center justify-between bg-slate-50 border border-slate-200 px-4 py-2 rounded-xl text-xs">
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => setShowAllPages(!showAllPages)}
                      className={`px-2.5 py-1 rounded-md border text-xs font-medium transition-colors ${
                        showAllPages
                          ? 'bg-blue-600 text-white border-blue-600'
                          : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-100'
                      }`}
                    >
                      {showAllPages ? 'Showing All Pages' : 'Paged View'}
                    </button>
                    <span className="text-slate-500">
                      {showAllPages ? `Total ${pageCount} pages` : `Page ${activePageIndex + 1} of ${pageCount}`}
                    </span>
                  </div>

                  {!showAllPages && (
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        disabled={activePageIndex === 0}
                        onClick={() => setActivePageIndex(p => Math.max(0, p - 1))}
                        className="p-1 rounded hover:bg-slate-200 disabled:opacity-40 disabled:hover:bg-transparent"
                      >
                        <ChevronLeft className="w-4 h-4" />
                      </button>
                      <button
                        type="button"
                        disabled={activePageIndex === pageCount - 1}
                        onClick={() => setActivePageIndex(p => Math.min(pageCount - 1, p + 1))}
                        className="p-1 rounded hover:bg-slate-200 disabled:opacity-40 disabled:hover:bg-transparent"
                      >
                        <ChevronRight className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </div>
              )}

              {/* Render SVG Page(s) */}
              {showAllPages || pageCount <= 1 ? (
                svgPages!.map((svgContent, idx) => (
                  <div key={idx} className="relative">
                    {pageCount > 1 && (
                      <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2">
                        Page {idx + 1} of {pageCount}
                      </div>
                    )}
                    <div
                      className="lilypond-score-page bg-white shadow-sm border border-slate-200 rounded-xl p-3 sm:p-6 overflow-hidden max-w-[900px] mx-auto"
                      dangerouslySetInnerHTML={{ __html: svgContent }}
                    />
                  </div>
                ))
              ) : (
                <div className="relative">
                  <div
                    className="lilypond-score-page bg-white shadow-sm border border-slate-200 rounded-xl p-3 sm:p-6 overflow-hidden max-w-[900px] mx-auto"
                    dangerouslySetInnerHTML={{ __html: svgPages![activePageIndex] }}
                  />
                </div>
              )}
            </div>
          ) : (
            <div className="py-12 text-center text-slate-500">
              <p className="text-sm">LilyPond vector SVG score not available for this score.</p>
              <button
                onClick={() => setViewMode('osmd')}
                className="mt-3 px-4 py-2 bg-blue-50 text-blue-600 rounded-lg text-xs font-semibold hover:bg-blue-100"
              >
                View with MusicXML
              </button>
            </div>
          )}
        </div>
      )}

      {/* OSMD MusicXML Fallback View */}
      {viewMode === 'osmd' && (
        <div className="overflow-x-auto">
          <div ref={osmdContainerRef} className="w-full min-w-[600px]" />
        </div>
      )}
    </div>
  );
};
