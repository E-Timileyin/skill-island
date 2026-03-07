"use client";

import React, { useState } from 'react';

interface SessionResultScreenProps {
  outcome?: 'win' | 'lose' | 'incomplete';
  starsEarned: number;
  xpEarned: number;
  totalXP?: number;
  unlockedZones: string[];
  onPlayAgain: () => void;
  onGoToIsland: () => void;
}

const encouragement = [
  'Keep going! You did your best.',
  'Nice effort! Every round helps.',
  'Great job! You’re improving.',
  'Amazing! You mastered Memory Cove!'
];

export default function SessionResultScreen({ outcome, starsEarned, xpEarned, totalXP, unlockedZones, onPlayAgain, onGoToIsland }: SessionResultScreenProps) {
  const [showStars, setShowStars] = useState(0);

  React.useEffect(() => {
    let i = 0;
    const interval = setInterval(() => {
      setShowStars(i + 1);
      i++;
      if (i >= starsEarned) clearInterval(interval);
    }, 400);
    return () => clearInterval(interval);
  }, [starsEarned]);

  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center bg-black bg-opacity-70 z-50">
      <div className="bg-white rounded-2xl p-8 shadow-xl flex flex-col items-center">
        <div className="flex mb-4">
          {Array.from({ length: 3 }, (_, i) => (
            <span key={i} className={`text-4xl mx-2 ${i < showStars ? 'text-yellow-400 animate-bounce' : 'text-gray-300'}`}>⭐</span>
          ))}
        </div>
        <div className="text-2xl font-bold text-green-600 mb-2">+{xpEarned} XP</div>
        <div className="text-lg mb-2">Total XP: {totalXP}</div>
        {unlockedZones.length > 0 && <div className="text-xl text-blue-500 mb-2 animate-pulse">New zone unlocked!</div>}
        <div className="text-lg mb-6">{encouragement[starsEarned]}</div>
        <div className="flex gap-4">
          <button className="btn bg-yellow-400 hover:bg-yellow-500 text-white px-6 py-2 rounded-xl font-semibold" onClick={onPlayAgain}>Play Again</button>
          <button className="btn bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-xl font-semibold" onClick={onGoToIsland}>Back to Island</button>
        </div>
      </div>
    </div>
  );
}
