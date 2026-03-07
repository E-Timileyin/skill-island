import React from 'react';

interface WaitingForPartnerProps {
  onCancel: () => void;
  waitingSeconds: number;
}

export default function WaitingForPartner({ onCancel, waitingSeconds }: WaitingForPartnerProps) {
  const isExtendedWait = waitingSeconds > 60;

  return (
    <div className="absolute inset-0 z-50 flex items-center justify-center bg-slate-900/95 backdrop-blur-md">
      <div className="flex max-w-sm flex-col items-center rounded-3xl bg-slate-800 p-8 text-center shadow-2xl border border-slate-700">
        
        {/* Pulsing Avatar Icons */}
        <div className="mb-6 flex space-x-4">
          <div className="h-16 w-16 animate-pulse rounded-full bg-cyan-400/80 shadow-[0_0_20px_rgba(34,211,238,0.5)] flex items-center justify-center text-2xl font-bold text-white">
            1
          </div>
          <div className="h-16 w-16 animate-pulse rounded-full bg-slate-600 flex items-center justify-center text-slate-400">
            ?
          </div>
        </div>

        <h2 className="mb-2 text-2xl font-bold text-slate-100">
          {!isExtendedWait ? "Finding a partner..." : "No partner found right now 🏝️"}
        </h2>
        
        <p className="mb-8 text-[15px] font-medium text-slate-400 leading-relaxed">
          {!isExtendedWait 
            ? "This usually takes just a moment 😊" 
            : "Try again in a few minutes! Other players might be online soon."
          }
        </p>

        <button
          onClick={onCancel}
          className="w-full rounded-2xl bg-slate-700 py-3 font-semibold text-white transition-colors hover:bg-slate-600 active:bg-slate-500"
        >
          {isExtendedWait ? "Back to Island" : "Wait, I've changed my mind"}
        </button>
      </div>
    </div>
  );
}
