import React from 'react';
import { JourneySpan } from '../types';

export const HarmonicJourney: React.FC<{ spans: JourneySpan[] }> = ({ spans }) => {
  if (!spans || spans.length === 0) return null;

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-4">
      <div className="text-xs font-semibold uppercase tracking-wider text-slate-500 mb-2">
        Harmonic Modulation Trajectory
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {spans.map((s, idx) => (
          <React.Fragment key={idx}>
            <div
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-white shadow-sm transition-transform hover:scale-105"
              style={{ backgroundColor: s.colorHex || '#1e293b' }}
            >
              <span className="text-sm font-bold tracking-tight">{s.center}</span>
              <span className="opacity-80 text-[10px] font-normal">
                ({s.startBar === s.endBar ? `bar ${s.startBar}` : `bars ${s.startBar}–${s.endBar}`})
              </span>
            </div>
            {idx < spans.length - 1 && (
              <span className="text-slate-400 font-bold select-none text-sm">➔</span>
            )}
          </React.Fragment>
        ))}
      </div>
    </div>
  );
};
