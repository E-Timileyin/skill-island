import React, { useState, useEffect } from 'react';

interface CoopSessionResultScreenProps {
  outcome: 'win' | 'lose' | 'incomplete';
  starsEarned: number;
  groupXPEarned: number;
  myXPEarned: number;
  totalXP: number;
  unlockedZones: string[];
  onPlayAgain: () => void;
  onGoToIsland: () => void;
}

const outcomeMessages = {
  win: "Amazing Teamwork! 🏆",
  lose: "Great Effort! The tower wobbled — try again! 🏗️",
  incomplete: "Good Try! Your partner disconnected this time. 💪"
};

export default function CoopSessionResultScreen({
  outcome,
  starsEarned,
  groupXPEarned,
  myXPEarned,
  totalXP,
  unlockedZones,
  onPlayAgain,
  onGoToIsland
}: CoopSessionResultScreenProps) {
  
  // Guarantee min 1 star visually as per SEND
  const displayStars = Math.max(1, starsEarned);
  const [showStars, setShowStars] = useState(0);

  useEffect(() => {
    let i = 0;
    const interval = setInterval(() => {
      setShowStars(prev => prev + 1);
      i++;
      if (i >= displayStars) clearInterval(interval);
    }, 400); // 0.3s-0.4s gap between each
    return () => clearInterval(interval);
  }, [displayStars]);

  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center bg-slate-900/80 backdrop-blur-md z-50">
      
      {unlockedZones.length > 0 && (
        <div className="absolute top-10 flex w-full animate-[pulse_3s_ease-in-out_infinite] items-center justify-center pointer-events-none">
           <div className="bg-blue-500 text-white font-bold text-2xl py-3 px-8 rounded-full shadow-[0_0_30px_rgba(59,130,246,0.8)] border border-blue-300">
             🔓 New Zone Unlocked: {unlockedZones.join(', ')}!
           </div>
        </div>
      )}

      <div className="bg-white rounded-3xl p-10 shadow-2xl flex flex-col items-center max-w-lg w-full text-center transform hover:-translate-y-1 transition duration-500">
        <h2 className="text-3xl font-extrabold text-slate-800 mb-8">{outcomeMessages[outcome]}</h2>
        
        <div className="flex mb-8 space-x-4">
          {Array.from({ length: 3 }, (_, i) => (
            <span 
              key={i} 
              className={`text-6xl ${i < showStars ? 'text-yellow-400 drop-shadow-md animate-bounce' : 'text-slate-200'} transition-all duration-300`}
            >
              ⭐
            </span>
          ))}
        </div>
        
        <div className="bg-sky-50 rounded-2xl p-6 w-full mb-8 border border-sky-100">
           <div className="text-3xl font-bold text-sky-600 mb-2 animate-[bounce_2s_ease-in-out_infinite]">
             +{myXPEarned} XP each
           </div>
           <div className="text-lg text-sky-800 font-medium">
             Your team earned {groupXPEarned} XP together!
           </div>
        </div>

        <div className="text-md text-slate-400 mb-10 font-medium">
          New Total XP: {totalXP}
        </div>
        
        <div className="flex w-full gap-4">
          <button 
            className="flex-1 bg-yellow-400 hover:bg-yellow-500 text-white px-6 py-4 rounded-2xl font-bold text-lg shadow-lg shadow-yellow-400/30 transition-all hover:scale-105" 
            onClick={onPlayAgain}
          >
            Play Again
          </button>
          <button 
            className="flex-1 bg-slate-200 hover:bg-slate-300 text-slate-700 px-6 py-4 rounded-2xl font-bold text-lg transition-all" 
            onClick={onGoToIsland}
          >
            Back to Island
          </button>
        </div>
      </div>
    </div>
  );
}
