import React from 'react';

interface FocusForestUIProps {
  timeRemainingMs: number;
  totalDurationMs: number;
  butterfliesHit: number;
  phase: 'playing' | 'complete';
}

export default function FocusForestUI({
  timeRemainingMs,
  totalDurationMs,
  butterfliesHit,
  phase
}: FocusForestUIProps) {
  const progressRatio = totalDurationMs > 0 ? timeRemainingMs / totalDurationMs : 0;
  
  // SEND Rule: Countdown bar fades colour gently (green -> amber -> soft coral)
  let barColour = 'bg-green-500';
  if (progressRatio <= 0.2) {
    barColour = 'bg-rose-400'; // soft coral
  } else if (progressRatio <= 0.5) {
    barColour = 'bg-amber-400';
  }

  return (
    <div className="pointer-events-none absolute inset-0 z-10 flex flex-col justify-between">
      {/* Top Header */}
      <div className="flex w-full items-start justify-between p-4">
        {/* Left: Zone Title / Pause Placeholder */}
        <div className="flex items-center space-x-2 rounded-full bg-slate-900/50 px-4 py-2 backdrop-blur">
          <span className="text-xl font-bold tracking-tight text-white drop-shadow-md">
            Focus Forest
          </span>
        </div>
        
        {/* Right: Butterfly Counter */}
        <div className="rounded-full bg-slate-900/50 px-4 py-2 backdrop-blur">
          <span className="text-xl font-bold text-white drop-shadow-md">
            🦋 × {butterfliesHit}
          </span>
        </div>
      </div>

      {/* Top Countdown Bar (Absolute full-width across top instead of inside p-4 if preferred) */}
      <div className="absolute top-0 w-full h-3 bg-slate-800/50 overflow-hidden rounded-t-xl" role="progressbar">
        <div 
          className={`h-full ${barColour} transition-all duration-1000 ease-linear shadow-sm`}
          style={{ width: `${Math.max(0, progressRatio * 100)}%` }}
        />
      </div>
      
      {/* Complete Overlay */}
      {phase === 'complete' && (
        <div className="pointer-events-auto absolute inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="rounded-2xl bg-white p-8 text-center shadow-2xl">
            <h2 className="mb-2 text-3xl font-bold text-slate-800">Time's Up!</h2>
            <p className="mb-6 text-lg text-slate-600">Great focus today!</p>
            <div className="animate-pulse text-sm text-slate-400">Saving your progress...</div>
          </div>
        </div>
      )}
    </div>
  );
}
