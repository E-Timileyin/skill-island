import React from 'react';

interface MemoryCoveUIProps {
  currentRound: number;
  totalRounds: number;
  stars: number;
  phase: 'watching' | 'your_turn' | 'complete';
  playerNickname: string;
}

export default function MemoryCoveUI({ currentRound, totalRounds, stars, phase, playerNickname }: MemoryCoveUIProps) {
  const starIcons = Array.from({ length: 3 }, (_, i) => (
    <span key={i} className={i < stars ? 'text-yellow-400' : 'text-gray-300'}>⭐</span>
  ));
  const phaseText = {
    watching: 'Watch carefully!',
    your_turn: 'Your turn!',
    complete: 'Amazing!'
  }[phase];

  return (
    <div className="pointer-events-none fixed inset-0 flex flex-col justify-between p-6 z-40">
      <div className="flex justify-between">
        <div className="bg-white bg-opacity-80 rounded-xl px-4 py-2 text-lg font-semibold">Round {currentRound} / {totalRounds}</div>
        <div className="bg-white bg-opacity-80 rounded-xl px-4 py-2 text-lg">{starIcons}</div>
      </div>
      <div className="flex flex-col items-center">
        <div className="bg-yellow-100 bg-opacity-90 rounded-2xl px-8 py-4 text-2xl font-bold mb-4">{phaseText}</div>
        <div className="bg-white bg-opacity-80 rounded-xl px-4 py-2 text-lg">Player: {playerNickname}</div>
      </div>
    </div>
  );
}
