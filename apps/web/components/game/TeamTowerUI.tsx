import React from 'react';

interface TeamTowerUIProps {
  phase: 'init' | 'waiting' | 'ready' | 'playing' | 'partner_disconnected' | 'idle_warning' | 'complete' | 'reconnecting' | 'partner_reconnected' | 'saving' | 'results' | 'error';
  groupXP: number;
  groupXPTarget?: number;
  activePlayer: string;
  myRole: string;
  turnNumber: number;
  opponentAvatarId?: number;
  partnerStatus?: 'connected' | 'disconnected';
  idleSecondsRemaining?: number;
}

export default function TeamTowerUI({
  phase,
  groupXP,
  groupXPTarget = 100,
  activePlayer,
  myRole,
  turnNumber,
  idleSecondsRemaining
}: TeamTowerUIProps) {

  // SEND Rule: Avoid numbers and pressure on the idle warning
  // Avoid red UI
  
  const isMyTurn = activePlayer === myRole;
  const progressRatio = groupXPTarget > 0 ? groupXP / groupXPTarget : 0;

  return (
    <div className="pointer-events-none absolute inset-0 z-10 flex flex-col justify-between">
      {/* Top Banner Area */}
      {phase === 'playing' && (
        <div className="flex w-full items-start justify-between p-4">
          <div className="flex flex-col space-y-2">
            {/* Turn Indicator */}
            <div className={`rounded-full px-6 py-3 font-bold shadow-lg transition-colors ${
              isMyTurn ? 'bg-green-500 text-white' : 'bg-blue-300 text-slate-800'
            }`}>
              {isMyTurn ? "Your Turn! 🎯" : "Partner's Turn..."}
            </div>
          </div>
          
          <div className="flex items-center space-x-3 rounded-full bg-slate-900/50 px-4 py-2 backdrop-blur">
            <span className="text-lg font-bold text-white drop-shadow">
              Turn {turnNumber}
            </span>
          </div>
        </div>
      )}

      {/* Disconnection Overlay */}
      {phase === 'partner_disconnected' && (
        <div className="pointer-events-auto absolute inset-0 flex items-center justify-center bg-amber-900/40 backdrop-blur-sm">
          <div className="rounded-2xl border-4 border-amber-400 bg-white p-8 text-center shadow-2xl">
            <h2 className="mb-2 text-2xl font-bold text-amber-700">Your partner will be right back... ⏳</h2>
            <p className="text-lg text-slate-600">The game is paused for 30 seconds</p>
          </div>
        </div>
      )}
      
      {/* Reconnection Client Overlay */}
      {phase === 'reconnecting' && (
        <div className="pointer-events-auto absolute inset-0 flex items-center justify-center bg-blue-900/40 backdrop-blur-sm">
          <div className="rounded-2xl border-4 border-blue-400 bg-white p-8 text-center shadow-2xl">
            <h2 className="mb-2 text-2xl font-bold text-blue-700 flex items-center gap-3">
              <svg className="animate-spin h-6 w-6 text-blue-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Reconnecting... 🔄
            </h2>
            <p className="text-lg text-slate-600">Please wait a moment while we rebuild the tower link.</p>
          </div>
        </div>
      )}

      {/* Partner Returned Banner */}
      {phase === 'partner_reconnected' && (
        <div className="absolute top-20 w-full flex justify-center animate-bounce">
           <div className="rounded-full bg-green-500 px-6 py-2 text-white font-bold shadow-lg">
             Partner is back! ▶️
           </div>
        </div>
      )}

      {/* Idle Warning Overlay */}
      {phase === 'idle_warning' && (
        <div className="absolute inset-0 pointer-events-none border-8 border-amber-300 animate-pulse rounded-2xl">
          <div className="absolute bottom-16 w-full flex justify-center mb-6">
            <div className="bg-amber-100 border-2 border-amber-400 rounded-full px-6 py-3 shadow-lg font-bold text-amber-800">
              Still there? Place a block to continue! 😊
            </div>
          </div>
        </div>
      )}

      {/* Bottom Group XP UI */}
      {phase === 'playing' && (
        <div className="w-full bg-slate-900/80 p-4 backdrop-blur-sm rounded-b-xl border-t border-slate-700">
          <div className="flex justify-between text-white text-sm font-semibold mb-2">
            <span>Group XP</span>
            <span>{groupXP} / {groupXPTarget}</span>
          </div>
          <div className="h-3 w-full bg-slate-700 rounded-full overflow-hidden">
             <div 
               className="h-full bg-cyan-400 transition-all duration-500 ease-out" 
               style={{ width: `${Math.min(100, Math.max(0, progressRatio * 100))}%` }} 
             />
          </div>
        </div>
      )}
    </div>
  );
}
