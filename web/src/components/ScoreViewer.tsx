import React, { useEffect, useRef } from 'react';
import { OpenSheetMusicDisplay } from 'opensheetmusicdisplay';

interface Props {
  xmlContent: string;
}

export const ScoreViewer: React.FC<Props> = ({ xmlContent }) => {
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
        drawPartNames: false,
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
        <h3 className="text-base font-semibold text-slate-800">Voice-Led Companion Score (3rds & 7ths)</h3>
        <span className="text-xs text-slate-500">Rendered via in-browser SVG</span>
      </div>
      <div ref={containerRef} className="w-full min-w-[600px]" />
    </div>
  );
};
