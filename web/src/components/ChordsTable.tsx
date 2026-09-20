import React from 'react';
import { MeasureDTO } from '../types';

export const ChordsTable: React.FC<{ measures: MeasureDTO[] }> = ({ measures }) => {
  if (!measures || measures.length === 0) return null;

  const hasAnyChords = measures.some((m) => m.chords && m.chords.length > 0);
  if (!hasAnyChords) return null;

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
      <div className="px-5 py-3 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
        <h3 className="text-sm font-semibold text-slate-800">Measure & Voice Leading Detail</h3>
        <span className="text-xs text-slate-500">{measures.length} measures</span>
      </div>
      <div className="overflow-x-auto max-h-96">
        <table className="w-full text-left text-xs border-collapse">
          <thead className="bg-slate-100/75 text-slate-600 font-semibold sticky top-0 shadow-sm z-10">
            <tr>
              <th className="py-2.5 px-4">Bar</th>
              <th className="py-2.5 px-4">Beat</th>
              <th className="py-2.5 px-4">Chord</th>
              <th className="py-2.5 px-4">Hold</th>
              <th className="py-2.5 px-4">Voice 1 (3/7)</th>
              <th className="py-2.5 px-4">Voice 2 (3/7)</th>
              <th className="py-2.5 px-4">Tonal Center</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 font-mono">
            {measures.flatMap((m) =>
              (m.chords || []).map((c, cIdx) => (
                <tr key={`${m.number}-${cIdx}`} className="hover:bg-slate-50/80 transition-colors">
                  <td className="py-2 px-4 font-semibold text-slate-700">
                    {cIdx === 0 ? m.number : ''}
                  </td>
                  <td className="py-2 px-4 text-slate-500">{(c.beatOffset + 1).toFixed(1)}</td>
                  <td className="py-2 px-4 font-bold text-slate-900">{c.symbol}</td>
                  <td className="py-2 px-4 text-slate-600">{c.durationBeats.toFixed(1)} b</td>
                  <td className="py-2 px-4 text-blue-600 font-semibold">{c.voice1}</td>
                  <td className="py-2 px-4 text-indigo-600 font-semibold">{c.voice2}</td>
                  <td className="py-2 px-4">
                    <span
                      className="inline-block px-2 py-0.5 rounded text-[11px] font-bold text-white shadow-xs"
                      style={{ backgroundColor: c.colorHex || '#334155' }}
                    >
                      {c.tonalCenter}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
