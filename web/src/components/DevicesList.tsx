import React from 'react';
import { HarmonicDevice } from '../types';

export const DevicesList: React.FC<{ devices: HarmonicDevice[] }> = ({ devices }) => {
  if (!devices || devices.length === 0) return null;

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-4">
      <div className="text-xs font-semibold uppercase tracking-wider text-slate-500 mb-2">
        Identified Berklee Harmonic Devices ({devices.length})
      </div>
      <div className="flex flex-wrap gap-2">
        {devices.map((d, idx) => (
          <div
            key={idx}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg border text-xs ${
              d.IsDashed
                ? 'border-dashed border-amber-300 bg-amber-50/50 text-amber-900'
                : 'border-slate-200 bg-slate-50 text-slate-800'
            }`}
          >
            <span className="font-bold text-sm">{d.Label}</span>
            <span className="text-[11px] text-slate-500">
              bars {d.StartMeasure + 1}–{d.EndMeasure + 1}
            </span>
            {d.Resolves && (
              <span className="px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800 text-[10px] font-bold">
                Resolves ➔ {d.TargetKey}
              </span>
            )}
            {!d.Resolves && (
              <span className="px-1.5 py-0.5 rounded bg-slate-200 text-slate-600 text-[10px] font-medium">
                Turnaround ┐
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
