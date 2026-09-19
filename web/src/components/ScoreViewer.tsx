import React, { useEffect, useRef } from 'react';
import { OpenSheetMusicDisplay } from 'opensheetmusicdisplay';

interface Props {
  xmlContent: string;
  hasMelody?: boolean;
}

export const ScoreViewer: React.FC<Props> = ({ xmlContent, hasMelody }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const osmdRef = useRef<OpenSheetMusicDisplay | null>(null);

  useEffect(() => {
    if (!containerRef.current || !xmlContent) return;

    if (!osmdRef.current) {
      osmdRef.current = new OpenSheetMusicDisplay(containerRef.current, {
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

    osmdRef.current.load(xmlContent).then(() => {
      osmdRef.current?.render();
    }).catch(err => {
      console.error("OSMD render error:", err);
    });
  }, [xmlContent]);

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6 overflow-x-auto">
      <div className="flex items-center justify-between mb-4 pb-2 border-b border-slate-100">
        <div className="flex items-center gap-2">
          <h3 className="text-base font-semibold text-slate-800">
            {hasMelody ? 'Lead Sheet & Companion Score (Melody + 3rds & 7ths)' : 'Voice-Led Companion Score (3rds & 7ths)'}
          </h3>
          {hasMelody && (
            <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
              Dual Staff System
            </span>
          )}
        </div>
        <span className="text-xs text-slate-500">Rendered via in-browser SVG</span>
      </div>
      <div ref={containerRef} className="w-full min-w-[600px]" />
    </div>
  );
};
